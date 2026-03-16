package dedup

import (
	"context"
	"fmt"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

const defaultRecentLimit = 200

// Service checks source items for duplicates against recently collected items.
type Service struct {
	repo  domain.SourceItemRepository
	limit int
}

// New creates a deduplication service.
func New(repo domain.SourceItemRepository, recentLimit int) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("source item repository is nil")
	}
	if recentLimit <= 0 {
		recentLimit = defaultRecentLimit
	}
	return &Service{repo: repo, limit: recentLimit}, nil
}

// IsDuplicate reports whether the given item already exists among recent source items.
func (s *Service) IsDuplicate(ctx context.Context, item domain.SourceItem) (bool, error) {
	if s == nil {
		return false, fmt.Errorf("dedup service is nil")
	}
	if s.repo == nil {
		return false, fmt.Errorf("source item repository is nil")
	}
	if ctx == nil {
		return false, fmt.Errorf("context is nil")
	}
	if item.SourceID <= 0 {
		return false, fmt.Errorf("source id is invalid")
	}

	externalID := strings.TrimSpace(item.ExternalID)
	url := strings.TrimSpace(item.URL)
	title := strings.TrimSpace(item.Title)
	if externalID == "" && url == "" && title == "" {
		return false, fmt.Errorf("item must have at least one dedup key")
	}

	recent, err := s.repo.ListBySourceID(ctx, item.SourceID, s.limit)
	if err != nil {
		return false, fmt.Errorf("list source items: %w", err)
	}

	for _, existing := range recent {
		if item.ID > 0 && existing.ID == item.ID {
			continue
		}
		if externalID != "" && strings.TrimSpace(existing.ExternalID) == externalID {
			return true, nil
		}
		if url != "" && strings.TrimSpace(existing.URL) == url {
			return true, nil
		}
		if title != "" && strings.TrimSpace(existing.Title) == title {
			return true, nil
		}
	}

	return false, nil
}
