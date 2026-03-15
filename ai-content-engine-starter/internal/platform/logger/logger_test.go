package logger

import (
	"log/slog"
	"testing"
)

func TestLevelForEnvDevelopment(t *testing.T) {
	level := levelForEnv("development")
	if level != slog.LevelDebug {
		t.Fatalf("levelForEnv(development) = %v, want %v", level, slog.LevelDebug)
	}
}

func TestLevelForEnvDevelopmentCaseInsensitive(t *testing.T) {
	level := levelForEnv("Development")
	if level != slog.LevelDebug {
		t.Fatalf("levelForEnv(Development) = %v, want %v", level, slog.LevelDebug)
	}
}

func TestLevelForEnvNonDevelopment(t *testing.T) {
	level := levelForEnv("production")
	if level != slog.LevelInfo {
		t.Fatalf("levelForEnv(production) = %v, want %v", level, slog.LevelInfo)
	}
}
