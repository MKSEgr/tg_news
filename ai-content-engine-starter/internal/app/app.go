package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"ai-content-engine-starter/internal/admin"
	"ai-content-engine-starter/internal/collector"
	githubcollector "ai-content-engine-starter/internal/collector/github"
	producthuntcollector "ai-content-engine-starter/internal/collector/producthunt"
	redditcollector "ai-content-engine-starter/internal/collector/reddit"
	rsscollector "ai-content-engine-starter/internal/collector/rss"
	"ai-content-engine-starter/internal/dedup"
	"ai-content-engine-starter/internal/domain"
	"ai-content-engine-starter/internal/editorial"
	"ai-content-engine-starter/internal/editorialplanner"
	"ai-content-engine-starter/internal/generator"
	"ai-content-engine-starter/internal/normalizer"
	"ai-content-engine-starter/internal/orchestration"
	"ai-content-engine-starter/internal/platform/config"
	"ai-content-engine-starter/internal/platform/logger"
	"ai-content-engine-starter/internal/platform/postgres"
	"ai-content-engine-starter/internal/platform/redis"
	"ai-content-engine-starter/internal/platform/yandexai"
	"ai-content-engine-starter/internal/publisher"
	"ai-content-engine-starter/internal/router"
	"ai-content-engine-starter/internal/scheduler"
	"ai-content-engine-starter/internal/scorer"
	"ai-content-engine-starter/internal/seed"
	"ai-content-engine-starter/internal/webui"
	_ "github.com/lib/pq"
)

const shutdownTimeout = 5 * time.Second

// App is the top-level application entry point.
type App struct {
	cfg         config.Config
	logger      *slog.Logger
	db          *sql.DB
	drafts      domain.DraftRepository
	runtime     *runtimeLoops
	stats       *runtimeStats
	statusCheck func(context.Context) error
	openDB      func(driverName, dsn string) (*sql.DB, error)
	pingRedisFn func(context.Context, string) error
}

type runtimeLoops struct {
	collector *scheduler.Scheduler
	pipeline  *scheduler.Scheduler
	publisher *scheduler.Scheduler
}

type runtimeStats struct {
	collectedItems    atomic.Int64
	pipelineRuns      atomic.Int64
	plannerFailures   atomic.Int64
	assetFailures     atomic.Int64
	clusterFailures   atomic.Int64
	publishedDrafts   atomic.Int64
	publishFailures   atomic.Int64
	collectorFailures atomic.Int64
	pipelineFailures  atomic.Int64
	loopRestarts      atomic.Int64
	lastSuccessUnix   atomic.Int64
	lastFailureUnix   atomic.Int64
}

type publisherClient interface {
	PublishDraft(ctx context.Context, draft domain.Draft, chatID string) (int64, error)
}

type publisherLoop struct {
	logger       *slog.Logger
	drafts       domain.DraftRepository
	channels     domain.ChannelRepository
	publisher    publisherClient
	chatIDBySlug map[string]string
	stats        *runtimeStats
	limit        int
}

type claimDraftRepository interface {
	domain.DraftRepository
	UpdateStatusIfCurrent(ctx context.Context, id int64, current domain.DraftStatus, next domain.DraftStatus) (bool, error)
}

// New creates a new application instance.
func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	log := logger.New(cfg.AppEnv)
	app := &App{
		cfg:         cfg,
		logger:      log,
		stats:       &runtimeStats{},
		drafts:      adminFallbackDraftRepository{},
		openDB:      sql.Open,
		pingRedisFn: pingRedis,
	}
	if err := app.initRuntime(context.Background()); err != nil {
		if app.db != nil {
			_ = app.db.Close()
		}
		return nil, err
	}
	return app, nil
}

