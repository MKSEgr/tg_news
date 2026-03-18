package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

// ChannelRepository is a PostgreSQL implementation of domain.ChannelRepository.
type ChannelRepository struct {
	db *sql.DB
}

// SourceRepository is a PostgreSQL implementation of domain.SourceRepository.
type SourceRepository struct {
	db *sql.DB
}

// SourceItemRepository is a PostgreSQL implementation of domain.SourceItemRepository.
type SourceItemRepository struct {
	db *sql.DB
}

// DraftRepository is a PostgreSQL implementation of domain.DraftRepository.
type DraftRepository struct {
	db *sql.DB
}

// PublishIntentRepository is a PostgreSQL implementation of domain.PublishIntentRepository.
type PublishIntentRepository struct {
	db *sql.DB
}

// ContentAssetRepository is a PostgreSQL implementation of domain.ContentAssetRepository.
type ContentAssetRepository struct {
	db *sql.DB
}

// AssetRelationshipRepository is a PostgreSQL implementation of domain.AssetRelationshipRepository.
type AssetRelationshipRepository struct {
	db *sql.DB
}

// StoryClusterRepository is a PostgreSQL implementation of domain.StoryClusterRepository.
type StoryClusterRepository struct {
	db *sql.DB
}

// MonetizationHookRepository is a PostgreSQL implementation of domain.MonetizationHookRepository.
type MonetizationHookRepository struct {
	db *sql.DB
}

// ClusterEventRepository is a PostgreSQL implementation of domain.ClusterEventRepository.
type ClusterEventRepository struct {
	db *sql.DB
}

// TopicMemoryRepository is a PostgreSQL implementation of domain.TopicMemoryRepository.
type TopicMemoryRepository struct {
	db *sql.DB
}

// ContentRuleRepository is a PostgreSQL implementation of domain.ContentRuleRepository.
type ContentRuleRepository struct {
	db *sql.DB
}

// PerformanceFeedbackRepository is a PostgreSQL implementation of domain.PerformanceFeedbackRepository.
type PerformanceFeedbackRepository struct {
	db *sql.DB
}

func NewChannelRepository(db *sql.DB) *ChannelRepository       { return &ChannelRepository{db: db} }
func NewSourceRepository(db *sql.DB) *SourceRepository         { return &SourceRepository{db: db} }
func NewSourceItemRepository(db *sql.DB) *SourceItemRepository { return &SourceItemRepository{db: db} }
func NewDraftRepository(db *sql.DB) *DraftRepository           { return &DraftRepository{db: db} }
func NewPublishIntentRepository(db *sql.DB) *PublishIntentRepository {
	return &PublishIntentRepository{db: db}
}
func NewContentAssetRepository(db *sql.DB) *ContentAssetRepository {
	return &ContentAssetRepository{db: db}
}
func NewAssetRelationshipRepository(db *sql.DB) *AssetRelationshipRepository {
	return &AssetRelationshipRepository{db: db}
}
func NewStoryClusterRepository(db *sql.DB) *StoryClusterRepository {
	return &StoryClusterRepository{db: db}
}
func NewMonetizationHookRepository(db *sql.DB) *MonetizationHookRepository {
	return &MonetizationHookRepository{db: db}
}
func NewClusterEventRepository(db *sql.DB) *ClusterEventRepository {
	return &ClusterEventRepository{db: db}
}
func NewTopicMemoryRepository(db *sql.DB) *TopicMemoryRepository {
	return &TopicMemoryRepository{db: db}
}
func NewContentRuleRepository(db *sql.DB) *ContentRuleRepository {
	return &ContentRuleRepository{db: db}
}
func NewPerformanceFeedbackRepository(db *sql.DB) *PerformanceFeedbackRepository {
	return &PerformanceFeedbackRepository{db: db}
}

func ensureDB(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("postgres repository: nil db")
	}
	return nil
}

func (r *ChannelRepository) Create(ctx context.Context, channel domain.Channel) (domain.Channel, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.Channel{}, err
	}

	const q = `INSERT INTO channels (slug, name) VALUES ($1, $2) RETURNING id, slug, name, created_at`
	row := r.db.QueryRowContext(ctx, q, channel.Slug, channel.Name)
	if err := row.Scan(&channel.ID, &channel.Slug, &channel.Name, &channel.CreatedAt); err != nil {
		return domain.Channel{}, fmt.Errorf("create channel: %w", err)
	}
	return channel, nil
}

func (r *ChannelRepository) GetByID(ctx context.Context, id int64) (domain.Channel, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.Channel{}, err
	}

	const q = `SELECT id, slug, name, created_at FROM channels WHERE id = $1`
	var channel domain.Channel
	row := r.db.QueryRowContext(ctx, q, id)
	if err := row.Scan(&channel.ID, &channel.Slug, &channel.Name, &channel.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Channel{}, domain.ErrNotFound
		}
		return domain.Channel{}, fmt.Errorf("get channel by id: %w", err)
	}
	return channel, nil
}

func (r *ChannelRepository) List(ctx context.Context) ([]domain.Channel, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}

	const q = `SELECT id, slug, name, created_at FROM channels ORDER BY id`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	defer rows.Close()

	channels := make([]domain.Channel, 0)
	for rows.Next() {
		var channel domain.Channel
		if err := rows.Scan(&channel.ID, &channel.Slug, &channel.Name, &channel.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		channels = append(channels, channel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate channels: %w", err)
	}
	return channels, nil
}

func (r *SourceRepository) Create(ctx context.Context, source domain.Source) (domain.Source, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.Source{}, err
	}

	const q = `INSERT INTO sources (kind, name, endpoint, enabled) VALUES ($1, $2, $3, $4) RETURNING id, kind, name, endpoint, enabled, created_at`
	row := r.db.QueryRowContext(ctx, q, source.Kind, source.Name, source.Endpoint, source.Enabled)
	if err := row.Scan(&source.ID, &source.Kind, &source.Name, &source.Endpoint, &source.Enabled, &source.CreatedAt); err != nil {
		return domain.Source{}, fmt.Errorf("create source: %w", err)
	}
	return source, nil
}

