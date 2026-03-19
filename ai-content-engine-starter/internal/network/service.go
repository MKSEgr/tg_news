package network

import (
	"context"
	"fmt"
	"sort"
	"time"

	"ai-content-engine-starter/internal/domain"
)

const defaultRelationshipLimit = 32

// ScheduleEntry is the relationship-aware execution recommendation for one channel.
type ScheduleEntry struct {
	ChannelID int64
	Score     float64
	Delay     time.Duration
	Reasons   []string
}

// Config controls relationship scoring and deterministic delay spacing.
type Config struct {
	RelationshipLimit int
	BaseDelay         time.Duration
	TypeWeights       map[domain.ChannelRelationshipType]float64
}

// Service produces simple network-aware scheduling recommendations from channel links.
type Service struct {
	relationships domain.ChannelRelationshipRepository
	cfg           Config
}

// New creates a network scheduling service.
func New(relationships domain.ChannelRelationshipRepository, cfg Config) (*Service, error) {
	if relationships == nil {
		return nil, fmt.Errorf("channel relationship repository is nil")
	}
	if cfg.RelationshipLimit == 0 {
		cfg.RelationshipLimit = defaultRelationshipLimit
	}
	if cfg.RelationshipLimit < 0 {
		return nil, fmt.Errorf("relationship limit must be greater than or equal to zero")
	}
	if cfg.BaseDelay < 0 {
		return nil, fmt.Errorf("base delay must be greater than or equal to zero")
	}
	if cfg.TypeWeights == nil {
		cfg.TypeWeights = defaultTypeWeights()
	}
	return &Service{relationships: relationships, cfg: cfg}, nil
}

// Build ranks channels by relationship score and assigns deterministic delays.
func (s *Service) Build(ctx context.Context, channels []domain.Channel) ([]ScheduleEntry, error) {
	if s == nil {
		return nil, fmt.Errorf("network scheduler service is nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	entries := make([]ScheduleEntry, 0, len(channels))
	for _, channel := range channels {
		if channel.ID <= 0 {
			return nil, fmt.Errorf("channel id must be greater than zero")
		}
		rels, err := s.relationships.ListByChannel(ctx, channel.ID, s.cfg.RelationshipLimit)
		if err != nil {
			return nil, fmt.Errorf("list relationships for channel %d: %w", channel.ID, err)
		}
		entry := ScheduleEntry{ChannelID: channel.ID}
		for _, rel := range rels {
			weight := s.cfg.TypeWeights[rel.RelationshipType]
			entry.Score += rel.Strength * weight
			entry.Reasons = append(entry.Reasons, fmt.Sprintf("%s:%d:%.2f", rel.RelationshipType, rel.RelatedChannelID, rel.Strength))
		}
		if len(entry.Reasons) == 0 {
			entry.Reasons = []string{"no_relationships"}
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Score == entries[j].Score {
			return entries[i].ChannelID < entries[j].ChannelID
		}
		return entries[i].Score > entries[j].Score
	})
	for i := range entries {
		entries[i].Delay = time.Duration(i) * s.cfg.BaseDelay
	}
	return entries, nil
}

func defaultTypeWeights() map[domain.ChannelRelationshipType]float64 {
	return map[domain.ChannelRelationshipType]float64{
		domain.ChannelRelationshipTypeParent:          3,
		domain.ChannelRelationshipTypePromotionTarget: 2,
		domain.ChannelRelationshipTypeSibling:         1,
	}
}
