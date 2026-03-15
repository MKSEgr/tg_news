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
	Title        string
	Body         string
	Status       DraftStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