func (r *SourceRepository) GetByID(ctx context.Context, id int64) (domain.Source, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.Source{}, err
	}

	const q = `SELECT id, kind, name, endpoint, enabled, created_at FROM sources WHERE id = $1`
	var source domain.Source
	row := r.db.QueryRowContext(ctx, q, id)
	if err := row.Scan(&source.ID, &source.Kind, &source.Name, &source.Endpoint, &source.Enabled, &source.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Source{}, domain.ErrNotFound
		}
		return domain.Source{}, fmt.Errorf("get source by id: %w", err)
	}
	return source, nil
}

func (r *SourceRepository) List(ctx context.Context) ([]domain.Source, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}

	const q = `SELECT id, kind, name, endpoint, enabled, created_at FROM sources ORDER BY id`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	defer rows.Close()

	sources := make([]domain.Source, 0)
	for rows.Next() {
		var source domain.Source
		if err := rows.Scan(&source.ID, &source.Kind, &source.Name, &source.Endpoint, &source.Enabled, &source.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, source)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sources: %w", err)
	}
	return sources, nil
}

func (r *SourceRepository) ListEnabled(ctx context.Context) ([]domain.Source, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}

	const q = `SELECT id, kind, name, endpoint, enabled, created_at FROM sources WHERE enabled = TRUE ORDER BY id`
	rows, err := r.db.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list enabled sources: %w", err)
	}
	defer rows.Close()

	sources := make([]domain.Source, 0)
	for rows.Next() {
		var source domain.Source
		if err := rows.Scan(&source.ID, &source.Kind, &source.Name, &source.Endpoint, &source.Enabled, &source.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan source: %w", err)
		}
		sources = append(sources, source)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sources: %w", err)
	}
	return sources, nil
}

func (r *SourceItemRepository) Create(ctx context.Context, item domain.SourceItem) (domain.SourceItem, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.SourceItem{}, err
	}

	const q = `INSERT INTO source_items (source_id, external_id, url, title, body, published_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (source_id, external_id) DO UPDATE
		SET url = EXCLUDED.url,
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			published_at = EXCLUDED.published_at,
			collected_at = NOW()
		RETURNING id, source_id, external_id, url, title, body, published_at, collected_at, created_at`
	row := r.db.QueryRowContext(ctx, q, item.SourceID, item.ExternalID, item.URL, item.Title, item.Body, item.PublishedAt)
	if err := scanSourceItem(row, &item); err != nil {
		return domain.SourceItem{}, fmt.Errorf("create source item: %w", err)
	}
	return item, nil
}

func (r *SourceItemRepository) GetByID(ctx context.Context, id int64) (domain.SourceItem, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.SourceItem{}, err
	}

	const q = `SELECT id, source_id, external_id, url, title, body, published_at, collected_at, created_at FROM source_items WHERE id = $1`
	var item domain.SourceItem
	row := r.db.QueryRowContext(ctx, q, id)
	if err := scanSourceItem(row, &item); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SourceItem{}, domain.ErrNotFound
		}
		return domain.SourceItem{}, fmt.Errorf("get source item by id: %w", err)
	}
	return item, nil
}

func (r *SourceItemRepository) ListBySourceID(ctx context.Context, sourceID int64, limit int) ([]domain.SourceItem, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}

	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}

	const q = `SELECT id, source_id, external_id, url, title, body, published_at, collected_at, created_at FROM source_items WHERE source_id = $1 ORDER BY collected_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, sourceID, limit)
	if err != nil {
		return nil, fmt.Errorf("list source items by source id: %w", err)
	}
	defer rows.Close()

	items := make([]domain.SourceItem, 0)
	for rows.Next() {
		var item domain.SourceItem
		if err := scanSourceItem(rows, &item); err != nil {
			return nil, fmt.Errorf("scan source item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate source items: %w", err)
	}
	return items, nil
}

func (r *SourceItemRepository) ListRecent(ctx context.Context, limit int) ([]domain.SourceItem, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}

	const q = `SELECT id, source_id, external_id, url, title, body, published_at, collected_at, created_at
		FROM source_items
		ORDER BY collected_at DESC, id DESC
		LIMIT $1`
	rows, err := r.db.QueryContext(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent source items: %w", err)
	}
	defer rows.Close()

	items := make([]domain.SourceItem, 0)
	for rows.Next() {
		var item domain.SourceItem
		if err := scanSourceItem(rows, &item); err != nil {
			return nil, fmt.Errorf("scan source item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate source items: %w", err)
	}
	return items, nil
}

func (r *DraftRepository) Create(ctx context.Context, draft domain.Draft) (domain.Draft, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.Draft{}, err
	}

	draft.Variant = strings.ToUpper(strings.TrimSpace(draft.Variant))
	if draft.Variant == "" {
		draft.Variant = "A"
	}
	if draft.Variant != "A" && draft.Variant != "B" {
		return domain.Draft{}, fmt.Errorf("draft variant is invalid")
	}

	const q = `INSERT INTO drafts (source_item_id, channel_id, variant, title, body, image_url, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, source_item_id, channel_id, variant, title, body, image_url, status, created_at, updated_at`
	row := r.db.QueryRowContext(ctx, q, draft.SourceItemID, draft.ChannelID, draft.Variant, draft.Title, draft.Body, draft.ImageURL, draft.Status)
	if err := row.Scan(&draft.ID, &draft.SourceItemID, &draft.ChannelID, &draft.Variant, &draft.Title, &draft.Body, &draft.ImageURL, &draft.Status, &draft.CreatedAt, &draft.UpdatedAt); err != nil {
		return domain.Draft{}, fmt.Errorf("create draft: %w", err)
	}
	return draft, nil
}

func (r *DraftRepository) GetByID(ctx context.Context, id int64) (domain.Draft, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.Draft{}, err
	}

	const q = `SELECT id, source_item_id, channel_id, variant, title, body, image_url, status, created_at, updated_at FROM drafts WHERE id = $1`
	var draft domain.Draft
	row := r.db.QueryRowContext(ctx, q, id)
	if err := row.Scan(&draft.ID, &draft.SourceItemID, &draft.ChannelID, &draft.Variant, &draft.Title, &draft.Body, &draft.ImageURL, &draft.Status, &draft.CreatedAt, &draft.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Draft{}, domain.ErrNotFound
		}
		return domain.Draft{}, fmt.Errorf("get draft by id: %w", err)
	}
	return draft, nil
}

func (r *DraftRepository) ListByStatus(ctx context.Context, status domain.DraftStatus, limit int) ([]domain.Draft, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}

	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}

	const q = `SELECT id, source_item_id, channel_id, variant, title, body, image_url, status, created_at, updated_at FROM drafts WHERE status = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, status, limit)
	if err != nil {
		return nil, fmt.Errorf("list drafts by status: %w", err)
	}
	defer rows.Close()

	drafts := make([]domain.Draft, 0)
	for rows.Next() {
		var draft domain.Draft
		if err := rows.Scan(&draft.ID, &draft.SourceItemID, &draft.ChannelID, &draft.Variant, &draft.Title, &draft.Body, &draft.ImageURL, &draft.Status, &draft.CreatedAt, &draft.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan draft: %w", err)
		}
		drafts = append(drafts, draft)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate drafts: %w", err)
	}
	return drafts, nil
}

