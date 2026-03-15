package scorer

import (
	"math"
	"strings"
	"time"

	"ai-content-engine-starter/internal/domain"
)

const (
	maxScore           = 100
	recencyWindowHours = 48
	maxRecencyPoints   = 60
	maxRelevancePoints = 40
	maxMemoryPoints    = 20
)

// Service calculates a simple trend score for normalized source items.
type Service struct {
	nowFn func() time.Time
}

// New creates a trend scoring service.
func New(nowFn func() time.Time) *Service {
	if nowFn == nil {
		nowFn = time.Now
	}
	return &Service{nowFn: nowFn}
}

// Score returns an integer trend score in range [0..100].
func (s *Service) Score(item domain.SourceItem) int {
	return s.ScoreWithMemory(item, nil)
}

// ScoreWithMemory applies deterministic score plus small explainable topic-memory boost.
func (s *Service) ScoreWithMemory(item domain.SourceItem, memories []domain.TopicMemory) int {
	if s == nil {
		return 0
	}
	if s.nowFn == nil {
		s.nowFn = time.Now
	}

	now := s.nowFn().UTC()
	recency := scoreRecency(item.PublishedAt, now)
	relevance := scoreRelevance(item.Title, item.Body)
	memory := scoreMemoryRelevance(item.Title, item.Body, memories)

	total := recency + relevance + memory
	if total < 0 {
		return 0
	}
	if total > maxScore {
		return maxScore
	}
	return total
}

func scoreRecency(publishedAt *time.Time, now time.Time) int {
	if publishedAt == nil {
		return maxRecencyPoints / 3
	}

	ageHours := now.Sub(publishedAt.UTC()).Hours()
	if ageHours <= 0 {
		return maxRecencyPoints
	}
	if ageHours >= recencyWindowHours {
		return 0
	}

	ratio := 1 - (ageHours / recencyWindowHours)
	return int(math.Round(ratio * maxRecencyPoints))
}

func scoreRelevance(title string, body *string) int {
	text := strings.ToLower(strings.TrimSpace(title))
	if body != nil {
		text += " " + strings.ToLower(strings.TrimSpace(*body))
	}

	keywords := []string{
		"ai", "llm", "gpt", "release", "launch", "open source", "automation", "workflow",
	}

	points := 0
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			points += 5
		}
	}
	if points > maxRelevancePoints {
		return maxRelevancePoints
	}
	return points
}

func scoreMemoryRelevance(title string, body *string, memories []domain.TopicMemory) int {
	if len(memories) == 0 {
		return 0
	}
	text := strings.ToLower(strings.TrimSpace(title))
	if body != nil {
		text += " " + strings.ToLower(strings.TrimSpace(*body))
	}
	if text == "" {
		return 0
	}

	points := 0
	for _, memory := range memories {
		topic := strings.ToLower(strings.TrimSpace(memory.Topic))
		if topic == "" {
			continue
		}
		if strings.Contains(text, topic) {
			points += 4
		}
		if points >= maxMemoryPoints {
			return maxMemoryPoints
		}
	}
	return points
}
