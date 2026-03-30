package tools

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"time"

	"noc-mcp/pkg/config"
	"noc-mcp/pkg/logger"
	"noc-mcp/pkg/parser"
	"noc-mcp/pkg/util"

	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

var nmapSemaphore chan struct{}

// InitNmapSemaphore debe llamarse al iniciar la app
func InitNmapSemaphore() {
	nmapSemaphore = make(chan struct{}, config.App.MaxNmapWorkers)
}

// NmapHandler ejecuta un escaneo de puertos contra un equipo de la red
func NmapHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 1. Adquirir un token del semáforo (Bloquea si ya hay MAX_NMAP_WORKERS corriendo)
	nmapSemaphore <- struct{}{}
	// 2. Liberar el token al terminar la función
	defer func() { <-nmapSemaphore }()
	toolArgs := request.GetArguments()
	// Extraer y validar el target
	targetRaw, ok := toolArgs["target"].(string)
	if !ok {
		return mcp.NewToolResultError("El parámetro 'target' es obligatorio y debe ser un string."), nil
	}

	if !util.IsValidTarget(targetRaw) {
		return mcp.NewToolResultError("Destino inválido por políticas de seguridad."), nil
	}

	// Construir los argumentos básicos.
	// -Pn: Vital en Telco para escanear interfaces de gestión que dropean ICMP.
	args := []string{"-Pn"}

	// Validar e inyectar puertos si fueron solicitados
	if portsRaw, ok := toolArgs["ports"].(string); ok && portsRaw != "" {
		// Validar que solo contenga números, comas o guiones (ej. "80,443" o "1-1024")
		portRegex := regexp.MustCompile(`^[0-9,-]+$`)
		if !portRegex.MatchString(portsRaw) {
			return mcp.NewToolResultError("Formato de puertos no permitido. Usa formatos como '80,443' o '1-100'."), nil
		}
		args = append(args, "-p", portsRaw)
	}

	// Añadir el target al final de los argumentos
	args = append(args, targetRaw)

	start := time.Now()
	cmdCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "nmap", args...)
	out, err := cmd.CombinedOutput()

	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		status := "failed"
		var summary string
		if cmdCtx.Err() == context.DeadlineExceeded {
			summary = fmt.Sprintf("[TIMEOUT] nmap a %s excedió 60s", targetRaw)
		} else {
			summary = fmt.Sprintf("[ERROR] nmap a %s: %s", targetRaw, err.Error())
		}
		logger.AuditEvent("network_nmap", targetRaw, elapsed, status, "ai-agent-mcp", summary)
		return mcp.NewToolResultError(summary), nil
	}

	summary := parser.SummarizeNmap(string(out))
	logger.AuditEvent("network_nmap", targetRaw, elapsed, "success", "ai-agent-mcp", summary)
	logger.Log.Debug("nmap raw output", zap.String("target", targetRaw), zap.String("raw", string(out)))

	return mcp.NewToolResultText(summary), nil
}
