CREATE TABLE performance_feedback (
    id BIGSERIAL PRIMARY KEY,
    draft_id BIGINT NOT NULL REFERENCES drafts(id) ON DELETE CASCADE,
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    views_count BIGINT NOT NULL DEFAULT 0,
    clicks_count BIGINT NOT NULL DEFAULT 0,
    reactions_count BIGINT NOT NULL DEFAULT 0,
    shares_count BIGINT NOT NULL DEFAULT 0,
    score DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(draft_id),
    CHECK (views_count >= 0),
    CHECK (clicks_count >= 0),
    CHECK (reactions_count >= 0),
    CHECK (shares_count >= 0)
);

CREATE INDEX idx_performance_feedback_channel_score ON performance_feedback(channel_id, score DESC);
