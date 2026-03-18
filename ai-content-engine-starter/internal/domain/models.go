package domain

import "time"

// Channel describes a Telegram channel managed by the system.
type Channel struct {
	ID        int64
	Slug      string
	Name      string
	CreatedAt time.Time
}

// Source describes an external content source.
type Source struct {
	ID        int64
	Kind      string
	Name      string
	Endpoint  string
	Enabled   bool
	CreatedAt time.Time
}

// SourceItem is a collected raw item from a source.
type SourceItem struct {
	ID          int64
	SourceID    int64
	ExternalID  string
	URL         string
	Title       string
	Body        *string
	ImageURL    *string
	PublishedAt *time.Time
	CollectedAt time.Time
	CreatedAt   time.Time
}

// DraftStatus defines editorial workflow state for generated drafts.
type DraftStatus string

const (
	DraftStatusPending  DraftStatus = "pending"
	DraftStatusApproved DraftStatus = "approved"
	DraftStatusRejected DraftStatus = "rejected"
	DraftStatusPosted   DraftStatus = "posted"
)

// Draft is a channel-targeted generated post draft.
type Draft struct {
	ID           int64
	SourceItemID int64
	ChannelID    int64
	Variant      string
	Title        string
	Body         string
	ImageURL     *string
	Status       DraftStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// PublishIntentStatus defines a minimal, explainable lifecycle for planner output.
//
// Semantics:
//   - planned: set by editorial planner when item should proceed to downstream pipeline.
//   - skipped: set by editorial planner (or future pipeline checks) when item should not proceed.
//   - cancelled: reserved for future pipeline/operator cancellation before execution.
//
// V3 intentionally keeps this as a lightweight status model (no state machine yet).
type PublishIntentStatus string

const (
	PublishIntentStatusPlanned   PublishIntentStatus = "planned"
	PublishIntentStatusSkipped   PublishIntentStatus = "skipped"
	PublishIntentStatusCancelled PublishIntentStatus = "cancelled"
)

// PublishIntent is a separate control-layer entity produced by editorial planning.
// It decouples planner decisions from immediate draft generation and will be consumed
// by future V3 pipeline integration (V3-003) without changing V2 flow contracts.
type PublishIntent struct {
	ID        int64
	RawItemID int64
	ChannelID int64
	Format    string
	Priority  int
	Status    PublishIntentStatus
	CreatedAt time.Time
}

// ContentAssetStatus defines lifecycle state for generated content assets.
type ContentAssetStatus string

const (
	ContentAssetStatusPending ContentAssetStatus = "pending"
)

// ContentAsset stores per-item, per-channel generated assets for future content packaging.
type ContentAsset struct {
	ID        int64
	RawItemID int64
	ChannelID int64
	AssetType string
	Title     string
	Body      string
	Status    ContentAssetStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// AssetRelationshipType defines explicit supported links between assets.
type AssetRelationshipType string

const (
	AssetRelationshipTypeDerivedFrom AssetRelationshipType = "derived_from"
	AssetRelationshipTypeFollowupTo  AssetRelationshipType = "followup_to"
)

// AssetRelationship stores a direct relationship between two content assets.
type AssetRelationship struct {
	ID               int64
	FromAssetID      int64
	ToAssetID        int64
	RelationshipType AssetRelationshipType
	CreatedAt        time.Time
}

// StoryCluster groups related content under a stable cluster key.
type StoryCluster struct {
	ID         int64
	ClusterKey string
	Title      string
	Summary    string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// MonetizationHookType defines the supported lightweight monetization hook kinds.
type MonetizationHookType string

const (
	MonetizationHookTypeAffiliateCTA MonetizationHookType = "affiliate_cta"
	MonetizationHookTypeSponsoredCTA MonetizationHookType = "sponsored_cta"
)

// MonetizationHook stores a future-ready monetization attachment for one draft.
// V3 intentionally keeps this lightweight: disclosure + CTA text + target URL.
type MonetizationHook struct {
	ID         int64
	DraftID    int64
	ChannelID  int64
	HookType   MonetizationHookType
	Disclosure string
	CTAText    string
	TargetURL  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// ClusterEventType defines the supported append-only cluster event kinds.
type ClusterEventType string

const (
	ClusterEventTypeSignalAdded ClusterEventType = "signal_added"
	ClusterEventTypeAssetAdded  ClusterEventType = "asset_added"
)

// ClusterEvent stores append-only links between a story cluster and observed signals/assets.
type ClusterEvent struct {
	ID             int64
	StoryClusterID int64
	RawItemID      *int64
	AssetID        *int64
	EventType      ClusterEventType
	EventTime      time.Time
	MetadataJSON   string
	CreatedAt      time.Time
}

// TopicMemory stores deterministic per-channel topic frequency memory.
type TopicMemory struct {
	ID           int64
	ChannelID    int64
	Topic        string
	MentionCount int
	LastSeenAt   time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ContentRuleKind defines rule mode for deterministic filtering.
type ContentRuleKind string

const (
	ContentRuleKindBlacklist ContentRuleKind = "blacklist"
	ContentRuleKindWhitelist ContentRuleKind = "whitelist"
)

// ContentRule defines a simple channel-scoped or global text matching rule.
type ContentRule struct {
	ID        int64
	ChannelID *int64
	Kind      ContentRuleKind
	Pattern   string
	Enabled   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PerformanceFeedback stores explainable engagement metrics for a published draft.
type PerformanceFeedback struct {
	ID             int64
	DraftID        int64
	ChannelID      int64
	Variant        string
	ViewsCount     int64
	ClicksCount    int64
	ReactionsCount int64
	SharesCount    int64
	Score          float64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// RankingFeature stores flexible numeric ranking signals for later adaptive scoring.
type RankingFeature struct {
	ID           int64
	EntityType   string
	EntityID     int64
	FeatureName  string
	FeatureValue float64
	CalculatedAt time.Time
}
