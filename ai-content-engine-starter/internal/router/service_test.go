package router

import (
	"testing"

	"ai-content-engine-starter/internal/domain"
)

func testChannels() []domain.Channel {
	return []domain.Channel{
		{ID: 1, Slug: "ai-news", Name: "AI News"},
		{ID: 2, Slug: "ai-tools", Name: "AI Tools"},
		{ID: 3, Slug: "ai-workflows", Name: "AI Workflows"},
	}
}

func TestRouteMatchesMultipleChannels(t *testing.T) {
	svc := New()
	body := "Open source automation workflow guide"

	ids, err := svc.Route(domain.SourceItem{Title: "Tool launch", Body: &body}, testChannels())
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}
	if len(ids) < 2 {
		t.Fatalf("ids len = %d, want >= 2", len(ids))
	}
}

func TestRouteFallsBackToNews(t *testing.T) {
	svc := New()

	ids, err := svc.Route(domain.SourceItem{Title: "Weekly update"}, testChannels())
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}
	if len(ids) != 1 || ids[0] != 1 {
		t.Fatalf("ids = %v, want [1]", ids)
	}
}

func TestRouteValidation(t *testing.T) {
	var nilService *Service
	if _, err := nilService.Route(domain.SourceItem{}, testChannels()); err == nil {
		t.Fatalf("expected error for nil service")
	}

	svc := New()
	if _, err := svc.Route(domain.SourceItem{}, nil); err == nil {
		t.Fatalf("expected error for empty channels")
	}
	if _, err := svc.Route(domain.SourceItem{}, []domain.Channel{{Slug: "ai-news"}}); err == nil {
		t.Fatalf("expected error for unroutable channels")
	}
}

func TestRouteFallbackWithoutNewsIsDeterministic(t *testing.T) {
	svc := New()

	channels := []domain.Channel{
		{ID: 20, Slug: "ai-tools", Name: "AI Tools"},
		{ID: 30, Slug: "ai-workflows", Name: "AI Workflows"},
	}

	ids, err := svc.Route(domain.SourceItem{Title: "Weekly update"}, channels)
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}
	if len(ids) != 1 || ids[0] != 20 {
		t.Fatalf("ids = %v, want [20]", ids)
	}
}

func TestRouteWithMemoryAddsChannelMatch(t *testing.T) {
	svc := New()
	channels := []domain.Channel{
		{ID: 1, Slug: "ai-news", Name: "AI News"},
		{ID: 2, Slug: "ai-tools", Name: "AI Tools"},
	}

	ids, err := svc.RouteWithMemory(
		domain.SourceItem{Title: "Case study on ai inference optimization"},
		channels,
		map[int64][]domain.TopicMemory{2: {{Topic: "inference", MentionCount: 8}}},
	)
	if err != nil {
		t.Fatalf("RouteWithMemory() error = %v", err)
	}
	if len(ids) != 1 || ids[0] != 2 {
		t.Fatalf("ids = %v, want [2]", ids)
	}
}
