package storycluster

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"ai-content-engine-starter/internal/domain"
)

const maxClusterTokens = 6

// Service deterministically maps raw source items into stable story clusters.
type Service struct {
	clusters  domain.StoryClusterRepository
	events    domain.ClusterEventRepository
	stopWords map[string]struct{}
	nowFn     func() time.Time
}

// New creates a story clustering service.
func New(clusters domain.StoryClusterRepository, events domain.ClusterEventRepository) (*Service, error) {
	if clusters == nil {
		return nil, fmt.Errorf("story cluster repository is nil")
	}
	if events == nil {
		return nil, fmt.Errorf("cluster event repository is nil")
	}
	return &Service{
		clusters:  clusters,
		events:    events,
		stopWords: defaultStopWords(),
		nowFn:     func() time.Time { return time.Now().UTC() },
	}, nil
}

// ObserveSignal finds or creates a deterministic story cluster for a raw item and appends a signal event.
func (s *Service) ObserveSignal(ctx context.Context, item domain.SourceItem) (domain.StoryCluster, domain.ClusterEvent, error) {
	if s == nil {
		return domain.StoryCluster{}, domain.ClusterEvent{}, fmt.Errorf("story cluster service is nil")
	}
	if s.clusters == nil {
		return domain.StoryCluster{}, domain.ClusterEvent{}, fmt.Errorf("story cluster repository is nil")
	}
	if s.events == nil {
		return domain.StoryCluster{}, domain.ClusterEvent{}, fmt.Errorf("cluster event repository is nil")
	}
	if ctx == nil {
		return domain.StoryCluster{}, domain.ClusterEvent{}, fmt.Errorf("context is nil")
	}
	if item.ID <= 0 {
		return domain.StoryCluster{}, domain.ClusterEvent{}, fmt.Errorf("source item id is invalid")
	}
	if item.SourceID <= 0 {
		return domain.StoryCluster{}, domain.ClusterEvent{}, fmt.Errorf("source id is invalid")
	}

	key := s.ClusterKey(item)
	if key == "" {
		return domain.StoryCluster{}, domain.ClusterEvent{}, fmt.Errorf("cluster key is empty")
	}

	cluster, err := s.clusters.FindByKey(ctx, key)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return domain.StoryCluster{}, domain.ClusterEvent{}, fmt.Errorf("find story cluster: %w", err)
		}
		cluster, err = s.clusters.Create(ctx, domain.StoryCluster{
			ClusterKey: key,
			Title:      clusterTitle(item),
			Summary:    clusterSummary(item),
		})
		if err != nil {
			return domain.StoryCluster{}, domain.ClusterEvent{}, fmt.Errorf("create story cluster: %w", err)
		}
	}

	rawItemID := item.ID
	event, err := s.events.Create(ctx, domain.ClusterEvent{
		StoryClusterID: cluster.ID,
		RawItemID:      &rawItemID,
		EventType:      domain.ClusterEventTypeSignalAdded,
		EventTime:      observedAt(item, s.nowFn),
		MetadataJSON:   buildSignalMetadata(item),
	})
	if err != nil {
		return domain.StoryCluster{}, domain.ClusterEvent{}, fmt.Errorf("create cluster event: %w", err)
	}
	return cluster, event, nil
}

// ClusterKey derives a deterministic cluster key for a source item.
func (s *Service) ClusterKey(item domain.SourceItem) string {
	if s == nil {
		return ""
	}
	for _, text := range clusterKeyCandidates(item) {
		tokens := tokenize(text, s.stopWords)
		if len(tokens) == 0 {
			continue
		}
		sort.Strings(tokens)
		if len(tokens) > maxClusterTokens {
			tokens = tokens[:maxClusterTokens]
		}
		return strings.Join(tokens, "-")
	}
	return ""
}

func tokenize(text string, stopWords map[string]struct{}) []string {
	fields := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	seen := make(map[string]struct{}, len(fields))
	out := make([]string, 0, len(fields))
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if len([]rune(field)) < 3 {
			continue
		}
		if _, blocked := stopWords[field]; blocked {
			continue
		}
		if _, ok := seen[field]; ok {
			continue
		}
		seen[field] = struct{}{}
		out = append(out, field)
	}
	return out
}

func clusterTitle(item domain.SourceItem) string {
	if title := strings.TrimSpace(item.Title); title != "" {
		return title
	}
	if item.Body != nil {
		if body := strings.TrimSpace(*item.Body); body != "" {
			return body
		}
	}
	if url := strings.TrimSpace(item.URL); url != "" {
		return url
	}
	return strings.TrimSpace(item.ExternalID)
}

func clusterSummary(item domain.SourceItem) string {
	if item.Body != nil {
		if body := strings.TrimSpace(*item.Body); body != "" {
			return body
		}
	}
	return strings.TrimSpace(item.Title)
}

func observedAt(item domain.SourceItem, nowFn func() time.Time) time.Time {
	if item.PublishedAt != nil && !item.PublishedAt.IsZero() {
		return item.PublishedAt.UTC()
	}
	if !item.CollectedAt.IsZero() {
		return item.CollectedAt.UTC()
	}
	if nowFn == nil {
		return time.Now().UTC()
	}
	return nowFn().UTC()
}

func buildSignalMetadata(item domain.SourceItem) string {
	parts := make([]string, 0, 3)
	if item.ExternalID != "" {
		parts = append(parts, fmt.Sprintf(`"external_id":%q`, strings.TrimSpace(item.ExternalID)))
	}
	if item.URL != "" {
		parts = append(parts, fmt.Sprintf(`"url":%q`, strings.TrimSpace(item.URL)))
	}
	if item.Title != "" {
		parts = append(parts, fmt.Sprintf(`"title":%q`, strings.TrimSpace(item.Title)))
	}
	if len(parts) == 0 {
		return "{}"
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func defaultStopWords() map[string]struct{} {
	words := []string{"the", "and", "for", "with", "this", "that", "from", "into", "after", "over", "get", "gets", "как", "для", "что", "это", "или", "при", "без"}
	out := make(map[string]struct{}, len(words))
	for _, word := range words {
		out[word] = struct{}{}
	}
	return out
}

func clusterKeyCandidates(item domain.SourceItem) []string {
	candidates := make([]string, 0, 4)
	if title := strings.TrimSpace(item.Title); title != "" {
		candidates = append(candidates, title)
	}
	if item.Body != nil {
		if body := strings.TrimSpace(*item.Body); body != "" {
			candidates = append(candidates, body)
		}
	}
	if url := strings.TrimSpace(item.URL); url != "" {
		candidates = append(candidates, url)
	}
	if externalID := strings.TrimSpace(item.ExternalID); externalID != "" {
		candidates = append(candidates, externalID)
	}
	return candidates
}
