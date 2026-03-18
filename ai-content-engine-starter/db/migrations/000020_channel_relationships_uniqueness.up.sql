CREATE UNIQUE INDEX IF NOT EXISTS idx_channel_relationships_unique_link
    ON channel_relationships(channel_id, related_channel_id, relationship_type);