func (a *App) initRuntime(ctx context.Context) error {
	if a == nil {
		return fmt.Errorf("app is nil")
	}
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	if err := postgres.ValidateDSN(a.cfg.PostgresDSN); err != nil {
		return fmt.Errorf("validate postgres dsn: %w", err)
	}
	if err := redis.ValidateAddr(a.cfg.RedisAddr); err != nil {
		return fmt.Errorf("validate redis addr: %w", err)
	}
	if a.openDB == nil {
		a.openDB = sql.Open
	}
	if a.pingRedisFn == nil {
		a.pingRedisFn = pingRedis
	}
	db, err := a.openDB("postgres", a.cfg.PostgresDSN)
	if err != nil {
		return fmt.Errorf("open postgres: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return fmt.Errorf("ping postgres: %w", err)
	}
	if err := a.pingRedisFn(ctx, a.cfg.RedisAddr); err != nil {
		_ = db.Close()
		return fmt.Errorf("ping redis: %w", err)
	}
	a.db = db
	if !a.cfg.EnablePipeline && !a.cfg.EnablePublisher {
		drafts, err := newAdminFileDraftRepository(filepath.Join(os.TempDir(), "ai-content-engine-starter-admin-drafts.json"))
		if err != nil {
			return fmt.Errorf("create admin draft repository: %w", err)
		}
		a.drafts = drafts
		a.statusCheck = func(ctx context.Context) error {
			if err := a.db.PingContext(ctx); err != nil {
				return fmt.Errorf("ping postgres: %w", err)
			}
			return a.pingRedisFn(ctx, a.cfg.RedisAddr)
		}
		return nil
	}

	a.statusCheck = func(ctx context.Context) error {
		if err := a.db.PingContext(ctx); err != nil {
			return fmt.Errorf("ping postgres: %w", err)
		}
		return a.pingRedisFn(ctx, a.cfg.RedisAddr)
	}

	channels := postgres.NewChannelRepository(a.db)
	sources := postgres.NewSourceRepository(a.db)
	items := postgres.NewSourceItemRepository(a.db)
	drafts := postgres.NewDraftRepository(a.db)
	publishIntents := postgres.NewPublishIntentRepository(a.db)
	assets := postgres.NewContentAssetRepository(a.db)

	if err := seed.New(channels, sources).Seed(ctx); err != nil {
		return fmt.Errorf("seed defaults: %w", err)
	}
	a.drafts = drafts

	if a.cfg.EnablePipeline {
		collectorFramework, pipelineJob, err := a.buildPipelineRuntime(sources, items, channels, drafts, publishIntents, assets)
		if err != nil {
			return err
		}
		collectorSched, err := scheduler.New(a.cfg.LoopInterval, a.wrapLoop("collector", func(ctx context.Context) error {
			if err := collectorFramework.RunOnce(ctx); err != nil {
				a.stats.collectorFailures.Add(1)
				return err
			}
			recent, listErr := items.ListRecent(ctx, a.cfg.RecentItemsLimit)
			if listErr == nil {
				a.stats.collectedItems.Store(int64(len(recent)))
			}
			return nil
		}))
		if err != nil {
			return fmt.Errorf("create collector scheduler: %w", err)
		}
		pipelineSched, err := scheduler.New(a.cfg.LoopInterval, a.wrapLoop("pipeline", func(ctx context.Context) error {
			a.stats.pipelineRuns.Add(1)
			if err := pipelineJob.Run(ctx); err != nil {
				a.stats.pipelineFailures.Add(1)
				return err
			}
			return nil
		}))
		if err != nil {
			return fmt.Errorf("create pipeline scheduler: %w", err)
		}
		a.runtime = &runtimeLoops{collector: collectorSched, pipeline: pipelineSched}
	}
	if a.cfg.EnablePublisher {
		publishClient, err := publisher.New(http.DefaultClient, a.cfg.TelegramBotToken)
		if err != nil {
			return fmt.Errorf("create publisher client: %w", err)
		}
		loop := &publisherLoop{
			logger:       a.logger,
			drafts:       drafts,
			channels:     channels,
			publisher:    publishClient,
			chatIDBySlug: a.cfg.ChannelChatMap,
			stats:        a.stats,
			limit:        a.cfg.PublisherBatchSize,
		}
		publisherSched, err := scheduler.NewWithRetry(a.cfg.LoopInterval, a.wrapLoop("publisher", loop.RunOnce), scheduler.RetryPolicy{MaxAttempts: 2, Backoff: time.Second})
		if err != nil {
			return fmt.Errorf("create publisher scheduler: %w", err)
		}
		if a.runtime == nil {
			a.runtime = &runtimeLoops{}
		}
		a.runtime.publisher = publisherSched
	}
	return nil
}

func (a *App) buildPipelineRuntime(
	sources domain.SourceRepository,
	items domain.SourceItemRepository,
	channels domain.ChannelRepository,
	drafts domain.DraftRepository,
	publishIntents domain.PublishIntentRepository,
	assets domain.ContentAssetRepository,
) (*collector.Framework, *orchestration.PipelineJob, error) {
	aiClient, err := yandexai.New(http.DefaultClient, yandexai.Config{APIKey: a.cfg.YandexAIAPIKey, ModelURI: a.cfg.YandexAIModelURI})
	if err != nil {
		return nil, nil, fmt.Errorf("create yandex ai client: %w", err)
	}
	gen, err := generator.New(aiClient)
	if err != nil {
		return nil, nil, fmt.Errorf("create generator: %w", err)
	}
	dedupSvc, err := dedup.New(items, 200)
	if err != nil {
		return nil, nil, fmt.Errorf("create dedup service: %w", err)
	}
	planner, err := editorialplanner.New(publishIntents, channels, scorer.New(nil), router.New())
	if err != nil {
		return nil, nil, fmt.Errorf("create editorial planner: %w", err)
	}
	assetGen, err := newRuntimeAssetGenerator(items, assets)
	if err != nil {
		return nil, nil, fmt.Errorf("create asset generator: %w", err)
	}
	pipelineJob, err := orchestration.NewPipelineJob(
		sources,
		items,
		channels,
		drafts,
		normalizer.New(),
		dedupSvc,
		scorer.New(nil),
		router.New(),
		gen,
		editorial.NewGuard(),
		nil,
		nil,
		nil,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create pipeline job: %w", err)
	}
	pipelineJob.WithEditorialPlanner(planner)
	pipelineJob.WithIntentAssetGenerator(assetGen)
	pipelineJob.WithBatchLimits(a.cfg.RecentItemsLimit, a.cfg.DraftScanLimit)
	pipelineJob.WithPlannerErrorHook(func(item domain.SourceItem, err error) {
		a.stats.plannerFailures.Add(1)
		a.logger.Error("planner failed", "source_item_id", item.ID, "error", err)
	})
	pipelineJob.WithAssetErrorHook(func(intent domain.PublishIntent, err error) {
		a.stats.assetFailures.Add(1)
		a.logger.Error("asset generation failed", "publish_intent_id", intent.ID, "raw_item_id", intent.RawItemID, "channel_id", intent.ChannelID, "error", err)
	})
	pipelineJob.WithStoryClusterErrorHook(func(item domain.SourceItem, err error) {
		a.stats.clusterFailures.Add(1)
		a.logger.Error("cluster observation failed", "source_item_id", item.ID, "error", err)
	})

	collectorFramework, err := collector.New(
		sources,
		items,
		rsscollector.New(http.DefaultClient),
		githubcollector.New(http.DefaultClient),
		redditcollector.New(http.DefaultClient),
		producthuntcollector.New(http.DefaultClient),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create collector framework: %w", err)
	}
	return collectorFramework, pipelineJob, nil
}

// Run starts the application lifecycle.
func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", a.cfg.HTTPPort),
		Handler: a.routes(),
	}

	errCh := make(chan error, 4)
	go func() {
		a.logger.Info("http server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen and serve: %w", err)
		}
	}()
	if a.runtime != nil {
		if a.runtime.collector != nil {
			go a.superviseLoop(ctx, "collector", a.runtime.collector)
		}
		if a.runtime.pipeline != nil {
			go a.superviseLoop(ctx, "pipeline", a.runtime.pipeline)
		}
		if a.runtime.publisher != nil {
			go a.superviseLoop(ctx, "publisher", a.runtime.publisher)
		}
	}

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	case <-ctx.Done():
		a.logger.Info("shutdown signal received", "reason", ctx.Err())
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			return fmt.Errorf("close postgres: %w", err)
		}
	}
	a.logger.Info("http server stopped")
	return nil
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/health/runtime", a.runtimeHealthHandler)

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

