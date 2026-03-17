package editorialplanner

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-content-engine-starter/internal/domain"
)

type testPublishIntentRepo struct {
	created     []domain.PublishIntent
	err         error
	listErr     error
	listByRawID map[int64][]domain.PublishIntent
}

func (r *testPublishIntentRepo) Create(_ context.Context, intent domain.PublishIntent) (domain.PublishIntent, error) {
	if r.err != nil {
		return domain.PublishIntent{}, r.err
	}
	intent.ID = int64(len(r.created) + 1)
	intent.CreatedAt = time.Now().UTC()
	r.created = append(r.created, intent)
	return intent, nil
}

func (r *testPublishIntentRepo) ListByRawItemID(_ context.Context, rawItemID int64, _ int) ([]domain.PublishIntent, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	if r.listByRawID == nil {
		return nil, nil
	}
	return r.listByRawID[rawItemID], nil
}

func (r *testPublishIntentRepo) UpdateStatus(context.Context, int64, domain.PublishIntentStatus) error {
	return nil
}

type testChannelsRepo struct {
	channels []domain.Channel
	err      error
}

func (r testChannelsRepo) Create(context.Context, domain.Channel) (domain.Channel, error) {
	return domain.Channel{}, nil
}
func (r testChannelsRepo) GetByID(context.Context, int64) (domain.Channel, error) {
	return domain.Channel{}, nil
}
func (r testChannelsRepo) List(context.Context) ([]domain.Channel, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.channels, nil
}

type testScorer struct{ score int }

func (s testScorer) Score(domain.SourceItem) int { return s.score }

type testRouter struct {
	ids []int64
	err error
}

func (r testRouter) Route(domain.SourceItem, []domain.Channel) ([]int64, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.ids, nil
}

func TestPlanForItemCreatesSingleIntent(t *testing.T) {
	repo := &testPublishIntentRepo{}
	svc, err := New(repo, testChannelsRepo{channels: []domain.Channel{{ID: 1}, {ID: 2}}}, testScorer{score: 7}, testRouter{ids: []int64{2, 1}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	intents, err := svc.PlanForItem(context.Background(), RawItem{ID: 99, Title: "x", URL: "https://example.com"})
	if err != nil {
		t.Fatalf("PlanForItem() error = %v", err)
	}
	if len(intents) != 1 {
		t.Fatalf("len(intents) = %d, want 1", len(intents))
	}
	if intents[0].RawItemID != 99 {
		t.Fatalf("RawItemID = %d, want 99", intents[0].RawItemID)
	}
	if intents[0].ChannelID != 2 {
		t.Fatalf("ChannelID = %d, want 2", intents[0].ChannelID)
	}
	if intents[0].Status != domain.PublishIntentStatusPlanned {
		t.Fatalf("Status = %q, want %q", intents[0].Status, domain.PublishIntentStatusPlanned)
	}
	if intents[0].Format != "text" {
		t.Fatalf("Format = %q, want text", intents[0].Format)
	}
}

func TestPlanForItemReturnsEmptyForNonPositiveScore(t *testing.T) {
	svc, err := New(&testPublishIntentRepo{}, testChannelsRepo{channels: []domain.Channel{{ID: 1}}}, testScorer{score: 0}, testRouter{ids: []int64{1}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	intents, err := svc.PlanForItem(context.Background(), RawItem{ID: 1})
	if err != nil {
		t.Fatalf("PlanForItem() error = %v", err)
	}
	if len(intents) != 0 {
		t.Fatalf("len(intents) = %d, want 0", len(intents))
	}
}

func TestPlanForItemPropagatesRepoError(t *testing.T) {
	repoErr := errors.New("boom")
	svc, err := New(&testPublishIntentRepo{err: repoErr}, testChannelsRepo{channels: []domain.Channel{{ID: 1}}}, testScorer{score: 2}, testRouter{ids: []int64{1}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = svc.PlanForItem(context.Background(), RawItem{ID: 1})
	if err == nil {
		t.Fatalf("PlanForItem() expected error")
	}
}

func TestPlanForItemSkipsWhenIntentAlreadyExists(t *testing.T) {
	repo := &testPublishIntentRepo{listByRawID: map[int64][]domain.PublishIntent{1: {{ID: 10, RawItemID: 1, ChannelID: 1}}}}
	svc, err := New(repo, testChannelsRepo{channels: []domain.Channel{{ID: 1}}}, testScorer{score: 4}, testRouter{ids: []int64{1}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	intents, err := svc.PlanForItem(context.Background(), RawItem{ID: 1})
	if err != nil {
		t.Fatalf("PlanForItem() error = %v", err)
	}
	if len(intents) != 0 {
		t.Fatalf("len(intents) = %d, want 0", len(intents))
	}
}

func TestPlanForItemAllowsDifferentChannelForSameRawItem(t *testing.T) {
	repo := &testPublishIntentRepo{listByRawID: map[int64][]domain.PublishIntent{1: {{ID: 10, RawItemID: 1, ChannelID: 1}}}}
	svc, err := New(repo, testChannelsRepo{channels: []domain.Channel{{ID: 1}, {ID: 2}}}, testScorer{score: 4}, testRouter{ids: []int64{2}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	intents, err := svc.PlanForItem(context.Background(), RawItem{ID: 1})
	if err != nil {
		t.Fatalf("PlanForItem() error = %v", err)
	}
	if len(intents) != 1 {
		t.Fatalf("len(intents) = %d, want 1", len(intents))
	}
	if intents[0].ChannelID != 2 {
		t.Fatalf("ChannelID = %d, want 2", intents[0].ChannelID)
	}
}

func TestPlanForItemPropagatesListError(t *testing.T) {
	repoErr := errors.New("list failed")
	svc, err := New(&testPublishIntentRepo{listErr: repoErr}, testChannelsRepo{channels: []domain.Channel{{ID: 1}}}, testScorer{score: 1}, testRouter{ids: []int64{1}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = svc.PlanForItem(context.Background(), RawItem{ID: 1})
	if err == nil {
		t.Fatalf("PlanForItem() expected error")
	}
}
