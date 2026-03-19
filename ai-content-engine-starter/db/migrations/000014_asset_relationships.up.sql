CREATE TABLE IF NOT EXISTS asset_relationships (
    id BIGSERIAL PRIMARY KEY,
    from_asset_id BIGINT NOT NULL,
    to_asset_id BIGINT NOT NULL,
    relationship_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (from_asset_id <> to_asset_id),
    CHECK (relationship_type IN ('derived_from', 'followup_to'))
);

CREATE INDEX IF NOT EXISTS idx_asset_relationships_from_asset_id ON asset_relationships(from_asset_id);
CREATE INDEX IF NOT EXISTS idx_asset_relationships_to_asset_id ON asset_relationships(to_asset_id);
CREATE INDEX IF NOT EXISTS idx_asset_relationships_type ON asset_relationships(relationship_type);