func (a *App) runtimeHealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	status := map[string]any{
		"status":             "ok",
		"enable_pipeline":    a != nil && a.cfg.EnablePipeline,
		"enable_publisher":   a != nil && a.cfg.EnablePublisher,
		"collector_failures": int64(0),
		"pipeline_runs":      int64(0),
		"pipeline_failures":  int64(0),
		"planner_failures":   int64(0),
		"asset_failures":     int64(0),
		"cluster_failures":   int64(0),
		"published_drafts":   int64(0),
		"publish_failures":   int64(0),
		"loop_restarts":      int64(0),
		"last_success_unix":  int64(0),
		"last_failure_unix":  int64(0),
	}
	if a != nil && a.stats != nil {
		status["collector_failures"] = a.stats.collectorFailures.Load()
		status["pipeline_runs"] = a.stats.pipelineRuns.Load()
		status["pipeline_failures"] = a.stats.pipelineFailures.Load()
		status["planner_failures"] = a.stats.plannerFailures.Load()
		status["asset_failures"] = a.stats.assetFailures.Load()
		status["cluster_failures"] = a.stats.clusterFailures.Load()
		status["published_drafts"] = a.stats.publishedDrafts.Load()
		status["publish_failures"] = a.stats.publishFailures.Load()
		status["loop_restarts"] = a.stats.loopRestarts.Load()
		status["last_success_unix"] = a.stats.lastSuccessUnix.Load()
		status["last_failure_unix"] = a.stats.lastFailureUnix.Load()
	}
	if a != nil && a.statusCheck != nil {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		if err := a.statusCheck(ctx); err != nil {
			status["status"] = "degraded"
			status["error"] = err.Error()
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}

func (a *App) wrapLoop(name string, fn func(context.Context) error) scheduler.Job {
	return func(ctx context.Context) error {
		started := time.Now()
		a.logger.Info("runtime loop started", "loop", name)
		err := fn(ctx)
		if err != nil {
			a.logger.Error("runtime loop failed", "loop", name, "duration", time.Since(started), "error", err)
			a.stats.lastFailureUnix.Store(time.Now().UTC().Unix())
			return err
		}
		a.stats.lastSuccessUnix.Store(time.Now().UTC().Unix())
		a.logger.Info("runtime loop completed", "loop", name, "duration", time.Since(started))
		return nil
	}
}

func (a *App) superviseLoop(ctx context.Context, name string, loop *scheduler.Scheduler) {
	if a == nil || loop == nil {
		return
	}
	for {
		err := loop.Run(ctx)
		if err == nil || errors.Is(err, context.Canceled) || ctx.Err() != nil {
			return
		}
		a.stats.loopRestarts.Add(1)
		a.logger.Error("runtime loop crashed; continuing in degraded mode", "loop", name, "error", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func pingRedis(ctx context.Context, addr string) error {
	d := net.Dialer{Timeout: time.Second}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	return conn.Close()
}

func newRuntimeAssetGenerator(items domain.SourceItemRepository, assets domain.ContentAssetRepository) (*runtimeAssetGenerator, error) {
	if items == nil {
		return nil, fmt.Errorf("source item repository is nil")
	}
	if assets == nil {
		return nil, fmt.Errorf("content asset repository is nil")
	}
	return &runtimeAssetGenerator{items: items, assets: assets}, nil
}

type runtimeAssetGenerator struct {
	items  domain.SourceItemRepository
	assets domain.ContentAssetRepository
}

func (g *runtimeAssetGenerator) GenerateFromIntent(ctx context.Context, intent domain.PublishIntent) (domain.ContentAsset, error) {
	item, err := g.items.GetByID(ctx, intent.RawItemID)
	if err != nil {
		return domain.ContentAsset{}, err
	}
	title := strings.TrimSpace(item.Title)
	body := ""
	if item.Body != nil {
		body = strings.TrimSpace(*item.Body)
	}
	return g.assets.Create(ctx, domain.ContentAsset{RawItemID: intent.RawItemID, ChannelID: intent.ChannelID, AssetType: intent.Format, Title: title, Body: body, Status: domain.ContentAssetStatusPending})
}

func (p *publisherLoop) RunOnce(ctx context.Context) error {
	if p == nil {
		return fmt.Errorf("publisher loop is nil")
	}
	claimer, ok := p.drafts.(claimDraftRepository)
	if !ok {
		return fmt.Errorf("draft repository does not support conditional status updates")
	}
	approved, err := p.drafts.ListByStatus(ctx, domain.DraftStatusApproved, p.limit)
	if err != nil {
		return fmt.Errorf("list approved drafts: %w", err)
	}
	channels, err := p.channels.List(ctx)
	if err != nil {
		return fmt.Errorf("list channels: %w", err)
	}
	channelByID := make(map[int64]domain.Channel, len(channels))
	for _, channel := range channels {
		channelByID[channel.ID] = channel
	}
	for _, draft := range approved {
		channel, ok := channelByID[draft.ChannelID]
		if !ok {
			p.stats.publishFailures.Add(1)
			p.logger.Error("publisher skipped draft with unknown channel", "draft_id", draft.ID, "channel_id", draft.ChannelID)
			continue
		}
		chatID := p.chatID(channel)
		if chatID == "" {
			p.stats.publishFailures.Add(1)
			p.logger.Error("publisher missing chat id", "draft_id", draft.ID, "channel_slug", channel.Slug)
			continue
		}
		claimed, err := claimer.UpdateStatusIfCurrent(ctx, draft.ID, domain.DraftStatusApproved, domain.DraftStatusPublishing)
		if err != nil {
			p.stats.publishFailures.Add(1)
			return fmt.Errorf("claim draft %d for publishing: %w", draft.ID, err)
		}
		if !claimed {
			continue
		}
		messageID, err := p.publisher.PublishDraft(ctx, draft, chatID)
		if err != nil {
			p.stats.publishFailures.Add(1)
			rolledBack, rollbackErr := claimer.UpdateStatusIfCurrent(ctx, draft.ID, domain.DraftStatusPublishing, domain.DraftStatusApproved)
			if rollbackErr != nil {
				return fmt.Errorf("rollback draft %d to approved: %w", draft.ID, rollbackErr)
			}
			if !rolledBack {
				return fmt.Errorf("rollback draft %d to approved: draft status changed concurrently", draft.ID)
			}
			p.logger.Error("publish draft failed", "draft_id", draft.ID, "channel_slug", channel.Slug, "error", err)
			continue
		}
		advanced, err := claimer.UpdateStatusIfCurrent(ctx, draft.ID, domain.DraftStatusPublishing, domain.DraftStatusPosted)
		if err != nil {
			p.stats.publishFailures.Add(1)
			return fmt.Errorf("mark draft %d posted: %w", draft.ID, err)
		}
		if !advanced {
			p.stats.publishFailures.Add(1)
			return fmt.Errorf("mark draft %d posted: draft status changed concurrently", draft.ID)
		}
		p.stats.publishedDrafts.Add(1)
		p.logger.Info("draft published", "draft_id", draft.ID, "channel_slug", channel.Slug, "chat_id", chatID, "message_id", messageID)
	}
	return nil
}

func (p *publisherLoop) chatID(channel domain.Channel) string {
	if p == nil {
		return ""
	}
	if chatID := strings.TrimSpace(p.chatIDBySlug[channel.Slug]); chatID != "" {
		return chatID
	}
	return "@" + strings.TrimSpace(channel.Slug)
}

// adminFileDraftRepository is a small persisted repository used by app runtime wiring.
// It keeps moderation state durable across restarts without adding broader pipeline wiring here.
type adminFileDraftRepository struct {
	mu     sync.RWMutex
	path   string
	drafts map[int64]domain.Draft
}

func newAdminFileDraftRepository(path string) (*adminFileDraftRepository, error) {
	path = filepath.Clean(path)
	repo := &adminFileDraftRepository{path: path, drafts: map[int64]domain.Draft{}}
	if err := repo.load(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *adminFileDraftRepository) Create(_ context.Context, draft domain.Draft) (domain.Draft, error) {
	if r == nil {
		return domain.Draft{}, errors.New("draft repository is unavailable")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if draft.ID <= 0 {
		draft.ID = int64(len(r.drafts) + 1)
	}
	r.drafts[draft.ID] = draft
	if err := r.saveLocked(); err != nil {
		return domain.Draft{}, err
	}
	return draft, nil
}

func (r *adminFileDraftRepository) GetByID(_ context.Context, id int64) (domain.Draft, error) {
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

func (r *adminFileDraftRepository) ListByStatus(_ context.Context, status domain.DraftStatus, limit int) ([]domain.Draft, error) {
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

func (r *adminFileDraftRepository) UpdateStatus(_ context.Context, id int64, status domain.DraftStatus) error {
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
	return r.saveLocked()
}

func (r *adminFileDraftRepository) UpdateStatusIfCurrent(_ context.Context, id int64, current domain.DraftStatus, next domain.DraftStatus) (bool, error) {
	if r == nil {
		return false, errors.New("draft repository is unavailable")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	draft, ok := r.drafts[id]
	if !ok {
		return false, domain.ErrNotFound
	}
	if draft.Status != current {
		return false, nil
	}
	draft.Status = next
	r.drafts[id] = draft
	return true, r.saveLocked()
}

func (r *adminFileDraftRepository) load() error {
	if r == nil {
		return errors.New("draft repository is unavailable")
	}
	body, err := os.ReadFile(r.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read admin drafts file: %w", err)
	}
	if len(body) == 0 {
		return nil
	}
	var drafts []domain.Draft
	if err := json.Unmarshal(body, &drafts); err != nil {
		return fmt.Errorf("unmarshal admin drafts file: %w", err)
	}
	for _, draft := range drafts {
		r.drafts[draft.ID] = draft
	}
	return nil
}

func (r *adminFileDraftRepository) saveLocked() error {
	list := make([]domain.Draft, 0, len(r.drafts))
	for _, draft := range r.drafts {
		list = append(list, draft)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	body, err := json.Marshal(list)
	if err != nil {
		return fmt.Errorf("marshal admin drafts file: %w", err)
	}
	if err := os.WriteFile(r.path, body, 0o644); err != nil {
		return fmt.Errorf("write admin drafts file: %w", err)
	}
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

func (adminFallbackDraftRepository) UpdateStatusIfCurrent(context.Context, int64, domain.DraftStatus, domain.DraftStatus) (bool, error) {
	return false, errors.New("draft repository is unavailable")
}

type runtimeChannelRepository struct {
	mu       sync.RWMutex
	channels map[int64]domain.Channel
	nextID   int64
}

func newRuntimeChannelRepository() *runtimeChannelRepository {
	return &runtimeChannelRepository{channels: map[int64]domain.Channel{}, nextID: 1}
}

func (r *runtimeChannelRepository) Create(_ context.Context, channel domain.Channel) (domain.Channel, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	channel.ID = r.nextID
	r.nextID++
	channel.CreatedAt = time.Now().UTC()
	r.channels[channel.ID] = channel
	return channel, nil
}
func (r *runtimeChannelRepository) GetByID(_ context.Context, id int64) (domain.Channel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	channel, ok := r.channels[id]
	if !ok {
		return domain.Channel{}, domain.ErrNotFound
	}
	return channel, nil
}
func (r *runtimeChannelRepository) List(context.Context) ([]domain.Channel, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Channel, 0, len(r.channels))
	for _, channel := range r.channels {
		out = append(out, channel)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

type runtimeSourceRepository struct {
	mu      sync.RWMutex
	sources map[int64]domain.Source
	nextID  int64
}

func newRuntimeSourceRepository() *runtimeSourceRepository {
	return &runtimeSourceRepository{sources: map[int64]domain.Source{}, nextID: 1}
}
func (r *runtimeSourceRepository) Create(_ context.Context, source domain.Source) (domain.Source, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	source.ID = r.nextID
	r.nextID++
	source.CreatedAt = time.Now().UTC()
	r.sources[source.ID] = source
	return source, nil
}
func (r *runtimeSourceRepository) GetByID(_ context.Context, id int64) (domain.Source, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	source, ok := r.sources[id]
	if !ok {
		return domain.Source{}, domain.ErrNotFound
	}
	return source, nil
}
func (r *runtimeSourceRepository) List(context.Context) ([]domain.Source, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.Source, 0, len(r.sources))
	for _, source := range r.sources {
		out = append(out, source)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
func (r *runtimeSourceRepository) ListEnabled(ctx context.Context) ([]domain.Source, error) {
	sources, err := r.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Source, 0, len(sources))
	for _, source := range sources {
		if source.Enabled {
			out = append(out, source)
		}
	}
	return out, nil
}

type runtimeSourceItemRepository struct {
	mu     sync.RWMutex
	items  map[int64]domain.SourceItem
	byKey  map[string]int64
	nextID int64
}

func newRuntimeSourceItemRepository() *runtimeSourceItemRepository {
	return &runtimeSourceItemRepository{items: map[int64]domain.SourceItem{}, byKey: map[string]int64{}, nextID: 1}
}
func (r *runtimeSourceItemRepository) Create(_ context.Context, item domain.SourceItem) (domain.SourceItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := fmt.Sprintf("%d:%s", item.SourceID, strings.TrimSpace(item.ExternalID))
	if existingID, ok := r.byKey[key]; ok {
		existing := r.items[existingID]
		existing.URL = item.URL
		existing.Title = item.Title
		existing.Body = item.Body
		existing.ImageURL = item.ImageURL
		existing.PublishedAt = item.PublishedAt
		existing.CollectedAt = time.Now().UTC()
		r.items[existingID] = existing
		return existing, nil
	}
	item.ID = r.nextID
	r.nextID++
	item.CollectedAt = time.Now().UTC()
	item.CreatedAt = item.CollectedAt
	r.items[item.ID] = item
	r.byKey[key] = item.ID
	return item, nil
}
func (r *runtimeSourceItemRepository) GetByID(_ context.Context, id int64) (domain.SourceItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok {
		return domain.SourceItem{}, domain.ErrNotFound
	}
	return item, nil
}
func (r *runtimeSourceItemRepository) ListBySourceID(_ context.Context, sourceID int64, limit int) ([]domain.SourceItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.SourceItem{}
	for _, item := range r.items {
		if item.SourceID == sourceID {
			out = append(out, item)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CollectedAt.After(out[j].CollectedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
func (r *runtimeSourceItemRepository) ListRecent(_ context.Context, limit int) ([]domain.SourceItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]domain.SourceItem, 0, len(r.items))
	for _, item := range r.items {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CollectedAt.Equal(out[j].CollectedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].CollectedAt.After(out[j].CollectedAt)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

type runtimePublishIntentRepository struct {
	mu      sync.RWMutex
	intents map[int64]domain.PublishIntent
	nextID  int64
}

func newRuntimePublishIntentRepository() *runtimePublishIntentRepository {
	return &runtimePublishIntentRepository{intents: map[int64]domain.PublishIntent{}, nextID: 1}
}
func (r *runtimePublishIntentRepository) Create(_ context.Context, intent domain.PublishIntent) (domain.PublishIntent, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.intents {
		if existing.RawItemID == intent.RawItemID && existing.ChannelID == intent.ChannelID {
			return existing, nil
		}
	}
	intent.ID = r.nextID
	r.nextID++
	intent.CreatedAt = time.Now().UTC()
	r.intents[intent.ID] = intent
	return intent, nil
}
func (r *runtimePublishIntentRepository) ListByRawItemID(_ context.Context, rawItemID int64, limit int) ([]domain.PublishIntent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.PublishIntent{}
	for _, intent := range r.intents {
		if intent.RawItemID == rawItemID {
			out = append(out, intent)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
func (r *runtimePublishIntentRepository) UpdateStatus(_ context.Context, id int64, status domain.PublishIntentStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	intent, ok := r.intents[id]
	if !ok {
		return domain.ErrNotFound
	}
	intent.Status = status
	r.intents[id] = intent
	return nil
}

type runtimeContentAssetRepository struct {
	mu     sync.RWMutex
	assets map[int64]domain.ContentAsset
	nextID int64
}

func newRuntimeContentAssetRepository() *runtimeContentAssetRepository {
	return &runtimeContentAssetRepository{assets: map[int64]domain.ContentAsset{}, nextID: 1}
}
func (r *runtimeContentAssetRepository) Create(_ context.Context, asset domain.ContentAsset) (domain.ContentAsset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	asset.ID = r.nextID
	r.nextID++
	asset.CreatedAt = time.Now().UTC()
	asset.UpdatedAt = asset.CreatedAt
	r.assets[asset.ID] = asset
	return asset, nil
}
func (r *runtimeContentAssetRepository) GetByID(_ context.Context, id int64) (domain.ContentAsset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	asset, ok := r.assets[id]
	if !ok {
		return domain.ContentAsset{}, domain.ErrNotFound
	}
	return asset, nil
}
func (r *runtimeContentAssetRepository) ListByRawItemID(_ context.Context, rawItemID int64, limit int) ([]domain.ContentAsset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := []domain.ContentAsset{}
	for _, asset := range r.assets {
		if asset.RawItemID == rawItemID {
			out = append(out, asset)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
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
