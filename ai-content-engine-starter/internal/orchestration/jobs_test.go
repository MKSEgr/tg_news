package orchestration

import (
	"context"
	"errors"
	"testing"

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
func (s *draftRepoStub) UpdateStatus(context.Context, int64, domain.DraftStatus) error { return nil }

type normalizerStub struct{ item domain.SourceItem }

func (n *normalizerStub) Normalize(domain.SourceItem) (domain.SourceItem, error) { return n.item, nil }

type dedupStub struct{ duplicate bool }

func (d *dedupStub) IsDuplicate(context.Context, domain.SourceItem) (bool, error) {
	return d.duplicate, nil
}

type scorerStub struct{ score int }

func (s *scorerStub) Score(domain.SourceItem) int { return s.score }

type routerStub struct{ ids []int64 }

func (r *routerStub) Route(domain.SourceItem, []domain.Channel) ([]int64, error) { return r.ids, nil }

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

type guardStub struct{ result editorial.Result }

func (g *guardStub) Check(domain.Draft) (editorial.Result, error) { return g.result, nil }

func TestCollectorJobRun(t *testing.T) {
	job, err := NewCollectorJob(&collectorStub{})
	if err != nil {
		t.Fatalf("NewCollectorJob() error = %v", err)
	}
	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Run() error = %v", err)
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

	if _, err := NewPipelineJob(nil, validItems, validChannels, validDrafts, validNormalizer, validDedup, validScorer, validRouter, validGenerator, validGuard); err == nil {
		t.Fatalf("expected error for nil sources")
	}
	if _, err := NewPipelineJob(validSources, nil, validChannels, validDrafts, validNormalizer, validDedup, validScorer, validRouter, validGenerator, validGuard); err == nil {
		t.Fatalf("expected error for nil items")
	}
	if _, err := NewPipelineJob(validSources, validItems, nil, validDrafts, validNormalizer, validDedup, validScorer, validRouter, validGenerator, validGuard); err == nil {
		t.Fatalf("expected error for nil channels")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, nil, validNormalizer, validDedup, validScorer, validRouter, validGenerator, validGuard); err == nil {
		t.Fatalf("expected error for nil drafts")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, validDrafts, nil, validDedup, validScorer, validRouter, validGenerator, validGuard); err == nil {
		t.Fatalf("expected error for nil normalizer")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, validDrafts, validNormalizer, nil, validScorer, validRouter, validGenerator, validGuard); err == nil {
		t.Fatalf("expected error for nil dedup")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, validDrafts, validNormalizer, validDedup, nil, validRouter, validGenerator, validGuard); err == nil {
		t.Fatalf("expected error for nil scorer")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, validDrafts, validNormalizer, validDedup, validScorer, nil, validGenerator, validGuard); err == nil {
		t.Fatalf("expected error for nil router")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, validDrafts, validNormalizer, validDedup, validScorer, validRouter, nil, validGuard); err == nil {
		t.Fatalf("expected error for nil generator")
	}
	if _, err := NewPipelineJob(validSources, validItems, validChannels, validDrafts, validNormalizer, validDedup, validScorer, validRouter, validGenerator, nil); err == nil {
		t.Fatalf("expected error for nil guard")
	}
}

func TestNewPipelineJobUsesUnboundedDraftScanLimit(t *testing.T) {
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
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	if job.existingDraftLimit != defaultExistingLimit {
		t.Fatalf("existingDraftLimit = %d, want %d", job.existingDraftLimit, defaultExistingLimit)
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
	)
	if err != nil {
		t.Fatalf("NewPipelineJob() error = %v", err)
	}
	if err := job.Run(nil); err == nil {
		t.Fatalf("expected error for nil context")
	}
}
