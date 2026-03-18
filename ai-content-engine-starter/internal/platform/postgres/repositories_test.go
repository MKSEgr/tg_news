package postgres

import (
	"context"
	"database/sql"
	"math"
	"testing"
	"time"

	"ai-content-engine-starter/internal/domain"
)

func TestRepositoriesImplementDomainInterfaces(t *testing.T) {
	var _ domain.ChannelRepository = (*ChannelRepository)(nil)
	var _ domain.ChannelRelationshipRepository = (*ChannelRelationshipRepository)(nil)
	var _ domain.SourceRepository = (*SourceRepository)(nil)
	var _ domain.SourceItemRepository = (*SourceItemRepository)(nil)
	var _ domain.DraftRepository = (*DraftRepository)(nil)
	var _ domain.PublishIntentRepository = (*PublishIntentRepository)(nil)
	var _ domain.ContentAssetRepository = (*ContentAssetRepository)(nil)
	var _ domain.AssetRelationshipRepository = (*AssetRelationshipRepository)(nil)
	var _ domain.StoryClusterRepository = (*StoryClusterRepository)(nil)
	var _ domain.MonetizationHookRepository = (*MonetizationHookRepository)(nil)
	var _ domain.SponsorRepository = (*SponsorRepository)(nil)
	var _ domain.AdCampaignRepository = (*AdCampaignRepository)(nil)
	var _ domain.AdSlotRepository = (*AdSlotRepository)(nil)
	var _ domain.ClusterEventRepository = (*ClusterEventRepository)(nil)
	var _ domain.TopicMemoryRepository = (*TopicMemoryRepository)(nil)
	var _ domain.ContentRuleRepository = (*ContentRuleRepository)(nil)
	var _ domain.PerformanceFeedbackRepository = (*PerformanceFeedbackRepository)(nil)
}

func TestPublishIntentRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewPublishIntentRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.PublishIntent{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	if _, err := repo.Create(context.Background(), domain.PublishIntent{RawItemID: 1, ChannelID: 1, Format: "text", Priority: 1, Status: "invalid"}); err == nil {
		t.Fatalf("Create expected status validation error")
	}
	if _, err := repo.ListByRawItemID(context.Background(), 0, 10); err == nil {
		t.Fatalf("ListByRawItemID expected validation error for raw item id")
	}
	if _, err := repo.ListByRawItemID(context.Background(), 1, 0); err == nil {
		t.Fatalf("ListByRawItemID expected validation error for limit")
	}
	if err := repo.UpdateStatus(context.Background(), 0, domain.PublishIntentStatusPlanned); err == nil {
		t.Fatalf("UpdateStatus expected id validation error")
	}
	if err := repo.UpdateStatus(context.Background(), 1, "invalid"); err == nil {
		t.Fatalf("UpdateStatus expected status validation error")
	}
}

func TestSponsorRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewSponsorRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.Sponsor{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	if _, err := repo.Create(context.Background(), domain.Sponsor{Name: "ACME", Status: "invalid", ContactInfo: "sales@acme.test"}); err == nil {
		t.Fatalf("Create expected status validation error")
	}
	if _, err := repo.GetByID(context.Background(), 0); err == nil {
		t.Fatalf("GetByID expected id validation error")
	}
	if _, err := repo.List(context.Background(), 0); err == nil {
		t.Fatalf("List expected limit validation error")
	}
}

func TestAdCampaignRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewAdCampaignRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.AdCampaign{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	start := time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC)
	if _, err := repo.Create(context.Background(), domain.AdCampaign{
		SponsorID:    1,
		CampaignName: "Launch",
		CampaignType: "invalid",
		Status:       domain.AdCampaignStatusDraft,
		StartAt:      start,
		EndAt:        start.Add(time.Hour),
	}); err == nil {
		t.Fatalf("Create expected campaign type validation error")
	}
	if _, err := repo.Create(context.Background(), domain.AdCampaign{
		SponsorID:    1,
		CampaignName: "Launch",
		CampaignType: domain.AdCampaignTypeSponsoredPost,
		Status:       "invalid",
		StartAt:      start,
		EndAt:        start.Add(time.Hour),
	}); err == nil {
		t.Fatalf("Create expected campaign status validation error")
	}
	if _, err := repo.Create(context.Background(), domain.AdCampaign{
		SponsorID:    1,
		CampaignName: "Launch",
		CampaignType: domain.AdCampaignTypeSponsoredPost,
		Status:       domain.AdCampaignStatusDraft,
		StartAt:      start.Add(time.Hour),
		EndAt:        start,
	}); err == nil {
		t.Fatalf("Create expected time ordering validation error")
	}
	if _, err := repo.GetByID(context.Background(), 0); err == nil {
		t.Fatalf("GetByID expected id validation error")
	}
	if _, err := repo.List(context.Background(), 0); err == nil {
		t.Fatalf("List expected limit validation error")
	}
}

func TestAdSlotRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewAdSlotRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.AdSlot{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	if _, err := repo.Create(context.Background(), domain.AdSlot{
		ChannelID:   1,
		ScheduledAt: time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC),
		SlotType:    "invalid",
		CampaignID:  1,
		Status:      domain.AdSlotStatusScheduled,
	}); err == nil {
		t.Fatalf("Create expected slot type validation error")
	}
	if _, err := repo.Create(context.Background(), domain.AdSlot{
		ChannelID:   1,
		ScheduledAt: time.Date(2026, 3, 18, 12, 0, 0, 0, time.UTC),
		SlotType:    domain.AdSlotTypeSponsoredPost,
		CampaignID:  1,
		Status:      "invalid",
	}); err == nil {
		t.Fatalf("Create expected slot status validation error")
	}
	if _, err := repo.ListByChannel(context.Background(), 0, 10); err == nil {
		t.Fatalf("ListByChannel expected channel validation error")
	}
	if _, err := repo.ListByChannel(context.Background(), 1, 0); err == nil {
		t.Fatalf("ListByChannel expected limit validation error")
	}
}

func TestListBySourceIDRejectsInvalidLimit(t *testing.T) {
	repo := NewSourceItemRepository(&sql.DB{})
	_, err := repo.ListBySourceID(context.Background(), 1, 0)
	if err == nil {
		t.Fatalf("ListBySourceID expected error for invalid limit")
	}
}

func TestListRecentRejectsInvalidLimit(t *testing.T) {
	repo := NewSourceItemRepository(&sql.DB{})
	if _, err := repo.ListRecent(context.Background(), 0); err == nil {
		t.Fatalf("ListRecent expected error for invalid limit")
	}
}

func TestListByStatusRejectsInvalidLimit(t *testing.T) {
	repo := NewDraftRepository(&sql.DB{})
	_, err := repo.ListByStatus(context.Background(), domain.DraftStatusPending, 0)
	if err == nil {
		t.Fatalf("ListByStatus expected error for invalid limit")
	}
}

func TestRepositoryRejectsNilDB(t *testing.T) {
	repo := NewChannelRepository(nil)
	_, err := repo.GetByID(context.Background(), 1)
	if err == nil {
		t.Fatalf("GetByID expected error when db is nil")
	}
}

func TestListTopByChannelRejectsInvalidInput(t *testing.T) {
	repo := NewTopicMemoryRepository(&sql.DB{})
	if _, err := repo.ListTopByChannel(context.Background(), 0, 10); err == nil {
		t.Fatalf("ListTopByChannel expected error for invalid channel id")
	}
	if _, err := repo.ListTopByChannel(context.Background(), 1, 0); err == nil {
		t.Fatalf("ListTopByChannel expected error for invalid limit")
	}
}

