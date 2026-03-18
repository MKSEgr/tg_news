package app

import (
	"context"
	"net/http"
	"os"
	"path/filepath"

	"ai-content-engine-starter/internal/domain"
	"ai-content-engine-starter/internal/platform/config"
	"net/http/httptest"
	"testing"
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