func (r *DraftRepository) UpdateStatus(ctx context.Context, id int64, status domain.DraftStatus) error {
	if err := ensureDB(r.db); err != nil {
		return err
	}

	const q = `UPDATE drafts SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.db.ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("update draft status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("draft status rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PublishIntentRepository) Create(ctx context.Context, intent domain.PublishIntent) (domain.PublishIntent, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.PublishIntent{}, err
	}
	if intent.RawItemID <= 0 {
		return domain.PublishIntent{}, fmt.Errorf("raw item id must be greater than zero")
	}
	if intent.ChannelID <= 0 {
		return domain.PublishIntent{}, fmt.Errorf("channel id must be greater than zero")
	}
	intent.Format = strings.ToLower(strings.TrimSpace(intent.Format))
	if intent.Format == "" {
		return domain.PublishIntent{}, fmt.Errorf("intent format is empty")
	}
	if intent.Priority <= 0 {
		return domain.PublishIntent{}, fmt.Errorf("intent priority must be greater than zero")
	}
	intent.Status = domain.PublishIntentStatus(strings.TrimSpace(string(intent.Status)))
	if intent.Status == "" {
		intent.Status = domain.PublishIntentStatusPlanned
	}
	if intent.Status != domain.PublishIntentStatusPlanned && intent.Status != domain.PublishIntentStatusSkipped && intent.Status != domain.PublishIntentStatusCancelled {
		return domain.PublishIntent{}, fmt.Errorf("intent status is invalid")
	}

	const q = `INSERT INTO publish_intents (raw_item_id, channel_id, format, priority, status)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (raw_item_id, channel_id) DO UPDATE
		SET format = publish_intents.format
		RETURNING id, raw_item_id, channel_id, format, priority, status, created_at`
	row := r.db.QueryRowContext(ctx, q, intent.RawItemID, intent.ChannelID, intent.Format, intent.Priority, intent.Status)
	if err := row.Scan(&intent.ID, &intent.RawItemID, &intent.ChannelID, &intent.Format, &intent.Priority, &intent.Status, &intent.CreatedAt); err != nil {
		return domain.PublishIntent{}, fmt.Errorf("create publish intent: %w", err)
	}
	return intent, nil
}

func (r *PublishIntentRepository) UpdateStatus(ctx context.Context, id int64, status domain.PublishIntentStatus) error {
	if err := ensureDB(r.db); err != nil {
		return err
	}
	if id <= 0 {
		return fmt.Errorf("publish intent id must be greater than zero")
	}
	status = domain.PublishIntentStatus(strings.TrimSpace(string(status)))
	if status != domain.PublishIntentStatusPlanned && status != domain.PublishIntentStatusSkipped && status != domain.PublishIntentStatusCancelled {
		return fmt.Errorf("publish intent status is invalid")
	}

	const q = `UPDATE publish_intents SET status = $1 WHERE id = $2`
	res, err := r.db.ExecContext(ctx, q, status, id)
	if err != nil {
		return fmt.Errorf("update publish intent status: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("publish intent status rows affected: %w", err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PublishIntentRepository) ListByRawItemID(ctx context.Context, rawItemID int64, limit int) ([]domain.PublishIntent, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}
	if rawItemID <= 0 {
		return nil, fmt.Errorf("raw item id must be greater than zero")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}

	const q = `SELECT id, raw_item_id, channel_id, format, priority, status, created_at
		FROM publish_intents
		WHERE raw_item_id = $1
		ORDER BY id DESC
		LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, rawItemID, limit)
	if err != nil {
		return nil, fmt.Errorf("list publish intents by raw item id: %w", err)
	}
	defer rows.Close()

	intents := make([]domain.PublishIntent, 0)
	for rows.Next() {
		var intent domain.PublishIntent
		if err := rows.Scan(&intent.ID, &intent.RawItemID, &intent.ChannelID, &intent.Format, &intent.Priority, &intent.Status, &intent.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan publish intent: %w", err)
		}
		intents = append(intents, intent)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate publish intents: %w", err)
	}
	return intents, nil
}

func (r *TopicMemoryRepository) UpsertMention(ctx context.Context, memory domain.TopicMemory) (domain.TopicMemory, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.TopicMemory{}, err
	}
	if memory.ChannelID <= 0 {
		return domain.TopicMemory{}, fmt.Errorf("channel id must be greater than zero")
	}
	memory.Topic = strings.TrimSpace(memory.Topic)
	if memory.Topic == "" {
		return domain.TopicMemory{}, fmt.Errorf("topic is empty")
	}
	if memory.MentionCount <= 0 {
		return domain.TopicMemory{}, fmt.Errorf("mention count must be greater than zero")
	}
	if memory.LastSeenAt.IsZero() {
		return domain.TopicMemory{}, fmt.Errorf("last seen at is zero")
	}

	const q = `INSERT INTO topic_memory (channel_id, topic, mention_count, last_seen_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (channel_id, topic) DO UPDATE
		SET mention_count = topic_memory.mention_count + EXCLUDED.mention_count,
			last_seen_at = GREATEST(topic_memory.last_seen_at, EXCLUDED.last_seen_at),
			updated_at = NOW()
		RETURNING id, channel_id, topic, mention_count, last_seen_at, created_at, updated_at`
	row := r.db.QueryRowContext(ctx, q, memory.ChannelID, memory.Topic, memory.MentionCount, memory.LastSeenAt)
	if err := row.Scan(&memory.ID, &memory.ChannelID, &memory.Topic, &memory.MentionCount, &memory.LastSeenAt, &memory.CreatedAt, &memory.UpdatedAt); err != nil {
		return domain.TopicMemory{}, fmt.Errorf("upsert topic memory: %w", err)
	}
	return memory, nil
}

