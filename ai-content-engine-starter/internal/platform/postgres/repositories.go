package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

// TopicMemoryRepository is a PostgreSQL implementation of domain.TopicMemoryRepository.
type TopicMemoryRepository struct {
	db *sql.DB
}

func NewChannelRepository(db *sql.DB) *ChannelRepository       { return &ChannelRepository{db: db} }
func NewSourceRepository(db *sql.DB) *SourceRepository         { return &SourceRepository{db: db} }
func NewSourceItemRepository(db *sql.DB) *SourceItemRepository { return &SourceItemRepository{db: db} }
func NewDraftRepository(db *sql.DB) *DraftRepository           { return &DraftRepository{db: db} }
func NewTopicMemoryRepository(db *sql.DB) *TopicMemoryRepository {
	return &TopicMemoryRepository{db: db}
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

func (r *DraftRepository) Create(ctx context.Context, draft domain.Draft) (domain.Draft, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.Draft{}, err
	}

	const q = `INSERT INTO drafts (source_item_id, channel_id, title, body, status) VALUES ($1, $2, $3, $4, $5) RETURNING id, source_item_id, channel_id, title, body, status, created_at, updated_at`
	row := r.db.QueryRowContext(ctx, q, draft.SourceItemID, draft.ChannelID, draft.Title, draft.Body, draft.Status)
	if err := row.Scan(&draft.ID, &draft.SourceItemID, &draft.ChannelID, &draft.Title, &draft.Body, &draft.Status, &draft.CreatedAt, &draft.UpdatedAt); err != nil {
		return domain.Draft{}, fmt.Errorf("create draft: %w", err)
	}
	return draft, nil
}

func (r *DraftRepository) GetByID(ctx context.Context, id int64) (domain.Draft, error) {
	if err := ensureDB(r.db); err != nil {
		return domain.Draft{}, err
	}

	const q = `SELECT id, source_item_id, channel_id, title, body, status, created_at, updated_at FROM drafts WHERE id = $1`
	var draft domain.Draft
	row := r.db.QueryRowContext(ctx, q, id)
	if err := row.Scan(&draft.ID, &draft.SourceItemID, &draft.ChannelID, &draft.Title, &draft.Body, &draft.Status, &draft.CreatedAt, &draft.UpdatedAt); err != nil {
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

	const q = `SELECT id, source_item_id, channel_id, title, body, status, created_at, updated_at FROM drafts WHERE status = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, q, status, limit)
	if err != nil {
		return nil, fmt.Errorf("list drafts by status: %w", err)
	}
	defer rows.Close()

	drafts := make([]domain.Draft, 0)
	for rows.Next() {
		var draft domain.Draft
		if err := rows.Scan(&draft.ID, &draft.SourceItemID, &draft.ChannelID, &draft.Title, &draft.Body, &draft.Status, &draft.CreatedAt, &draft.UpdatedAt); err != nil {
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

type sourceItemScanner interface {
	Scan(dest ...any) error
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
