package orchestration

import (
	"context"
	"fmt"
	"math"

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
	for _, channel := range channels {
		channelByID[channel.ID] = channel
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

			if j.scorer.Score(normalized) <= 0 {
				continue
			}

			targetIDs, err := j.router.Route(normalized, channels)
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

				draft, err := j.generator.GenerateDraft(ctx, normalized, channel)
				if err != nil {
					return fmt.Errorf("generate draft for item %d, channel %d: %w", normalized.ID, channelID, err)
				}
				result, err := j.guard.Check(draft)
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