func (r *TopicMemoryRepository) ListTopByChannel(ctx context.Context, channelID int64, limit int) ([]domain.TopicMemory, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}
	if channelID <= 0 {
		return nil, fmt.Errorf("channel id must be greater than zero")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}

	const q = `SELECT id, channel_id, topic, mention_count, last_seen_at, created_at, updated_at
		FROM topic_memory
		WHERE channel_id = $1
		ORDER BY mention_count DESC, last_seen_at DESC, topic ASC
		LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, channelID, limit)
	if err != nil {
		return nil, fmt.Errorf("list topic memory by channel: %w", err)
	}
	defer rows.Close()

	out := make([]domain.TopicMemory, 0)
	for rows.Next() {
		var memory domain.TopicMemory
		if err := rows.Scan(&memory.ID, &memory.ChannelID, &memory.Topic, &memory.MentionCount, &memory.LastSeenAt, &memory.CreatedAt, &memory.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan topic memory: %w", err)
		}
		out = append(out, memory)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate topic memory: %w", err)
	}
	return out, nil
}

func (r *ContentRuleRepository) Create(ctx context.Context, rule domain.ContentRule) (domain.ContentRule, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.ContentRule{}, err
	}
	rule.Pattern = strings.TrimSpace(strings.ToLower(rule.Pattern))
	if rule.Pattern == "" {
		return domain.ContentRule{}, fmt.Errorf("rule pattern is empty")
	}
	if rule.Kind != domain.ContentRuleKindBlacklist && rule.Kind != domain.ContentRuleKindWhitelist {
		return domain.ContentRule{}, fmt.Errorf("rule kind is invalid")
	}
	const q = `INSERT INTO content_rules (channel_id, kind, pattern, enabled)
		VALUES ($1, $2, $3, $4)
		RETURNING id, channel_id, kind, pattern, enabled, created_at, updated_at`
	row := r.db.QueryRowContext(ctx, q, rule.ChannelID, rule.Kind, rule.Pattern, rule.Enabled)
	if err := scanContentRule(row, &rule); err != nil {
		return domain.ContentRule{}, fmt.Errorf("create content rule: %w", err)
	}
	return rule, nil
}

func (r *ContentRuleRepository) ListEnabled(ctx context.Context, channelID *int64) ([]domain.ContentRule, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}
	if channelID != nil && *channelID <= 0 {
		return nil, fmt.Errorf("channel id must be greater than zero")
	}

	const q = `SELECT id, channel_id, kind, pattern, enabled, created_at, updated_at
		FROM content_rules
		WHERE enabled = TRUE AND ($1::BIGINT IS NULL OR channel_id IS NULL OR channel_id = $1)
		ORDER BY kind ASC, pattern ASC`
	rows, err := r.db.QueryContext(ctx, q, channelID)
	if err != nil {
		return nil, fmt.Errorf("list enabled content rules: %w", err)
	}
	defer rows.Close()

	out := make([]domain.ContentRule, 0)
	for rows.Next() {
		var rule domain.ContentRule
		if err := scanContentRule(rows, &rule); err != nil {
			return nil, fmt.Errorf("scan content rule: %w", err)
		}
		out = append(out, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate content rules: %w", err)
	}
	return out, nil
}

func (r *PerformanceFeedbackRepository) Upsert(ctx context.Context, feedback domain.PerformanceFeedback) (domain.PerformanceFeedback, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.PerformanceFeedback{}, err
	}
	if feedback.DraftID <= 0 {
		return domain.PerformanceFeedback{}, fmt.Errorf("draft id must be greater than zero")
	}
	if feedback.ChannelID <= 0 {
		return domain.PerformanceFeedback{}, fmt.Errorf("channel id must be greater than zero")
	}
	if feedback.ViewsCount < 0 || feedback.ClicksCount < 0 || feedback.ReactionsCount < 0 || feedback.SharesCount < 0 {
		return domain.PerformanceFeedback{}, fmt.Errorf("feedback metrics must be non-negative")
	}
	if feedback.ViewsCount == 0 && (feedback.ClicksCount > 0 || feedback.ReactionsCount > 0 || feedback.SharesCount > 0) {
		return domain.PerformanceFeedback{}, fmt.Errorf("views count must be positive when engagement metrics are present")
	}
	if feedback.Score < 0 || math.IsNaN(feedback.Score) || math.IsInf(feedback.Score, 0) {
		return domain.PerformanceFeedback{}, fmt.Errorf("feedback score is invalid")
	}
	feedback.Variant = normalizeFeedbackVariant(feedback.Variant)
	if feedback.Variant != "" && feedback.Variant != "A" && feedback.Variant != "B" {
		return domain.PerformanceFeedback{}, fmt.Errorf("feedback variant is invalid")
	}

	const q = `INSERT INTO performance_feedback (draft_id, channel_id, variant, views_count, clicks_count, reactions_count, shares_count, score)
		SELECT d.id, d.channel_id, COALESCE(NULLIF($3, ''), d.variant, 'A'), $4, $5, $6, $7, $8
		FROM drafts d
		WHERE d.id = $1 AND d.channel_id = $2
		ON CONFLICT (draft_id) DO UPDATE
		SET channel_id = EXCLUDED.channel_id,
			variant = EXCLUDED.variant,
			views_count = EXCLUDED.views_count,
			clicks_count = EXCLUDED.clicks_count,
			reactions_count = EXCLUDED.reactions_count,
			shares_count = EXCLUDED.shares_count,
			score = EXCLUDED.score,
			updated_at = NOW()
		RETURNING id, draft_id, channel_id, variant, views_count, clicks_count, reactions_count, shares_count, score, created_at, updated_at`
	row := r.db.QueryRowContext(ctx, q, feedback.DraftID, feedback.ChannelID, feedback.Variant, feedback.ViewsCount, feedback.ClicksCount, feedback.ReactionsCount, feedback.SharesCount, feedback.Score)
	if err := row.Scan(&feedback.ID, &feedback.DraftID, &feedback.ChannelID, &feedback.Variant, &feedback.ViewsCount, &feedback.ClicksCount, &feedback.ReactionsCount, &feedback.SharesCount, &feedback.Score, &feedback.CreatedAt, &feedback.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.PerformanceFeedback{}, domain.ErrNotFound
		}
		return domain.PerformanceFeedback{}, fmt.Errorf("upsert performance feedback: %w", err)
	}
	return feedback, nil
}

func (r *PerformanceFeedbackRepository) GetByDraftID(ctx context.Context, draftID int64) (domain.PerformanceFeedback, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.PerformanceFeedback{}, err
	}
	if draftID <= 0 {
		return domain.PerformanceFeedback{}, fmt.Errorf("draft id must be greater than zero")
	}

	const q = `SELECT id, draft_id, channel_id, variant, views_count, clicks_count, reactions_count, shares_count, score, created_at, updated_at
		FROM performance_feedback WHERE draft_id = $1`
	var feedback domain.PerformanceFeedback
	row := r.db.QueryRowContext(ctx, q, draftID)
	if err := row.Scan(&feedback.ID, &feedback.DraftID, &feedback.ChannelID, &feedback.Variant, &feedback.ViewsCount, &feedback.ClicksCount, &feedback.ReactionsCount, &feedback.SharesCount, &feedback.Score, &feedback.CreatedAt, &feedback.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.PerformanceFeedback{}, domain.ErrNotFound
		}
		return domain.PerformanceFeedback{}, fmt.Errorf("get performance feedback by draft id: %w", err)
	}
	return feedback, nil
}

func normalizeFeedbackVariant(raw string) string {
	return strings.ToUpper(strings.TrimSpace(raw))
}

type sourceItemScanner interface {
	Scan(dest ...any) error
}

func scanClusterEvent(scanner sourceItemScanner, event *domain.ClusterEvent) error {
	var rawItemID sql.NullInt64
	var assetID sql.NullInt64
	if err := scanner.Scan(
		&event.ID,
		&event.StoryClusterID,
		&rawItemID,
		&assetID,
		&event.EventType,
		&event.EventTime,
		&event.MetadataJSON,
		&event.CreatedAt,
	); err != nil {
		return err
	}

	event.RawItemID = nil
	if rawItemID.Valid {
		value := rawItemID.Int64
		event.RawItemID = &value
	}
	event.AssetID = nil
	if assetID.Valid {
		value := assetID.Int64
		event.AssetID = &value
	}
	return nil
}

func scanContentRule(scanner sourceItemScanner, rule *domain.ContentRule) error {
	var channelID sql.NullInt64
	if err := scanner.Scan(&rule.ID, &channelID, &rule.Kind, &rule.Pattern, &rule.Enabled, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
		return err
	}
	rule.ChannelID = nil
	if channelID.Valid {
		value := channelID.Int64
		rule.ChannelID = &value
	}
	return nil
}

func scanSourceItem(scanner sourceItemScanner, item *domain.SourceItem) error {
	var body sql.NullString
	var publishedAt sql.NullTime

	if err := scanner.Scan(
		&item.ID,
		&item.SourceID,
		&item.ExternalID,
		&item.URL,
		&item.Title,
		&body,
		&publishedAt,
		&item.CollectedAt,
		&item.CreatedAt,
	); err != nil {
		return err
	}

	item.Body = nil
	if body.Valid {
		value := body.String
		item.Body = &value
	}

	item.PublishedAt = nil
	if publishedAt.Valid {
		value := publishedAt.Time
		item.PublishedAt = &value
	}

	return nil
}

func (r *ContentAssetRepository) Create(ctx context.Context, asset domain.ContentAsset) (domain.ContentAsset, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.ContentAsset{}, err
	}
	if asset.RawItemID <= 0 {
		return domain.ContentAsset{}, fmt.Errorf("raw item id must be greater than zero")
	}
	if asset.ChannelID <= 0 {
		return domain.ContentAsset{}, fmt.Errorf("channel id must be greater than zero")
	}
	asset.AssetType = strings.ToLower(strings.TrimSpace(asset.AssetType))
	if asset.AssetType == "" {
		return domain.ContentAsset{}, fmt.Errorf("asset type is empty")
	}
	asset.Status = domain.ContentAssetStatus(strings.TrimSpace(string(asset.Status)))
	if asset.Status == "" {
		asset.Status = domain.ContentAssetStatusPending
	}
	if asset.Status != domain.ContentAssetStatusPending {
		return domain.ContentAsset{}, fmt.Errorf("asset status is invalid")
	}

	const q = `INSERT INTO content_assets (raw_item_id, channel_id, asset_type, title, body, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, raw_item_id, channel_id, asset_type, title, body, status, created_at, updated_at`
	row := r.db.QueryRowContext(ctx, q, asset.RawItemID, asset.ChannelID, asset.AssetType, asset.Title, asset.Body, asset.Status)
	if err := row.Scan(&asset.ID, &asset.RawItemID, &asset.ChannelID, &asset.AssetType, &asset.Title, &asset.Body, &asset.Status, &asset.CreatedAt, &asset.UpdatedAt); err != nil {
		return domain.ContentAsset{}, fmt.Errorf("create content asset: %w", err)
	}
	return asset, nil
}

func (r *ContentAssetRepository) GetByID(ctx context.Context, id int64) (domain.ContentAsset, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.ContentAsset{}, err
	}
	if id <= 0 {
		return domain.ContentAsset{}, fmt.Errorf("content asset id must be greater than zero")
	}

	const q = `SELECT id, raw_item_id, channel_id, asset_type, title, body, status, created_at, updated_at
		FROM content_assets WHERE id = $1`
	var asset domain.ContentAsset
	row := r.db.QueryRowContext(ctx, q, id)
	if err := row.Scan(&asset.ID, &asset.RawItemID, &asset.ChannelID, &asset.AssetType, &asset.Title, &asset.Body, &asset.Status, &asset.CreatedAt, &asset.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ContentAsset{}, domain.ErrNotFound
		}
		return domain.ContentAsset{}, fmt.Errorf("get content asset by id: %w", err)
	}
	return asset, nil
}

func (r *ContentAssetRepository) ListByRawItemID(ctx context.Context, rawItemID int64, limit int) ([]domain.ContentAsset, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}
	if rawItemID <= 0 {
		return nil, fmt.Errorf("raw item id must be greater than zero")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}

	const q = `SELECT id, raw_item_id, channel_id, asset_type, title, body, status, created_at, updated_at
		FROM content_assets
		WHERE raw_item_id = $1
		ORDER BY id DESC
		LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, rawItemID, limit)
	if err != nil {
		return nil, fmt.Errorf("list content assets by raw item id: %w", err)
	}
	defer rows.Close()

	assets := make([]domain.ContentAsset, 0)
	for rows.Next() {
		var asset domain.ContentAsset
		if err := rows.Scan(&asset.ID, &asset.RawItemID, &asset.ChannelID, &asset.AssetType, &asset.Title, &asset.Body, &asset.Status, &asset.CreatedAt, &asset.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan content asset: %w", err)
		}
		assets = append(assets, asset)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate content assets: %w", err)
	}
	return assets, nil
}

