package scorer

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-content-engine-starter/internal/domain"
)

type rankingFeatureReaderStub struct {
	byEntity map[string][]domain.RankingFeature
	err      error
}

func (r *rankingFeatureReaderStub) ListByEntity(_ context.Context, entityType string, entityID int64, _ int) ([]domain.RankingFeature, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.byEntity[entityType+":"+stringKey(entityID)], nil
}

func stringKey(id int64) string { return time.Unix(id, 0).UTC().Format(time.RFC3339Nano) }

func entityKey(entityType string, entityID int64) string {
	return entityType + ":" + stringKey(entityID)
}

func TestNewAdaptiveValidation(t *testing.T) {
	base := New(nil)
	if _, err := NewAdaptive(nil, &rankingFeatureReaderStub{}); err == nil {
		t.Fatalf("expected nil base scorer error")
	}
	if _, err := NewAdaptive(base, nil); err == nil {
		t.Fatalf("expected nil feature reader error")
	}
}

func TestExplainAdaptiveAppliesBoundedAdjustments(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	base := New(func() time.Time { return now })
	body := "Transformers workflow for launch teams"
	reader := &rankingFeatureReaderStub{byEntity: map[string][]domain.RankingFeature{
		entityKey(rankingEntityTypeChannel, 7): {
			{FeatureName: rankingFeatureChannelPerformance, FeatureValue: 2.5},
		},
		entityKey(rankingEntityTypeTopic, topicEntityID("transformers")): {
			{FeatureName: rankingFeatureTopicPerformance, FeatureValue: 1.5},
		},
		entityKey(rankingEntityTypeFormat, formatEntityID("text")): {
			{FeatureName: rankingFeatureFormatSuccess, FeatureValue: 1.0},
		},
	}}
	svc, err := NewAdaptive(base, reader)
	if err != nil {
		t.Fatalf("NewAdaptive() error = %v", err)
	}

	result, err := svc.ExplainAdaptive(context.Background(), domain.SourceItem{Title: "AI release", Body: &body}, 7, "text")
	if err != nil {
		t.Fatalf("ExplainAdaptive() error = %v", err)
	}
	if result.ChannelAdjustment <= 0 || result.FormatAdjustment <= 0 {
		t.Fatalf("expected positive channel/format adjustments, got %+v", result)
	}
	if result.TopicAdjustment < 0 {
		t.Fatalf("expected non-negative topic adjustment, got %+v", result)
	}
	if result.TotalAdjustment > maxAdaptiveTotalPoints {
		t.Fatalf("total adjustment = %d, want <= %d", result.TotalAdjustment, maxAdaptiveTotalPoints)
	}
	if result.FinalScore <= result.BaseScore {
		t.Fatalf("final score = %d, base score = %d, want increase", result.FinalScore, result.BaseScore)
	}
}

func TestExplainAdaptiveValidationAndErrors(t *testing.T) {
	base := New(nil)
	reader := &rankingFeatureReaderStub{}
	svc, _ := NewAdaptive(base, reader)
	if _, err := svc.ExplainAdaptive(nil, domain.SourceItem{Title: "AI"}, 1, "text"); err == nil {
		t.Fatalf("expected nil context error")
	}
	if _, err := svc.ExplainAdaptive(context.Background(), domain.SourceItem{Title: "AI"}, 0, "text"); err == nil {
		t.Fatalf("expected invalid channel id error")
	}
	reader.err = errors.New("db")
	if _, err := svc.ExplainAdaptive(context.Background(), domain.SourceItem{Title: "AI"}, 1, "text"); err == nil {
		t.Fatalf("expected repository error")
	}
}

func TestScoreAdaptiveUsesDefaultFormatAndCapsNegativeScores(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	base := New(func() time.Time { return now })
	reader := &rankingFeatureReaderStub{byEntity: map[string][]domain.RankingFeature{
		entityKey(rankingEntityTypeChannel, 1):                                 {{FeatureName: rankingFeatureChannelPerformance, FeatureValue: -5}},
		entityKey(rankingEntityTypeFormat, formatEntityID(defaultDraftFormat)): {{FeatureName: rankingFeatureFormatSuccess, FeatureValue: -5}},
	}}
	svc, _ := NewAdaptive(base, reader)
	score, err := svc.ScoreAdaptive(context.Background(), domain.SourceItem{Title: "Weekly digest"}, 1, "")
	if err != nil {
		t.Fatalf("ScoreAdaptive() error = %v", err)
	}
	if score < 0 || score > maxScore {
		t.Fatalf("score = %d, want bounded [0,%d]", score, maxScore)
	}
}

func TestExplainAdaptiveUsesTopicPerformance(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	base := New(func() time.Time { return now })
	body := "transformers transformers transformers"
	reader := &rankingFeatureReaderStub{byEntity: map[string][]domain.RankingFeature{
		entityKey(rankingEntityTypeTopic, topicEntityID("transformers")): {{FeatureName: rankingFeatureTopicPerformance, FeatureValue: 1.5}},
	}}
	svc, _ := NewAdaptive(base, reader)
	result, err := svc.ExplainAdaptive(context.Background(), domain.SourceItem{Title: "Notes", Body: &body}, 7, "text")
	if err != nil {
		t.Fatalf("ExplainAdaptive() error = %v", err)
	}
	if result.TopicAdjustment <= 0 {
		t.Fatalf("expected positive topic adjustment, got %+v", result)
	}
}

func TestAdaptiveFeatureEntityHelpersAreStableAndNormalized(t *testing.T) {
	if got, want := TopicEntityID(" Transformers "), TopicEntityID("transformers"); got != want {
		t.Fatalf("TopicEntityID normalization mismatch: got %d want %d", got, want)
	}
	if got, want := FormatEntityID(" Text "), FormatEntityID("text"); got != want {
		t.Fatalf("FormatEntityID normalization mismatch: got %d want %d", got, want)
	}
	if TopicEntityID("transformers") == 0 {
		t.Fatalf("TopicEntityID returned zero")
	}
	if FormatEntityID("text") == 0 {
		t.Fatalf("FormatEntityID returned zero")
	}
}
