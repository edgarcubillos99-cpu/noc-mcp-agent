package main

import (
	"noc-mcp/internal/server"
	"noc-mcp/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	// Inicializar observabilidad
	logger.InitLogger()
	defer logger.Log.Sync() // Asegurar que los logs en buffer se escriban al salir

	logger.Log.Info("Iniciando Agente MCP para NOC Telco...")

	// Arrancar servidor
	if err := server.SetupAndRun(); err != nil {
		logger.Log.Fatal("Error fatal iniciando el servidor MCP", zap.Error(err))
	}
}