func (r *AssetRelationshipRepository) Create(ctx context.Context, rel domain.AssetRelationship) (domain.AssetRelationship, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.AssetRelationship{}, err
	}
	if rel.FromAssetID <= 0 {
		return domain.AssetRelationship{}, fmt.Errorf("from asset id must be greater than zero")
	}
	if rel.ToAssetID <= 0 {
		return domain.AssetRelationship{}, fmt.Errorf("to asset id must be greater than zero")
	}
	if rel.FromAssetID == rel.ToAssetID {
		return domain.AssetRelationship{}, fmt.Errorf("from and to asset ids must differ")
	}
	rel.RelationshipType = domain.AssetRelationshipType(strings.ToLower(strings.TrimSpace(string(rel.RelationshipType))))
	if rel.RelationshipType != domain.AssetRelationshipTypeDerivedFrom && rel.RelationshipType != domain.AssetRelationshipTypeFollowupTo {
		return domain.AssetRelationship{}, fmt.Errorf("relationship type is invalid")
	}

	const q = `INSERT INTO asset_relationships (from_asset_id, to_asset_id, relationship_type)
		VALUES ($1, $2, $3)
		RETURNING id, from_asset_id, to_asset_id, relationship_type, created_at`
	row := r.db.QueryRowContext(ctx, q, rel.FromAssetID, rel.ToAssetID, rel.RelationshipType)
	if err := row.Scan(&rel.ID, &rel.FromAssetID, &rel.ToAssetID, &rel.RelationshipType, &rel.CreatedAt); err != nil {
		return domain.AssetRelationship{}, fmt.Errorf("create asset relationship: %w", err)
	}
	return rel, nil
}

