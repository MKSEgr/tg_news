package sourcediscovery

import (
	"context"
	"errors"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

type sourceRepoStub struct {
	sources []domain.Source
	err     error
}

func (s *sourceRepoStub) Create(context.Context, domain.Source) (domain.Source, error) {
	return domain.Source{}, nil
}
func (s *sourceRepoStub) GetByID(context.Context, int64) (domain.Source, error) {
	return domain.Source{}, nil
}
func (s *sourceRepoStub) ListEnabled(context.Context) ([]domain.Source, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.sources, nil
}

func TestNewValidation(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatalf("expected nil repository error")
	}
}

func TestDiscoverValidation(t *testing.T) {
	svc, err := New(&sourceRepoStub{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := svc.Discover(nil, nil); err == nil {
		t.Fatalf("expected nil context error")
	}

	var nilSvc *Service
	if _, err := nilSvc.Discover(context.Background(), nil); err == nil {
		t.Fatalf("expected nil service error")
	}
}

func TestDiscoverReturnsDeterministicCandidates(t *testing.T) {
	svc, err := New(&sourceRepoStub{sources: []domain.Source{{Endpoint: "https://known.example/feed", Enabled: true}}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	items := []domain.SourceItem{
		{URL: "https://known.example/news/1"},
		{URL: "https://zeta.example/articles/1"},
		{URL: "https://alpha.example/rss.xml?from=utm"},
		{URL: "https://alpha.example/news/2"},
		{URL: "not-a-url"},
	}

	got, err := svc.Discover(context.Background(), items)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("candidates = %d, want 2", len(got))
	}
	if got[0].Endpoint != "https://alpha.example/rss.xml" {
		t.Fatalf("first endpoint = %q", got[0].Endpoint)
	}
	if got[1].Endpoint != "https://zeta.example/feed" {
		t.Fatalf("second endpoint = %q", got[1].Endpoint)
	}
	for _, source := range got {
		if source.Kind != "rss" {
			t.Fatalf("kind = %q, want rss", source.Kind)
		}
		if source.Enabled {
			t.Fatalf("enabled = true, want false for discovered candidates")
		}
	}
}

func TestDiscoverListEnabledError(t *testing.T) {
	svc, err := New(&sourceRepoStub{err: errors.New("boom")})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := svc.Discover(context.Background(), []domain.SourceItem{{URL: "https://a.example/news"}}); err == nil {
		t.Fatalf("expected error")
	}
}
