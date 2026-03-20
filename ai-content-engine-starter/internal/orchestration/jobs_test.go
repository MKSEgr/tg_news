package orchestration

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-content-engine-starter/internal/domain"
	"ai-content-engine-starter/internal/editorial"
)

type collectorStub struct{ err error }

func (s *collectorStub) RunOnce(context.Context) error { return s.err }

type sourceRepoStub struct {
	sources []domain.Source
	err     error
}

func (s *sourceRepoStub) Create(context.Context, domain.Source) (domain.Source, error) {
	return domain.Source{}, nil
}
func (s *sourceRepoStub) GetByID(context.Context, int64) (domain.Source, error) {
	return domain.Source{}, nil
}
func (s *sourceRepoStub) List(context.Context) ([]domain.Source, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.sources, nil
}
func (s *sourceRepoStub) ListEnabled(context.Context) ([]domain.Source, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.sources, nil
}

type sourceItemRepoStub struct {
	itemsBySource map[int64][]domain.SourceItem
	err           error
}

func (s *sourceItemRepoStub) Create(context.Context, domain.SourceItem) (domain.SourceItem, error) {
	return domain.SourceItem{}, nil
}
func (s *sourceItemRepoStub) GetByID(context.Context, int64) (domain.SourceItem, error) {
	return domain.SourceItem{}, nil
}
func (s *sourceItemRepoStub) ListBySourceID(_ context.Context, sourceID int64, _ int) ([]domain.SourceItem, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.itemsBySource[sourceID], nil
}
func (s *sourceItemRepoStub) ListRecent(_ context.Context, _ int) ([]domain.SourceItem, error) {
	if s.err != nil {
		return nil, s.err
	}
	items := make([]domain.SourceItem, 0)
	for _, sourceItems := range s.itemsBySource {
		items = append(items, sourceItems...)
	}
	return items, nil
}

type channelRepoStub struct {
	channels []domain.Channel
	err      error
}

func (s *channelRepoStub) Create(context.Context, domain.Channel) (domain.Channel, error) {
	return domain.Channel{}, nil
}
func (s *channelRepoStub) GetByID(context.Context, int64) (domain.Channel, error) {
	return domain.Channel{}, nil
}
func (s *channelRepoStub) List(context.Context) ([]domain.Channel, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.channels, nil
}

type draftRepoStub struct {
	created   []domain.Draft
	byStatus  map[domain.DraftStatus][]domain.Draft
	createErr error
	listErr   error
	updateErr error
	updated   map[int64]domain.DraftStatus
}

func (s *draftRepoStub) Create(_ context.Context, d domain.Draft) (domain.Draft, error) {
	if s.createErr != nil {
		return domain.Draft{}, s.createErr
	}
	s.created = append(s.created, d)
	return d, nil
}
func (s *draftRepoStub) GetByID(context.Context, int64) (domain.Draft, error) {
	return domain.Draft{}, nil
}
func (s *draftRepoStub) ListByStatus(_ context.Context, status domain.DraftStatus, _ int) ([]domain.Draft, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.byStatus[status], nil
}
func (s *draftRepoStub) UpdateStatus(_ context.Context, id int64, status domain.DraftStatus) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	if s.updated == nil {
		s.updated = make(map[int64]domain.DraftStatus)
	}
	s.updated[id] = status
	return nil
}

func (s *draftRepoStub) UpdateStatusIfCurrent(_ context.Context, id int64, current domain.DraftStatus, next domain.DraftStatus) (bool, error) {
	if s.updateErr != nil {
		return false, s.updateErr
	}
	if s.updated == nil {
		s.updated = make(map[int64]domain.DraftStatus)
	}
	if status, ok := s.updated[id]; ok && status != current {
		return false, nil
	}
	s.updated[id] = next
	return true, nil
}

type normalizerStub struct{ item domain.SourceItem }

func (n *normalizerStub) Normalize(domain.SourceItem) (domain.SourceItem, error) { return n.item, nil }

type imageEnricherStub struct {
	item domain.SourceItem
	err  error
}

func (e *imageEnricherStub) Enrich(domain.SourceItem) (domain.SourceItem, error) {
	if e.err != nil {
		return domain.SourceItem{}, e.err
	}
	return e.item, nil
}

type dedupStub struct{ duplicate bool }

func (d *dedupStub) IsDuplicate(context.Context, domain.SourceItem) (bool, error) {
	return d.duplicate, nil
}

type scorerStub struct{ score int }

func (s *scorerStub) Score(domain.SourceItem) int { return s.score }

type adaptiveScorerStub struct {
	score int
	err   error
	calls int
}

func (s *adaptiveScorerStub) Score(domain.SourceItem) int { return 5 }
func (s *adaptiveScorerStub) ScoreAdaptive(context.Context, domain.SourceItem, int64, string) (int, error) {
	s.calls++
	if s.err != nil {
		return 0, s.err
	}
	return s.score, nil
}

type adaptiveFeedbackScorerStub struct {
	base     int
	feedback int
	adaptive int
}

func (s *adaptiveFeedbackScorerStub) Score(domain.SourceItem) int { return s.base }

func (s *adaptiveFeedbackScorerStub) ScoreWithFeedback(domain.SourceItem, map[int64]float64) int {
	return s.feedback
}

func (s *adaptiveFeedbackScorerStub) ScoreAdaptive(context.Context, domain.SourceItem, int64, string) (int, error) {
	return s.adaptive, nil
}

type advancedScorerStub struct {
	base     int
	memory   int
	feedback int
}

func (s *advancedScorerStub) Score(domain.SourceItem) int { return s.base }

func (s *advancedScorerStub) ScoreWithMemory(domain.SourceItem, []domain.TopicMemory) int {
	return s.base + s.memory
}

func (s *advancedScorerStub) ScoreWithFeedback(domain.SourceItem, map[int64]float64) int {
	return s.base + s.feedback
}

type channelMemoryScorerStub struct{ base int }

