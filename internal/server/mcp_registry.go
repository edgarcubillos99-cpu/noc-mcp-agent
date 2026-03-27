package server

import (
	"noc-mcp/internal/tools"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SetupAndRun inicializa el servidor y mapea las herramientas
func SetupAndRun() error {
	s := server.NewMCPServer("noc-telco-agent", "1.0.0")

	// 1. Tool: Ping
	toolPing := mcp.NewTool("network_ping",
		mcp.WithDescription("Ejecuta pings concurrentes en el NOC."),
		mcp.WithArray("targets", mcp.Required(), mcp.Description("Array de IPs o FQDNs."), mcp.WithStringItems()),
		mcp.WithNumber("count", mcp.Description("Cantidad de paquetes por host.")),
	)
	s.AddTool(toolPing, tools.PingHandler)

	// 2. Tool: Nmap
	toolNmap := mcp.NewTool("network_nmap",
		mcp.WithDescription("Ejecuta un escaneo de puertos Nmap silencioso (-Pn)."),
		mcp.WithString("target", mcp.Required(), mcp.Description("IP o FQDN a escanear.")),
		mcp.WithString("ports", mcp.Description("Puertos opcionales (ej: '80,443' o '1-1000').")),
	)
	s.AddTool(toolNmap, tools.NmapHandler)

	// 3. Tool: Traceroute
	toolTraceroute := mcp.NewTool("network_traceroute",
		mcp.WithDescription("Traza la ruta de red hacia un destino para detectar cuellos de botella o caídas."),
		mcp.WithString("target", mcp.Required(), mcp.Description("IP o FQDN destino.")),
	)
	s.AddTool(toolTraceroute, tools.TracerouteHandler)

	// Iniciar servidor
	return server.ServeStdio(s)
}
