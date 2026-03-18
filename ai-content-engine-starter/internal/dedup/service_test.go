package dedup

import (
	"context"
	"errors"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

type sourceItemRepoStub struct {
	items      []domain.SourceItem
	err        error
	lastSource int64
	lastLimit  int
	listRecent bool
}

func (s *sourceItemRepoStub) Create(context.Context, domain.SourceItem) (domain.SourceItem, error) {
	return domain.SourceItem{}, nil
}

func (s *sourceItemRepoStub) GetByID(context.Context, int64) (domain.SourceItem, error) {
	return domain.SourceItem{}, nil
}

func (s *sourceItemRepoStub) ListBySourceID(_ context.Context, sourceID int64, limit int) ([]domain.SourceItem, error) {
	s.lastSource = sourceID
	s.lastLimit = limit
	if s.err != nil {
		return nil, s.err
	}
	return s.items, nil
}

func (s *sourceItemRepoStub) ListRecent(_ context.Context, limit int) ([]domain.SourceItem, error) {
	s.listRecent = true
	s.lastLimit = limit
	if s.err != nil {
		return nil, s.err
	}
	return s.items, nil
}

func TestNewRejectsNilRepo(t *testing.T) {
	if _, err := New(nil, 10); err == nil {
		t.Fatalf("expected error for nil repo")
	}
}

func TestIsDuplicateByExternalID(t *testing.T) {
	repo := &sourceItemRepoStub{items: []domain.SourceItem{{SourceID: 1, ExternalID: "ext-1"}}}
	svc, err := New(repo, 50)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	dup, err := svc.IsDuplicate(context.Background(), domain.SourceItem{SourceID: 1, ExternalID: "ext-1"})
	if err != nil {
		t.Fatalf("IsDuplicate() error = %v", err)
	}
	if !dup {
		t.Fatalf("expected duplicate")
	}
	if repo.lastLimit != 50 {
		t.Fatalf("limit = %d, want 50", repo.lastLimit)
	}
	if !repo.listRecent {
		t.Fatalf("expected ListRecent to be used")
	}
}

func TestIsDuplicateByURLOrTitle(t *testing.T) {
	repo := &sourceItemRepoStub{items: []domain.SourceItem{{SourceID: 1, URL: "https://example.com/a", Title: "Same"}}}
	svc, err := New(repo, 0)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	dupByURL, err := svc.IsDuplicate(context.Background(), domain.SourceItem{SourceID: 1, URL: "https://example.com/a"})
	if err != nil {
		t.Fatalf("IsDuplicate() error = %v", err)
	}
	if !dupByURL {
		t.Fatalf("expected duplicate by URL")
	}

	dupByTitle, err := svc.IsDuplicate(context.Background(), domain.SourceItem{SourceID: 1, Title: "Same"})
	if err != nil {
		t.Fatalf("IsDuplicate() error = %v", err)
	}
	if !dupByTitle {
		t.Fatalf("expected duplicate by title")
	}
	if repo.lastLimit != defaultRecentLimit {
		t.Fatalf("limit = %d, want default %d", repo.lastLimit, defaultRecentLimit)
	}
}

func TestIsDuplicateAcrossDifferentSources(t *testing.T) {
	repo := &sourceItemRepoStub{items: []domain.SourceItem{{SourceID: 2, URL: "https://example.com/shared"}}}
	svc, err := New(repo, 10)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	dup, err := svc.IsDuplicate(context.Background(), domain.SourceItem{SourceID: 1, URL: "https://example.com/shared"})
	if err != nil {
		t.Fatalf("IsDuplicate() error = %v", err)
	}
	if !dup {
		t.Fatalf("expected cross-source duplicate")
	}
}

func TestIsDuplicateFalseWhenNoMatches(t *testing.T) {
	repo := &sourceItemRepoStub{items: []domain.SourceItem{{SourceID: 1, ExternalID: "ext-1"}}}
	svc, err := New(repo, 10)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	dup, err := svc.IsDuplicate(context.Background(), domain.SourceItem{SourceID: 1, ExternalID: "ext-2", URL: "https://example.com/new"})
	if err != nil {
		t.Fatalf("IsDuplicate() error = %v", err)
	}
	if dup {
		t.Fatalf("expected non-duplicate")
	}
}

func TestIsDuplicateValidationAndErrors(t *testing.T) {
	repo := &sourceItemRepoStub{}
	svc, err := New(repo, 10)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if _, err := svc.IsDuplicate(nil, domain.SourceItem{SourceID: 1, ExternalID: "id"}); err == nil {
		t.Fatalf("expected error for nil context")
	}
	if _, err := svc.IsDuplicate(context.Background(), domain.SourceItem{ExternalID: "id"}); err == nil {
		t.Fatalf("expected error for invalid source id")
	}
	if _, err := svc.IsDuplicate(context.Background(), domain.SourceItem{SourceID: 1}); err == nil {
		t.Fatalf("expected error for missing dedup keys")
	}

	repo.err = errors.New("db down")
	if _, err := svc.IsDuplicate(context.Background(), domain.SourceItem{SourceID: 1, ExternalID: "id"}); err == nil {
		t.Fatalf("expected repository error")
	}
}

func TestIsDuplicateRejectsNilServiceOrRepo(t *testing.T) {
	var nilService *Service
	if _, err := nilService.IsDuplicate(context.Background(), domain.SourceItem{SourceID: 1, ExternalID: "id"}); err == nil {
		t.Fatalf("expected error for nil service")
	}

	svc := &Service{limit: 10}
	if _, err := svc.IsDuplicate(context.Background(), domain.SourceItem{SourceID: 1, ExternalID: "id"}); err == nil {
		t.Fatalf("expected error for nil repository")
	}
}

func TestIsDuplicateIgnoresSameItemID(t *testing.T) {
	repo := &sourceItemRepoStub{items: []domain.SourceItem{{ID: 10, SourceID: 1, ExternalID: "ext-1"}}}
	svc, err := New(repo, 10)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	dup, err := svc.IsDuplicate(context.Background(), domain.SourceItem{ID: 10, SourceID: 1, ExternalID: "ext-1"})
	if err != nil {
		t.Fatalf("IsDuplicate() error = %v", err)
	}
	if dup {
		t.Fatalf("expected non-duplicate when matching item has the same ID")
	}
}
