package app

// App is the top-level application entry point.
type App struct{}

// New creates a new application instance.
func New() *App {
	return &App{}
}

// Run starts the application lifecycle.
func (a *App) Run() error {
	return nil
}
