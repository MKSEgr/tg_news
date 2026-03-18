package feedbackloop

import (
	"context"
	"fmt"

	"ai-content-engine-starter/internal/domain"
)

// Service records deterministic performance feedback score from basic engagement metrics.
type Service struct {
	repo domain.PerformanceFeedbackRepository
}

// New creates performance feedback service.
func New(repo domain.PerformanceFeedbackRepository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("performance feedback repository is nil")
	}
	return &Service{repo: repo}, nil
}

// Record calculates score and persists feedback.
func (s *Service) Record(ctx context.Context, feedback domain.PerformanceFeedback) (domain.PerformanceFeedback, error) {
	if s == nil {
		return domain.PerformanceFeedback{}, fmt.Errorf("feedback loop service is nil")
	}
	if s.repo == nil {
		return domain.PerformanceFeedback{}, fmt.Errorf("performance feedback repository is nil")
	}
	if ctx == nil {
		return domain.PerformanceFeedback{}, fmt.Errorf("context is nil")
	}
	if feedback.DraftID <= 0 {
		return domain.PerformanceFeedback{}, fmt.Errorf("draft id is invalid")
	}
	if feedback.ChannelID <= 0 {
		return domain.PerformanceFeedback{}, fmt.Errorf("channel id is invalid")
	}
	if feedback.ViewsCount < 0 || feedback.ClicksCount < 0 || feedback.ReactionsCount < 0 || feedback.SharesCount < 0 {
		return domain.PerformanceFeedback{}, fmt.Errorf("feedback metrics must be non-negative")
	}
	if feedback.ViewsCount == 0 && (feedback.ClicksCount > 0 || feedback.ReactionsCount > 0 || feedback.SharesCount > 0) {
		return domain.PerformanceFeedback{}, fmt.Errorf("views count must be positive when engagement metrics are present")
	}

	feedback.Score = score(feedback)
	return s.repo.Upsert(ctx, feedback)
}

// Get returns persisted feedback for a draft.
func (s *Service) Get(ctx context.Context, draftID int64) (domain.PerformanceFeedback, error) {
	if s == nil {
		return domain.PerformanceFeedback{}, fmt.Errorf("feedback loop service is nil")
	}
	if s.repo == nil {
		return domain.PerformanceFeedback{}, fmt.Errorf("performance feedback repository is nil")
	}
	if ctx == nil {
		return domain.PerformanceFeedback{}, fmt.Errorf("context is nil")
	}
	if draftID <= 0 {
		return domain.PerformanceFeedback{}, fmt.Errorf("draft id is invalid")
	}
	return s.repo.GetByDraftID(ctx, draftID)
}

func score(feedback domain.PerformanceFeedback) float64 {
	if feedback.ViewsCount <= 0 {
		return 0
	}
	engagement := (2*feedback.ClicksCount + 3*feedback.ReactionsCount + 4*feedback.SharesCount)
	return float64(engagement) / float64(feedback.ViewsCount)
}
