package app

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ai-content-engine-starter/internal/domain"
	"ai-content-engine-starter/internal/platform/config"
	"ai-content-engine-starter/internal/platform/postgres"
	"ai-content-engine-starter/internal/scheduler"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	healthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q, want %q", got, "application/json")
	}
	if got := rr.Body.String(); got != `{"status":"ok"}` {
		t.Fatalf("body = %q, want %q", got, `{"status":"ok"}`)
	}
}

func TestHealthHandlerMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()

	healthHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestRoutesServeWebUIRoot(t *testing.T) {
	a := &App{cfg: config.Config{Features: config.FeatureFlags{WebUI: true}}}
	h := a.routes()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestRoutesDoNotServeWebUIWhenDisabled(t *testing.T) {
	a := &App{}
	h := a.routes()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusNotFound)
	}
}

func TestRoutesHealthMethodNotAllowed(t *testing.T) {
	a := &App{}
	h := a.routes()

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
}

func TestRoutesAdminDraftsIsRegistered(t *testing.T) {
	repo, err := newAdminFileDraftRepository(filepath.Join(t.TempDir(), "drafts.json"))
	if err != nil {
		t.Fatalf("newAdminFileDraftRepository() error = %v", err)
	}
	a := &App{drafts: repo}
	h := a.routes()

	req := httptest.NewRequest(http.MethodGet, "/admin/drafts", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestAdminFileDraftRepositoryPersistsAcrossReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "drafts.json")
	repo, err := newAdminFileDraftRepository(path)
	if err != nil {
		t.Fatalf("newAdminFileDraftRepository() error = %v", err)
	}
	draft, err := repo.Create(context.Background(), domain.Draft{ID: 5, Title: "t", Body: "b", Status: domain.DraftStatusPending})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := repo.UpdateStatus(context.Background(), draft.ID, domain.DraftStatusApproved); err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}

	reloaded, err := newAdminFileDraftRepository(path)
	if err != nil {
		t.Fatalf("reload repository error = %v", err)
	}
	got, err := reloaded.GetByID(context.Background(), draft.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Status != domain.DraftStatusApproved {
		t.Fatalf("status = %q, want %q", got.Status, domain.DraftStatusApproved)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("draft persistence file stat error = %v", err)
	}
}

func TestRuntimeHealthHandler(t *testing.T) {
	a := &App{cfg: config.Config{EnablePipeline: true, EnablePublisher: true}, stats: &runtimeStats{}, statusCheck: func(context.Context) error { return nil }}
	a.stats.pipelineRuns.Store(2)
	a.stats.publishedDrafts.Store(3)
	req := httptest.NewRequest(http.MethodGet, "/health/runtime", nil)
	rr := httptest.NewRecorder()

	a.runtimeHealthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q, want %q", got, "application/json")
	}
}

func TestPublisherLoopPublishesApprovedDrafts(t *testing.T) {
	drafts, err := newAdminFileDraftRepository(filepath.Join(t.TempDir(), "drafts.json"))
	if err != nil {
		t.Fatalf("newAdminFileDraftRepository() error = %v", err)
	}
	_, err = drafts.Create(context.Background(), domain.Draft{ID: 1, SourceItemID: 10, ChannelID: 1, Title: "t", Body: "b", Status: domain.DraftStatusApproved})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	channels := newRuntimeChannelRepository()
	if _, err := channels.Create(context.Background(), domain.Channel{Slug: "ai-news", Name: "AI News"}); err != nil {
		t.Fatalf("Create channel error = %v", err)
	}
	published := 0
	loop := &publisherLoop{
		logger:       loggerForTest(),
		drafts:       drafts,
		channels:     channels,
		publisher:    publisherClientFunc(func(context.Context, domain.Draft, string) (int64, error) { published++; return 42, nil }),
		chatIDBySlug: map[string]string{"ai-news": "@ai_news"},
		stats:        &runtimeStats{},
		limit:        10,
	}
	if err := loop.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
	if published != 1 {
		t.Fatalf("published = %d, want 1", published)
	}
	got, err := drafts.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.Status != domain.DraftStatusPosted {
		t.Fatalf("status = %q, want %q", got.Status, domain.DraftStatusPosted)
	}
}

