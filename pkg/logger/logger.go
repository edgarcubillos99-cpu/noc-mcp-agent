package logger

import (
	"noc-mcp/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// InitLogger configura un logger estructurado en formato JSON
func InitLogger() {
	var cfg zap.Config

	if config.App.AppEnv == "development" {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	// Ajustar el nivel de log dinámicamente
	switch config.App.LogLevel {
	case "debug":
		cfg.Level.SetLevel(zap.DebugLevel)
	case "warn":
		cfg.Level.SetLevel(zap.WarnLevel)
	case "error":
		cfg.Level.SetLevel(zap.ErrorLevel)
	default:
		cfg.Level.SetLevel(zap.InfoLevel)
	}

	var err error
	Log, err = cfg.Build()
	if err != nil {
		panic("Error crítico inicializando zap logger: " + err.Error())
	}
}

// AuditEvent emite un log de auditoría especializado para compliance
func AuditEvent(toolName, target string, durationMs int64, status, user, resultSummary string) {
	Log.Info("AUDIT_EVENT",
		zap.String("event_type", "noc_diagnostic_execution"),
		zap.String("tool_name", toolName),
		zap.String("target", target),
		zap.Int64("duration_ms", durationMs),
		zap.String("status", status),
		zap.String("requested_by", user),
		zap.String("result_summary", resultSummary),
	)
}