func (s *channelMemoryScorerStub) Score(domain.SourceItem) int { return s.base }

func (s *channelMemoryScorerStub) ScoreWithMemory(_ domain.SourceItem, memories []domain.TopicMemory) int {
	return s.base + len(memories)
}

type routerStub struct{ ids []int64 }

func (r *routerStub) Route(domain.SourceItem, []domain.Channel) ([]int64, error) { return r.ids, nil }

type advancedRouterStub struct {
	idsMemory   []int64
	idsFeedback []int64
}

func (r *advancedRouterStub) Route(domain.SourceItem, []domain.Channel) ([]int64, error) {
	return r.idsMemory, nil
}

func (r *advancedRouterStub) RouteWithMemory(domain.SourceItem, []domain.Channel, map[int64][]domain.TopicMemory) ([]int64, error) {
	return r.idsMemory, nil
}

func (r *advancedRouterStub) RouteWithFeedback(domain.SourceItem, []domain.Channel, map[int64]float64) ([]int64, error) {
	return r.idsFeedback, nil
}

type generatorStub struct {
	draft domain.Draft
	err   error
}

func (g *generatorStub) GenerateDraft(context.Context, domain.SourceItem, domain.Channel) (domain.Draft, error) {
	if g.err != nil {
		return domain.Draft{}, g.err
	}
	return g.draft, nil
}

type variantGeneratorStub struct {
	variants []domain.Draft
	calls    int
}

func (g *variantGeneratorStub) GenerateDraft(context.Context, domain.SourceItem, domain.Channel) (domain.Draft, error) {
	return domain.Draft{}, errors.New("base path should not be used")
}

func (g *variantGeneratorStub) GenerateDraftVariants(context.Context, domain.SourceItem, domain.Channel, float64) ([]domain.Draft, error) {
	g.calls++
	return g.variants, nil
}

type feedbackGeneratorStub struct {
	draftBase     domain.Draft
	draftFeedback domain.Draft
	baseCalls     int
	feedbackCalls int
}

func (g *feedbackGeneratorStub) GenerateDraft(context.Context, domain.SourceItem, domain.Channel) (domain.Draft, error) {
	g.baseCalls++
	return g.draftBase, nil
}

func (g *feedbackGeneratorStub) GenerateDraftWithFeedback(context.Context, domain.SourceItem, domain.Channel, float64) (domain.Draft, error) {
	g.feedbackCalls++
	return g.draftFeedback, nil
}

type clusterObserverStub struct {
	cluster domain.StoryCluster
	err     error
	calls   int
}

func (s *clusterObserverStub) ObserveSignal(context.Context, domain.SourceItem) (domain.StoryCluster, domain.ClusterEvent, error) {
	s.calls++
	if s.err != nil {
		return domain.StoryCluster{}, domain.ClusterEvent{}, s.err
	}
	return s.cluster, domain.ClusterEvent{StoryClusterID: s.cluster.ID}, nil
}

type clusterAwareRouterStub struct {
	baseIDs    []int64
	clusterIDs []int64
	calls      int
}

func (r *clusterAwareRouterStub) Route(domain.SourceItem, []domain.Channel) ([]int64, error) {
	return r.baseIDs, nil
}
func (r *clusterAwareRouterStub) RouteWithCluster(domain.SourceItem, []domain.Channel, domain.StoryCluster) ([]int64, error) {
	r.calls++
	return r.clusterIDs, nil
}

type clusterAwareGeneratorStub struct {
	draft domain.Draft
	calls int
}

func (g *clusterAwareGeneratorStub) GenerateDraft(context.Context, domain.SourceItem, domain.Channel) (domain.Draft, error) {
	return domain.Draft{}, errors.New("base path should not be used")
}
func (g *clusterAwareGeneratorStub) GenerateDraftWithCluster(_ context.Context, item domain.SourceItem, channel domain.Channel, _ domain.StoryCluster) (domain.Draft, error) {
	g.calls++
	draft := g.draft
	draft.SourceItemID = item.ID
	draft.ChannelID = channel.ID
	return draft, nil
}

type clusterAwareVariantGeneratorStub struct {
	calls int
}

func (g *clusterAwareVariantGeneratorStub) GenerateDraft(context.Context, domain.SourceItem, domain.Channel) (domain.Draft, error) {
	return domain.Draft{}, errors.New("base path should not be used")
}

func (g *clusterAwareVariantGeneratorStub) GenerateDraftVariantsWithCluster(_ context.Context, item domain.SourceItem, channel domain.Channel, _ float64, _ domain.StoryCluster) ([]domain.Draft, error) {
	g.calls++
	return []domain.Draft{
		{SourceItemID: item.ID, ChannelID: channel.ID, Variant: "A", Title: "a", Body: "body", Status: domain.DraftStatusPending},
		{SourceItemID: item.ID, ChannelID: channel.ID, Variant: "B", Title: "b", Body: "body", Status: domain.DraftStatusPending},
	}, nil
}

type guardStub struct{ result editorial.Result }

func (g *guardStub) Check(domain.Draft) (editorial.Result, error) { return g.result, nil }

type plannerStub struct {
	err   error
	calls int
}

func (p *plannerStub) PlanForSourceItem(context.Context, domain.SourceItem) ([]domain.PublishIntent, error) {
	p.calls++
	if p.err != nil {
		return nil, p.err
	}
	return nil, nil
}

type routedPlannerStub struct {
	channelIDs []int64
	calls      int
	result     []domain.PublishIntent
}

func (p *routedPlannerStub) PlanForSourceItem(context.Context, domain.SourceItem) ([]domain.PublishIntent, error) {
	p.calls++
	return p.result, nil
}

func (p *routedPlannerStub) PlanForSourceItemForChannels(_ context.Context, _ domain.SourceItem, channelIDs []int64) ([]domain.PublishIntent, error) {
	p.calls++
	p.channelIDs = append([]int64(nil), channelIDs...)
	return p.result, nil
}