func (r *AssetRelationshipRepository) ListByAssetID(ctx context.Context, assetID int64, limit int) ([]domain.AssetRelationship, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}
	if assetID <= 0 {
		return nil, fmt.Errorf("asset id must be greater than zero")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}

	const q = `SELECT id, from_asset_id, to_asset_id, relationship_type, created_at
		FROM asset_relationships
		WHERE from_asset_id = $1 OR to_asset_id = $1
		ORDER BY id DESC
		LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, assetID, limit)
	if err != nil {
		return nil, fmt.Errorf("list asset relationships by asset id: %w", err)
	}
	defer rows.Close()

	rels := make([]domain.AssetRelationship, 0)
	for rows.Next() {
		var rel domain.AssetRelationship
		if err := rows.Scan(&rel.ID, &rel.FromAssetID, &rel.ToAssetID, &rel.RelationshipType, &rel.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan asset relationship: %w", err)
		}
		rels = append(rels, rel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate asset relationships: %w", err)
	}
	return rels, nil
}

func (r *StoryClusterRepository) Create(ctx context.Context, cluster domain.StoryCluster) (domain.StoryCluster, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.StoryCluster{}, err
	}
	cluster.ClusterKey = strings.ToLower(strings.TrimSpace(cluster.ClusterKey))
	if cluster.ClusterKey == "" {
		return domain.StoryCluster{}, fmt.Errorf("cluster key is empty")
	}
	cluster.Title = strings.TrimSpace(cluster.Title)
	cluster.Summary = strings.TrimSpace(cluster.Summary)

	const q = `INSERT INTO story_clusters (cluster_key, title, summary)
		VALUES ($1, $2, $3)
		RETURNING id, cluster_key, title, summary, created_at, updated_at`
	row := r.db.QueryRowContext(ctx, q, cluster.ClusterKey, cluster.Title, cluster.Summary)
	if err := row.Scan(&cluster.ID, &cluster.ClusterKey, &cluster.Title, &cluster.Summary, &cluster.CreatedAt, &cluster.UpdatedAt); err != nil {
		return domain.StoryCluster{}, fmt.Errorf("create story cluster: %w", err)
	}
	return cluster, nil
}

func (r *StoryClusterRepository) GetByID(ctx context.Context, id int64) (domain.StoryCluster, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.StoryCluster{}, err
	}
	if id <= 0 {
		return domain.StoryCluster{}, fmt.Errorf("story cluster id must be greater than zero")
	}

	const q = `SELECT id, cluster_key, title, summary, created_at, updated_at FROM story_clusters WHERE id = $1`
	var cluster domain.StoryCluster
	row := r.db.QueryRowContext(ctx, q, id)
	if err := row.Scan(&cluster.ID, &cluster.ClusterKey, &cluster.Title, &cluster.Summary, &cluster.CreatedAt, &cluster.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.StoryCluster{}, domain.ErrNotFound
		}
		return domain.StoryCluster{}, fmt.Errorf("get story cluster by id: %w", err)
	}
	return cluster, nil
}

func (r *StoryClusterRepository) FindByKey(ctx context.Context, clusterKey string) (domain.StoryCluster, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.StoryCluster{}, err
	}
	clusterKey = strings.ToLower(strings.TrimSpace(clusterKey))
	if clusterKey == "" {
		return domain.StoryCluster{}, fmt.Errorf("cluster key is empty")
	}

	const q = `SELECT id, cluster_key, title, summary, created_at, updated_at FROM story_clusters WHERE cluster_key = $1`
	var cluster domain.StoryCluster
	row := r.db.QueryRowContext(ctx, q, clusterKey)
	if err := row.Scan(&cluster.ID, &cluster.ClusterKey, &cluster.Title, &cluster.Summary, &cluster.CreatedAt, &cluster.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.StoryCluster{}, domain.ErrNotFound
		}
		return domain.StoryCluster{}, fmt.Errorf("find story cluster by key: %w", err)
	}
	return cluster, nil
}

func (r *MonetizationHookRepository) Create(ctx context.Context, hook domain.MonetizationHook) (domain.MonetizationHook, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.MonetizationHook{}, err
	}
	if hook.DraftID <= 0 {
		return domain.MonetizationHook{}, fmt.Errorf("draft id must be greater than zero")
	}
	if hook.ChannelID <= 0 {
		return domain.MonetizationHook{}, fmt.Errorf("channel id must be greater than zero")
	}
	hook.HookType = domain.MonetizationHookType(strings.ToLower(strings.TrimSpace(string(hook.HookType))))
	if hook.HookType != domain.MonetizationHookTypeAffiliateCTA && hook.HookType != domain.MonetizationHookTypeSponsoredCTA {
		return domain.MonetizationHook{}, fmt.Errorf("hook type is invalid")
	}
	hook.Disclosure = strings.TrimSpace(hook.Disclosure)
	if hook.Disclosure == "" {
		return domain.MonetizationHook{}, fmt.Errorf("disclosure is empty")
	}
	hook.CTAText = strings.TrimSpace(hook.CTAText)
	if hook.CTAText == "" {
		return domain.MonetizationHook{}, fmt.Errorf("cta text is empty")
	}
	hook.TargetURL = strings.TrimSpace(hook.TargetURL)
	if hook.TargetURL == "" {
		return domain.MonetizationHook{}, fmt.Errorf("target url is empty")
	}

	const q = `INSERT INTO monetization_hooks (draft_id, channel_id, hook_type, disclosure, cta_text, target_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, draft_id, channel_id, hook_type, disclosure, cta_text, target_url, created_at, updated_at`
	row := r.db.QueryRowContext(ctx, q, hook.DraftID, hook.ChannelID, hook.HookType, hook.Disclosure, hook.CTAText, hook.TargetURL)
	if err := row.Scan(&hook.ID, &hook.DraftID, &hook.ChannelID, &hook.HookType, &hook.Disclosure, &hook.CTAText, &hook.TargetURL, &hook.CreatedAt, &hook.UpdatedAt); err != nil {
		return domain.MonetizationHook{}, fmt.Errorf("create monetization hook: %w", err)
	}
	return hook, nil
}

func (r *MonetizationHookRepository) GetByID(ctx context.Context, id int64) (domain.MonetizationHook, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.MonetizationHook{}, err
	}
	if id <= 0 {
		return domain.MonetizationHook{}, fmt.Errorf("monetization hook id must be greater than zero")
	}

	const q = `SELECT id, draft_id, channel_id, hook_type, disclosure, cta_text, target_url, created_at, updated_at
		FROM monetization_hooks WHERE id = $1`
	var hook domain.MonetizationHook
	row := r.db.QueryRowContext(ctx, q, id)
	if err := row.Scan(&hook.ID, &hook.DraftID, &hook.ChannelID, &hook.HookType, &hook.Disclosure, &hook.CTAText, &hook.TargetURL, &hook.CreatedAt, &hook.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.MonetizationHook{}, domain.ErrNotFound
		}
		return domain.MonetizationHook{}, fmt.Errorf("get monetization hook by id: %w", err)
	}
	return hook, nil
}

func (r *MonetizationHookRepository) ListByDraftID(ctx context.Context, draftID int64, limit int) ([]domain.MonetizationHook, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}
	if draftID <= 0 {
		return nil, fmt.Errorf("draft id must be greater than zero")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}

	const q = `SELECT id, draft_id, channel_id, hook_type, disclosure, cta_text, target_url, created_at, updated_at
		FROM monetization_hooks
		WHERE draft_id = $1
		ORDER BY id DESC
		LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, draftID, limit)
	if err != nil {
		return nil, fmt.Errorf("list monetization hooks by draft id: %w", err)
	}
	defer rows.Close()

	hooks := make([]domain.MonetizationHook, 0)
	for rows.Next() {
		var hook domain.MonetizationHook
		if err := rows.Scan(&hook.ID, &hook.DraftID, &hook.ChannelID, &hook.HookType, &hook.Disclosure, &hook.CTAText, &hook.TargetURL, &hook.CreatedAt, &hook.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan monetization hook: %w", err)
		}
		hooks = append(hooks, hook)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate monetization hooks: %w", err)
	}
	return hooks, nil
}

