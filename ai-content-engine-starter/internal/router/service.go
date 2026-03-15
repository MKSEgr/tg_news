package router

import (
	"fmt"
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
	if s == nil {
		return nil, fmt.Errorf("router service is nil")
	}
	if len(channels) == 0 {
		return nil, fmt.Errorf("channels are empty")
	}

	bySlug := make(map[string]domain.Channel, len(channels))
	for _, channel := range channels {
		slug := strings.TrimSpace(channel.Slug)
		if slug == "" || channel.ID <= 0 {
			continue
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

	if len(out) == 0 {
		if _, ok := bySlug["ai-news"]; ok {
			push("ai-news")
		} else {
			for _, channel := range bySlug {
				out = append(out, channel.ID)
				break
			}
		}
	}

	return out, nil
}

func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}
