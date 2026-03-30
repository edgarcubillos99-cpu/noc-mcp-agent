package main

import (
	"noc-mcp/internal/server"
	"noc-mcp/internal/tools"
	"noc-mcp/pkg/config"
	"noc-mcp/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	config.Load()
	logger.InitLogger()
	defer logger.Log.Sync()

	tools.InitNmapSemaphore()

	logger.Log.Info("Iniciando Agente MCP para NOC Telco...",
		zap.String("env", config.App.AppEnv),
		zap.String("port", config.App.HTTPPort),
	)

	if err := server.SetupAndRun(config.App.HTTPPort); err != nil {
		logger.Log.Fatal("Error fatal iniciando el servidor MCP", zap.Error(err))
	}
}
