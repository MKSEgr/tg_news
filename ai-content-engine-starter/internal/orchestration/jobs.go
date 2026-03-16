package orchestration

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"ai-content-engine-starter/internal/domain"
	"ai-content-engine-starter/internal/editorial"
)

const (
	defaultRecentItemsLimit = 50
	defaultExistingLimit    = math.MaxInt32
)

// CollectorJob runs content collection.
type CollectorJob struct {
	collector collectorRunner
}

// PipelineJob orchestrates draft generation from recently collected source items.
type PipelineJob struct {
	sources            domain.SourceRepository
	items              domain.SourceItemRepository
	channels           domain.ChannelRepository
	drafts             domain.DraftRepository
	normalizer         sourceItemNormalizer
	dedup              duplicateChecker
	scorer             trendScorer
	router             channelRouter
	generator          draftGenerator
	guard              draftGuard
	topicMemory        topicMemoryReader
	rules              contentRuleEvaluator
	feedback           feedbackReader
	recentItemsLimit   int
	existingDraftLimit int
}

type collectorRunner interface {
	RunOnce(ctx context.Context) error
}

type sourceItemNormalizer interface {
	Normalize(item domain.SourceItem) (domain.SourceItem, error)
}

type duplicateChecker interface {
	IsDuplicate(ctx context.Context, item domain.SourceItem) (bool, error)
}

type trendScorer interface {
	Score(item domain.SourceItem) int
}

type channelRouter interface {
	Route(item domain.SourceItem, channels []domain.Channel) ([]int64, error)
}

type draftGenerator interface {
	GenerateDraft(ctx context.Context, item domain.SourceItem, channel domain.Channel) (domain.Draft, error)
}

type draftGuard interface {
	Check(draft domain.Draft) (editorial.Result, error)
}

type topicMemoryReader interface {
	TopTopics(ctx context.Context, channelID int64, limit int) ([]domain.TopicMemory, error)
}

type contentRuleEvaluator interface {
	EvaluateAllowed(ctx context.Context, channelID int64, text string) (bool, error)
}

type feedbackReader interface {
	GetByDraftID(ctx context.Context, draftID int64) (domain.PerformanceFeedback, error)
}

type feedbackAwareScorer interface {
	ScoreWithFeedback(item domain.SourceItem, feedbackByChannel map[int64]float64) int
}

type feedbackAwareRouter interface {
	RouteWithFeedback(item domain.SourceItem, channels []domain.Channel, feedbackByChannel map[int64]float64) ([]int64, error)
}

type feedbackAwareGenerator interface {
	GenerateDraftWithFeedback(ctx context.Context, item domain.SourceItem, channel domain.Channel, channelFeedback float64) (domain.Draft, error)
}

type memoryAwareScorer interface {
	ScoreWithMemory(item domain.SourceItem, memories []domain.TopicMemory) int
}

type memoryAwareRouter interface {
	RouteWithMemory(item domain.SourceItem, channels []domain.Channel, memoryByChannel map[int64][]domain.TopicMemory) ([]int64, error)
}

type memoryAwareGuard interface {
	CheckWithMemory(draft domain.Draft, memories []domain.TopicMemory) (editorial.Result, error)
}

// NewCollectorJob creates a collection job.
func NewCollectorJob(collector collectorRunner) (*CollectorJob, error) {
	if collector == nil {
		return nil, fmt.Errorf("collector is nil")
	}
	return &CollectorJob{collector: collector}, nil
}

// Run executes collection once.
func (j *CollectorJob) Run(ctx context.Context) error {
	if j == nil {
		return fmt.Errorf("collector job is nil")
	}
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	if j.collector == nil {
		return fmt.Errorf("collector is nil")
	}
	if err := j.collector.RunOnce(ctx); err != nil {
		return fmt.Errorf("run collector: %w", err)
	}
	return nil
}

