package collector

import (
	"context"
	"fmt"

	"ai-content-engine-starter/internal/domain"
)

// Collector fetches source items for a specific source kind.
type Collector interface {
	Kind() string
	Collect(ctx context.Context, source domain.Source) ([]domain.SourceItem, error)
}

// Framework orchestrates source collection using registered collectors.
type Framework struct {
	collectors map[string]Collector
	sources    domain.SourceRepository
	items      domain.SourceItemRepository
}

// New creates a collector framework instance.
func New(sources domain.SourceRepository, items domain.SourceItemRepository, collectors ...Collector) (*Framework, error) {
	if sources == nil {
		return nil, fmt.Errorf("source repository is nil")
	}
	if items == nil {
		return nil, fmt.Errorf("source item repository is nil")
	}

	registry := make(map[string]Collector, len(collectors))
	for _, collector := range collectors {
		if collector == nil {
			return nil, fmt.Errorf("collector is nil")
		}
		kind := collector.Kind()
		if kind == "" {
			return nil, fmt.Errorf("collector kind is empty")
		}
		if _, exists := registry[kind]; exists {
			return nil, fmt.Errorf("duplicate collector kind: %s", kind)
		}
		registry[kind] = collector
	}

	return &Framework{collectors: registry, sources: sources, items: items}, nil
}

// RunOnce collects enabled sources and stores collected items.
func (f *Framework) RunOnce(ctx context.Context) error {
	if f == nil {
		return fmt.Errorf("framework is nil")
	}
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}

	sources, err := f.sources.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("list enabled sources: %w", err)
	}

	for _, source := range sources {
		collector, ok := f.collectors[source.Kind]
		if !ok {
			return fmt.Errorf("collector not found for source kind: %s", source.Kind)
		}

		items, err := collector.Collect(ctx, source)
		if err != nil {
			return fmt.Errorf("collect source %d (%s): %w", source.ID, source.Kind, err)
		}

		for _, item := range items {
			item.SourceID = source.ID
			if _, err := f.items.Create(ctx, item); err != nil {
				return fmt.Errorf("store item for source %d: %w", source.ID, err)
			}
		}
	}

	return nil
}
