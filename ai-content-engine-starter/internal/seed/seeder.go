package seed

import (
	"context"
	"fmt"

	"ai-content-engine-starter/internal/domain"
)

// DefaultChannels are created when no channels exist yet.
var DefaultChannels = []domain.Channel{
	{Slug: "ai-news", Name: "AI News"},
	{Slug: "ai-tools", Name: "AI Tools"},
	{Slug: "ai-workflows", Name: "AI Workflows"},
}

// DefaultSources are created when no enabled sources exist yet.
var DefaultSources = []domain.Source{
	{Kind: "rss", Name: "AI News RSS", Endpoint: "https://example.com/ai-news.rss", Enabled: true},
	{Kind: "github", Name: "GitHub AI", Endpoint: "https://api.github.com", Enabled: true},
	{Kind: "reddit", Name: "Reddit AI", Endpoint: "https://www.reddit.com/r/artificial/.json", Enabled: true},
	{Kind: "producthunt", Name: "Product Hunt", Endpoint: "https://api.producthunt.com/v2/api/graphql", Enabled: true},
}

// Seeder bootstraps initial channels and sources for a fresh database.
type Seeder struct {
	channels domain.ChannelRepository
	sources  domain.SourceRepository
}

// New creates a new Seeder.
func New(channels domain.ChannelRepository, sources domain.SourceRepository) *Seeder {
	return &Seeder{channels: channels, sources: sources}
}

// Seed creates default channels and sources only when corresponding tables are empty.
func (s *Seeder) Seed(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("seeder is nil")
	}
	if s.channels == nil {
		return fmt.Errorf("channel repository is nil")
	}
	if s.sources == nil {
		return fmt.Errorf("source repository is nil")
	}

	channels, err := s.channels.List(ctx)
	if err != nil {
		return fmt.Errorf("list channels: %w", err)
	}
	if len(channels) == 0 {
		for _, channel := range DefaultChannels {
			if _, err := s.channels.Create(ctx, channel); err != nil {
				return fmt.Errorf("create channel %s: %w", channel.Slug, err)
			}
		}
	}

	sources, err := s.sources.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("list enabled sources: %w", err)
	}
	if len(sources) == 0 {
		for _, source := range DefaultSources {
			if _, err := s.sources.Create(ctx, source); err != nil {
				return fmt.Errorf("create source %s: %w", source.Name, err)
			}
		}
	}

	return nil
}
