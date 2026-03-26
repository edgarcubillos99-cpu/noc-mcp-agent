package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"noc-mcp/pkg/util"
	"github.com/mark3labs/mcp-go/mcp"
)

// PingHandler maneja la ejecución concurrente de pings
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

	var wg sync.WaitGroup
	var mu sync.Mutex
	var results []string

	for _, tRaw := range targetsRaw {
		target, ok := tRaw.(string)
		if !ok || !util.IsValidTarget(target) {
			mu.Lock()
			results = append(results, fmt.Sprintf("⚠️ Destino inválido: %v", tRaw))
			mu.Unlock()
			continue
		}

		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			cmdCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			cmd := exec.CommandContext(cmdCtx, "ping", "-c", count, ip)
			out, err := cmd.CombinedOutput()

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if cmdCtx.Err() == context.DeadlineExceeded {
					results = append(results, fmt.Sprintf("❌ [%s] Timeout", ip))
				} else {
					results = append(results, fmt.Sprintf("❌ [%s] Fallo:\n%s", ip, string(out)))
				}
			} else {
				results = append(results, fmt.Sprintf("✅ [%s] OK:\n%s", ip, string(out)))
			}
		}(target)
	}

	wg.Wait()
	return mcp.NewToolResultText(strings.Join(results, "\n\n")), nil
}
