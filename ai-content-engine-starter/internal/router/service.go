package router

import (
	"fmt"
	"sort"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

// Service routes scored source items into target channels.
type Service struct {
	newsKeywords      []string
	toolsKeywords     []string
	workflowsKeywords []string
}

// New creates a channel router with minimal MVP keyword rules.
func New() *Service {
	return &Service{
		newsKeywords:      []string{"release", "funding", "acquisition", "research", "model", "company"},
		toolsKeywords:     []string{"tool", "saas", "open source", "repository", "product", "launch"},
		workflowsKeywords: []string{"workflow", "automation", "prompt", "how to", "use case"},
	}
}

// Route returns matching channel IDs for the given item and channel catalog.
func (s *Service) Route(item domain.SourceItem, channels []domain.Channel) ([]int64, error) {
	return s.RouteWithMemory(item, channels, nil)
}

// RouteWithMemory extends keyword routing with deterministic topic-memory matches.
func (s *Service) RouteWithMemory(item domain.SourceItem, channels []domain.Channel, memoryByChannel map[int64][]domain.TopicMemory) ([]int64, error) {
	if s == nil {
		return nil, fmt.Errorf("router service is nil")
	}
	if len(channels) == 0 {
		return nil, fmt.Errorf("channels are empty")
	}

	bySlug := make(map[string]domain.Channel, len(channels))
	firstRoutableID := int64(0)
	for _, channel := range channels {
		slug := strings.TrimSpace(channel.Slug)
		if slug == "" || channel.ID <= 0 {
			continue
		}
		if firstRoutableID == 0 {
			firstRoutableID = channel.ID
		}
		bySlug[slug] = channel
	}
	if len(bySlug) == 0 {
		return nil, fmt.Errorf("channels do not contain routable entries")
	}

	text := strings.ToLower(strings.TrimSpace(item.Title))
	if item.Body != nil {
		text += " " + strings.ToLower(strings.TrimSpace(*item.Body))
	}

	out := make([]int64, 0, 3)
	push := func(slug string) {
		if channel, ok := bySlug[slug]; ok {
			for _, id := range out {
				if id == channel.ID {
					return
				}
			}
			out = append(out, channel.ID)
		}
	}

	if containsAny(text, s.newsKeywords) {
		push("ai-news")
	}
	if containsAny(text, s.toolsKeywords) {
		push("ai-tools")
	}
	if containsAny(text, s.workflowsKeywords) {
		push("ai-workflows")
	}

	for _, channel := range channels {
		if channel.ID <= 0 {
			continue
		}
		memories := memoryByChannel[channel.ID]
		for _, memory := range memories {
			topic := strings.ToLower(strings.TrimSpace(memory.Topic))
			if topic == "" {
				continue
			}
			if strings.Contains(text, topic) {
				out = appendUnique(out, channel.ID)
				break
			}
		}
	}

	if len(out) == 0 {
		if _, ok := bySlug["ai-news"]; ok {
			push("ai-news")
		} else {
			out = append(out, firstRoutableID)
		}
	}

	return out, nil
}

// RouteWithFeedback returns channels ordered by historical feedback strength (desc) for deterministic prioritization.
func (s *Service) RouteWithFeedback(item domain.SourceItem, channels []domain.Channel, feedbackByChannel map[int64]float64) ([]int64, error) {
	ids, err := s.RouteWithMemory(item, channels, nil)
	if err != nil {
		return nil, err
	}
	sorted := append([]int64(nil), ids...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left := feedbackByChannel[sorted[i]]
		right := feedbackByChannel[sorted[j]]
		if left == right {
			return sorted[i] < sorted[j]
		}
		return left > right
	})
	return sorted, nil
}

func appendUnique(ids []int64, value int64) []int64 {
	for _, existing := range ids {
		if existing == value {
			return ids
		}
	}
	return append(ids, value)
}

func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

// RouteWithCluster adds optional cluster context as a soft hint without overriding base routing decisions.
func (s *Service) RouteWithCluster(item domain.SourceItem, channels []domain.Channel, cluster domain.StoryCluster) ([]int64, error) {
	base, err := s.Route(item, channels)
	if err != nil {
		return nil, err
	}
	if cluster.ID <= 0 {
		return base, nil
	}
	hintItem := domain.SourceItem{Title: strings.TrimSpace(cluster.Title)}
	if summary := strings.TrimSpace(cluster.Summary); summary != "" {
		hintItem.Body = &summary
	}
	hintIDs, err := s.Route(hintItem, channels)
	if err != nil {
		return base, nil
	}
	return mergeSoftHintRouteIDs(base, hintIDs), nil
}

func mergeSoftHintRouteIDs(base []int64, hints []int64) []int64 {
	out := append([]int64(nil), base...)
	for _, id := range hints {
		out = appendUnique(out, id)
	}
	return out
}
