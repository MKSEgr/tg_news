package orchestration

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"ai-content-engine-starter/internal/domain"
	"ai-content-engine-starter/internal/editorial"
)

const (
	defaultRecentItemsLimit = 50
	defaultExistingLimit    = math.MaxInt32
	defaultRepostLimit      = 1
	defaultRepostThreshold  = 1.0
	defaultRepostCooldown   = 72 * time.Hour
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
	imageEnricher      sourceItemImageEnricher
	dedup              duplicateChecker
	scorer             trendScorer
	router             channelRouter
	generator          draftGenerator
	guard              draftGuard
	topicMemory        topicMemoryReader
	rules              contentRuleEvaluator
	feedback           feedbackReader
	planner            editorialPlanner
	assetGenerator     intentAssetGenerator
	clusterObserver    storyClusterObserver
	recentItemsLimit   int
	existingDraftLimit int
}

// AutoRepostJob promotes strong posted drafts for reposting.
type AutoRepostJob struct {
	drafts       domain.DraftRepository
	feedback     feedbackReader
	minScore     float64
	maxPerRun    int
	minPostedFor time.Duration
	listLimit    int
}

type collectorRunner interface {
	RunOnce(ctx context.Context) error
}

type sourceItemNormalizer interface {
	Normalize(item domain.SourceItem) (domain.SourceItem, error)
}

