CREATE TABLE IF NOT EXISTS ad_campaigns (
    id BIGSERIAL PRIMARY KEY,
    sponsor_id BIGINT NOT NULL REFERENCES sponsors(id) ON DELETE CASCADE,
    campaign_name TEXT NOT NULL,
    campaign_type TEXT NOT NULL CHECK (campaign_type IN ('sponsored_post', 'branding')),
    status TEXT NOT NULL CHECK (status IN ('draft', 'active', 'paused', 'ended')),
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ad_campaigns_sponsor_id ON ad_campaigns(sponsor_id);
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_status ON ad_campaigns(status);
