package app

import (
	"log/slog"

	"ai-content-engine-starter/internal/platform/config"
	"ai-content-engine-starter/internal/platform/logger"
)

// App is the top-level application entry point.
type App struct {
	cfg    config.Config
	logger *slog.Logger
}

// New creates a new application instance.
func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	return &App{
		cfg:    cfg,
		logger: logger.New(cfg.AppEnv),
	}, nil
}

// Run starts the application lifecycle.
func (a *App) Run() error {
	a.logger.Info("application started", "app_env", a.cfg.AppEnv, "http_port", a.cfg.HTTPPort)
	return nil
}
