package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultAppEnv   = "development"
	defaultHTTPPort = 8080
)

// Config contains application runtime settings loaded from environment.
type Config struct {
	AppEnv      string
	HTTPPort    int
	PostgresDSN string
}

// Load reads configuration from environment variables and applies defaults.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:      getEnvOrDefault("APP_ENV", defaultAppEnv),
		HTTPPort:    defaultHTTPPort,
		PostgresDSN: strings.TrimSpace(os.Getenv("POSTGRES_DSN")),
	}

	if rawPort := os.Getenv("HTTP_PORT"); rawPort != "" {
		port, err := strconv.Atoi(rawPort)
		if err != nil {
			return Config{}, fmt.Errorf("parse HTTP_PORT: %w", err)
		}
		if port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("HTTP_PORT out of range: %d", port)
		}
		cfg.HTTPPort = port
	}

	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("POSTGRES_DSN is required")
	}

	return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