type assetGeneratorStub struct {
	calls   int
	intents []domain.PublishIntent
	err     error
}

func (g *assetGeneratorStub) GenerateFromIntent(_ context.Context, intent domain.PublishIntent) (domain.ContentAsset, error) {
	g.calls++
	g.intents = append(g.intents, intent)
	if g.err != nil {
		return domain.ContentAsset{}, g.err
	}
	return domain.ContentAsset{RawItemID: intent.RawItemID, ChannelID: intent.ChannelID, AssetType: intent.Format}, nil
}

type rulesStub struct {
	allowByChannel map[int64]bool
	err            error
}

func (r *rulesStub) EvaluateAllowed(_ context.Context, channelID int64, _ string) (bool, error) {
	if r.err != nil {
		return false, r.err
	}
	if r.allowByChannel == nil {
		return true, nil
	}
	allowed, ok := r.allowByChannel[channelID]
	if !ok {
		return true, nil
	}
	return allowed, nil
}

type feedbackRepoStub struct {
	byDraft map[int64]domain.PerformanceFeedback
	err     error
}

func (r *feedbackRepoStub) GetByDraftID(_ context.Context, draftID int64) (domain.PerformanceFeedback, error) {
	if r.err != nil {
		return domain.PerformanceFeedback{}, r.err
	}
	feedback, ok := r.byDraft[draftID]
	if !ok {
		return domain.PerformanceFeedback{}, domain.ErrNotFound
	}
	return feedback, nil
}

type topicMemoryStub struct {
	topicsByChannel map[int64][]domain.TopicMemory
	err             error
}

func (s *topicMemoryStub) TopTopics(_ context.Context, channelID int64, _ int) ([]domain.TopicMemory, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.topicsByChannel[channelID], nil
}

func TestCollectorJobRun(t *testing.T) {
	job, err := NewCollectorJob(&collectorStub{})
	if err != nil {
		t.Fatalf("NewCollectorJob() error = %v", err)
	}
	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestPipelineJobRunSkipsClusterObservationWhenNoDraftWorkOrIntentsRemain(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}
	observer := &clusterObserverStub{cluster: domain.StoryCluster{ID: 77, Title: "Tool launch"}}
	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{
		domain.DraftStatusPending: {{SourceItemID: 11, ChannelID: 7, Variant: "A", Status: domain.DraftStatusPending}},
	}}

	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	job.WithStoryClusterObserver(observer)

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if observer.calls != 0 {
		t.Fatalf("observer calls = %d, want 0", observer.calls)
	}
}

func TestPipelineJobRunPassesAllowedChannelsToRoutedPlanner(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channels := []domain.Channel{{ID: 1, Slug: "ai-news", Name: "AI News"}, {ID: 2, Slug: "ai-tools", Name: "AI Tools"}}
	planner := &routedPlannerStub{result: []domain.PublishIntent{{RawItemID: 11, ChannelID: 2, Format: "text", Status: domain.PublishIntentStatusPlanned}}}
	assetGen := &assetGeneratorStub{}

	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: channels},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{1, 2}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 2, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		&rulesStub{allowByChannel: map[int64]bool{1: false, 2: true}},
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	job = job.WithEditorialPlanner(planner).WithIntentAssetGenerator(assetGen)

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if planner.calls != 1 {
		t.Fatalf("planner calls = %d, want 1", planner.calls)
	}
	if len(planner.channelIDs) != 1 || planner.channelIDs[0] != 2 {
		t.Fatalf("planner channelIDs = %v, want [2]", planner.channelIDs)
	}
	if assetGen.calls != 1 || len(assetGen.intents) != 1 || assetGen.intents[0].ChannelID != 2 {
		t.Fatalf("asset generation intents = %+v", assetGen.intents)
	}
}

func TestPipelineJobRunClusterHintsCanAddAdditionalChannels(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "Weekly update"}
	channels := []domain.Channel{{ID: 1, Slug: "ai-news", Name: "AI News"}, {ID: 2, Slug: "ai-tools", Name: "AI Tools"}}
	observer := &clusterObserverStub{cluster: domain.StoryCluster{ID: 77, Title: "Tool launch"}}
	router := &clusterAwareRouterStub{baseIDs: []int64{1}, clusterIDs: []int64{1, 2}}
	gen := &clusterAwareGeneratorStub{draft: domain.Draft{Title: "t", Body: "b", Status: domain.DraftStatusPending}}
	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}

	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: channels},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		router,
		gen,
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	job.WithStoryClusterObserver(observer)

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 2 {
		t.Fatalf("created drafts = %d, want 2", len(drafts.created))
	}
}

func TestPipelineJobRunSkipsClusterAwareVariantGenerationWhenVariantsExist(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "Weekly update"}
	channel := domain.Channel{ID: 1, Slug: "ai-news", Name: "AI News"}
	observer := &clusterObserverStub{cluster: domain.StoryCluster{ID: 77, Title: "Tool launch"}}
	gen := &clusterAwareVariantGeneratorStub{}
	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{
		domain.DraftStatusPending: {
			{SourceItemID: 11, ChannelID: 1, Variant: "A", Status: domain.DraftStatusPending},
			{SourceItemID: 11, ChannelID: 1, Variant: "B", Status: domain.DraftStatusPending},
		},
	}}

	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{1}},
		gen,
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	job.WithStoryClusterObserver(observer)

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if gen.calls != 0 {
		t.Fatalf("generator calls = %d, want 0", gen.calls)
	}
}

func TestPipelineJobRunUsesOptionalClusterContextForRoutingAndGeneration(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "Weekly update"}
	channels := []domain.Channel{{ID: 1, Slug: "ai-news", Name: "AI News"}}
	observer := &clusterObserverStub{cluster: domain.StoryCluster{ID: 77, Title: "Tool launch"}}
	router := &clusterAwareRouterStub{baseIDs: []int64{1}, clusterIDs: []int64{1}}
	gen := &clusterAwareGeneratorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 1, Title: "t", Body: "b", Status: domain.DraftStatusPending}}
	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}

	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: channels},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		router,
		gen,
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	job.WithStoryClusterObserver(observer)

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if observer.calls != 1 {
		t.Fatalf("observer calls = %d, want 1", observer.calls)
	}
	if router.calls != 1 {
		t.Fatalf("router calls = %d, want 1", router.calls)
	}
	if gen.calls != 1 {
		t.Fatalf("generator calls = %d, want 1", gen.calls)
	}
}

