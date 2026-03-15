CREATE TABLE content_rules (
    id BIGSERIAL PRIMARY KEY,
    channel_id BIGINT REFERENCES channels(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    pattern TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (kind IN ('blacklist', 'whitelist')),
    UNIQUE(channel_id, kind, pattern)
);

CREATE INDEX idx_content_rules_enabled_kind ON content_rules(enabled, kind);
