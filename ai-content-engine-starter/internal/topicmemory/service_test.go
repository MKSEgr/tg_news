package topicmemory

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-content-engine-starter/internal/domain"
)

type repoStub struct {
	upserted []domain.TopicMemory
	listOut  []domain.TopicMemory
	err      error
}

func (r *repoStub) UpsertMention(_ context.Context, memory domain.TopicMemory) (domain.TopicMemory, error) {
	if r.err != nil {
		return domain.TopicMemory{}, r.err
	}
	r.upserted = append(r.upserted, memory)
	return memory, nil
}

func (r *repoStub) ListTopByChannel(context.Context, int64, int) ([]domain.TopicMemory, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.listOut, nil
}

func TestObserveStoresDeterministicTopics(t *testing.T) {
	repo := &repoStub{}
	svc, err := New(repo)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	seenAt := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)
	items, err := svc.Observe(context.Background(), 7, "AI launch launch for tools and automation", seenAt)
	if err != nil {
		t.Fatalf("Observe() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatalf("expected non-empty topics")
	}
	if len(repo.upserted) != len(items) {
		t.Fatalf("upserted count = %d, want %d", len(repo.upserted), len(items))
	}
	for i := 1; i < len(repo.upserted); i++ {
		if repo.upserted[i-1].Topic > repo.upserted[i].Topic {
			t.Fatalf("topics are not deterministic-sorted")
		}
	}
}

func TestObserveValidationAndRepoError(t *testing.T) {
	repo := &repoStub{}
	svc, _ := New(repo)

	if _, err := svc.Observe(nil, 1, "text", time.Now()); err == nil {
		t.Fatalf("expected error for nil context")
	}
	if _, err := svc.Observe(context.Background(), 0, "text", time.Now()); err == nil {
		t.Fatalf("expected error for invalid channel")
	}
	if _, err := svc.Observe(context.Background(), 1, "", time.Now()); err == nil {
		t.Fatalf("expected error for empty text")
	}

	repo.err = errors.New("db")
	if _, err := svc.Observe(context.Background(), 1, "ai topic", time.Now()); err == nil {
		t.Fatalf("expected repo error")
	}
}

func TestTopTopics(t *testing.T) {
	repo := &repoStub{listOut: []domain.TopicMemory{{ChannelID: 1, Topic: "ai", MentionCount: 5}}}
	svc, _ := New(repo)

	items, err := svc.TopTopics(context.Background(), 1, 0)
	if err != nil {
		t.Fatalf("TopTopics() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
}

func TestNewValidation(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatalf("expected error for nil repo")
	}
}

func TestObserveSkipsOnlyStopWordsAndShortTokens(t *testing.T) {
	repo := &repoStub{}
	svc, _ := New(repo)

	items, err := svc.Observe(context.Background(), 1, "the and for ai", time.Now())
	if err != nil {
		t.Fatalf("Observe() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items len = %d, want 0", len(items))
	}
	if len(repo.upserted) != 0 {
		t.Fatalf("upserted len = %d, want 0", len(repo.upserted))
	}
}
