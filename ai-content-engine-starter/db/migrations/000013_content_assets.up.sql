CREATE TABLE IF NOT EXISTS content_assets (
    id BIGSERIAL PRIMARY KEY,
    raw_item_id BIGINT NOT NULL,
    channel_id BIGINT NOT NULL,
    asset_type TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_content_assets_raw_item_id ON content_assets(raw_item_id);
CREATE INDEX IF NOT EXISTS idx_content_assets_channel_id ON content_assets(channel_id);
CREATE INDEX IF NOT EXISTS idx_content_assets_status ON content_assets(status);
