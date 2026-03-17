package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
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
	drafts domain.DraftRepository
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
		drafts: newAdminMemoryDraftRepository(),
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
	if a != nil && a.drafts != nil {
		adminHandler, err = admin.NewHandler(a.drafts)
	}
	if err == nil {
		_ = adminHandler.Register(mux)
	}

	if a != nil && a.cfg.Features.WebUI {
		_ = webui.Register(mux)
	}
	return mux
}

// adminMemoryDraftRepository is an in-process repository used by app runtime wiring.
// It keeps moderation endpoints functional when external storage wiring is not configured.
type adminMemoryDraftRepository struct {
	mu     sync.RWMutex
	drafts map[int64]domain.Draft
}

func newAdminMemoryDraftRepository() *adminMemoryDraftRepository {
	return &adminMemoryDraftRepository{drafts: map[int64]domain.Draft{}}
}

func (r *adminMemoryDraftRepository) Create(_ context.Context, draft domain.Draft) (domain.Draft, error) {
	if r == nil {
		return domain.Draft{}, errors.New("draft repository is unavailable")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if draft.ID <= 0 {
		draft.ID = int64(len(r.drafts) + 1)
	}
	r.drafts[draft.ID] = draft
	return draft, nil
}

func (r *adminMemoryDraftRepository) GetByID(_ context.Context, id int64) (domain.Draft, error) {
	if r == nil {
		return domain.Draft{}, errors.New("draft repository is unavailable")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	draft, ok := r.drafts[id]
	if !ok {
		return domain.Draft{}, domain.ErrNotFound
	}
	return draft, nil
}

func (r *adminMemoryDraftRepository) ListByStatus(_ context.Context, status domain.DraftStatus, limit int) ([]domain.Draft, error) {
	if r == nil {
		return nil, errors.New("draft repository is unavailable")
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if limit <= 0 {
		return nil, errors.New("limit must be greater than zero")
	}
	list := make([]domain.Draft, 0, len(r.drafts))
	for _, draft := range r.drafts {
		if draft.Status == status {
			list = append(list, draft)
		}
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID > list[j].ID })
	if len(list) > limit {
		list = list[:limit]
	}
	return list, nil
}

func (r *adminMemoryDraftRepository) UpdateStatus(_ context.Context, id int64, status domain.DraftStatus) error {
	if r == nil {
		return errors.New("draft repository is unavailable")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	draft, ok := r.drafts[id]
	if !ok {
		return domain.ErrNotFound
	}
	draft.Status = status
	r.drafts[id] = draft
	return nil
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