func TestPipelineJobRunIgnoresOptionalClusterObserverError(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}
	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}

	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	job.WithStoryClusterObserver(&clusterObserverStub{err: errors.New("cluster unavailable")})

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 1 {
		t.Fatalf("created drafts = %d, want 1", len(drafts.created))
	}
}

func TestPipelineJobRunReportsOptionalClusterObserverErrorViaHook(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}
	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}

	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	observerErr := errors.New("cluster unavailable")
	hookCalls := 0
	job.WithStoryClusterObserver(&clusterObserverStub{err: observerErr})
	job.WithStoryClusterErrorHook(func(gotItem domain.SourceItem, gotErr error) {
		hookCalls++
		if gotItem.ID != item.ID {
			t.Fatalf("hook item id = %d, want %d", gotItem.ID, item.ID)
		}
		if !errors.Is(gotErr, observerErr) {
			t.Fatalf("hook error = %v, want %v", gotErr, observerErr)
		}
	})

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if hookCalls != 1 {
		t.Fatalf("hook calls = %d, want 1", hookCalls)
	}
	if len(drafts.created) != 1 {
		t.Fatalf("created drafts = %d, want 1", len(drafts.created))
	}
}

func TestPipelineJobRunUsesAdaptiveScorePerChannel(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}
	scorer := &adaptiveScorerStub{score: 12}
	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		scorer,
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if scorer.calls != 1 {
		t.Fatalf("adaptive scorer calls = %d, want 1", scorer.calls)
	}
	if len(drafts.created) != 1 {
		t.Fatalf("created drafts = %d, want 1", len(drafts.created))
	}
}

func TestPipelineJobRunAddsAdaptiveDeltaToPriorScorePath(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}
	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}

	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&adaptiveFeedbackScorerStub{base: 5, feedback: 15, adaptive: 0},
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 1 {
		t.Fatalf("created drafts = %d, want 1", len(drafts.created))
	}
}

func TestPipelineJobRunCreatesDraft(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}
	generated := domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}

	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: generated},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	drafts := job.drafts.(*draftRepoStub)

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 1 {
		t.Fatalf("created drafts = %d, want 1", len(drafts.created))
	}
	if drafts.created[0].Status != domain.DraftStatusPending {
		t.Fatalf("status = %s, want pending", drafts.created[0].Status)
	}
}

