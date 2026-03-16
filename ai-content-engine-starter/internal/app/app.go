package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-content-engine-starter/internal/admin"
	"ai-content-engine-starter/internal/domain"
	"ai-content-engine-starter/internal/platform/config"
	"ai-content-engine-starter/internal/platform/logger"
	"ai-content-engine-starter/internal/platform/postgres"
	"ai-content-engine-starter/internal/platform/redis"
	"ai-content-engine-starter/internal/webui"
)

const shutdownTimeout = 5 * time.Second

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
	if err := postgres.ValidateDSN(cfg.PostgresDSN); err != nil {
		return nil, fmt.Errorf("validate postgres dsn: %w", err)
	}
	if err := redis.ValidateAddr(cfg.RedisAddr); err != nil {
		return nil, fmt.Errorf("validate redis addr: %w", err)
	}

	return &App{
		cfg:    cfg,
		logger: logger.New(cfg.AppEnv),
	}, nil
}

// Run starts the application lifecycle.
func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.cfg.HTTPPort),
		Handler: a.routes(),
	}

	errCh := make(chan error, 1)
	go func() {
		a.logger.Info("http server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("listen and serve: %w", err)
		}
	case <-ctx.Done():
		a.logger.Info("shutdown signal received", "reason", ctx.Err())
		select {
		case err := <-errCh:
			if err != nil {
				return fmt.Errorf("listen and serve: %w", err)
			}
		default:
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	a.logger.Info("http server stopped")
	return nil
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	adminHandler, err := admin.NewHandler(adminFallbackDraftRepository{})
	if err == nil {
		_ = adminHandler.Register(mux)
	}

	_ = webui.Register(mux)
	return mux
}

// adminFallbackDraftRepository keeps admin endpoints reachable even when storage wiring is not yet configured.
type adminFallbackDraftRepository struct{}

func (adminFallbackDraftRepository) Create(context.Context, domain.Draft) (domain.Draft, error) {
	return domain.Draft{}, errors.New("draft repository is unavailable")
}

func (adminFallbackDraftRepository) GetByID(context.Context, int64) (domain.Draft, error) {
	return domain.Draft{}, errors.New("draft repository is unavailable")
}

func (adminFallbackDraftRepository) ListByStatus(context.Context, domain.DraftStatus, int) ([]domain.Draft, error) {
	return nil, errors.New("draft repository is unavailable")
}

func (adminFallbackDraftRepository) UpdateStatus(context.Context, int64, domain.DraftStatus) error {
	return errors.New("draft repository is unavailable")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