func TestUpsertMentionRejectsInvalidInput(t *testing.T) {
	repo := NewTopicMemoryRepository(&sql.DB{})
	base := domain.TopicMemory{ChannelID: 1, Topic: "ai", MentionCount: 1, LastSeenAt: time.Now().UTC()}

	invalid := []domain.TopicMemory{
		{Topic: base.Topic, MentionCount: base.MentionCount, LastSeenAt: base.LastSeenAt},
		{ChannelID: base.ChannelID, MentionCount: base.MentionCount, LastSeenAt: base.LastSeenAt},
		{ChannelID: base.ChannelID, Topic: base.Topic, LastSeenAt: base.LastSeenAt},
		{ChannelID: base.ChannelID, Topic: base.Topic, MentionCount: base.MentionCount},
	}

	for _, item := range invalid {
		if _, err := repo.UpsertMention(context.Background(), item); err == nil {
			t.Fatalf("UpsertMention expected validation error")
		}
	}
}

func TestContentRuleRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewContentRuleRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.ContentRule{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	invalidChannel := int64(0)
	if _, err := repo.ListEnabled(context.Background(), &invalidChannel); err == nil {
		t.Fatalf("ListEnabled expected channel validation error")
	}
}

func TestPerformanceFeedbackRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewPerformanceFeedbackRepository(&sql.DB{})
	if _, err := repo.Upsert(context.Background(), domain.PerformanceFeedback{}); err == nil {
		t.Fatalf("Upsert expected validation error")
	}
	if _, err := repo.Upsert(context.Background(), domain.PerformanceFeedback{DraftID: 1, ChannelID: 1, Variant: "C"}); err == nil {
		t.Fatalf("Upsert expected variant validation error")
	}
	if _, err := repo.GetByDraftID(context.Background(), 0); err == nil {
		t.Fatalf("GetByDraftID expected validation error")
	}
}