func TestPipelineJobRunWithPlannerEnabledStillCreatesDraft(t *testing.T) {
	planner := &plannerStub{}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{{ID: 1, Enabled: true}}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {{ID: 10, URL: "https://example.com/a", Title: "AI update"}}}},
		&channelRepoStub{channels: []domain.Channel{{ID: 1, Slug: "ai-news"}}},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{item: domain.SourceItem{ID: 10, URL: "https://example.com/a", Title: "AI update"}},
		&dedupStub{duplicate: false},
		&scorerStub{score: 5},
		&routerStub{ids: []int64{1}},
		&generatorStub{draft: domain.Draft{SourceItemID: 10, ChannelID: 1, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	job = job.WithEditorialPlanner(planner)

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	drafts := job.drafts.(*draftRepoStub)
	if planner.calls != 1 {
		t.Fatalf("planner calls = %d, want 1", planner.calls)
	}
	if len(drafts.created) != 1 {
		t.Fatalf("created drafts = %d, want 1", len(drafts.created))
	}
}

func TestPipelineJobRunReportsPlannerErrorViaHook(t *testing.T) {
	planner := &plannerStub{err: errors.New("planner failed")}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{{ID: 1, Enabled: true}}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {{ID: 10, URL: "https://example.com/a", Title: "AI update"}}}},
		&channelRepoStub{channels: []domain.Channel{{ID: 1, Slug: "ai-news"}}},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{item: domain.SourceItem{ID: 10, URL: "https://example.com/a", Title: "AI update"}},
		&dedupStub{duplicate: false},
		&scorerStub{score: 5},
		&routerStub{ids: []int64{1}},
		&generatorStub{draft: domain.Draft{SourceItemID: 10, ChannelID: 1, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	plannerHookCalls := 0
	job = job.WithEditorialPlanner(planner).WithPlannerErrorHook(func(domain.SourceItem, error) { plannerHookCalls++ })

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	drafts := job.drafts.(*draftRepoStub)
	if planner.calls != 1 {
		t.Fatalf("planner calls = %d, want 1", planner.calls)
	}
	if plannerHookCalls != 1 {
		t.Fatalf("plannerHookCalls = %d, want 1", plannerHookCalls)
	}
	if len(drafts.created) != 1 {
		t.Fatalf("created drafts = %d, want 1", len(drafts.created))
	}
}

func TestPipelineJobRunDoesNotPlanDuplicateItems(t *testing.T) {
	planner := &plannerStub{}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{{ID: 1, Enabled: true}}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {{ID: 10, URL: "https://example.com/a", Title: "AI update"}}}},
		&channelRepoStub{channels: []domain.Channel{{ID: 1, Slug: "ai-news"}}},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{item: domain.SourceItem{ID: 10, URL: "https://example.com/a", Title: "AI update"}},
		&dedupStub{duplicate: true},
		&scorerStub{score: 5},
		&routerStub{ids: []int64{1}},
		&generatorStub{draft: domain.Draft{SourceItemID: 10, ChannelID: 1, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	job = job.WithEditorialPlanner(planner)

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if planner.calls != 0 {
		t.Fatalf("planner calls = %d, want 0", planner.calls)
	}
}

func TestPipelineJobRunGeneratesAssetsFromPlannedIntents(t *testing.T) {
	plannerCalls := 0
	assetGen := &assetGeneratorStub{}
	plannerResult := []domain.PublishIntent{{RawItemID: 10, ChannelID: 1, Format: "text", Status: domain.PublishIntentStatusPlanned}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{{ID: 1, Enabled: true}}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {{ID: 10, URL: "https://example.com/a", Title: "AI update"}}}},
		&channelRepoStub{channels: []domain.Channel{{ID: 1, Slug: "ai-news"}}},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{item: domain.SourceItem{ID: 10, URL: "https://example.com/a", Title: "AI update"}},
		&dedupStub{duplicate: false},
		&scorerStub{score: 5},
		&routerStub{ids: []int64{1}},
		&generatorStub{draft: domain.Draft{SourceItemID: 10, ChannelID: 1, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	job = job.WithEditorialPlanner(plannerFunc(func(context.Context, domain.SourceItem) ([]domain.PublishIntent, error) {
		plannerCalls++
		return plannerResult, nil
	})).WithIntentAssetGenerator(assetGen)

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if plannerCalls != 1 {
		t.Fatalf("planner calls = %d, want 1", plannerCalls)
	}
	if assetGen.calls != 1 {
		t.Fatalf("assetGen calls = %d, want 1", assetGen.calls)
	}
}

func TestPipelineJobRunReportsAssetGenerationErrorViaHook(t *testing.T) {
	assetGen := &assetGeneratorStub{err: errors.New("asset gen failed")}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{{ID: 1, Enabled: true}}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {{ID: 10, URL: "https://example.com/a", Title: "AI update"}}}},
		&channelRepoStub{channels: []domain.Channel{{ID: 1, Slug: "ai-news"}}},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{item: domain.SourceItem{ID: 10, URL: "https://example.com/a", Title: "AI update"}},
		&dedupStub{duplicate: false},
		&scorerStub{score: 5},
		&routerStub{ids: []int64{1}},
		&generatorStub{draft: domain.Draft{SourceItemID: 10, ChannelID: 1, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	assetHookCalls := 0
	job = job.WithEditorialPlanner(plannerFunc(func(context.Context, domain.SourceItem) ([]domain.PublishIntent, error) {
		return []domain.PublishIntent{{RawItemID: 10, ChannelID: 1, Format: "text", Status: domain.PublishIntentStatusPlanned}}, nil
	})).WithIntentAssetGenerator(assetGen).WithAssetErrorHook(func(domain.PublishIntent, error) { assetHookCalls++ })

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if assetGen.calls != 1 {
		t.Fatalf("assetGen calls = %d, want 1", assetGen.calls)
	}
	if assetHookCalls != 1 {
		t.Fatalf("assetHookCalls = %d, want 1", assetHookCalls)
	}
}

func TestPipelineJobRunDoesNotPlanOrGenerateAssetsWhenRulesBlockAllChannels(t *testing.T) {
	plannerCalls := 0
	assetGen := &assetGeneratorStub{}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{{ID: 1, Enabled: true}}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {{ID: 10, URL: "https://example.com/a", Title: "AI update"}}}},
		&channelRepoStub{channels: []domain.Channel{{ID: 1, Slug: "ai-news"}}},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{item: domain.SourceItem{ID: 10, URL: "https://example.com/a", Title: "AI update"}},
		&dedupStub{duplicate: false},
		&scorerStub{score: 5},
		&routerStub{ids: []int64{1}},
		&generatorStub{draft: domain.Draft{SourceItemID: 10, ChannelID: 1, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		&rulesStub{allowByChannel: map[int64]bool{1: false}},
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	job = job.WithEditorialPlanner(plannerFunc(func(context.Context, domain.SourceItem) ([]domain.PublishIntent, error) {
		plannerCalls++
		return []domain.PublishIntent{{RawItemID: 10, ChannelID: 1, Format: "text", Status: domain.PublishIntentStatusPlanned}}, nil
	})).WithIntentAssetGenerator(assetGen)

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if plannerCalls != 0 {
		t.Fatalf("planner calls = %d, want 0", plannerCalls)
	}
	if assetGen.calls != 0 {
		t.Fatalf("assetGen calls = %d, want 0", assetGen.calls)
	}
}

type plannerFunc func(context.Context, domain.SourceItem) ([]domain.PublishIntent, error)

func (f plannerFunc) PlanForSourceItem(ctx context.Context, item domain.SourceItem) ([]domain.PublishIntent, error) {
	return f(ctx, item)
}

func TestPipelineJobRunAppliesImageEnrichment(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	imageURL := "https://cdn.example.com/image.jpg"
	enriched := item
	enriched.ImageURL = &imageURL
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}

	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	job.WithImageEnricher(&imageEnricherStub{item: enriched})

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 1 {
		t.Fatalf("created drafts = %d, want 1", len(drafts.created))
	}
	if drafts.created[0].ImageURL == nil || *drafts.created[0].ImageURL != imageURL {
		t.Fatalf("draft image url not propagated")
	}
}

func TestPipelineJobRunReturnsErrorWhenImageEnrichmentFails(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}

	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	job.WithImageEnricher(&imageEnricherStub{err: errors.New("boom")})

	if err := job.Run(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPipelineJobRunSkipsDuplicate(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}
	generated := domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}

	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: true},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: generated},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 0 {
		t.Fatalf("created drafts = %d, want 0 when duplicate", len(drafts.created))
	}
}

func TestPipelineJobRunCombinesMemoryAndFeedbackScoreAdjustments(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}

	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{
		domain.DraftStatusPosted: {{ID: 100, ChannelID: 7, Status: domain.DraftStatusPosted}},
	}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&advancedScorerStub{base: -1, memory: 2, feedback: 1},
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		&feedbackRepoStub{byDraft: map[int64]domain.PerformanceFeedback{100: {DraftID: 100, ChannelID: 7, Score: 1}}},
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 1 {
		t.Fatalf("created drafts = %d, want 1", len(drafts.created))
	}
}

