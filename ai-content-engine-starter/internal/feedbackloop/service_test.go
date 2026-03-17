package feedbackloop

import (
	"context"
	"errors"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

type repoStub struct {
	stored domain.PerformanceFeedback
	err    error
}

func (r *repoStub) Upsert(_ context.Context, feedback domain.PerformanceFeedback) (domain.PerformanceFeedback, error) {
	if r.err != nil {
		return domain.PerformanceFeedback{}, r.err
	}
	r.stored = feedback
	return feedback, nil
}

func (r *repoStub) GetByDraftID(_ context.Context, draftID int64) (domain.PerformanceFeedback, error) {
	if r.err != nil {
		return domain.PerformanceFeedback{}, r.err
	}
	if r.stored.DraftID != draftID {
		return domain.PerformanceFeedback{}, domain.ErrNotFound
	}
	return r.stored, nil
}

func TestRecordCalculatesDeterministicScore(t *testing.T) {
	repo := &repoStub{}
	svc, _ := New(repo)
	feedback, err := svc.Record(context.Background(), domain.PerformanceFeedback{DraftID: 10, ChannelID: 1, ViewsCount: 100, ClicksCount: 20, ReactionsCount: 10, SharesCount: 5})
	if err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	want := float64(2*20+3*10+4*5) / 100.0
	if feedback.Score != want {
		t.Fatalf("score = %f, want %f", feedback.Score, want)
	}
}

func TestRecordValidationAndErrors(t *testing.T) {
	repo := &repoStub{}
	svc, _ := New(repo)
	if _, err := svc.Record(nil, domain.PerformanceFeedback{DraftID: 1, ChannelID: 1}); err == nil {
		t.Fatalf("expected nil context error")
	}
	if _, err := svc.Record(context.Background(), domain.PerformanceFeedback{ChannelID: 1}); err == nil {
		t.Fatalf("expected invalid draft id error")
	}
	if _, err := svc.Record(context.Background(), domain.PerformanceFeedback{DraftID: 1}); err == nil {
		t.Fatalf("expected invalid channel id error")
	}
	if _, err := svc.Record(context.Background(), domain.PerformanceFeedback{DraftID: 1, ChannelID: 1, ViewsCount: -1}); err == nil {
		t.Fatalf("expected invalid metrics error")
	}
	repo.err = errors.New("db")
	if _, err := svc.Record(context.Background(), domain.PerformanceFeedback{DraftID: 1, ChannelID: 1}); err == nil {
		t.Fatalf("expected repo error")
	}
}

func TestGet(t *testing.T) {
	repo := &repoStub{stored: domain.PerformanceFeedback{DraftID: 7, ChannelID: 1, ViewsCount: 10}}
	svc, _ := New(repo)
	item, err := svc.Get(context.Background(), 7)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if item.DraftID != 7 {
		t.Fatalf("draft id = %d, want 7", item.DraftID)
	}
}

func TestNewValidation(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatalf("expected error for nil repo")
	}
}