func (r *ClusterEventRepository) Create(ctx context.Context, event domain.ClusterEvent) (domain.ClusterEvent, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.ClusterEvent{}, err
	}
	if event.StoryClusterID <= 0 {
		return domain.ClusterEvent{}, fmt.Errorf("story cluster id must be greater than zero")
	}
	event.EventType = domain.ClusterEventType(strings.ToLower(strings.TrimSpace(string(event.EventType))))
	if event.EventType != domain.ClusterEventTypeSignalAdded && event.EventType != domain.ClusterEventTypeAssetAdded {
		return domain.ClusterEvent{}, fmt.Errorf("event type is invalid")
	}
	if event.EventTime.IsZero() {
		return domain.ClusterEvent{}, fmt.Errorf("event time is zero")
	}
	if event.RawItemID != nil && *event.RawItemID <= 0 {
		return domain.ClusterEvent{}, fmt.Errorf("raw item id must be greater than zero")
	}
	if event.AssetID != nil && *event.AssetID <= 0 {
		return domain.ClusterEvent{}, fmt.Errorf("asset id must be greater than zero")
	}
	switch event.EventType {
	case domain.ClusterEventTypeSignalAdded:
		if event.RawItemID == nil {
			return domain.ClusterEvent{}, fmt.Errorf("signal_added requires raw item id")
		}
		if event.AssetID != nil {
			return domain.ClusterEvent{}, fmt.Errorf("signal_added must not include asset id")
		}
	case domain.ClusterEventTypeAssetAdded:
		if event.AssetID == nil {
			return domain.ClusterEvent{}, fmt.Errorf("asset_added requires asset id")
		}
		if event.RawItemID != nil {
			return domain.ClusterEvent{}, fmt.Errorf("asset_added must not include raw item id")
		}
	}
	event.MetadataJSON = strings.TrimSpace(event.MetadataJSON)
	if event.MetadataJSON == "" {
		event.MetadataJSON = "{}"
	}

	const q = `INSERT INTO cluster_events (story_cluster_id, raw_item_id, asset_id, event_type, event_time, metadata_json)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb)
		RETURNING id, story_cluster_id, raw_item_id, asset_id, event_type, event_time, metadata_json, created_at`
	row := r.db.QueryRowContext(ctx, q, event.StoryClusterID, event.RawItemID, event.AssetID, event.EventType, event.EventTime, event.MetadataJSON)
	if err := scanClusterEvent(row, &event); err != nil {
		return domain.ClusterEvent{}, fmt.Errorf("create cluster event: %w", err)
	}
	return event, nil
}

