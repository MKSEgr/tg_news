package domain

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type stubChannelRepo struct{}

func (stubChannelRepo) Create(context.Context, Channel) (Channel, error) { return Channel{}, nil }
func (stubChannelRepo) GetByID(context.Context, int64) (Channel, error)  { return Channel{}, nil }
func (stubChannelRepo) List(context.Context) ([]Channel, error)          { return nil, nil }

type stubChannelRelationshipRepo struct{}

func (stubChannelRelationshipRepo) Create(context.Context, ChannelRelationship) (ChannelRelationship, error) {
	return ChannelRelationship{}, nil
}
func (stubChannelRelationshipRepo) ListByChannel(context.Context, int64, int) ([]ChannelRelationship, error) {
	return nil, nil
}

type stubSourceRepo struct{}

func (stubSourceRepo) Create(context.Context, Source) (Source, error) { return Source{}, nil }
func (stubSourceRepo) GetByID(context.Context, int64) (Source, error) { return Source{}, nil }
func (stubSourceRepo) List(context.Context) ([]Source, error)         { return nil, nil }
func (stubSourceRepo) ListEnabled(context.Context) ([]Source, error)  { return nil, nil }

type stubSourceItemRepo struct{}

func (stubSourceItemRepo) Create(context.Context, SourceItem) (SourceItem, error) {
	return SourceItem{}, nil
}
func (stubSourceItemRepo) GetByID(context.Context, int64) (SourceItem, error) {
	return SourceItem{}, nil
}
func (stubSourceItemRepo) ListBySourceID(context.Context, int64, int) ([]SourceItem, error) {
	return nil, nil
}
func (stubSourceItemRepo) ListRecent(context.Context, int) ([]SourceItem, error) {
	return nil, nil
}

type stubDraftRepo struct{}

func (stubDraftRepo) Create(context.Context, Draft) (Draft, error)  { return Draft{}, nil }
func (stubDraftRepo) GetByID(context.Context, int64) (Draft, error) { return Draft{}, nil }
func (stubDraftRepo) ListByStatus(context.Context, DraftStatus, int) ([]Draft, error) {
	return nil, nil
}
func (stubDraftRepo) UpdateStatus(context.Context, int64, DraftStatus) error { return nil }

type stubContentAssetRepo struct{}

func (stubContentAssetRepo) Create(context.Context, ContentAsset) (ContentAsset, error) {
	return ContentAsset{}, nil
}
func (stubContentAssetRepo) GetByID(context.Context, int64) (ContentAsset, error) {
	return ContentAsset{}, nil
}
func (stubContentAssetRepo) ListByRawItemID(context.Context, int64, int) ([]ContentAsset, error) {
	return nil, nil
}

type stubAssetRelationshipRepo struct{}

func (stubAssetRelationshipRepo) Create(context.Context, AssetRelationship) (AssetRelationship, error) {
	return AssetRelationship{}, nil
}
func (stubAssetRelationshipRepo) ListByAssetID(context.Context, int64, int) ([]AssetRelationship, error) {
	return nil, nil
}

type stubStoryClusterRepo struct{}

func (stubStoryClusterRepo) Create(context.Context, StoryCluster) (StoryCluster, error) {
	return StoryCluster{}, nil
}
func (stubStoryClusterRepo) GetByID(context.Context, int64) (StoryCluster, error) {
	return StoryCluster{}, nil
}
func (stubStoryClusterRepo) FindByKey(context.Context, string) (StoryCluster, error) {
	return StoryCluster{}, nil
}

type stubMonetizationHookRepo struct{}

func (stubMonetizationHookRepo) Create(context.Context, MonetizationHook) (MonetizationHook, error) {
	return MonetizationHook{}, nil
}
func (stubMonetizationHookRepo) GetByID(context.Context, int64) (MonetizationHook, error) {
	return MonetizationHook{}, nil
}
func (stubMonetizationHookRepo) ListByDraftID(context.Context, int64, int) ([]MonetizationHook, error) {
	return nil, nil
}

type stubClusterEventRepo struct{}

func (stubClusterEventRepo) Create(context.Context, ClusterEvent) (ClusterEvent, error) {
	return ClusterEvent{}, nil
}
func (stubClusterEventRepo) ListByClusterID(context.Context, int64, int) ([]ClusterEvent, error) {
	return nil, nil
}

type stubTopicMemoryRepo struct{}

func (stubTopicMemoryRepo) UpsertMention(context.Context, TopicMemory) (TopicMemory, error) {
	return TopicMemory{}, nil
}
func (stubTopicMemoryRepo) ListTopByChannel(context.Context, int64, int) ([]TopicMemory, error) {
	return nil, nil
}

type stubContentRuleRepo struct{}

func (stubContentRuleRepo) Create(context.Context, ContentRule) (ContentRule, error) {
	return ContentRule{}, nil
}
func (stubContentRuleRepo) ListEnabled(context.Context, *int64) ([]ContentRule, error) {
	return nil, nil
}

type stubPerformanceFeedbackRepo struct{}

type stubRankingFeatureRepo struct{}

func (stubRankingFeatureRepo) Create(context.Context, RankingFeature) (RankingFeature, error) {
	return RankingFeature{}, nil
}
func (stubRankingFeatureRepo) ListByEntity(context.Context, string, int64, int) ([]RankingFeature, error) {
	return nil, nil
}

func (stubPerformanceFeedbackRepo) Upsert(context.Context, PerformanceFeedback) (PerformanceFeedback, error) {
	return PerformanceFeedback{}, nil
}
func (stubPerformanceFeedbackRepo) GetByDraftID(context.Context, int64) (PerformanceFeedback, error) {
	return PerformanceFeedback{}, nil
}

type stubPublishIntentRepo struct{}

func (stubPublishIntentRepo) Create(context.Context, PublishIntent) (PublishIntent, error) {
	return PublishIntent{}, nil
}
func (stubPublishIntentRepo) ListByRawItemID(context.Context, int64, int) ([]PublishIntent, error) {
	return nil, nil
}
func (stubPublishIntentRepo) UpdateStatus(context.Context, int64, PublishIntentStatus) error {
	return nil
}

func TestRepositoryInterfacesImplementedByStubs(t *testing.T) {
	var _ ChannelRepository = stubChannelRepo{}
	var _ ChannelRelationshipRepository = stubChannelRelationshipRepo{}
	var _ SourceRepository = stubSourceRepo{}
	var _ SourceItemRepository = stubSourceItemRepo{}
	var _ DraftRepository = stubDraftRepo{}
	var _ TopicMemoryRepository = stubTopicMemoryRepo{}
	var _ ContentRuleRepository = stubContentRuleRepo{}
	var _ PerformanceFeedbackRepository = stubPerformanceFeedbackRepo{}
	var _ PublishIntentRepository = stubPublishIntentRepo{}
	var _ ContentAssetRepository = stubContentAssetRepo{}
	var _ AssetRelationshipRepository = stubAssetRelationshipRepo{}
	var _ StoryClusterRepository = stubStoryClusterRepo{}
	var _ MonetizationHookRepository = stubMonetizationHookRepo{}
	var _ ClusterEventRepository = stubClusterEventRepo{}
	var _ RankingFeatureRepository = stubRankingFeatureRepo{}
}

func TestErrNotFoundIsWrappable(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", ErrNotFound)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected wrapped error to match ErrNotFound")
	}
}