func TestPublisherLoopSkipsDraftThatCannotBeClaimed(t *testing.T) {
	loop := &publisherLoop{
		logger:   loggerForTest(),
		drafts:   &claimingDraftRepoStub{drafts: []domain.Draft{{ID: 1, SourceItemID: 10, ChannelID: 1, Status: domain.DraftStatusApproved}}},
		channels: &channelRepoStub{channels: []domain.Channel{{ID: 1, Slug: "ai-news", Name: "AI News"}}},
		publisher: publisherClientFunc(func(context.Context, domain.Draft, string) (int64, error) {
			t.Fatalf("publisher should not be called when claim fails")
			return 0, nil
		}),
		chatIDBySlug: map[string]string{"ai-news": "@ai_news"},
		stats:        &runtimeStats{},
		limit:        10,
	}
	if err := loop.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}
}

func TestSuperviseLoopRestartsAfterFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	attempts := 0
	sched, err := scheduler.New(5*time.Millisecond, func(context.Context) error {
		attempts++
		if attempts == 1 {
			return errors.New("boom")
		}
		cancel()
		return nil
	})
	if err != nil {
		t.Fatalf("scheduler.New() error = %v", err)
	}
	a := &App{logger: loggerForTest(), stats: &runtimeStats{}}
	done := make(chan struct{})
	go func() {
		defer close(done)
		a.superviseLoop(ctx, "collector", sched)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("superviseLoop did not exit")
	}
	if attempts < 2 {
		t.Fatalf("attempts = %d, want at least 2", attempts)
	}
	if got := a.stats.loopRestarts.Load(); got != 1 {
		t.Fatalf("loopRestarts = %d, want 1", got)
	}
}

func TestInitRuntimeUsesPostgresRepositories(t *testing.T) {
	now := time.Now().UTC()
	db := openStubDB(t, &stubDBHandler{
		pingFunc: func(context.Context) error { return nil },
		queryFunc: func(query string, args []driver.NamedValue) (driver.Rows, error) {
			switch {
			case query == "SELECT id, slug, name, created_at FROM channels ORDER BY id":
				return &stubRows{columns: []string{"id", "slug", "name", "created_at"}}, nil
			case query == "SELECT id, kind, name, endpoint, enabled, created_at FROM sources ORDER BY id":
				return &stubRows{columns: []string{"id", "kind", "name", "endpoint", "enabled", "created_at"}}, nil
			case len(args) == 2 && contains(query, "INSERT INTO channels"):
				return &stubRows{columns: []string{"id", "slug", "name", "created_at"}, values: [][]driver.Value{{int64(1), args[0].Value, args[1].Value, now}}}, nil
			case len(args) == 4 && contains(query, "INSERT INTO sources"):
				return &stubRows{columns: []string{"id", "kind", "name", "endpoint", "enabled", "created_at"}, values: [][]driver.Value{{int64(1), args[0].Value, args[1].Value, args[2].Value, args[3].Value, now}}}, nil
			default:
				return nil, fmt.Errorf("unexpected query: %s", query)
			}
		},
	})
	defer db.Close()

	a := &App{
		cfg: config.Config{
			PostgresDSN:        "postgres://user:pass@localhost/testdb?sslmode=disable",
			RedisAddr:          "127.0.0.1:6379",
			EnablePublisher:    true,
			LoopInterval:       time.Second,
			TelegramBotToken:   "token",
			PublisherBatchSize: 10,
			RecentItemsLimit:   10,
			DraftScanLimit:     20,
		},
		logger:      loggerForTest(),
		stats:       &runtimeStats{},
		openDB:      func(string, string) (*sql.DB, error) { return db, nil },
		pingRedisFn: func(context.Context, string) error { return nil },
		drafts:      adminFallbackDraftRepository{},
	}
	if err := a.initRuntime(context.Background()); err != nil {
		t.Fatalf("initRuntime() error = %v", err)
	}
	if _, ok := a.drafts.(*postgres.DraftRepository); !ok {
		t.Fatalf("draft repository type = %T, want *postgres.DraftRepository", a.drafts)
	}
	if a.runtime == nil || a.runtime.publisher == nil {
		t.Fatalf("publisher runtime was not initialized")
	}
}

