package analytics

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"ai-content-engine-starter/internal/domain"
)

type draftRepoStub struct {
	byStatus map[domain.DraftStatus][]domain.Draft
	err      error
}

func (s *draftRepoStub) Create(context.Context, domain.Draft) (domain.Draft, error) {
	return domain.Draft{}, nil
}
func (s *draftRepoStub) GetByID(context.Context, int64) (domain.Draft, error) {
	return domain.Draft{}, nil
}
func (s *draftRepoStub) ListByStatus(_ context.Context, status domain.DraftStatus, _ int) ([]domain.Draft, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.byStatus[status], nil
}
func (s *draftRepoStub) UpdateStatus(context.Context, int64, domain.DraftStatus) error { return nil }

type feedbackRepoStub struct {
	byDraft map[int64]domain.PerformanceFeedback
	err     error
}

func (s *feedbackRepoStub) Upsert(context.Context, domain.PerformanceFeedback) (domain.PerformanceFeedback, error) {
	return domain.PerformanceFeedback{}, nil
}
func (s *feedbackRepoStub) GetByDraftID(_ context.Context, draftID int64) (domain.PerformanceFeedback, error) {
	if s.err != nil {
		return domain.PerformanceFeedback{}, s.err
	}
	item, ok := s.byDraft[draftID]
	if !ok {
		return domain.PerformanceFeedback{}, domain.ErrNotFound
	}
	return item, nil
}

func TestNewValidation(t *testing.T) {
	if _, err := New(nil, &feedbackRepoStub{}); err == nil {
		t.Fatalf("expected nil drafts repository error")
	}
	if _, err := New(&draftRepoStub{}, nil); err == nil {
		t.Fatalf("expected nil feedback repository error")
	}
}

func TestBuildByChannel(t *testing.T) {
	now := time.Now().UTC()
	service, err := New(
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{
			domain.DraftStatusPosted: {
				{ID: 1, ChannelID: 7, UpdatedAt: now.Add(-2 * time.Hour)},
				{ID: 2, ChannelID: 7, UpdatedAt: now.Add(-1 * time.Hour)},
				{ID: 3, ChannelID: 9, UpdatedAt: now.Add(-3 * time.Hour)},
				{ID: -1, ChannelID: 9, UpdatedAt: now},
			},
		}},
		&feedbackRepoStub{byDraft: map[int64]domain.PerformanceFeedback{
			1: {DraftID: 1, ChannelID: 7, Variant: "A", Score: 1.0},
			2: {DraftID: 2, ChannelID: 7, Variant: "B", Score: 3.0},
			3: {DraftID: 3, ChannelID: 9, Variant: "A", Score: 2.0},
		}},
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	out, err := service.BuildByChannel(context.Background())
	if err != nil {
		t.Fatalf("BuildByChannel() error = %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("len(out) = %d, want 2", len(out))
	}
	if out[0].ChannelID != 7 || out[0].PostedDrafts != 2 || out[0].FeedbackDrafts != 2 {
		t.Fatalf("channel 7 summary = %+v", out[0])
	}
	if out[0].AvgScore != 2.0 || out[0].AvgScoreA != 1.0 || out[0].AvgScoreB != 3.0 {
		t.Fatalf("channel 7 score summary = %+v", out[0])
	}
	if !out[0].LastPostedAt.Equal(now.Add(-1 * time.Hour)) {
		t.Fatalf("channel 7 last posted = %v", out[0].LastPostedAt)
	}

	if out[1].ChannelID != 9 || out[1].PostedDrafts != 1 || out[1].FeedbackDrafts != 1 {
		t.Fatalf("channel 9 summary = %+v", out[1])
	}
	if out[1].AvgScore != 2.0 || out[1].AvgScoreA != 2.0 || out[1].AvgScoreB != 0 {
		t.Fatalf("channel 9 score summary = %+v", out[1])
	}
}

func TestBuildByChannelValidationAndErrors(t *testing.T) {
	var nilSvc *Service
	if _, err := nilSvc.BuildByChannel(context.Background()); err == nil {
		t.Fatalf("expected nil service error")
	}

	svc := &Service{}
	if _, err := svc.BuildByChannel(nil); err == nil {
		t.Fatalf("expected nil context error")
	}

	svc = &Service{drafts: &draftRepoStub{}, feedback: &feedbackRepoStub{}, limit: 0}
	if _, err := svc.BuildByChannel(context.Background()); err == nil {
		t.Fatalf("expected invalid limit error")
	}

	svc = &Service{drafts: &draftRepoStub{err: errors.New("boom")}, feedback: &feedbackRepoStub{}, limit: 1}
	if _, err := svc.BuildByChannel(context.Background()); err == nil {
		t.Fatalf("expected list error")
	}

	svc = &Service{drafts: &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{
		domain.DraftStatusPosted: {{ID: 1, ChannelID: 7}},
	}}, feedback: &feedbackRepoStub{err: errors.New("boom")}, limit: 1}
	if _, err := svc.BuildByChannel(context.Background()); err == nil {
		t.Fatalf("expected feedback error")
	}
}

func TestBuildByChannelSkipsInvalidFeedbackScores(t *testing.T) {
	service, err := New(
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{
			domain.DraftStatusPosted: {
				{ID: 1, ChannelID: 7},
				{ID: 2, ChannelID: 7},
			},
		}},
		&feedbackRepoStub{byDraft: map[int64]domain.PerformanceFeedback{
			1: {DraftID: 1, ChannelID: 7, Variant: "A", Score: math.NaN()},
			2: {DraftID: 2, ChannelID: 7, Variant: "B", Score: 2.0},
		}},
	)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	out, err := service.BuildByChannel(context.Background())
	if err != nil {
		t.Fatalf("BuildByChannel() error = %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("len(out) = %d, want 1", len(out))
	}
	if out[0].FeedbackDrafts != 1 || out[0].AvgScore != 2.0 || out[0].AvgScoreA != 0 || out[0].AvgScoreB != 2.0 {
		t.Fatalf("summary = %+v", out[0])
	}
}
