package scorer

import (
	"testing"
	"time"

	"ai-content-engine-starter/internal/domain"
)

func TestScoreRecentKeywordRichItem(t *testing.T) {
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	svc := New(func() time.Time { return now })
	body := "Open source automation workflow launch"
	published := now.Add(-2 * time.Hour)

	score := svc.Score(domain.SourceItem{
		Title:       "AI GPT release",
		Body:        &body,
		PublishedAt: &published,
	})

	if score <= 70 {
		t.Fatalf("score = %d, want > 70", score)
	}
}

func TestScoreOldItemWithNoKeywords(t *testing.T) {
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	svc := New(func() time.Time { return now })
	published := now.Add(-72 * time.Hour)

	score := svc.Score(domain.SourceItem{
		Title:       "Weekly digest",
		PublishedAt: &published,
	})

	if score != 0 {
		t.Fatalf("score = %d, want 0", score)
	}
}

func TestScoreFallbackWhenPublishedAtMissing(t *testing.T) {
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	svc := New(func() time.Time { return now })

	score := svc.Score(domain.SourceItem{Title: "New AI tools"})
	if score < 20 {
		t.Fatalf("score = %d, want >= 20", score)
	}
}

func TestNewUsesTimeNowWhenNilNowFn(t *testing.T) {
	svc := New(nil)
	score := svc.Score(domain.SourceItem{Title: "AI"})
	if score < 0 || score > 100 {
		t.Fatalf("score = %d, want in [0,100]", score)
	}
}

func TestScoreHandlesNilServiceOrNowFn(t *testing.T) {
	var nilService *Service
	if score := nilService.Score(domain.SourceItem{Title: "AI"}); score != 0 {
		t.Fatalf("nil service score = %d, want 0", score)
	}

	svc := &Service{}
	score := svc.Score(domain.SourceItem{Title: "AI"})
	if score < 0 || score > 100 {
		t.Fatalf("score = %d, want in [0,100]", score)
	}
}

func TestScoreWithMemoryAddsDeterministicBoost(t *testing.T) {
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	svc := New(func() time.Time { return now })
	body := "Deep dive into transformers for production"

	withoutMemory := svc.Score(domain.SourceItem{Title: "Engineering notes", Body: &body})
	withMemory := svc.ScoreWithMemory(domain.SourceItem{Title: "Engineering notes", Body: &body}, []domain.TopicMemory{
		{Topic: "transformers", MentionCount: 5},
		{Topic: "agents", MentionCount: 3},
	})

	if withMemory <= withoutMemory {
		t.Fatalf("withMemory=%d, withoutMemory=%d, want boost", withMemory, withoutMemory)
	}
}

func TestScoreWithFeedbackUsesChannelPrior(t *testing.T) {
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	svc := New(func() time.Time { return now })
	base := svc.Score(domain.SourceItem{Title: "AI update"})
	withFeedback := svc.ScoreWithFeedback(domain.SourceItem{Title: "AI update"}, map[int64]float64{1: 2.0, 2: 1.0})
	if withFeedback <= base {
		t.Fatalf("withFeedback=%d, base=%d, want feedback boost", withFeedback, base)
	}
}
