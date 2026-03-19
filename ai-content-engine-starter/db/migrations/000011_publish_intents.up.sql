CREATE TABLE IF NOT EXISTS publish_intents (
    id BIGSERIAL PRIMARY KEY,
    raw_item_id BIGINT NOT NULL REFERENCES source_items(id) ON DELETE CASCADE,
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE RESTRICT,
    format TEXT NOT NULL,
    priority INT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_publish_intents_raw_item_id ON publish_intents(raw_item_id);
CREATE INDEX IF NOT EXISTS idx_publish_intents_channel_id ON publish_intents(channel_id);
CREATE INDEX IF NOT EXISTS idx_publish_intents_status ON publish_intents(status);
