package logger

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New() (*zap.SugaredLogger, error) {
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))

	var lvl zapcore.Level
	switch levelStr {
	case "debug":
		lvl = zap.DebugLevel
	case "info":
		lvl = zap.InfoLevel
	case "warn":
		lvl = zap.WarnLevel
	case "error":
		lvl = zap.ErrorLevel
	default:
		lvl = zap.InfoLevel
	}

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(lvl),
		Development: true,
		Encoding:    "console",
	}

	zapLogger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("logger init error: %w", err)
	}
	return zapLogger.Sugar(), nil
}