func TestPipelineJobRunAppliesMemoryScorePerChannel(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channels := []domain.Channel{
		{ID: 7, Slug: "ai-news", Name: "AI News"},
		{ID: 8, Slug: "ai-tools", Name: "AI Tools"},
	}

	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: channels},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&channelMemoryScorerStub{base: 0},
		&routerStub{ids: []int64{7, 8}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		&topicMemoryStub{topicsByChannel: map[int64][]domain.TopicMemory{
			7: {{ChannelID: 7, Topic: "ai", MentionCount: 1}},
		}},
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 1 {
		t.Fatalf("created drafts = %d, want 1", len(drafts.created))
	}
	if drafts.created[0].ChannelID != 7 {
		t.Fatalf("created draft channel = %d, want 7", drafts.created[0].ChannelID)
	}
}

func TestPipelineJobRunMergesFeedbackAndMemoryRouting(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channels := []domain.Channel{{ID: 7, Slug: "ai-news", Name: "AI News"}, {ID: 8, Slug: "ai-tools", Name: "AI Tools"}}

	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: channels},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&advancedRouterStub{idsMemory: []int64{7, 8}, idsFeedback: []int64{8, 7}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 2 {
		t.Fatalf("created drafts = %d, want 2", len(drafts.created))
	}
}

func TestPipelineJobRunFeedbackRoutingDoesNotInjectFallbackOutsideMemoryMatches(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channels := []domain.Channel{{ID: 7, Slug: "ai-news", Name: "AI News"}, {ID: 8, Slug: "ai-tools", Name: "AI Tools"}}

	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: channels},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&advancedRouterStub{idsMemory: []int64{8}, idsFeedback: []int64{7}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 8, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 1 {
		t.Fatalf("created drafts = %d, want 1", len(drafts.created))
	}
	if drafts.created[0].ChannelID != 8 {
		t.Fatalf("created draft channel = %d, want 8", drafts.created[0].ChannelID)
	}
}

func TestPipelineJobRunUsesOnlyFeedbackAwareGeneratorCall(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}

	gen := &feedbackGeneratorStub{
		draftBase:     domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "base", Body: "base", Status: domain.DraftStatusPending},
		draftFeedback: domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "fb", Body: "fb", Status: domain.DraftStatusPending},
	}
	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		gen,
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if gen.baseCalls != 0 || gen.feedbackCalls != 1 {
		t.Fatalf("baseCalls=%d feedbackCalls=%d, want 0 and 1", gen.baseCalls, gen.feedbackCalls)
	}
	if len(drafts.created) != 1 || drafts.created[0].Title != "fb" {
		t.Fatalf("created drafts = %+v", drafts.created)
	}
}

func TestPipelineJobRunStoresABVariants(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}

	gen := &variantGeneratorStub{variants: []domain.Draft{
		{SourceItemID: 11, ChannelID: 7, Variant: "A", Title: "A", Body: "Body A", Status: domain.DraftStatusPending},
		{SourceItemID: 11, ChannelID: 7, Variant: "B", Title: "B", Body: "Body B", Status: domain.DraftStatusPending},
	}}
	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		gen,
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 2 {
		t.Fatalf("created drafts = %d, want 2", len(drafts.created))
	}
}

func TestPipelineJobRunSkipsVariantGenerationWhenABAlreadyExist(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}

	gen := &variantGeneratorStub{variants: []domain.Draft{
		{SourceItemID: 11, ChannelID: 7, Variant: "A", Title: "A", Body: "Body A", Status: domain.DraftStatusPending},
		{SourceItemID: 11, ChannelID: 7, Variant: "B", Title: "B", Body: "Body B", Status: domain.DraftStatusPending},
	}}
	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{
		domain.DraftStatusPending: {
			{SourceItemID: 11, ChannelID: 7, Variant: "A", Status: domain.DraftStatusPending},
			{SourceItemID: 11, ChannelID: 7, Variant: "B", Status: domain.DraftStatusPending},
		},
	}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		gen,
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if gen.calls != 0 {
		t.Fatalf("variant generation calls = %d, want 0", gen.calls)
	}
	if len(drafts.created) != 0 {
		t.Fatalf("created drafts = %d, want 0", len(drafts.created))
	}
}

func TestPipelineJobRunStoresRejectedDraft(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}
	generated := domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}

	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: generated},
		&guardStub{result: editorial.Result{Accepted: false}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 1 {
		t.Fatalf("created drafts = %d, want 1", len(drafts.created))
	}
	if drafts.created[0].Status != domain.DraftStatusRejected {
		t.Fatalf("status = %s, want rejected", drafts.created[0].Status)
	}
}

func TestPipelineJobRunReturnsUpstreamError(t *testing.T) {
	job, err := NewPipelineJob(
		&sourceRepoStub{err: errors.New("boom")},
		&sourceItemRepoStub{},
		&channelRepoStub{},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{},
		&dedupStub{},
		&scorerStub{},
		&routerStub{},
		&generatorStub{},
		&guardStub{},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestPipelineJobRunSkipsWhenRejectedDraftAlreadyExists(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}
	generated := domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}

	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{
		domain.DraftStatusRejected: {{SourceItemID: 11, ChannelID: 7, Status: domain.DraftStatusRejected}},
	}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: generated},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 0 {
		t.Fatalf("created drafts = %d, want 0 when rejected draft already exists", len(drafts.created))
	}
}

func TestPipelineJobRunSkipsDraftWhenRuleBlocksChannel(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}

	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}
	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		drafts,
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		&rulesStub{allowByChannel: map[int64]bool{7: false}},
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.created) != 0 {
		t.Fatalf("created drafts = %d, want 0", len(drafts.created))
	}
}

