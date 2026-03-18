package storycluster

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-content-engine-starter/internal/domain"
)

type clusterRepoStub struct {
	findByKey map[string]domain.StoryCluster
	findErr   error
	created   []domain.StoryCluster
	createErr error
}

func (r *clusterRepoStub) Create(_ context.Context, cluster domain.StoryCluster) (domain.StoryCluster, error) {
	if r.createErr != nil {
		return domain.StoryCluster{}, r.createErr
	}
	cluster.ID = int64(len(r.created) + 1)
	r.created = append(r.created, cluster)
	if r.findByKey == nil {
		r.findByKey = make(map[string]domain.StoryCluster)
	}
	r.findByKey[cluster.ClusterKey] = cluster
	return cluster, nil
}
func (r *clusterRepoStub) GetByID(context.Context, int64) (domain.StoryCluster, error) {
	return domain.StoryCluster{}, nil
}
func (r *clusterRepoStub) FindByKey(_ context.Context, key string) (domain.StoryCluster, error) {
	if r.findErr != nil {
		return domain.StoryCluster{}, r.findErr
	}
	cluster, ok := r.findByKey[key]
	if !ok {
		return domain.StoryCluster{}, domain.ErrNotFound
	}
	return cluster, nil
}

type eventRepoStub struct {
	created   []domain.ClusterEvent
	createErr error
}

func (r *eventRepoStub) Create(_ context.Context, event domain.ClusterEvent) (domain.ClusterEvent, error) {
	if r.createErr != nil {
		return domain.ClusterEvent{}, r.createErr
	}
	event.ID = int64(len(r.created) + 1)
	r.created = append(r.created, event)
	return event, nil
}
func (r *eventRepoStub) ListByClusterID(context.Context, int64, int) ([]domain.ClusterEvent, error) {
	return nil, nil
}

func TestNewValidation(t *testing.T) {
	if _, err := New(nil, &eventRepoStub{}); err == nil {
		t.Fatalf("expected error for nil cluster repo")
	}
	if _, err := New(&clusterRepoStub{}, nil); err == nil {
		t.Fatalf("expected error for nil event repo")
	}
}

func TestClusterKeyDeterministicAcrossWordOrder(t *testing.T) {
	svc, _ := New(&clusterRepoStub{}, &eventRepoStub{})
	key1 := svc.ClusterKey(domain.SourceItem{Title: "OpenAI launches new AI tools for teams"})
	key2 := svc.ClusterKey(domain.SourceItem{Title: "Teams get new tools as OpenAI launches AI"})
	if key1 == "" || key2 == "" {
		t.Fatalf("expected non-empty keys")
	}
	if key1 != key2 {
		t.Fatalf("keys differ: %q != %q", key1, key2)
	}
}

func TestClusterKeyFallsBackWhenEarlierFieldTokenizesEmpty(t *testing.T) {
	svc, _ := New(&clusterRepoStub{}, &eventRepoStub{})
	body := "OpenAI launches platform updates"
	key := svc.ClusterKey(domain.SourceItem{
		Title: "AI",
		Body:  &body,
	})
	if key == "" {
		t.Fatalf("expected fallback key from body")
	}
}

func TestObserveSignalAcceptsWrappedNotFound(t *testing.T) {
	clusters := &clusterRepoStub{findErr: errors.Join(domain.ErrNotFound, errors.New("wrapped"))}
	events := &eventRepoStub{}
	svc, _ := New(clusters, events)

	cluster, _, err := svc.ObserveSignal(context.Background(), domain.SourceItem{
		ID:       13,
		SourceID: 4,
		Title:    "OpenAI launches new AI tools",
	})
	if err != nil {
		t.Fatalf("ObserveSignal() error = %v", err)
	}
	if cluster.ID == 0 {
		t.Fatalf("expected cluster to be created for wrapped not found")
	}
}

