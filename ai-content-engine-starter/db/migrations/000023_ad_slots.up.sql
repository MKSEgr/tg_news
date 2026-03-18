CREATE TABLE IF NOT EXISTS ad_slots (
    id BIGSERIAL PRIMARY KEY,
    channel_id BIGINT NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    scheduled_at TIMESTAMPTZ NOT NULL,
    slot_type TEXT NOT NULL CHECK (slot_type IN ('sponsored_post', 'branding')),
    campaign_id BIGINT NOT NULL REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('scheduled', 'cancelled'))
);

CREATE INDEX IF NOT EXISTS idx_ad_slots_channel_id ON ad_slots(channel_id);
CREATE INDEX IF NOT EXISTS idx_ad_slots_scheduled_at ON ad_slots(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_ad_slots_campaign_id ON ad_slots(campaign_id);
