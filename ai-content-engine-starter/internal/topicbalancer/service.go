package topicbalancer

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"ai-content-engine-starter/internal/domain"
)

// Action is the balancing recommendation for one candidate publish intent.
type Action string

const (
	// ActionAllow means the intent can proceed normally.
	ActionAllow Action = "allow"
	// ActionDelay means the intent should wait until a computed time.
	ActionDelay Action = "delay"
	// ActionDeprioritize means the intent can proceed, but after fresher/other topics.
	ActionDeprioritize Action = "deprioritize"
)

// Config defines simple, explainable balancing rules.
type Config struct {
	MaxSameTopicPosts int
	Window            time.Duration
	Cooldown          time.Duration
	ChannelPriority   map[int64]int
}

// Candidate is the intent under evaluation.
type Candidate struct {
	Intent     domain.PublishIntent
	Topic      string
	ClusterKey string
}

// PublishedSignal is a recent same-system publish signal used for balancing.
type PublishedSignal struct {
	Intent      domain.PublishIntent
	Topic       string
	ClusterKey  string
	PublishedAt time.Time
}

// Decision is the balancing result.
type Decision struct {
	Action     Action
	TopicKey   string
	DelayUntil *time.Time
	Reason     string
	Signals    []string
}

// Service applies cross-channel topic balancing rules.
type Service struct {
	cfg Config
}

// New creates a balancing service.
func New(cfg Config) (*Service, error) {
	if cfg.MaxSameTopicPosts <= 0 {
		return nil, fmt.Errorf("max same-topic posts must be greater than zero")
	}
	if cfg.Window <= 0 {
		return nil, fmt.Errorf("window must be greater than zero")
	}
	if cfg.Cooldown < 0 {
		return nil, fmt.Errorf("cooldown must be greater than or equal to zero")
	}
	if cfg.ChannelPriority == nil {
		cfg.ChannelPriority = map[int64]int{}
	}
	return &Service{cfg: cfg}, nil
}

// Balance returns an allow/delay/deprioritize recommendation for a publish intent.
func (s *Service) Balance(candidate Candidate, recent []PublishedSignal, now time.Time) (Decision, error) {
	if s == nil {
		return Decision{}, fmt.Errorf("topic balancer service is nil")
	}
	if candidate.Intent.ID <= 0 {
		return Decision{}, fmt.Errorf("publish intent id is invalid")
	}
	if candidate.Intent.ChannelID <= 0 {
		return Decision{}, fmt.Errorf("channel id is invalid")
	}
	if now.IsZero() {
		return Decision{}, fmt.Errorf("current time is required")
	}

	topicKey := signalKey(candidate.Topic, candidate.ClusterKey)
	if topicKey == "" {
		return Decision{}, fmt.Errorf("topic or cluster key is required")
	}

	windowStart := now.Add(-s.cfg.Window)
	cooldownStart := now.Add(-s.cfg.Cooldown)
	relevant := make([]PublishedSignal, 0, len(recent))
	for _, signal := range recent {
		if signal.Intent.ChannelID <= 0 || signal.PublishedAt.IsZero() {
			continue
		}
		if signal.Intent.ChannelID == candidate.Intent.ChannelID {
			continue
		}
		if signalKey(signal.Topic, signal.ClusterKey) != topicKey {
			continue
		}
		if signal.PublishedAt.After(now) {
			continue
		}
		if signal.PublishedAt.Before(windowStart) {
			continue
		}
		relevant = append(relevant, signal)
	}
	sort.Slice(relevant, func(i, j int) bool {
		return relevant[i].PublishedAt.Before(relevant[j].PublishedAt)
	})

	signals := []string{
		fmt.Sprintf("topic=%s", topicKey),
		fmt.Sprintf("window_posts=%d", len(relevant)),
		fmt.Sprintf("candidate_priority=%d", s.channelPriority(candidate.Intent.ChannelID)),
	}

	if len(relevant) == 0 {
		return Decision{
			Action:   ActionAllow,
			TopicKey: topicKey,
			Reason:   "no recent same-topic posts in balancing window",
			Signals:  signals,
		}, nil
	}

	latest := relevant[len(relevant)-1]
	highestRecentPriority := s.highestPriority(relevant)
	candidatePriority := s.channelPriority(candidate.Intent.ChannelID)
	if s.cfg.Cooldown > 0 && !latest.PublishedAt.Before(cooldownStart) {
		signals = append(signals, fmt.Sprintf("cooldown_until=%s", latest.PublishedAt.Add(s.cfg.Cooldown).UTC().Format(time.RFC3339)))
		if candidatePriority > highestRecentPriority {
			return Decision{
				Action:   ActionDeprioritize,
				TopicKey: topicKey,
				Reason:   "same topic is in cooldown, but candidate channel has higher priority",
				Signals:  signals,
			}, nil
		}
		delayUntil := latest.PublishedAt.Add(s.cfg.Cooldown).UTC()
		return Decision{
			Action:     ActionDelay,
			TopicKey:   topicKey,
			DelayUntil: &delayUntil,
			Reason:     "same topic was posted recently in another channel cooldown window",
			Signals:    signals,
		}, nil
	}

	if len(relevant) >= s.cfg.MaxSameTopicPosts {
		delayUntil := relevant[0].PublishedAt.Add(s.cfg.Window).UTC()
		signals = append(signals, fmt.Sprintf("window_limit=%d", s.cfg.MaxSameTopicPosts))
		if candidatePriority > highestRecentPriority {
			return Decision{
				Action:   ActionDeprioritize,
				TopicKey: topicKey,
				Reason:   "same-topic window is saturated, but candidate channel has higher priority",
				Signals:  signals,
			}, nil
		}
		return Decision{
			Action:     ActionDelay,
			TopicKey:   topicKey,
			DelayUntil: &delayUntil,
			Reason:     "same-topic window limit reached across channels",
			Signals:    signals,
		}, nil
	}

	if candidatePriority < highestRecentPriority {
		return Decision{
			Action:   ActionDeprioritize,
			TopicKey: topicKey,
			Reason:   "same topic already posted by a higher-priority channel in the current window",
			Signals:  signals,
		}, nil
	}

	return Decision{
		Action:   ActionAllow,
		TopicKey: topicKey,
		Reason:   "same-topic volume is within limits for this channel priority",
		Signals:  signals,
	}, nil
}

func (s *Service) channelPriority(channelID int64) int {
	return s.cfg.ChannelPriority[channelID]
}

func (s *Service) highestPriority(signals []PublishedSignal) int {
	highest := 0
	for _, signal := range signals {
		if priority := s.channelPriority(signal.Intent.ChannelID); priority > highest {
			highest = priority
		}
	}
	return highest
}

func signalKey(topic, clusterKey string) string {
	if normalized := normalizeKey(topic); normalized != "" {
		return normalized
	}
	return normalizeKey(clusterKey)
}

func normalizeKey(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}
