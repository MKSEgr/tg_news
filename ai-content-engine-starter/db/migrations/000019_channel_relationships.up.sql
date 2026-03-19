CREATE TABLE IF NOT EXISTS channel_relationships (
    id BIGSERIAL PRIMARY KEY,
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    related_channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    relationship_type TEXT NOT NULL CHECK (relationship_type IN ('parent', 'sibling', 'promotion_target')),
    strength DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (strength >= 0 AND strength <= 1),
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (channel_id <> related_channel_id)
);

CREATE INDEX IF NOT EXISTS idx_channel_relationships_channel_id ON channel_relationships(channel_id);
CREATE INDEX IF NOT EXISTS idx_channel_relationships_related_channel_id ON channel_relationships(related_channel_id);
CREATE INDEX IF NOT EXISTS idx_channel_relationships_type ON channel_relationships(relationship_type);