// NewPipelineJob creates a draft orchestration job.
func NewPipelineJob(
	sources domain.SourceRepository,
	items domain.SourceItemRepository,
	channels domain.ChannelRepository,
	drafts domain.DraftRepository,
	normalizer sourceItemNormalizer,
	dedup duplicateChecker,
	scorer trendScorer,
	router channelRouter,
	generator draftGenerator,
	guard draftGuard,
	topicMemory topicMemoryReader,
	rules contentRuleEvaluator,
	feedback feedbackReader,
) (*PipelineJob, error) {
	if sources == nil {
		return nil, fmt.Errorf("source repository is nil")
	}
	if items == nil {
		return nil, fmt.Errorf("source item repository is nil")
	}
	if channels == nil {
		return nil, fmt.Errorf("channel repository is nil")
	}
	if drafts == nil {
		return nil, fmt.Errorf("draft repository is nil")
	}
	if normalizer == nil {
		return nil, fmt.Errorf("normalizer is nil")
	}
	if dedup == nil {
		return nil, fmt.Errorf("dedup service is nil")
	}
	if scorer == nil {
		return nil, fmt.Errorf("scorer service is nil")
	}
	if router == nil {
		return nil, fmt.Errorf("router service is nil")
	}
	if generator == nil {
		return nil, fmt.Errorf("generator service is nil")
	}
	if guard == nil {
		return nil, fmt.Errorf("editorial guard is nil")
	}

	return &PipelineJob{
		sources:            sources,
		items:              items,
		channels:           channels,
		drafts:             drafts,
		normalizer:         normalizer,
		dedup:              dedup,
		scorer:             scorer,
		router:             router,
		generator:          generator,
		guard:              guard,
		topicMemory:        topicMemory,
		rules:              rules,
		feedback:           feedback,
		recentItemsLimit:   defaultRecentItemsLimit,
		existingDraftLimit: defaultExistingLimit,
	}, nil
}

// Run executes one orchestration cycle.
func (j *PipelineJob) Run(ctx context.Context) error {
	if j == nil {
		return fmt.Errorf("pipeline job is nil")
	}
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}

	sources, err := j.sources.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("list enabled sources: %w", err)
	}
	channels, err := j.channels.List(ctx)
	if err != nil {
		return fmt.Errorf("list channels: %w", err)
	}
	channelByID := make(map[int64]domain.Channel, len(channels))
	memoryByChannel := make(map[int64][]domain.TopicMemory, len(channels))
	for _, channel := range channels {
		channelByID[channel.ID] = channel
		if j.topicMemory == nil || channel.ID <= 0 {
			continue
		}
		topics, err := j.topicMemory.TopTopics(ctx, channel.ID, 10)
		if err != nil {
			return fmt.Errorf("top topics for channel %d: %w", channel.ID, err)
		}
		memoryByChannel[channel.ID] = topics
	}

	feedbackByChannel, err := j.channelFeedbackAverages(ctx)
	if err != nil {
		return err
	}

	existing, err := j.existingDraftKeys(ctx)
	if err != nil {
		return err
	}

	for _, source := range sources {
		items, err := j.items.ListBySourceID(ctx, source.ID, j.recentItemsLimit)
		if err != nil {
			return fmt.Errorf("list items for source %d: %w", source.ID, err)
		}

		for _, item := range items {
			normalized, err := j.normalizer.Normalize(item)
			if err != nil {
				continue
			}

			duplicate, err := j.dedup.IsDuplicate(ctx, normalized)
			if err != nil {
				return fmt.Errorf("dedup item %d: %w", item.ID, err)
			}
			if duplicate {
				continue
			}

			baseScore := j.scorer.Score(normalized)
			score := baseScore
			if memoryScorer, ok := j.scorer.(memoryAwareScorer); ok {
				score = memoryScorer.ScoreWithMemory(normalized, flattenMemory(memoryByChannel))
			}
			if feedbackScorer, ok := j.scorer.(feedbackAwareScorer); ok {
				feedbackOnlyDelta := feedbackScorer.ScoreWithFeedback(normalized, feedbackByChannel) - baseScore
				score += feedbackOnlyDelta
			}
			if score <= 0 {
				continue
			}

			targetIDs, err := j.router.Route(normalized, channels)
			if memoryRouter, ok := j.router.(memoryAwareRouter); ok {
				targetIDs, err = memoryRouter.RouteWithMemory(normalized, channels, memoryByChannel)
			}
			if feedbackRouter, ok := j.router.(feedbackAwareRouter); ok {
				targetIDs, err = feedbackRouter.RouteWithFeedback(normalized, channels, feedbackByChannel)
			}
			if err != nil {
				return fmt.Errorf("route item %d: %w", item.ID, err)
			}

			for _, channelID := range targetIDs {
				key := draftKey{SourceItemID: normalized.ID, ChannelID: channelID}
				if _, ok := existing[key]; ok {
					continue
				}
				channel, ok := channelByID[channelID]
				if !ok {
					continue
				}
				if j.rules != nil {
					allowed, err := j.rules.EvaluateAllowed(ctx, channelID, pipelineRuleText(normalized))
					if err != nil {
						return fmt.Errorf("evaluate content rules for channel %d: %w", channelID, err)
					}
					if !allowed {
						continue
					}
				}

				draft, err := j.generator.GenerateDraft(ctx, normalized, channel)
				if feedbackGenerator, ok := j.generator.(feedbackAwareGenerator); ok {
					draft, err = feedbackGenerator.GenerateDraftWithFeedback(ctx, normalized, channel, feedbackByChannel[channelID])
				}
				if err != nil {
					return fmt.Errorf("generate draft for item %d, channel %d: %w", normalized.ID, channelID, err)
				}
				result, err := j.guard.Check(draft)
				if memoryGuard, ok := j.guard.(memoryAwareGuard); ok {
					result, err = memoryGuard.CheckWithMemory(draft, memoryByChannel[channelID])
				}
				if err != nil {
					return fmt.Errorf("editorial guard for item %d, channel %d: %w", normalized.ID, channelID, err)
				}
				if !result.Accepted {
					draft.Status = domain.DraftStatusRejected
				}
				if _, err := j.drafts.Create(ctx, draft); err != nil {
					return fmt.Errorf("store draft for item %d, channel %d: %w", normalized.ID, channelID, err)
				}
				existing[key] = struct{}{}
			}
		}
	}

	return nil
}

