package app

import "ai-content-engine-starter/internal/platform/config"

// App is the top-level application entry point.
type App struct {
	cfg config.Config
}

// New creates a new application instance.
func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	return &App{cfg: cfg}, nil
}

// Run starts the application lifecycle.
func (a *App) Run() error {
	return nil
}
