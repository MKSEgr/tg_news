package editorialplanner

import (
	"context"
	"fmt"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

const (
	defaultIntentFormat = "text"
)

// RawItem represents a minimal raw content unit for editorial planning.
type RawItem struct {
	ID    int64
	URL   string
	Title string
	Body  *string
}

// PublishIntent is a planned publishing intent created from raw content.
type PublishIntent = domain.PublishIntent

// EditorialPlanner plans publish intents for incoming raw items.
type EditorialPlanner interface {
	PlanForItem(ctx context.Context, item RawItem) ([]PublishIntent, error)
}

type trendScorer interface {
	Score(item domain.SourceItem) int
}

type channelRouter interface {
	Route(item domain.SourceItem, channels []domain.Channel) ([]int64, error)
}

// Service is a deterministic editorial planner implementation.
type Service struct {
	repo     domain.PublishIntentRepository
	channels domain.ChannelRepository
	scorer   trendScorer
	router   channelRouter
}

// New creates editorial planner service.
func New(repo domain.PublishIntentRepository, channels domain.ChannelRepository, scorer trendScorer, router channelRouter) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("publish intent repository is nil")
	}
	if channels == nil {
		return nil, fmt.Errorf("channel repository is nil")
	}
	if scorer == nil {
		return nil, fmt.Errorf("scorer is nil")
	}
	if router == nil {
		return nil, fmt.Errorf("router is nil")
	}
	return &Service{repo: repo, channels: channels, scorer: scorer, router: router}, nil
}

// PlanForItem produces exactly one intent per raw item when score and routing allow it.
func (s *Service) PlanForItem(ctx context.Context, item RawItem) ([]PublishIntent, error) {
	if s == nil {
		return nil, fmt.Errorf("editorial planner service is nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if item.ID <= 0 {
		return nil, fmt.Errorf("raw item id is invalid")
	}

	existing, err := s.repo.ListByRawItemID(ctx, item.ID, 1)
	if err != nil {
		return nil, fmt.Errorf("list publish intents by raw item id: %w", err)
	}
	if len(existing) > 0 {
		return []PublishIntent{}, nil
	}

	normalized := domain.SourceItem{
		ID:    item.ID,
		URL:   strings.TrimSpace(item.URL),
		Title: strings.TrimSpace(item.Title),
		Body:  item.Body,
	}

	priority := s.scorer.Score(normalized)
	if priority <= 0 {
		return []PublishIntent{}, nil
	}

	channels, err := s.channels.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	routes, err := s.router.Route(normalized, channels)
	if err != nil {
		return nil, fmt.Errorf("route raw item: %w", err)
	}

	channelID, ok := firstValidChannelID(routes)
	if !ok {
		return []PublishIntent{}, nil
	}

	intent := domain.PublishIntent{
		RawItemID: item.ID,
		ChannelID: channelID,
		Format:    defaultIntentFormat,
		Priority:  priority,
		Status:    domain.PublishIntentStatusPlanned,
	}
	created, err := s.repo.Create(ctx, intent)
	if err != nil {
		return nil, fmt.Errorf("create publish intent: %w", err)
	}

	// TODO(v3): support story clusters when V3-002 is implemented.
	// TODO(v3): support multi-asset planning after asset generation modules land.
	// TODO(v3): support advanced ranking beyond scorer-derived priority.
	return []PublishIntent{created}, nil
}

// PlanForSourceItem adapts existing domain SourceItem to planner RawItem.
func (s *Service) PlanForSourceItem(ctx context.Context, item domain.SourceItem) ([]PublishIntent, error) {
	return s.PlanForItem(ctx, RawItem{ID: item.ID, URL: item.URL, Title: item.Title, Body: item.Body})
}

func firstValidChannelID(ids []int64) (int64, bool) {
	for _, id := range ids {
		if id > 0 {
			return id, true
		}
	}
	return 0, false
}