func TestPostgresDriverIsRegistered(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://user:pass@127.0.0.1:1/testdb?sslmode=disable&connect_timeout=1")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err == nil {
		t.Fatal("expected ping failure against invalid test address")
	} else if strings.Contains(err.Error(), "unknown driver") {
		t.Fatalf("expected registered postgres driver, got %v", err)
	}
}

type claimingDraftRepoStub struct {
	drafts   []domain.Draft
	claimOK  bool
	statuses map[int64]domain.DraftStatus
}

func (s *claimingDraftRepoStub) Create(context.Context, domain.Draft) (domain.Draft, error) {
	return domain.Draft{}, nil
}
func (s *claimingDraftRepoStub) GetByID(context.Context, int64) (domain.Draft, error) {
	return domain.Draft{}, nil
}
func (s *claimingDraftRepoStub) ListByStatus(context.Context, domain.DraftStatus, int) ([]domain.Draft, error) {
	return s.drafts, nil
}
func (s *claimingDraftRepoStub) UpdateStatus(context.Context, int64, domain.DraftStatus) error {
	return nil
}
func (s *claimingDraftRepoStub) UpdateStatusIfCurrent(context.Context, int64, domain.DraftStatus, domain.DraftStatus) (bool, error) {
	return s.claimOK, nil
}

type channelRepoStub struct{ channels []domain.Channel }

func (s *channelRepoStub) Create(context.Context, domain.Channel) (domain.Channel, error) {
	return domain.Channel{}, nil
}
func (s *channelRepoStub) GetByID(context.Context, int64) (domain.Channel, error) {
	return domain.Channel{}, nil
}
func (s *channelRepoStub) List(context.Context) ([]domain.Channel, error) { return s.channels, nil }

func seedSourcesForTest() []domain.Source {
	return []domain.Source{
		{Kind: "rss", Name: "AI News RSS", Endpoint: "https://example.com/ai-news.rss", Enabled: false},
		{Kind: "github", Name: "GitHub AI", Endpoint: "https://api.github.com", Enabled: false},
		{Kind: "reddit", Name: "Reddit AI", Endpoint: "https://www.reddit.com/r/artificial/.json", Enabled: false},
		{Kind: "producthunt", Name: "Product Hunt", Endpoint: "https://api.producthunt.com/v2/api/graphql", Enabled: false},
	}
}

type stubDBHandler struct {
	pingFunc  func(context.Context) error
	queryFunc func(query string, args []driver.NamedValue) (driver.Rows, error)
	execFunc  func(query string, args []driver.NamedValue) (driver.Result, error)
}

type stubDriver struct{ handler *stubDBHandler }

func (d *stubDriver) Open(string) (driver.Conn, error) {
	return &stubConn{handler: d.handler}, nil
}

type stubConn struct{ handler *stubDBHandler }

func (c *stubConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not implemented") }
func (c *stubConn) Close() error                        { return nil }
func (c *stubConn) Begin() (driver.Tx, error)           { return nil, errors.New("not implemented") }
func (c *stubConn) Ping(ctx context.Context) error {
	if c.handler != nil && c.handler.pingFunc != nil {
		return c.handler.pingFunc(ctx)
	}
	return nil
}
func (c *stubConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if c.handler == nil || c.handler.queryFunc == nil {
		return nil, fmt.Errorf("unexpected query: %s", query)
	}
	return c.handler.queryFunc(query, args)
}
func (c *stubConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if c.handler == nil || c.handler.execFunc == nil {
		return nil, fmt.Errorf("unexpected exec: %s", query)
	}
	return c.handler.execFunc(query, args)
}

type stubRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *stubRows) Columns() []string { return r.columns }
func (r *stubRows) Close() error      { return nil }
func (r *stubRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}

func openStubDB(t *testing.T, handler *stubDBHandler) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("stubdb-%d", time.Now().UnixNano())
	sql.Register(name, &stubDriver{handler: handler})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	return db
}

func contains(value string, needle string) bool {
	return strings.Contains(value, needle)
}

type publisherClientFunc func(context.Context, domain.Draft, string) (int64, error)

func (f publisherClientFunc) PublishDraft(ctx context.Context, draft domain.Draft, chatID string) (int64, error) {
	return f(ctx, draft, chatID)
}

func loggerForTest() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
