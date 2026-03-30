package tools

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"noc-mcp/pkg/logger"
	"noc-mcp/pkg/parser"
	"noc-mcp/pkg/util"

	"github.com/mark3labs/mcp-go/mcp"
	"go.uber.org/zap"
)

// TracerouteHandler ejecuta un traceroute para analizar saltos de red
func TracerouteHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toolArgs := request.GetArguments()
	targetRaw, ok := toolArgs["target"].(string)
	if !ok {
		return mcp.NewToolResultError("El parámetro 'target' es obligatorio y debe ser un string."), nil
	}

	if !util.IsValidTarget(targetRaw) {
		return mcp.NewToolResultError("Destino inválido por políticas de seguridad."), nil
	}

	start := time.Now()
	cmdCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "traceroute", "-n", "-m", "30", targetRaw)
	out, err := cmd.CombinedOutput()
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		status := "failed"
		var summary string
		if cmdCtx.Err() == context.DeadlineExceeded {
			summary = fmt.Sprintf("[TIMEOUT] traceroute a %s excedió 45s", targetRaw)
		} else {
			summary = fmt.Sprintf("[ERROR] traceroute a %s: %s", targetRaw, err.Error())
		}
		logger.AuditEvent("network_traceroute", targetRaw, elapsed, status, "ai-agent-mcp", summary)
		return mcp.NewToolResultError(summary), nil
	}

	summary := parser.SummarizeTraceroute(string(out))
	logger.AuditEvent("network_traceroute", targetRaw, elapsed, "success", "ai-agent-mcp", summary)
	logger.Log.Debug("traceroute raw output", zap.String("target", targetRaw), zap.String("raw", string(out)))

	return mcp.NewToolResultText(summary), nil
}