type sourceItemImageEnricher interface {
	Enrich(item domain.SourceItem) (domain.SourceItem, error)
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

type variantAwareGenerator interface {
	GenerateDraftVariants(ctx context.Context, item domain.SourceItem, channel domain.Channel, channelFeedback float64) ([]domain.Draft, error)
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

type editorialPlanner interface {
	PlanForSourceItem(ctx context.Context, item domain.SourceItem) ([]domain.PublishIntent, error)
}

type routedEditorialPlanner interface {
	PlanForSourceItemForChannels(ctx context.Context, item domain.SourceItem, channelIDs []int64) ([]domain.PublishIntent, error)
}

type intentAssetGenerator interface {
	GenerateFromIntent(ctx context.Context, intent domain.PublishIntent) (domain.ContentAsset, error)
}

type storyClusterObserver interface {
	ObserveSignal(ctx context.Context, item domain.SourceItem) (domain.StoryCluster, domain.ClusterEvent, error)
}

type clusterAwareRouter interface {
	RouteWithCluster(item domain.SourceItem, channels []domain.Channel, cluster domain.StoryCluster) ([]int64, error)
}

type clusterAwareGenerator interface {
	GenerateDraftWithCluster(ctx context.Context, item domain.SourceItem, channel domain.Channel, cluster domain.StoryCluster) (domain.Draft, error)
}

type clusterAwareFeedbackGenerator interface {
	GenerateDraftWithFeedbackAndCluster(ctx context.Context, item domain.SourceItem, channel domain.Channel, channelFeedback float64, cluster domain.StoryCluster) (domain.Draft, error)
}

type clusterAwareVariantGenerator interface {
	GenerateDraftVariantsWithCluster(ctx context.Context, item domain.SourceItem, channel domain.Channel, channelFeedback float64, cluster domain.StoryCluster) ([]domain.Draft, error)
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

// NewAutoRepostJob creates an auto-repost job with deterministic defaults.
func NewAutoRepostJob(drafts domain.DraftRepository, feedback feedbackReader) (*AutoRepostJob, error) {
	if drafts == nil {
		return nil, fmt.Errorf("draft repository is nil")
	}
	if feedback == nil {
		return nil, fmt.Errorf("feedback reader is nil")
	}
	return &AutoRepostJob{
		drafts:       drafts,
		feedback:     feedback,
		minScore:     defaultRepostThreshold,
		maxPerRun:    defaultRepostLimit,
		minPostedFor: defaultRepostCooldown,
		listLimit:    defaultExistingLimit,
	}, nil
}

// Run promotes top-performing posted drafts back to approved for scheduled reposting.
func (j *AutoRepostJob) Run(ctx context.Context) error {
	if j == nil {
		return fmt.Errorf("auto repost job is nil")
	}
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	if j.drafts == nil {
		return fmt.Errorf("draft repository is nil")
	}
	if j.feedback == nil {
		return fmt.Errorf("feedback reader is nil")
	}
	if j.maxPerRun <= 0 {
		return fmt.Errorf("max per run must be greater than zero")
	}
	if j.listLimit <= 0 {
		return fmt.Errorf("list limit must be greater than zero")
	}

	posted, err := j.drafts.ListByStatus(ctx, domain.DraftStatusPosted, j.listLimit)
	if err != nil {
		return fmt.Errorf("list posted drafts: %w", err)
	}

	now := time.Now().UTC()
	type candidate struct {
		draft domain.Draft
		score float64
	}
	candidates := make([]candidate, 0)
	for _, draft := range posted {
		if draft.ID <= 0 {
			continue
		}
		if j.minPostedFor > 0 && draft.UpdatedAt.IsZero() {
			continue
		}
		if j.minPostedFor > 0 && now.Sub(draft.UpdatedAt.UTC()) < j.minPostedFor {
			continue
		}
		feedback, err := j.feedback.GetByDraftID(ctx, draft.ID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				continue
			}
			return fmt.Errorf("get feedback for draft %d: %w", draft.ID, err)
		}
		if feedback.Score < j.minScore {
			continue
		}
		candidates = append(candidates, candidate{draft: draft, score: feedback.Score})
	}

	sort.Slice(candidates, func(i, k int) bool {
		if candidates[i].score == candidates[k].score {
			if candidates[i].draft.UpdatedAt.Equal(candidates[k].draft.UpdatedAt) {
				return candidates[i].draft.ID < candidates[k].draft.ID
			}
			return candidates[i].draft.UpdatedAt.Before(candidates[k].draft.UpdatedAt)
		}
		return candidates[i].score > candidates[k].score
	})

	limit := j.maxPerRun
	if limit > len(candidates) {
		limit = len(candidates)
	}
	for i := 0; i < limit; i++ {
		if err := j.drafts.UpdateStatus(ctx, candidates[i].draft.ID, domain.DraftStatusApproved); err != nil {
			return fmt.Errorf("promote draft %d for repost: %w", candidates[i].draft.ID, err)
		}
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

// WithImageEnricher sets optional image enrichment for normalized items.
func (j *PipelineJob) WithImageEnricher(enricher sourceItemImageEnricher) *PipelineJob {
	if j == nil {
		return nil
	}
	j.imageEnricher = enricher
	return j
}

// WithEditorialPlanner sets optional editorial planner integration.
func (j *PipelineJob) WithEditorialPlanner(planner editorialPlanner) *PipelineJob {
	if j == nil {
		return nil
	}
	j.planner = planner
	return j
}

// WithIntentAssetGenerator sets optional publish-intent to asset generation.
func (j *PipelineJob) WithIntentAssetGenerator(generator intentAssetGenerator) *PipelineJob {
	if j == nil {
		return nil
	}
	j.assetGenerator = generator
	return j
}

// WithStoryClusterObserver sets optional story cluster observation for routing/generation context.
func (j *PipelineJob) WithStoryClusterObserver(observer storyClusterObserver) *PipelineJob {
	if j == nil {
		return nil
	}
	j.clusterObserver = observer
	return j
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
			if j.imageEnricher != nil {
				normalized, err = j.imageEnricher.Enrich(normalized)
				if err != nil {
					return fmt.Errorf("enrich images for item %d: %w", item.ID, err)
				}
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
			if feedbackScorer, ok := j.scorer.(feedbackAwareScorer); ok {
				feedbackOnlyDelta := feedbackScorer.ScoreWithFeedback(normalized, feedbackByChannel) - baseScore
				score += feedbackOnlyDelta
			}
			memoryScorer, hasMemoryScorer := j.scorer.(memoryAwareScorer)
			if score <= 0 && !hasMemoryScorer {
				continue
			}

			cluster := domain.StoryCluster{}

			targetIDs, err := j.router.Route(normalized, channels)
			if memoryRouter, ok := j.router.(memoryAwareRouter); ok {
				targetIDs, err = memoryRouter.RouteWithMemory(normalized, channels, memoryByChannel)
			}
			if feedbackRouter, ok := j.router.(feedbackAwareRouter); ok {
				feedbackIDs, feedbackErr := feedbackRouter.RouteWithFeedback(normalized, channels, feedbackByChannel)
				if feedbackErr != nil {
					return fmt.Errorf("route item %d: %w", item.ID, feedbackErr)
				}
				targetIDs = mergeRankedRouteIDs(filterRouteIDs(feedbackIDs, targetIDs), targetIDs)
			}
			if err != nil {
				return fmt.Errorf("route item %d: %w", item.ID, err)
			}

			allowedTargetIDs, err := filterAllowedRouteIDs(targetIDs, channelByID, ctx, j.rules, normalized)
			if err != nil {
				return fmt.Errorf("filter allowed route ids for item %d: %w", item.ID, err)
			}
			if len(allowedTargetIDs) == 0 {
				continue
			}

			intents := []domain.PublishIntent(nil)
			if j.planner != nil {
				if routedPlanner, ok := j.planner.(routedEditorialPlanner); ok {
					intents, _ = routedPlanner.PlanForSourceItemForChannels(ctx, normalized, allowedTargetIDs)
				} else {
					intents, _ = j.planner.PlanForSourceItem(ctx, normalized)
				}
				if j.assetGenerator != nil {
					allowedIntentChannels := make(map[int64]struct{}, len(allowedTargetIDs))
					for _, channelID := range allowedTargetIDs {
						allowedIntentChannels[channelID] = struct{}{}
					}
					for _, intent := range intents {
						if _, ok := allowedIntentChannels[intent.ChannelID]; !ok {
							continue
						}
						_, _ = j.assetGenerator.GenerateFromIntent(ctx, intent)
					}
				}
			}

			if len(intents) == 0 && !hasDraftWorkForChannels(normalized.ID, allowedTargetIDs, existing, j.generator) {
				continue
			}

			if j.clusterObserver != nil {
				observedCluster, _, observeErr := j.clusterObserver.ObserveSignal(ctx, normalized)
				if observeErr == nil {
					cluster = observedCluster
				}
			}
			if clusterRouter, ok := j.router.(clusterAwareRouter); ok && cluster.ID > 0 {
				clusterIDs, clusterErr := clusterRouter.RouteWithCluster(normalized, channels, cluster)
				if clusterErr != nil {
					return fmt.Errorf("route item %d: %w", item.ID, clusterErr)
				}
				targetIDs = mergeRankedRouteIDs(targetIDs, clusterIDs)
				allowedTargetIDs, err = filterAllowedRouteIDs(targetIDs, channelByID, ctx, j.rules, normalized)
				if err != nil {
					return fmt.Errorf("filter allowed route ids for item %d: %w", item.ID, err)
				}
				if len(allowedTargetIDs) == 0 {
					continue
				}
			}

			for _, channelID := range allowedTargetIDs {
				key := draftKey{SourceItemID: normalized.ID, ChannelID: channelID, Variant: "A"}
				if _, isVariantGenerator := j.generator.(variantAwareGenerator); isVariantGenerator || implementsClusterAwareVariants(j.generator) {
					if _, hasA := existing[draftKey{SourceItemID: normalized.ID, ChannelID: channelID, Variant: "A"}]; hasA {
						if _, hasB := existing[draftKey{SourceItemID: normalized.ID, ChannelID: channelID, Variant: "B"}]; hasB {
							continue
						}
					}
				} else {
					if _, ok := existing[key]; ok {
						continue
					}
				}
				channel := channelByID[channelID]
				channelScore := score
				if hasMemoryScorer {
					channelScore = score + (memoryScorer.ScoreWithMemory(normalized, memoryByChannel[channelID]) - baseScore)
				}
				if channelScore <= 0 {
					continue
				}

				variants := make([]domain.Draft, 0, 2)
				if clusterVariantGenerator, ok := j.generator.(clusterAwareVariantGenerator); ok && cluster.ID > 0 {
					variants, err = clusterVariantGenerator.GenerateDraftVariantsWithCluster(ctx, normalized, channel, feedbackByChannel[channelID], cluster)
					if err != nil {
						return fmt.Errorf("generate draft variants for item %d, channel %d: %w", normalized.ID, channelID, err)
					}
				} else if variantGenerator, ok := j.generator.(variantAwareGenerator); ok {
					variants, err = variantGenerator.GenerateDraftVariants(ctx, normalized, channel, feedbackByChannel[channelID])
					if err != nil {
						return fmt.Errorf("generate draft variants for item %d, channel %d: %w", normalized.ID, channelID, err)
					}
				} else {
					var draft domain.Draft
					if clusterFeedbackGenerator, ok := j.generator.(clusterAwareFeedbackGenerator); ok && cluster.ID > 0 {
						draft, err = clusterFeedbackGenerator.GenerateDraftWithFeedbackAndCluster(ctx, normalized, channel, feedbackByChannel[channelID], cluster)
					} else if feedbackGenerator, ok := j.generator.(feedbackAwareGenerator); ok {
						draft, err = feedbackGenerator.GenerateDraftWithFeedback(ctx, normalized, channel, feedbackByChannel[channelID])
					} else if clusterGenerator, ok := j.generator.(clusterAwareGenerator); ok && cluster.ID > 0 {
						draft, err = clusterGenerator.GenerateDraftWithCluster(ctx, normalized, channel, cluster)
					} else {
						draft, err = j.generator.GenerateDraft(ctx, normalized, channel)
					}
					if err != nil {
						return fmt.Errorf("generate draft for item %d, channel %d: %w", normalized.ID, channelID, err)
					}
					variants = append(variants, draft)
				}

				for _, draft := range variants {
					draft.ImageURL = normalized.ImageURL
					variant := normalizeVariant(draft.Variant)
					variantKey := draftKey{SourceItemID: normalized.ID, ChannelID: channelID, Variant: variant}
					if _, ok := existing[variantKey]; ok {
						continue
					}
					draft.Variant = variant
					result, err := j.guard.Check(draft)
					if memoryGuard, ok := j.guard.(memoryAwareGuard); ok {
						result, err = memoryGuard.CheckWithMemory(draft, memoryByChannel[channelID])
					}
					if err != nil {
						return fmt.Errorf("editorial guard for item %d, channel %d, variant %s: %w", normalized.ID, channelID, variant, err)
					}
					if !result.Accepted {
						draft.Status = domain.DraftStatusRejected
					}
					if _, err := j.drafts.Create(ctx, draft); err != nil {
						return fmt.Errorf("store draft for item %d, channel %d, variant %s: %w", normalized.ID, channelID, variant, err)
					}
					existing[variantKey] = struct{}{}
				}
			}
		}
	}

	return nil
}

func mergeRankedRouteIDs(primary []int64, fallback []int64) []int64 {
	out := make([]int64, 0, len(primary)+len(fallback))
	seen := make(map[int64]struct{}, len(primary)+len(fallback))
	for _, id := range primary {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	for _, id := range fallback {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

func filterRouteIDs(ids []int64, allowed []int64) []int64 {
	if len(ids) == 0 || len(allowed) == 0 {
		return nil
	}
	allowedSet := make(map[int64]struct{}, len(allowed))
	for _, id := range allowed {
		if id > 0 {
			allowedSet[id] = struct{}{}
		}
	}
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if _, ok := allowedSet[id]; ok {
			out = append(out, id)
		}
	}
	return out
}

func hasDraftWorkForChannels(sourceItemID int64, channelIDs []int64, existing map[draftKey]struct{}, generator draftGenerator) bool {
	variantAware := false
	if _, ok := generator.(variantAwareGenerator); ok {
		variantAware = true
	}
	if !variantAware {
		variantAware = implementsClusterAwareVariants(generator)
	}
	for _, channelID := range channelIDs {
		keyA := draftKey{SourceItemID: sourceItemID, ChannelID: channelID, Variant: "A"}
		if !variantAware {
			if _, ok := existing[keyA]; !ok {
				return true
			}
			continue
		}
		_, hasA := existing[keyA]
		_, hasB := existing[draftKey{SourceItemID: sourceItemID, ChannelID: channelID, Variant: "B"}]
		if !hasA || !hasB {
			return true
		}
	}
	return false
}

func filterAllowedRouteIDs(targetIDs []int64, channelByID map[int64]domain.Channel, ctx context.Context, rules contentRuleEvaluator, item domain.SourceItem) ([]int64, error) {
	allowedTargetIDs := make([]int64, 0, len(targetIDs))
	for _, channelID := range targetIDs {
		channel, ok := channelByID[channelID]
		if !ok {
			continue
		}
		if rules != nil {
			allowed, err := rules.EvaluateAllowed(ctx, channelID, pipelineRuleText(item))
			if err != nil {
				return nil, fmt.Errorf("evaluate content rules for channel %d: %w", channelID, err)
			}
			if !allowed {
				continue
			}
		}
		allowedTargetIDs = append(allowedTargetIDs, channel.ID)
	}
	return allowedTargetIDs, nil
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

func normalizeVariant(variant string) string {
	value := strings.ToUpper(strings.TrimSpace(variant))
	if value != "B" {
		return "A"
	}
	return value
}

type draftKey struct {
	SourceItemID int64
	ChannelID    int64
	Variant      string
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
			keys[draftKey{SourceItemID: draft.SourceItemID, ChannelID: draft.ChannelID, Variant: normalizeVariant(draft.Variant)}] = struct{}{}
		}
	}
	return keys, nil
}

func implementsClusterAwareVariants(generator draftGenerator) bool {
	if generator == nil {
		return false
	}
	_, ok := generator.(clusterAwareVariantGenerator)
	return ok
}
