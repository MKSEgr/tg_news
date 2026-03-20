package analytics

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"ai-content-engine-starter/internal/domain"
)

const defaultPostedLimit = math.MaxInt32

// ChannelSummary contains deterministic per-channel analytics snapshot.
type ChannelSummary struct {
	ChannelID      int64
	PostedDrafts   int
	FeedbackDrafts int
	AvgScore       float64
	AvgScoreA      float64
	AvgScoreB      float64
	LastPostedAt   time.Time
}

// Service computes per-channel analytics from posted drafts and feedback.
type Service struct {
	drafts   domain.DraftRepository
	feedback domain.PerformanceFeedbackRepository
	limit    int
}

// New creates analytics service.
func New(drafts domain.DraftRepository, feedback domain.PerformanceFeedbackRepository) (*Service, error) {
	if drafts == nil {
		return nil, fmt.Errorf("draft repository is nil")
	}
	if feedback == nil {
		return nil, fmt.Errorf("performance feedback repository is nil")
	}
	return &Service{drafts: drafts, feedback: feedback, limit: defaultPostedLimit}, nil
}

// BuildByChannel returns deterministic per-channel analytics from posted drafts.
func (s *Service) BuildByChannel(ctx context.Context) ([]ChannelSummary, error) {
	if s == nil {
		return nil, fmt.Errorf("analytics service is nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if s.drafts == nil {
		return nil, fmt.Errorf("draft repository is nil")
	}
	if s.feedback == nil {
		return nil, fmt.Errorf("performance feedback repository is nil")
	}
	if s.limit <= 0 {
		return nil, fmt.Errorf("posted drafts limit must be greater than zero")
	}

	posted, err := s.drafts.ListByStatus(ctx, domain.DraftStatusPosted, s.limit)
	if err != nil {
		return nil, fmt.Errorf("list posted drafts: %w", err)
	}

	type accum struct {
		ChannelSummary
		totalScore float64
		totalA     float64
		totalB     float64
		countA     int
		countB     int
	}

	byChannel := make(map[int64]*accum)
	for _, draft := range posted {
		if draft.ID <= 0 || draft.ChannelID <= 0 {
			continue
		}
		item, ok := byChannel[draft.ChannelID]
		if !ok {
			item = &accum{ChannelSummary: ChannelSummary{ChannelID: draft.ChannelID}}
			byChannel[draft.ChannelID] = item
		}
		item.PostedDrafts++
		if draft.UpdatedAt.After(item.LastPostedAt) {
			item.LastPostedAt = draft.UpdatedAt
		}

		feedback, err := s.feedback.GetByDraftID(ctx, draft.ID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				continue
			}
			return nil, fmt.Errorf("get feedback for draft %d: %w", draft.ID, err)
		}
		if math.IsNaN(feedback.Score) || math.IsInf(feedback.Score, 0) {
			continue
		}
		item.FeedbackDrafts++
		item.totalScore += feedback.Score
		switch strings.ToUpper(strings.TrimSpace(feedback.Variant)) {
		case "B":
			item.countB++
			item.totalB += feedback.Score
		default:
			item.countA++
			item.totalA += feedback.Score
		}
	}

	channelIDs := make([]int64, 0, len(byChannel))
	for id := range byChannel {
		channelIDs = append(channelIDs, id)
	}
	sort.Slice(channelIDs, func(i, j int) bool { return channelIDs[i] < channelIDs[j] })

	out := make([]ChannelSummary, 0, len(channelIDs))
	for _, channelID := range channelIDs {
		item := byChannel[channelID]
		if item.FeedbackDrafts > 0 {
			item.AvgScore = item.totalScore / float64(item.FeedbackDrafts)
		}
		if item.countA > 0 {
			item.AvgScoreA = item.totalA / float64(item.countA)
		}
		if item.countB > 0 {
			item.AvgScoreB = item.totalB / float64(item.countB)
		}
		out = append(out, item.ChannelSummary)
	}

	return out, nil
}

// BuildByChannelMetrics adapts analytics summaries for source discovery integration.
func (s *Service) BuildByChannelMetrics(ctx context.Context) ([]ChannelSummary, error) {
	return s.BuildByChannel(ctx)
}
