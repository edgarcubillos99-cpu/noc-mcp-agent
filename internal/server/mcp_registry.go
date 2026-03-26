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

	// Aquí agregarías traceroute y nmap siguiendo la misma lógica:
	// s.AddTool(toolTraceroute, tools.TracerouteHandler)
	// s.AddTool(toolNmap, tools.NmapHandler)

	// Iniciar servidor
	return server.ServeStdio(s)
}