func pipelineRuleText(item domain.SourceItem) string {
	text := strings.TrimSpace(item.Title)
	if item.Body != nil {
		body := strings.TrimSpace(*item.Body)
		if body != "" {
			if text == "" {
				text = body
			} else {
				text += " " + body
			}
		}
	}
	return text
}

func flattenMemory(memoryByChannel map[int64][]domain.TopicMemory) []domain.TopicMemory {
	if len(memoryByChannel) == 0 {
		return nil
	}

	channelIDs := make([]int64, 0, len(memoryByChannel))
	for channelID := range memoryByChannel {
		channelIDs = append(channelIDs, channelID)
	}
	sort.Slice(channelIDs, func(i, j int) bool { return channelIDs[i] < channelIDs[j] })

	out := make([]domain.TopicMemory, 0)
	for _, channelID := range channelIDs {
		topics := append([]domain.TopicMemory(nil), memoryByChannel[channelID]...)
		sort.Slice(topics, func(i, j int) bool {
			left := strings.TrimSpace(strings.ToLower(topics[i].Topic))
			right := strings.TrimSpace(strings.ToLower(topics[j].Topic))
			if left == right {
				return topics[i].MentionCount > topics[j].MentionCount
			}
			return left < right
		})
		out = append(out, topics...)
	}
	return out
}

func (j *PipelineJob) channelFeedbackAverages(ctx context.Context) (map[int64]float64, error) {
	out := make(map[int64]float64)
	if j.feedback == nil {
		return out, nil
	}

	posted, err := j.drafts.ListByStatus(ctx, domain.DraftStatusPosted, j.existingDraftLimit)
	if err != nil {
		return nil, fmt.Errorf("list posted drafts: %w", err)
	}
	totals := make(map[int64]float64)
	counts := make(map[int64]int)
	for _, draft := range posted {
		if draft.ID <= 0 || draft.ChannelID <= 0 {
			continue
		}
		feedback, err := j.feedback.GetByDraftID(ctx, draft.ID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				continue
			}
			return nil, fmt.Errorf("get feedback for draft %d: %w", draft.ID, err)
		}
		totals[draft.ChannelID] += feedback.Score
		counts[draft.ChannelID]++
	}
	for channelID, total := range totals {
		count := counts[channelID]
		if count > 0 {
			out[channelID] = total / float64(count)
		}
	}
	return out, nil
}

type draftKey struct {
	SourceItemID int64
	ChannelID    int64
}

func (j *PipelineJob) existingDraftKeys(ctx context.Context) (map[draftKey]struct{}, error) {
	statuses := []domain.DraftStatus{domain.DraftStatusPending, domain.DraftStatusApproved, domain.DraftStatusRejected, domain.DraftStatusPosted}
	keys := make(map[draftKey]struct{})
	for _, status := range statuses {
		drafts, err := j.drafts.ListByStatus(ctx, status, j.existingDraftLimit)
		if err != nil {
			return nil, fmt.Errorf("list %s drafts: %w", status, err)
		}
		for _, draft := range drafts {
			if draft.SourceItemID <= 0 || draft.ChannelID <= 0 {
				continue
			}
			keys[draftKey{SourceItemID: draft.SourceItemID, ChannelID: draft.ChannelID}] = struct{}{}
		}
	}
	return keys, nil
}
