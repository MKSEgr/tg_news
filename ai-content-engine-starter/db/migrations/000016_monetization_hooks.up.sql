CREATE TABLE IF NOT EXISTS monetization_hooks (
    id BIGSERIAL PRIMARY KEY,
    draft_id BIGINT NOT NULL REFERENCES drafts(id) ON DELETE CASCADE,
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    hook_type TEXT NOT NULL CHECK (hook_type IN ('affiliate_cta', 'sponsored_cta')),
    disclosure TEXT NOT NULL,
    cta_text TEXT NOT NULL,
    target_url TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_monetization_hooks_draft_id ON monetization_hooks(draft_id);
CREATE INDEX IF NOT EXISTS idx_monetization_hooks_channel_id ON monetization_hooks(channel_id);
CREATE INDEX IF NOT EXISTS idx_monetization_hooks_type ON monetization_hooks(hook_type);
