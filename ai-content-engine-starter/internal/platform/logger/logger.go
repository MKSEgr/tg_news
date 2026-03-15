package logger

import (
	"log/slog"
	"os"
	"strings"
)

const developmentEnv = "development"

// New creates a structured logger configured for the application environment.
func New(appEnv string) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: levelForEnv(appEnv),
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

func levelForEnv(appEnv string) slog.Level {
	if strings.EqualFold(appEnv, developmentEnv) {
		return slog.LevelDebug
	}

	return slog.LevelInfo
}