func TestNormalizeFeedbackVariant(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "spaces", in: "   ", want: ""},
		{name: "lower a", in: "a", want: "A"},
		{name: "mixed b", in: " b ", want: "B"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeFeedbackVariant(tc.in); got != tc.want {
				t.Fatalf("normalizeFeedbackVariant(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestContentAssetRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewContentAssetRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.ContentAsset{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	if _, err := repo.Create(context.Background(), domain.ContentAsset{RawItemID: 1, ChannelID: 1, AssetType: "text", Status: "invalid"}); err == nil {
		t.Fatalf("Create expected status validation error")
	}
	if _, err := repo.GetByID(context.Background(), 0); err == nil {
		t.Fatalf("GetByID expected id validation error")
	}
	if _, err := repo.ListByRawItemID(context.Background(), 0, 10); err == nil {
		t.Fatalf("ListByRawItemID expected raw item id validation error")
	}
	if _, err := repo.ListByRawItemID(context.Background(), 1, 0); err == nil {
		t.Fatalf("ListByRawItemID expected limit validation error")
	}
}

func TestAssetRelationshipRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewAssetRelationshipRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.AssetRelationship{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	if _, err := repo.Create(context.Background(), domain.AssetRelationship{FromAssetID: 1, ToAssetID: 1, RelationshipType: domain.AssetRelationshipTypeDerivedFrom}); err == nil {
		t.Fatalf("Create expected self-link validation error")
	}
	if _, err := repo.Create(context.Background(), domain.AssetRelationship{FromAssetID: 1, ToAssetID: 2, RelationshipType: "invalid"}); err == nil {
		t.Fatalf("Create expected relationship type validation error")
	}
	if _, err := repo.ListByAssetID(context.Background(), 0, 10); err == nil {
		t.Fatalf("ListByAssetID expected asset id validation error")
	}
	if _, err := repo.ListByAssetID(context.Background(), 1, 0); err == nil {
		t.Fatalf("ListByAssetID expected limit validation error")
	}
}

func TestChannelRelationshipRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewChannelRelationshipRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.ChannelRelationship{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	if _, err := repo.Create(context.Background(), domain.ChannelRelationship{
		ChannelID:        1,
		RelatedChannelID: 1,
		RelationshipType: domain.ChannelRelationshipTypeSibling,
	}); err == nil {
		t.Fatalf("Create expected self-link validation error")
	}
	if _, err := repo.Create(context.Background(), domain.ChannelRelationship{
		ChannelID:        1,
		RelatedChannelID: 2,
		RelationshipType: "invalid",
	}); err == nil {
		t.Fatalf("Create expected relationship type validation error")
	}
	if _, err := repo.Create(context.Background(), domain.ChannelRelationship{
		ChannelID:        1,
		RelatedChannelID: 2,
		RelationshipType: domain.ChannelRelationshipTypeParent,
		Strength:         -0.1,
	}); err == nil {
		t.Fatalf("Create expected negative strength validation error")
	}
	if _, err := repo.Create(context.Background(), domain.ChannelRelationship{
		ChannelID:        1,
		RelatedChannelID: 2,
		RelationshipType: domain.ChannelRelationshipTypeParent,
		Strength:         1.1,
	}); err == nil {
		t.Fatalf("Create expected strength upper-bound validation error")
	}
	if _, err := repo.ListByChannel(context.Background(), 0, 10); err == nil {
		t.Fatalf("ListByChannel expected channel id validation error")
	}
	if _, err := repo.ListByChannel(context.Background(), 1, 0); err == nil {
		t.Fatalf("ListByChannel expected limit validation error")
	}
}

func TestStoryClusterRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewStoryClusterRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.StoryCluster{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	if _, err := repo.GetByID(context.Background(), 0); err == nil {
		t.Fatalf("GetByID expected id validation error")
	}
	if _, err := repo.FindByKey(context.Background(), " "); err == nil {
		t.Fatalf("FindByKey expected key validation error")
	}
}

func TestMonetizationHookRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewMonetizationHookRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.MonetizationHook{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	valid := domain.MonetizationHook{
		DraftID:    1,
		ChannelID:  1,
		HookType:   domain.MonetizationHookTypeAffiliateCTA,
		Disclosure: "affiliate link",
		CTAText:    "Try it",
		TargetURL:  "https://example.com",
	}
	invalid := []domain.MonetizationHook{
		{ChannelID: valid.ChannelID, HookType: valid.HookType, Disclosure: valid.Disclosure, CTAText: valid.CTAText, TargetURL: valid.TargetURL},
		{DraftID: valid.DraftID, HookType: valid.HookType, Disclosure: valid.Disclosure, CTAText: valid.CTAText, TargetURL: valid.TargetURL},
		{DraftID: valid.DraftID, ChannelID: valid.ChannelID, HookType: "invalid", Disclosure: valid.Disclosure, CTAText: valid.CTAText, TargetURL: valid.TargetURL},
		{DraftID: valid.DraftID, ChannelID: valid.ChannelID, HookType: valid.HookType, CTAText: valid.CTAText, TargetURL: valid.TargetURL},
		{DraftID: valid.DraftID, ChannelID: valid.ChannelID, HookType: valid.HookType, Disclosure: valid.Disclosure, TargetURL: valid.TargetURL},
		{DraftID: valid.DraftID, ChannelID: valid.ChannelID, HookType: valid.HookType, Disclosure: valid.Disclosure, CTAText: valid.CTAText},
	}
	for _, hook := range invalid {
		if _, err := repo.Create(context.Background(), hook); err == nil {
			t.Fatalf("Create expected validation error for %#v", hook)
		}
	}
	if _, err := repo.GetByID(context.Background(), 0); err == nil {
		t.Fatalf("GetByID expected id validation error")
	}
	if _, err := repo.ListByDraftID(context.Background(), 0, 10); err == nil {
		t.Fatalf("ListByDraftID expected draft id validation error")
	}
	if _, err := repo.ListByDraftID(context.Background(), 1, 0); err == nil {
		t.Fatalf("ListByDraftID expected limit validation error")
	}
}

func TestClusterEventRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewClusterEventRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.ClusterEvent{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	invalidRawItemID := int64(0)
	invalidAssetID := int64(0)
	validRawItemID := int64(1)
	validAssetID := int64(2)
	if _, err := repo.Create(context.Background(), domain.ClusterEvent{
		StoryClusterID: 1,
		EventType:      domain.ClusterEventTypeSignalAdded,
		EventTime:      time.Now().UTC(),
		RawItemID:      &invalidRawItemID,
	}); err == nil {
		t.Fatalf("Create expected raw item id validation error")
	}
	if _, err := repo.Create(context.Background(), domain.ClusterEvent{
		StoryClusterID: 1,
		EventType:      domain.ClusterEventTypeAssetAdded,
		EventTime:      time.Now().UTC(),
		AssetID:        &invalidAssetID,
	}); err == nil {
		t.Fatalf("Create expected asset id validation error")
	}
	if _, err := repo.Create(context.Background(), domain.ClusterEvent{
		StoryClusterID: 1,
		EventType:      "invalid",
		EventTime:      time.Now().UTC(),
	}); err == nil {
		t.Fatalf("Create expected event type validation error")
	}
	if _, err := repo.ListByClusterID(context.Background(), 0, 10); err == nil {
		t.Fatalf("ListByClusterID expected cluster id validation error")
	}
	if _, err := repo.ListByClusterID(context.Background(), 1, 0); err == nil {
		t.Fatalf("ListByClusterID expected limit validation error")
	}
	if _, err := repo.Create(context.Background(), domain.ClusterEvent{
		StoryClusterID: 1,
		EventType:      domain.ClusterEventTypeSignalAdded,
		EventTime:      time.Now().UTC(),
	}); err == nil {
		t.Fatalf("Create expected missing raw item id validation error")
	}
	if _, err := repo.Create(context.Background(), domain.ClusterEvent{
		StoryClusterID: 1,
		EventType:      domain.ClusterEventTypeAssetAdded,
		EventTime:      time.Now().UTC(),
	}); err == nil {
		t.Fatalf("Create expected missing asset id validation error")
	}
	if _, err := repo.Create(context.Background(), domain.ClusterEvent{
		StoryClusterID: 1,
		EventType:      domain.ClusterEventTypeSignalAdded,
		EventTime:      time.Now().UTC(),
		RawItemID:      &validRawItemID,
		AssetID:        &validAssetID,
	}); err == nil {
		t.Fatalf("Create expected signal_added asset id validation error")
	}
	if _, err := repo.Create(context.Background(), domain.ClusterEvent{
		StoryClusterID: 1,
		EventType:      domain.ClusterEventTypeAssetAdded,
		EventTime:      time.Now().UTC(),
		RawItemID:      &validRawItemID,
		AssetID:        &validAssetID,
	}); err == nil {
		t.Fatalf("Create expected asset_added raw item id validation error")
	}
}

func TestRankingFeatureRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewRankingFeatureRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.RankingFeature{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	if _, err := repo.Create(context.Background(), domain.RankingFeature{
		EntityType:   "draft",
		EntityID:     1,
		FeatureName:  "score",
		FeatureValue: math.NaN(),
		CalculatedAt: time.Now().UTC(),
	}); err == nil {
		t.Fatalf("Create expected invalid feature value error")
	}
	if _, err := repo.Create(context.Background(), domain.RankingFeature{
		EntityType:   "draft",
		EntityID:     1,
		FeatureName:  "score",
		FeatureValue: math.Inf(1),
		CalculatedAt: time.Now().UTC(),
	}); err == nil {
		t.Fatalf("Create expected infinite feature value error")
	}
	if _, err := repo.ListByEntity(context.Background(), "", 1, 10); err == nil {
		t.Fatalf("ListByEntity expected entity type validation error")
	}
	if _, err := repo.ListByEntity(context.Background(), "draft", 0, 10); err == nil {
		t.Fatalf("ListByEntity expected entity id validation error")
	}
	if _, err := repo.ListByEntity(context.Background(), "draft", 1, 0); err == nil {
		t.Fatalf("ListByEntity expected limit validation error")
	}
}
