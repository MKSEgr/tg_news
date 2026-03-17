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
