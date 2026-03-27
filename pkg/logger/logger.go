package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

// InitLogger configura un logger estructurado en formato JSON
func InitLogger() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	Log, err = config.Build()
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
