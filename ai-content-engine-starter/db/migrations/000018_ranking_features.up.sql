CREATE TABLE IF NOT EXISTS ranking_features (
    id BIGSERIAL PRIMARY KEY,
    entity_type TEXT NOT NULL,
    entity_id BIGINT NOT NULL,
    feature_name TEXT NOT NULL,
    feature_value DOUBLE PRECISION NOT NULL,
    calculated_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ranking_features_entity ON ranking_features(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_ranking_features_feature_name ON ranking_features(feature_name);
CREATE INDEX IF NOT EXISTS idx_ranking_features_calculated_at ON ranking_features(calculated_at);
