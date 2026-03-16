package topicmemory

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode"

	"ai-content-engine-starter/internal/domain"
)

const defaultTopLimit = 20

// Service provides deterministic and explainable topic memory operations.
type Service struct {
	repo      domain.TopicMemoryRepository
	stopWords map[string]struct{}
}

// New creates topic memory service.
func New(repo domain.TopicMemoryRepository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("topic memory repository is nil")
	}
	return &Service{repo: repo, stopWords: defaultStopWords()}, nil
}

// Observe extracts topics from text and updates memory for a channel.
func (s *Service) Observe(ctx context.Context, channelID int64, text string, seenAt time.Time) ([]domain.TopicMemory, error) {
	if s == nil {
		return nil, fmt.Errorf("topic memory service is nil")
	}
	if s.repo == nil {
		return nil, fmt.Errorf("topic memory repository is nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if channelID <= 0 {
		return nil, fmt.Errorf("channel id is invalid")
	}
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text is empty")
	}
	if seenAt.IsZero() {
		seenAt = time.Now().UTC()
	}

	counts := extractTopicCounts(text, s.stopWords)
	if len(counts) == 0 {
		return nil, nil
	}

	topics := sortedTopics(counts)
	out := make([]domain.TopicMemory, 0, len(topics))
	for _, topic := range topics {
		memory, err := s.repo.UpsertMention(ctx, domain.TopicMemory{
			ChannelID:    channelID,
			Topic:        topic,
			MentionCount: counts[topic],
			LastSeenAt:   seenAt.UTC(),
		})
		if err != nil {
			return nil, fmt.Errorf("upsert topic %q: %w", topic, err)
		}
		out = append(out, memory)
	}
	return out, nil
}

// TopTopics returns top remembered topics for a channel.
func (s *Service) TopTopics(ctx context.Context, channelID int64, limit int) ([]domain.TopicMemory, error) {
	if s == nil {
		return nil, fmt.Errorf("topic memory service is nil")
	}
	if s.repo == nil {
		return nil, fmt.Errorf("topic memory repository is nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if channelID <= 0 {
		return nil, fmt.Errorf("channel id is invalid")
	}
	if limit <= 0 {
		limit = defaultTopLimit
	}
	return s.repo.ListTopByChannel(ctx, channelID, limit)
}

func extractTopicCounts(text string, stopWords map[string]struct{}) map[string]int {
	text = strings.ToLower(text)
	tokens := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	counts := make(map[string]int)
	for _, token := range tokens {
		token = strings.TrimSpace(token)
		if len([]rune(token)) < 3 {
			continue
		}
		if _, blocked := stopWords[token]; blocked {
			continue
		}
		counts[token]++
	}
	return counts
}

func sortedTopics(counts map[string]int) []string {
	topics := make([]string, 0, len(counts))
	for topic := range counts {
		topics = append(topics, topic)
	}
	sort.Strings(topics)
	return topics
}

func defaultStopWords() map[string]struct{} {
	words := []string{"the", "and", "for", "with", "this", "that", "from", "как", "для", "что", "это", "или", "при", "без"}
	out := make(map[string]struct{}, len(words))
	for _, w := range words {
		out[w] = struct{}{}
	}
	return out
}