func TestPipelineJobRunReturnsRuleEvaluationError(t *testing.T) {
	source := domain.Source{ID: 1, Enabled: true}
	item := domain.SourceItem{ID: 11, SourceID: 1, ExternalID: "x", URL: "https://example.com", Title: "AI launch"}
	channel := domain.Channel{ID: 7, Slug: "ai-news", Name: "AI News"}

	job, err := NewPipelineJob(
		&sourceRepoStub{sources: []domain.Source{source}},
		&sourceItemRepoStub{itemsBySource: map[int64][]domain.SourceItem{1: {item}}},
		&channelRepoStub{channels: []domain.Channel{channel}},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{item: item},
		&dedupStub{duplicate: false},
		&scorerStub{score: 10},
		&routerStub{ids: []int64{7}},
		&generatorStub{draft: domain.Draft{SourceItemID: 11, ChannelID: 7, Title: "t", Body: "b", Status: domain.DraftStatusPending}},
		&guardStub{result: editorial.Result{Accepted: true}},
		nil,
		&rulesStub{err: errors.New("rules db down")},
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	if err := job.Run(context.Background()); err == nil {
		t.Fatalf("expected error")
	}
}

func TestNewCollectorJobValidation(t *testing.T) {
	if _, err := NewCollectorJob(nil); err == nil {
		t.Fatalf("expected error for nil collector")
	}

	job := &CollectorJob{}
	if err := job.Run(nil); err == nil {
		t.Fatalf("expected error for nil context")
	}
	if err := job.Run(context.Background()); err == nil {
		t.Fatalf("expected error for nil collector dependency")
	}
}

func TestNewPipelineJobValidation(t *testing.T) {
	validSources := &sourceRepoStub{}
	validItems := &sourceItemRepoStub{}
	validChannels := &channelRepoStub{}
	validDrafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}
	validNormalizer := &normalizerStub{}
	validDedup := &dedupStub{}
	validScorer := &scorerStub{}
	validRouter := &routerStub{}
	validGenerator := &generatorStub{}
	validGuard := &guardStub{}

	if _, err := NewPipelineJob(nil, validItems, validChannels, validDrafts, validNormalizer, validDedup, validScorer, validRouter, validGenerator, validGuard, nil, nil, nil); err == nil {
		t.Fatalf("expected error for nil sources")
	}
	if _, err := NewPipelineJob(validSources, nil, validChannels, validDrafts, validNormalizer, validDedup, validScorer, validRouter, validGenerator, validGuard, nil, nil, nil); err == nil {
		t.Fatalf("expected error for nil items")
	}
	if _, err := NewPipelineJob(validSources, validItems, nil, validDrafts, validNormalizer, validDedup, validScorer, validRouter, validGenerator, validGuard, nil, nil, nil); err == nil {
		t.Fatalf("expected error for nil channels")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, nil, validNormalizer, validDedup, validScorer, validRouter, validGenerator, validGuard, nil, nil, nil); err == nil {
		t.Fatalf("expected error for nil drafts")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, validDrafts, nil, validDedup, validScorer, validRouter, validGenerator, validGuard, nil, nil, nil); err == nil {
		t.Fatalf("expected error for nil normalizer")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, validDrafts, validNormalizer, nil, validScorer, validRouter, validGenerator, validGuard, nil, nil, nil); err == nil {
		t.Fatalf("expected error for nil dedup")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, validDrafts, validNormalizer, validDedup, nil, validRouter, validGenerator, validGuard, nil, nil, nil); err == nil {
		t.Fatalf("expected error for nil scorer")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, validDrafts, validNormalizer, validDedup, validScorer, nil, validGenerator, validGuard, nil, nil, nil); err == nil {
		t.Fatalf("expected error for nil router")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, validDrafts, validNormalizer, validDedup, validScorer, validRouter, nil, validGuard, nil, nil, nil); err == nil {
		t.Fatalf("expected error for nil generator")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, validDrafts, validNormalizer, validDedup, validScorer, validRouter, validGenerator, nil, nil, nil, nil); err == nil {
		t.Fatalf("expected error for nil guard")
	}
}

func TestNewPipelineJobUsesBoundedDraftScanLimit(t *testing.T) {
	job, err := NewPipelineJob(
		&sourceRepoStub{},
		&sourceItemRepoStub{},
		&channelRepoStub{},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{},
		&dedupStub{},
		&scorerStub{},
		&routerStub{},
		&generatorStub{},
		&guardStub{},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	if job.existingDraftLimit != defaultExistingLimit {
		t.Fatalf("existingDraftLimit = %d, want %d", job.existingDraftLimit, defaultExistingLimit)
	}
}

func TestPipelineJobWithBatchLimitsOverridesDefaults(t *testing.T) {
	job, err := NewPipelineJob(
		&sourceRepoStub{},
		&sourceItemRepoStub{},
		&channelRepoStub{},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{},
		&dedupStub{},
		&scorerStub{},
		&routerStub{},
		&generatorStub{},
		&guardStub{},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	job.WithBatchLimits(12, 34)
	if job.recentItemsLimit != 12 {
		t.Fatalf("recentItemsLimit = %d, want 12", job.recentItemsLimit)
	}
	if job.existingDraftLimit != 34 {
		t.Fatalf("existingDraftLimit = %d, want 34", job.existingDraftLimit)
	}
}

func TestMergeRankedRouteIDsPreservesFallbackMatches(t *testing.T) {
	got := mergeRankedRouteIDs([]int64{8}, []int64{7})
	if len(got) != 2 || got[0] != 8 || got[1] != 7 {
		t.Fatalf("got = %v, want [8 7]", got)
	}
}

func TestFilterRouteIDsKeepsOnlyAllowedMatches(t *testing.T) {
	got := filterRouteIDs([]int64{7, 8, 9}, []int64{8, 9})
	if len(got) != 2 || got[0] != 8 || got[1] != 9 {
		t.Fatalf("got = %v, want [8 9]", got)
	}
}

func TestChannelFeedbackAverages(t *testing.T) {
	job, err := NewPipelineJob(
		&sourceRepoStub{},
		&sourceItemRepoStub{},
		&channelRepoStub{},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{domain.DraftStatusPosted: {
			{ID: 1, ChannelID: 7, Status: domain.DraftStatusPosted},
			{ID: 2, ChannelID: 7, Status: domain.DraftStatusPosted},
			{ID: 3, ChannelID: 9, Status: domain.DraftStatusPosted},
		}}},
		&normalizerStub{},
		&dedupStub{},
		&scorerStub{},
		&routerStub{},
		&generatorStub{},
		&guardStub{},
		nil,
		nil,
		&feedbackRepoStub{byDraft: map[int64]domain.PerformanceFeedback{
			1: {DraftID: 1, ChannelID: 7, Score: 1.0},
			2: {DraftID: 2, ChannelID: 7, Score: 3.0},
			3: {DraftID: 3, ChannelID: 9, Score: 2.0},
		}},
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}

	avg, err := job.channelFeedbackAverages(context.Background())
	if err != nil {
		t.Fatalf("channelFeedbackAverages() error = %v", err)
	}
	if avg[7] != 2.0 || avg[9] != 2.0 {
		t.Fatalf("avg = %v", avg)
	}
}

func TestPipelineJobRunValidation(t *testing.T) {
	var nilJob *PipelineJob
	if err := nilJob.Run(context.Background()); err == nil {
		t.Fatalf("expected error for nil job")
	}

	job, err := NewPipelineJob(
		&sourceRepoStub{},
		&sourceItemRepoStub{},
		&channelRepoStub{},
		&draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}},
		&normalizerStub{},
		&dedupStub{},
		&scorerStub{},
		&routerStub{},
		&generatorStub{},
		&guardStub{},
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	if err := job.Run(nil); err == nil {
		t.Fatalf("expected error for nil context")
	}
}

