CREATE TABLE IF NOT EXISTS cluster_events (
    id BIGSERIAL PRIMARY KEY,
    story_cluster_id BIGINT NOT NULL REFERENCES story_clusters(id) ON DELETE CASCADE,
    raw_item_id BIGINT REFERENCES source_items(id) ON DELETE SET NULL,
    asset_id BIGINT REFERENCES content_assets(id) ON DELETE SET NULL,
    event_type TEXT NOT NULL CHECK (event_type IN ('signal_added', 'asset_added')),
    event_time TIMESTAMPTZ NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cluster_events_story_cluster_id ON cluster_events(story_cluster_id);
CREATE INDEX IF NOT EXISTS idx_cluster_events_event_time ON cluster_events(event_time);
CREATE INDEX IF NOT EXISTS idx_cluster_events_event_type ON cluster_events(event_type);
