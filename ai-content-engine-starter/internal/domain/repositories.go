package domain

import (
	"context"
	"errors"
)

// ErrNotFound indicates entity absence in repository lookups.
var ErrNotFound = errors.New("domain: entity not found")

// ChannelRepository defines persistence operations for channels.
type ChannelRepository interface {
	Create(ctx context.Context, channel Channel) (Channel, error)
	GetByID(ctx context.Context, id int64) (Channel, error)
	List(ctx context.Context) ([]Channel, error)
}

// SourceRepository defines persistence operations for sources.
type SourceRepository interface {
	Create(ctx context.Context, source Source) (Source, error)
	GetByID(ctx context.Context, id int64) (Source, error)
	List(ctx context.Context) ([]Source, error)
	ListEnabled(ctx context.Context) ([]Source, error)
}

// SourceItemRepository defines persistence operations for collected source items.
type SourceItemRepository interface {
	Create(ctx context.Context, item SourceItem) (SourceItem, error)
	GetByID(ctx context.Context, id int64) (SourceItem, error)
	ListBySourceID(ctx context.Context, sourceID int64, limit int) ([]SourceItem, error)
}

// DraftRepository defines persistence operations for generated drafts.
type DraftRepository interface {
	Create(ctx context.Context, draft Draft) (Draft, error)
	GetByID(ctx context.Context, id int64) (Draft, error)
	ListByStatus(ctx context.Context, status DraftStatus, limit int) ([]Draft, error)
	UpdateStatus(ctx context.Context, id int64, status DraftStatus) error
}

// PublishIntentRepository defines persistence operations for editorial planner intents.
type PublishIntentRepository interface {
	Create(ctx context.Context, intent PublishIntent) (PublishIntent, error)
	ListByRawItemID(ctx context.Context, rawItemID int64, limit int) ([]PublishIntent, error)
	UpdateStatus(ctx context.Context, id int64, status PublishIntentStatus) error
}

// ContentAssetRepository defines persistence operations for content assets.
type ContentAssetRepository interface {
	Create(ctx context.Context, asset ContentAsset) (ContentAsset, error)
	GetByID(ctx context.Context, id int64) (ContentAsset, error)
	ListByRawItemID(ctx context.Context, rawItemID int64, limit int) ([]ContentAsset, error)
}

// TopicMemoryRepository defines persistence operations for topic memory.
type TopicMemoryRepository interface {
	UpsertMention(ctx context.Context, memory TopicMemory) (TopicMemory, error)
	ListTopByChannel(ctx context.Context, channelID int64, limit int) ([]TopicMemory, error)
}

// ContentRuleRepository defines persistence operations for blacklist/whitelist rules.
type ContentRuleRepository interface {
	Create(ctx context.Context, rule ContentRule) (ContentRule, error)
	ListEnabled(ctx context.Context, channelID *int64) ([]ContentRule, error)
}

// PerformanceFeedbackRepository defines persistence operations for feedback loop data.
type PerformanceFeedbackRepository interface {
	Upsert(ctx context.Context, feedback PerformanceFeedback) (PerformanceFeedback, error)
	GetByDraftID(ctx context.Context, draftID int64) (PerformanceFeedback, error)
}