func TestObserveSignalCreatesClusterAndEvent(t *testing.T) {
	clusters := &clusterRepoStub{}
	events := &eventRepoStub{}
	svc, _ := New(clusters, events)
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	svc.nowFn = func() time.Time { return now }

	cluster, event, err := svc.ObserveSignal(context.Background(), domain.SourceItem{
		ID:         11,
		SourceID:   4,
		ExternalID: "abc",
		URL:        "https://example.com/story",
		Title:      "OpenAI launches new AI tools for teams",
	})
	if err != nil {
		t.Fatalf("ObserveSignal() error = %v", err)
	}
	if cluster.ID == 0 {
		t.Fatalf("expected created cluster id")
	}
	if len(clusters.created) != 1 {
		t.Fatalf("created clusters = %d, want 1", len(clusters.created))
	}
	if len(events.created) != 1 {
		t.Fatalf("created events = %d, want 1", len(events.created))
	}
	if event.RawItemID == nil || *event.RawItemID != 11 {
		t.Fatalf("raw item id = %#v, want 11", event.RawItemID)
	}
	if event.EventType != domain.ClusterEventTypeSignalAdded {
		t.Fatalf("event type = %q, want %q", event.EventType, domain.ClusterEventTypeSignalAdded)
	}
	if !event.EventTime.Equal(now) {
		t.Fatalf("event time = %v, want %v", event.EventTime, now)
	}
}

func TestObserveSignalReusesExistingCluster(t *testing.T) {
	clusters := &clusterRepoStub{}
	events := &eventRepoStub{}
	svc, _ := New(clusters, events)
	existingKey := svc.ClusterKey(domain.SourceItem{Title: "Teams tools OpenAI launches AI"})
	existing := domain.StoryCluster{ID: 9, ClusterKey: existingKey}
	clusters.findByKey = map[string]domain.StoryCluster{existing.ClusterKey: existing}

	cluster, _, err := svc.ObserveSignal(context.Background(), domain.SourceItem{
		ID:       12,
		SourceID: 4,
		Title:    "Teams tools OpenAI launches AI",
	})
	if err != nil {
		t.Fatalf("ObserveSignal() error = %v", err)
	}
	if cluster.ID != existing.ID {
		t.Fatalf("cluster id = %d, want %d", cluster.ID, existing.ID)
	}
	if len(clusters.created) != 0 {
		t.Fatalf("expected no cluster creation")
	}
}

func TestObserveSignalValidationAndErrors(t *testing.T) {
	clusters := &clusterRepoStub{}
	events := &eventRepoStub{}
	svc, _ := New(clusters, events)

	if _, _, err := svc.ObserveSignal(nil, domain.SourceItem{ID: 1, SourceID: 1, Title: "x"}); err == nil {
		t.Fatalf("expected nil context error")
	}
	if _, _, err := svc.ObserveSignal(context.Background(), domain.SourceItem{SourceID: 1, Title: "x"}); err == nil {
		t.Fatalf("expected invalid item id error")
	}
	if _, _, err := svc.ObserveSignal(context.Background(), domain.SourceItem{ID: 1, Title: "x"}); err == nil {
		t.Fatalf("expected invalid source id error")
	}

	clusters.findErr = errors.New("db")
	if _, _, err := svc.ObserveSignal(context.Background(), domain.SourceItem{ID: 1, SourceID: 1, Title: "valid title"}); err == nil {
		t.Fatalf("expected repository error")
	}
	clusters.findErr = nil
	events.createErr = errors.New("event db")
	if _, _, err := svc.ObserveSignal(context.Background(), domain.SourceItem{ID: 1, SourceID: 1, Title: "valid title"}); err == nil {
		t.Fatalf("expected event repository error")
	}
}

func TestObservedAtPrefersPublishedAtThenCollectedAt(t *testing.T) {
	publishedAt := time.Date(2026, 3, 17, 8, 0, 0, 0, time.UTC)
	collectedAt := time.Date(2026, 3, 18, 8, 0, 0, 0, time.UTC)
	if got := observedAt(domain.SourceItem{PublishedAt: &publishedAt, CollectedAt: collectedAt}, func() time.Time { return time.Now().UTC() }); !got.Equal(publishedAt) {
		t.Fatalf("publishedAt precedence mismatch")
	}
	if got := observedAt(domain.SourceItem{CollectedAt: collectedAt}, func() time.Time { return time.Now().UTC() }); !got.Equal(collectedAt) {
		t.Fatalf("collectedAt precedence mismatch")
	}
}
