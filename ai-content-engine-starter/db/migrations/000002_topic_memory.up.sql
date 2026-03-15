CREATE TABLE topic_memory (
    id BIGSERIAL PRIMARY KEY,
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    topic TEXT NOT NULL,
    mention_count INT NOT NULL DEFAULT 0,
    last_seen_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(channel_id, topic)
);

CREATE INDEX idx_topic_memory_channel_count ON topic_memory(channel_id, mention_count DESC);
