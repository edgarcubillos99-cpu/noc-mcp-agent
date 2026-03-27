package tools

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"noc-mcp/pkg/util"

	"github.com/mark3labs/mcp-go/mcp"
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

	// Timeout de 45 segundos para traceroute (puede ser lento en redes Telco)
	cmdCtx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Usamos -n para evitar resolución DNS lenta y -m 30 para max saltos
	cmd := exec.CommandContext(cmdCtx, "traceroute", "-n", "-m", "30", targetRaw)
	out, err := cmd.CombinedOutput()

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			return mcp.NewToolResultError(fmt.Sprintf("Timeout excedido trazando la ruta a %s.\nSalida parcial:\n%s", targetRaw, string(out))), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Error ejecutando traceroute hacia %s:\n%s\n%s", targetRaw, err.Error(), string(out))), nil
	}

	return mcp.NewToolResultText(string(out)), nil
}
