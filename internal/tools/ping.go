package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"noc-mcp/pkg/logger"
	"noc-mcp/pkg/util"

	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

const maxPingWorkers = 50 // Límite de concurrencia para proteger el servidor

type pingJob struct {
	IP    string
	Count string
}

type pingResult struct {
	IP     string
	Output string
	Error  error
	TimeMs int64
}

// PingHandler maneja la ejecución de pings controlados por un Worker Pool
func PingHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toolArgs := request.GetArguments()
	targetsRaw, ok := toolArgs["targets"].([]interface{})
	if !ok {
		return mcp.NewToolResultError("El parámetro 'targets' debe ser una lista."), nil
	}

	count := "4"
	if c, ok := toolArgs["count"].(float64); ok {
		count = fmt.Sprintf("%.0f", c)
	}

	jobs := make(chan pingJob, len(targetsRaw))
	resultsChan := make(chan pingResult, len(targetsRaw))
	var wg sync.WaitGroup

	// 1. Iniciar los Workers (Máximo 50)
	for i := 0; i < maxPingWorkers && i < len(targetsRaw); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				start := time.Now()
				cmdCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

				cmd := exec.CommandContext(cmdCtx, "ping", "-c", job.Count, job.IP)
				out, err := cmd.CombinedOutput()

				if err != nil && cmdCtx.Err() == context.DeadlineExceeded {
					err = fmt.Errorf("timeout")
				}

				resultsChan <- pingResult{
					IP:     job.IP,
					Output: string(out),
					Error:  err,
					TimeMs: time.Since(start).Milliseconds(),
				}
				cancel()
			}
		}()
	}

	// 2. Enviar los trabajos a la cola
	var finalResults []string
	var mu sync.Mutex

	for _, tRaw := range targetsRaw {
		target, ok := tRaw.(string)
		if !ok || !util.IsValidTarget(target) {
			mu.Lock()
			finalResults = append(finalResults, fmt.Sprintf("⚠️ Destino inválido: %v", tRaw))
			mu.Unlock()
			logger.Log.Warn("Intento de ping a destino inválido", zap.Any("target", tRaw))
			continue
		}
		jobs <- pingJob{IP: target, Count: count}
	}
	close(jobs) // Cierra el canal para indicar a los workers que no hay más tareas

	// 3. Esperar a los workers en background para cerrar el canal de resultados
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// 4. Recolectar resultados y emitir Auditoría
	for res := range resultsChan {
		status := "success"
		var resultStr string

		if res.Error != nil {
			status = "failed"
			if res.Error.Error() == "timeout" {
				resultStr = fmt.Sprintf("❌ [%s] Timeout", res.IP)
			} else {
				resultStr = fmt.Sprintf("❌ [%s] Fallo:\n%s", res.IP, res.Output)
			}
		} else {
			resultStr = fmt.Sprintf("✅ [%s] OK:\n%s", res.IP, res.Output)
		}

		mu.Lock()
		finalResults = append(finalResults, resultStr)
		mu.Unlock()

		// Emisión del log de cumplimiento (Compliance Audit)
		logger.AuditEvent(
			"network_ping",
			res.IP,
			res.TimeMs,
			status,
			"ai-agent-mcp", // Aquí en el futuro puedes inyectar el ID de sesión del usuario en Teams o WhatsApp
			resultStr,
		)
	}

	return mcp.NewToolResultText(strings.Join(finalResults, "\n\n")), nil
}
