package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

type draftRepoStub struct {
	listResult        []domain.Draft
	listErr           error
	updateErr         error
	lastListStatus    domain.DraftStatus
	lastListLimit     int
	lastUpdatedID     int64
	lastUpdatedStatus domain.DraftStatus
}

func (s *draftRepoStub) Create(context.Context, domain.Draft) (domain.Draft, error) {
	return domain.Draft{}, nil
}

func (s *draftRepoStub) GetByID(context.Context, int64) (domain.Draft, error) {
	return domain.Draft{}, nil
}

func (s *draftRepoStub) ListByStatus(_ context.Context, status domain.DraftStatus, limit int) ([]domain.Draft, error) {
	s.lastListStatus = status
	s.lastListLimit = limit
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.listResult, nil
}

func (s *draftRepoStub) UpdateStatus(_ context.Context, id int64, status domain.DraftStatus) error {
	s.lastUpdatedID = id
	s.lastUpdatedStatus = status
	return s.updateErr
}

func TestNewHandlerValidation(t *testing.T) {
	if _, err := NewHandler(nil); err == nil {
		t.Fatalf("expected error for nil repository")
	}
}

func TestRegisterValidation(t *testing.T) {
	h, err := NewHandler(&draftRepoStub{})
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	if err := h.Register(nil); err == nil {
		t.Fatalf("expected error for nil mux")
	}
}

func TestListDrafts(t *testing.T) {
	repo := &draftRepoStub{listResult: []domain.Draft{{ID: 1, Status: domain.DraftStatusPending}}}
	h, _ := NewHandler(repo)
	mux := http.NewServeMux()
	if err := h.Register(mux); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/drafts?status=approved&limit=10", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if repo.lastListStatus != domain.DraftStatusApproved {
		t.Fatalf("status = %s, want approved", repo.lastListStatus)
	}
	if repo.lastListLimit != 10 {
		t.Fatalf("limit = %d, want 10", repo.lastListLimit)
	}
}

func TestListDraftsValidation(t *testing.T) {
	repo := &draftRepoStub{}
	h, _ := NewHandler(repo)
	mux := http.NewServeMux()
	_ = h.Register(mux)

	cases := []string{
		"/admin/drafts?status=bad",
		"/admin/drafts?limit=0",
		"/admin/drafts?limit=abc",
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc, nil)
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("url %s status = %d, want 400", tc, rr.Code)
		}
	}
}

func TestDraftActionApproveAndReject(t *testing.T) {
	repo := &draftRepoStub{}
	h, _ := NewHandler(repo)
	mux := http.NewServeMux()
	_ = h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/drafts/5/approve", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if repo.lastUpdatedID != 5 || repo.lastUpdatedStatus != domain.DraftStatusApproved {
		t.Fatalf("unexpected update args: id=%d status=%s", repo.lastUpdatedID, repo.lastUpdatedStatus)
	}

	req = httptest.NewRequest(http.MethodPost, "/admin/drafts/6/reject", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if repo.lastUpdatedID != 6 || repo.lastUpdatedStatus != domain.DraftStatusRejected {
		t.Fatalf("unexpected update args: id=%d status=%s", repo.lastUpdatedID, repo.lastUpdatedStatus)
	}
}

func TestDraftActionErrors(t *testing.T) {
	repo := &draftRepoStub{}
	h, _ := NewHandler(repo)
	mux := http.NewServeMux()
	_ = h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/drafts/0/approve", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/admin/drafts/7/publish", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}

	repo.updateErr = domain.ErrNotFound
	req = httptest.NewRequest(http.MethodPost, "/admin/drafts/7/approve", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}

	repo.updateErr = errors.New("boom")
	req = httptest.NewRequest(http.MethodPost, "/admin/drafts/7/approve", nil)
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}
}

func TestDraftActionWrappedNotFound(t *testing.T) {
	repo := &draftRepoStub{updateErr: fmt.Errorf("wrap: %w", domain.ErrNotFound)}
	h, _ := NewHandler(repo)
	mux := http.NewServeMux()
	_ = h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/admin/drafts/8/approve", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rr.Code)
	}
}

func TestListDraftsRepoError(t *testing.T) {
	repo := &draftRepoStub{listErr: errors.New("db")}
	h, _ := NewHandler(repo)
	mux := http.NewServeMux()
	_ = h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/admin/drafts", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "failed to list drafts") {
		t.Fatalf("unexpected body: %s", rr.Body.String())
	}
}
