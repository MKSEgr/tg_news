package topicbalancer

import (
	"testing"
	"time"

	"ai-content-engine-starter/internal/domain"
)

func TestNewValidation(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatalf("expected validation error")
	}
	if _, err := New(Config{MaxSameTopicPosts: 1, Window: 0}); err == nil {
		t.Fatalf("expected window validation error")
	}
	if _, err := New(Config{MaxSameTopicPosts: 1, Window: time.Hour, Cooldown: -time.Minute}); err == nil {
		t.Fatalf("expected cooldown validation error")
	}
}

func TestBalanceAllowsWhenNoRecentSameTopicSignals(t *testing.T) {
	svc, err := New(Config{MaxSameTopicPosts: 2, Window: time.Hour, Cooldown: 10 * time.Minute})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	decision, err := svc.Balance(Candidate{
		Intent: domain.PublishIntent{ID: 1, ChannelID: 10},
		Topic:  "AI Agents",
	}, nil, time.Now().UTC())
	if err != nil {
		t.Fatalf("Balance() error = %v", err)
	}
	if decision.Action != ActionAllow {
		t.Fatalf("Action = %q, want %q", decision.Action, ActionAllow)
	}
}

func TestBalanceIgnoresSameChannelSignals(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	svc, err := New(Config{
		MaxSameTopicPosts: 1,
		Window:            time.Hour,
		Cooldown:          30 * time.Minute,
		ChannelPriority:   map[int64]int{2: 5},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	decision, err := svc.Balance(Candidate{
		Intent: domain.PublishIntent{ID: 1, ChannelID: 2},
		Topic:  "AI Agents",
	}, []PublishedSignal{
		{Intent: domain.PublishIntent{ID: 2, ChannelID: 2}, Topic: "AI Agents", PublishedAt: now.Add(-5 * time.Minute)},
	}, now)
	if err != nil {
		t.Fatalf("Balance() error = %v", err)
	}
	if decision.Action != ActionAllow {
		t.Fatalf("Action = %q, want %q", decision.Action, ActionAllow)
	}
}

func TestBalanceDelaysDuringCooldownForLowerPriorityChannel(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	svc, err := New(Config{
		MaxSameTopicPosts: 3,
		Window:            2 * time.Hour,
		Cooldown:          30 * time.Minute,
		ChannelPriority:   map[int64]int{1: 10, 2: 5},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	decision, err := svc.Balance(Candidate{
		Intent: domain.PublishIntent{ID: 1, ChannelID: 2},
		Topic:  "AI Agents",
	}, []PublishedSignal{
		{Intent: domain.PublishIntent{ID: 2, ChannelID: 1}, Topic: "AI Agents", PublishedAt: now.Add(-10 * time.Minute)},
	}, now)
	if err != nil {
		t.Fatalf("Balance() error = %v", err)
	}
	if decision.Action != ActionDelay {
		t.Fatalf("Action = %q, want %q", decision.Action, ActionDelay)
	}
	if decision.DelayUntil == nil || !decision.DelayUntil.Equal(now.Add(20*time.Minute)) {
		t.Fatalf("DelayUntil = %v, want %v", decision.DelayUntil, now.Add(20*time.Minute))
	}
}

func TestBalanceDeprioritizesDuringCooldownForHigherPriorityChannel(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	svc, err := New(Config{
		MaxSameTopicPosts: 3,
		Window:            2 * time.Hour,
		Cooldown:          30 * time.Minute,
		ChannelPriority:   map[int64]int{1: 5, 2: 10},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	decision, err := svc.Balance(Candidate{
		Intent: domain.PublishIntent{ID: 1, ChannelID: 2},
		Topic:  "AI Agents",
	}, []PublishedSignal{
		{Intent: domain.PublishIntent{ID: 2, ChannelID: 1}, Topic: "AI Agents", PublishedAt: now.Add(-10 * time.Minute)},
	}, now)
	if err != nil {
		t.Fatalf("Balance() error = %v", err)
	}
	if decision.Action != ActionDeprioritize {
		t.Fatalf("Action = %q, want %q", decision.Action, ActionDeprioritize)
	}
}

func TestBalanceDelaysWhenWindowLimitReached(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	svc, err := New(Config{
		MaxSameTopicPosts: 2,
		Window:            time.Hour,
		Cooldown:          5 * time.Minute,
		ChannelPriority:   map[int64]int{1: 10, 2: 5, 3: 1},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	decision, err := svc.Balance(Candidate{
		Intent:     domain.PublishIntent{ID: 3, ChannelID: 3},
		ClusterKey: "launch-agents",
	}, []PublishedSignal{
		{Intent: domain.PublishIntent{ID: 11, ChannelID: 1}, ClusterKey: "launch-agents", PublishedAt: now.Add(-50 * time.Minute)},
		{Intent: domain.PublishIntent{ID: 12, ChannelID: 2}, ClusterKey: "launch-agents", PublishedAt: now.Add(-20 * time.Minute)},
	}, now)
	if err != nil {
		t.Fatalf("Balance() error = %v", err)
	}
	if decision.Action != ActionDelay {
		t.Fatalf("Action = %q, want %q", decision.Action, ActionDelay)
	}
	if decision.DelayUntil == nil || !decision.DelayUntil.Equal(now.Add(10*time.Minute)) {
		t.Fatalf("DelayUntil = %v, want %v", decision.DelayUntil, now.Add(10*time.Minute))
	}
}

func TestBalanceDeprioritizesLowerPriorityWithinWindow(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	svc, err := New(Config{
		MaxSameTopicPosts: 3,
		Window:            time.Hour,
		Cooldown:          0,
		ChannelPriority:   map[int64]int{1: 10, 2: 2},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	decision, err := svc.Balance(Candidate{
		Intent: domain.PublishIntent{ID: 2, ChannelID: 2},
		Topic:  "AI Agents",
	}, []PublishedSignal{
		{Intent: domain.PublishIntent{ID: 1, ChannelID: 1}, Topic: "AI Agents", PublishedAt: now.Add(-40 * time.Minute)},
	}, now)
	if err != nil {
		t.Fatalf("Balance() error = %v", err)
	}
	if decision.Action != ActionDeprioritize {
		t.Fatalf("Action = %q, want %q", decision.Action, ActionDeprioritize)
	}
}

func TestBalanceValidation(t *testing.T) {
	svc, err := New(Config{MaxSameTopicPosts: 1, Window: time.Hour})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := svc.Balance(Candidate{}, nil, time.Now().UTC()); err == nil {
		t.Fatalf("expected invalid intent error")
	}
	if _, err := svc.Balance(Candidate{Intent: domain.PublishIntent{ID: 1}}, nil, time.Now().UTC()); err == nil {
		t.Fatalf("expected invalid channel error")
	}
	if _, err := svc.Balance(Candidate{Intent: domain.PublishIntent{ID: 1, ChannelID: 1}}, nil, time.Time{}); err == nil {
		t.Fatalf("expected zero time error")
	}
	if _, err := svc.Balance(Candidate{Intent: domain.PublishIntent{ID: 1, ChannelID: 1}}, nil, time.Now().UTC()); err == nil {
		t.Fatalf("expected missing topic signal error")
	}
}

func TestBalanceIgnoresFutureSignals(t *testing.T) {
	now := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	svc, err := New(Config{
		MaxSameTopicPosts: 1,
		Window:            time.Hour,
		Cooldown:          30 * time.Minute,
		ChannelPriority:   map[int64]int{1: 10, 2: 5},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	decision, err := svc.Balance(Candidate{
		Intent: domain.PublishIntent{ID: 1, ChannelID: 2},
		Topic:  "AI Agents",
	}, []PublishedSignal{
		{Intent: domain.PublishIntent{ID: 2, ChannelID: 1}, Topic: "AI Agents", PublishedAt: now.Add(5 * time.Minute)},
	}, now)
	if err != nil {
		t.Fatalf("Balance() error = %v", err)
	}
	if decision.Action != ActionAllow {
		t.Fatalf("Action = %q, want %q", decision.Action, ActionAllow)
	}
}
