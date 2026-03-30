package server

import (
	"fmt"
	"net/http"
	"noc-mcp/internal/middleware"
	"noc-mcp/internal/tools"
	"noc-mcp/pkg/config"
	"noc-mcp/pkg/logger"

	"go.uber.org/zap"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func SetupAndRun(port string) error {
	s := server.NewMCPServer("noc-telco-agent", "1.0.0")

	toolPing := mcp.NewTool("network_ping",
		mcp.WithDescription("Ejecuta pings concurrentes a uno o más hosts y devuelve un resumen de conectividad (pérdida de paquetes, RTT)."),
		mcp.WithArray("targets", mcp.Required(), mcp.Description("Array de IPs o FQDNs a verificar."), mcp.WithStringItems()),
		mcp.WithNumber("count", mcp.Description("Cantidad de paquetes ICMP por host (default: 4).")),
	)
	s.AddTool(toolPing, tools.PingHandler)

	toolNmap := mcp.NewTool("network_nmap",
		mcp.WithDescription("Escanea puertos de un host con nmap (-Pn) y devuelve un resumen de puertos abiertos/filtrados."),
		mcp.WithString("target", mcp.Required(), mcp.Description("IP o FQDN a escanear.")),
		mcp.WithString("ports", mcp.Description("Puertos opcionales (ej: '80,443' o '1-1000').")),
	)
	s.AddTool(toolNmap, tools.NmapHandler)

	toolTraceroute := mcp.NewTool("network_traceroute",
		mcp.WithDescription("Traza la ruta de red hacia un destino y devuelve un resumen de saltos, latencias y timeouts."),
		mcp.WithString("target", mcp.Required(), mcp.Description("IP o FQDN destino.")),
	)
	s.AddTool(toolTraceroute, tools.TracerouteHandler)

	sse := server.NewSSEServer(s)

	var handler http.Handler = sse
	if config.App.APIKey != "" {
		handler = middleware.BearerAuth(config.App.APIKey, sse)
		logger.Log.Info("Autenticación Bearer activada para el servidor MCP")
	} else {
		logger.Log.Warn("API_KEY no configurada — servidor MCP sin autenticación (no recomendado en producción)")
	}

	addr := fmt.Sprintf(":%s", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	logger.Log.Info("Servidor MCP HTTP/SSE en ejecución",
		zap.String("addr", addr),
		zap.Bool("auth_enabled", config.App.APIKey != ""),
	)

	return srv.ListenAndServe()
}