func (r *ClusterEventRepository) ListByClusterID(ctx context.Context, storyClusterID int64, limit int) ([]domain.ClusterEvent, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}
	if storyClusterID <= 0 {
		return nil, fmt.Errorf("story cluster id must be greater than zero")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}

	const q = `SELECT id, story_cluster_id, raw_item_id, asset_id, event_type, event_time, metadata_json, created_at
		FROM cluster_events
		WHERE story_cluster_id = $1
		ORDER BY event_time ASC, id ASC
		LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, storyClusterID, limit)
	if err != nil {
		return nil, fmt.Errorf("list cluster events by cluster id: %w", err)
	}
	defer rows.Close()

	events := make([]domain.ClusterEvent, 0)
	for rows.Next() {
		var event domain.ClusterEvent
		if err := scanClusterEvent(rows, &event); err != nil {
			return nil, fmt.Errorf("scan cluster event: %w", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cluster events: %w", err)
	}
	return events, nil
}

type RankingFeatureRepository struct{ db *sql.DB }

func NewRankingFeatureRepository(db *sql.DB) *RankingFeatureRepository {
	return &RankingFeatureRepository{db: db}
}

func (r *RankingFeatureRepository) Create(ctx context.Context, feature domain.RankingFeature) (domain.RankingFeature, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.RankingFeature{}, err
	}
	feature.EntityType = strings.ToLower(strings.TrimSpace(feature.EntityType))
	if feature.EntityType == "" {
		return domain.RankingFeature{}, fmt.Errorf("entity type is empty")
	}
	if feature.EntityID <= 0 {
		return domain.RankingFeature{}, fmt.Errorf("entity id must be greater than zero")
	}
	feature.FeatureName = strings.TrimSpace(feature.FeatureName)
	if feature.FeatureName == "" {
		return domain.RankingFeature{}, fmt.Errorf("feature name is empty")
	}
	if math.IsNaN(feature.FeatureValue) || math.IsInf(feature.FeatureValue, 0) {
		return domain.RankingFeature{}, fmt.Errorf("feature value is invalid")
	}
	if feature.CalculatedAt.IsZero() {
		return domain.RankingFeature{}, fmt.Errorf("calculated at is zero")
	}

	const q = `INSERT INTO ranking_features (entity_type, entity_id, feature_name, feature_value, calculated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, entity_type, entity_id, feature_name, feature_value, calculated_at`
	row := r.db.QueryRowContext(ctx, q, feature.EntityType, feature.EntityID, feature.FeatureName, feature.FeatureValue, feature.CalculatedAt)
	if err := row.Scan(&feature.ID, &feature.EntityType, &feature.EntityID, &feature.FeatureName, &feature.FeatureValue, &feature.CalculatedAt); err != nil {
		return domain.RankingFeature{}, fmt.Errorf("create ranking feature: %w", err)
	}
	return feature, nil
}

func (r *RankingFeatureRepository) ListByEntity(ctx context.Context, entityType string, entityID int64, limit int) ([]domain.RankingFeature, error) {
	if err := ensureDB(r.db); err != nil {
		return nil, err
	}
	entityType = strings.ToLower(strings.TrimSpace(entityType))
	if entityType == "" {
		return nil, fmt.Errorf("entity type is empty")
	}
	if entityID <= 0 {
		return nil, fmt.Errorf("entity id must be greater than zero")
	}
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be greater than zero")
	}

	const q = `SELECT id, entity_type, entity_id, feature_name, feature_value, calculated_at
		FROM ranking_features
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY calculated_at DESC, id DESC
		LIMIT $3`
	rows, err := r.db.QueryContext(ctx, q, entityType, entityID, limit)
	if err != nil {
		return nil, fmt.Errorf("list ranking features by entity: %w", err)
	}
	defer rows.Close()

	features := make([]domain.RankingFeature, 0)
	for rows.Next() {
		var feature domain.RankingFeature
		if err := rows.Scan(&feature.ID, &feature.EntityType, &feature.EntityID, &feature.FeatureName, &feature.FeatureValue, &feature.CalculatedAt); err != nil {
			return nil, fmt.Errorf("scan ranking feature: %w", err)
		}
		features = append(features, feature)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ranking features: %w", err)
	}
	return features, nil
}
