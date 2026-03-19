package domain

import "testing"

func TestDraftStatusesAreStable(t *testing.T) {
	if DraftStatusPending != "pending" {
		t.Fatalf("DraftStatusPending = %q, want %q", DraftStatusPending, "pending")
	}
	if DraftStatusApproved != "approved" {
		t.Fatalf("DraftStatusApproved = %q, want %q", DraftStatusApproved, "approved")
	}
	if DraftStatusPublishing != "publishing" {
		t.Fatalf("DraftStatusPublishing = %q, want %q", DraftStatusPublishing, "publishing")
	}
	if DraftStatusRejected != "rejected" {
		t.Fatalf("DraftStatusRejected = %q, want %q", DraftStatusRejected, "rejected")
	}
	if DraftStatusPosted != "posted" {
		t.Fatalf("DraftStatusPosted = %q, want %q", DraftStatusPosted, "posted")
	}
}

func TestSourceItemAllowsNullableBody(t *testing.T) {
	item := SourceItem{}
	if item.Body != nil {
		t.Fatalf("SourceItem.Body = %v, want nil", item.Body)
	}
}

func TestContentRuleKindsAreStable(t *testing.T) {
	if ContentRuleKindBlacklist != "blacklist" {
		t.Fatalf("ContentRuleKindBlacklist = %q, want %q", ContentRuleKindBlacklist, "blacklist")
	}
	if ContentRuleKindWhitelist != "whitelist" {
		t.Fatalf("ContentRuleKindWhitelist = %q, want %q", ContentRuleKindWhitelist, "whitelist")
	}
}

func TestPerformanceFeedbackDefaults(t *testing.T) {
	feedback := PerformanceFeedback{}
	if feedback.Score != 0 {
		t.Fatalf("PerformanceFeedback.Score = %f, want 0", feedback.Score)
	}
}

func TestContentAssetStatusIsStable(t *testing.T) {
	if ContentAssetStatusPending != "pending" {
		t.Fatalf("ContentAssetStatusPending = %q, want %q", ContentAssetStatusPending, "pending")
	}
}

func TestAssetRelationshipTypesAreStable(t *testing.T) {
	if AssetRelationshipTypeDerivedFrom != "derived_from" {
		t.Fatalf("AssetRelationshipTypeDerivedFrom = %q, want %q", AssetRelationshipTypeDerivedFrom, "derived_from")
	}
	if AssetRelationshipTypeFollowupTo != "followup_to" {
		t.Fatalf("AssetRelationshipTypeFollowupTo = %q, want %q", AssetRelationshipTypeFollowupTo, "followup_to")
	}
}

func TestStoryClusterDefaults(t *testing.T) {
	cluster := StoryCluster{}
	if cluster.ClusterKey != "" {
		t.Fatalf("StoryCluster.ClusterKey = %q, want empty", cluster.ClusterKey)
	}
}

func TestMonetizationHookTypesAreStable(t *testing.T) {
	if MonetizationHookTypeAffiliateCTA != "affiliate_cta" {
		t.Fatalf("MonetizationHookTypeAffiliateCTA = %q, want %q", MonetizationHookTypeAffiliateCTA, "affiliate_cta")
	}
	if MonetizationHookTypeSponsoredCTA != "sponsored_cta" {
		t.Fatalf("MonetizationHookTypeSponsoredCTA = %q, want %q", MonetizationHookTypeSponsoredCTA, "sponsored_cta")
	}
}

func TestClusterEventTypesAreStable(t *testing.T) {
	if ClusterEventTypeSignalAdded != "signal_added" {
		t.Fatalf("ClusterEventTypeSignalAdded = %q, want %q", ClusterEventTypeSignalAdded, "signal_added")
	}
	if ClusterEventTypeAssetAdded != "asset_added" {
		t.Fatalf("ClusterEventTypeAssetAdded = %q, want %q", ClusterEventTypeAssetAdded, "asset_added")
	}
}

func TestRankingFeatureDefaults(t *testing.T) {
	feature := RankingFeature{}
	if feature.FeatureValue != 0 {
		t.Fatalf("RankingFeature.FeatureValue = %f, want 0", feature.FeatureValue)
	}
}

func TestChannelRelationshipTypesAreStable(t *testing.T) {
	if ChannelRelationshipTypeParent != "parent" {
		t.Fatalf("ChannelRelationshipTypeParent = %q, want %q", ChannelRelationshipTypeParent, "parent")
	}
	if ChannelRelationshipTypeSibling != "sibling" {
		t.Fatalf("ChannelRelationshipTypeSibling = %q, want %q", ChannelRelationshipTypeSibling, "sibling")
	}
	if ChannelRelationshipTypePromotionTarget != "promotion_target" {
		t.Fatalf("ChannelRelationshipTypePromotionTarget = %q, want %q", ChannelRelationshipTypePromotionTarget, "promotion_target")
	}
}
