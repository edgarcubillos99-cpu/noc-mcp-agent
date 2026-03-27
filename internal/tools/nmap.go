package tools

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"time"

	"noc-mcp/pkg/config"
	"noc-mcp/pkg/util"

	"github.com/mark3labs/mcp-go/mcp"
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

	// Los escaneos nmap pueden ser lentos, asignamos 60 segundos de timeout
	cmdCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "nmap", args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		if cmdCtx.Err() == context.DeadlineExceeded {
			return mcp.NewToolResultError(fmt.Sprintf("Timeout excedido (60s) escaneando %s.\nSalida parcial:\n%s", targetRaw, string(out))), nil
		}
		return mcp.NewToolResultError(fmt.Sprintf("Error escaneando %s:\n%s\n%s", targetRaw, err.Error(), string(out))), nil
	}

	return mcp.NewToolResultText(string(out)), nil
}