func TestNewAutoRepostJobValidation(t *testing.T) {
	if _, err := NewAutoRepostJob(nil, &feedbackRepoStub{}); err == nil {
		t.Fatalf("expected nil draft repository error")
	}
	if _, err := NewAutoRepostJob(&draftRepoStub{}, nil); err == nil {
		t.Fatalf("expected nil feedback reader error")
	}
}

func TestAutoRepostJobRunPromotesTopEligibleDrafts(t *testing.T) {
	now := time.Now().UTC()
	drafts := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{
		domain.DraftStatusPosted: {
			{ID: 1, Status: domain.DraftStatusPosted, UpdatedAt: now.Add(-96 * time.Hour)},
			{ID: 2, Status: domain.DraftStatusPosted, UpdatedAt: now.Add(-120 * time.Hour)},
			{ID: 3, Status: domain.DraftStatusPosted, UpdatedAt: now.Add(-10 * time.Hour)},
		},
	}}
	feedback := &feedbackRepoStub{byDraft: map[int64]domain.PerformanceFeedback{
		1: {DraftID: 1, Score: 1.1},
		2: {DraftID: 2, Score: 2.0},
		3: {DraftID: 3, Score: 3.0},
	}}

	job, err := NewAutoRepostJob(drafts, feedback)
	if err != nil {
		t.Fatalf("NewAutoRepostJob() error = %v", err)
	}
	job.maxPerRun = 2

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(drafts.updated) != 2 {
		t.Fatalf("updated drafts = %d, want 2", len(drafts.updated))
	}
	if drafts.updated[2] != domain.DraftStatusApproved {
		t.Fatalf("draft 2 status = %q, want approved", drafts.updated[2])
	}
	if drafts.updated[1] != domain.DraftStatusApproved {
		t.Fatalf("draft 1 status = %q, want approved", drafts.updated[1])
	}
	if _, ok := drafts.updated[3]; ok {
		t.Fatalf("draft 3 should be skipped by cooldown")
	}
}

func TestAutoRepostJobRunValidationAndErrors(t *testing.T) {
	job := &AutoRepostJob{}
	if err := job.Run(nil); err == nil {
		t.Fatalf("expected nil context error")
	}

	job.drafts = &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{}}
	job.feedback = &feedbackRepoStub{}
	job.maxPerRun = 0
	job.listLimit = 1
	if err := job.Run(context.Background()); err == nil {
		t.Fatalf("expected invalid maxPerRun error")
	}

	job.maxPerRun = 1
	job.listLimit = 0
	if err := job.Run(context.Background()); err == nil {
		t.Fatalf("expected invalid listLimit error")
	}

	job.listLimit = 1
	job.drafts = &draftRepoStub{listErr: errors.New("boom")}
	if err := job.Run(context.Background()); err == nil {
		t.Fatalf("expected list error")
	}

	job.drafts = &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{
		domain.DraftStatusPosted: {{ID: 10}},
	}}
	job.feedback = &feedbackRepoStub{byDraft: map[int64]domain.PerformanceFeedback{10: {DraftID: 10, Score: 5.0}}}
	job.maxPerRun = 1
	job.listLimit = 1
	job.minPostedFor = time.Hour
	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if len(job.drafts.(*draftRepoStub).updated) != 0 {
		t.Fatalf("draft with zero updated_at should be skipped by cooldown")
	}

	job.drafts = &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{
		domain.DraftStatusPosted: {{ID: 9, UpdatedAt: time.Now().UTC().Add(-96 * time.Hour)}},
	}}
	job.feedback = &feedbackRepoStub{err: errors.New("boom")}
	if err := job.Run(context.Background()); err == nil {
		t.Fatalf("expected feedback error")
	}

	job.feedback = &feedbackRepoStub{byDraft: map[int64]domain.PerformanceFeedback{9: {DraftID: 9, Score: 2.0}}}
	job.drafts = &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{
		domain.DraftStatusPosted: {{ID: 9, UpdatedAt: time.Now().UTC().Add(-96 * time.Hour)}},
	}, updateErr: errors.New("boom")}
	if err := job.Run(context.Background()); err == nil {
		t.Fatalf("expected update error")
	}
}
