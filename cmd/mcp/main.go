package main

import (
	"noc-mcp/internal/server"
	"noc-mcp/internal/tools"
	"noc-mcp/pkg/config"
	"noc-mcp/pkg/logger"

	"go.uber.org/zap"
)

func main() {
	// 1. Cargar variables de entorno
	config.Load()

	// 2. Inicializar observabilidad
	logger.InitLogger()
	defer logger.Log.Sync()

	// 3. Inicializar semáforos de concurrencia
	tools.InitNmapSemaphore()

	logger.Log.Info("Iniciando Agente MCP para NOC Telco...", zap.String("env", config.App.AppEnv))

	// 4. Arrancar servidor en el puerto configurado
	if err := server.SetupAndRun(config.App.HTTPPort); err != nil {
		logger.Log.Fatal("Error fatal iniciando el servidor MCP", zap.Error(err))
	}
}
