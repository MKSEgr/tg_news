package migrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContentAssetsMigrationUpContainsTableAndIndexes(t *testing.T) {
	path := filepath.Join("000013_content_assets.up.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)

	checks := []string{
		"CREATE TABLE IF NOT EXISTS content_assets",
		"raw_item_id BIGINT NOT NULL",
		"channel_id BIGINT NOT NULL",
		"asset_type TEXT NOT NULL",
		"title TEXT NOT NULL",
		"body TEXT NOT NULL",
		"status TEXT NOT NULL",
		"created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"CREATE INDEX IF NOT EXISTS idx_content_assets_raw_item_id",
		"CREATE INDEX IF NOT EXISTS idx_content_assets_channel_id",
		"CREATE INDEX IF NOT EXISTS idx_content_assets_status",
	}
	for _, want := range checks {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration up missing %q", want)
		}
	}
}

func TestContentAssetsMigrationDownDropsTable(t *testing.T) {
	path := filepath.Join("000013_content_assets.down.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)
	if !strings.Contains(sql, "DROP TABLE IF EXISTS content_assets") {
		t.Fatalf("migration down missing DROP TABLE for content_assets")
	}
}

func TestAssetRelationshipsMigrationUpContainsTableAndIndexes(t *testing.T) {
	path := filepath.Join("000014_asset_relationships.up.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)

	checks := []string{
		"CREATE TABLE IF NOT EXISTS asset_relationships",
		"from_asset_id BIGINT NOT NULL",
		"to_asset_id BIGINT NOT NULL",
		"relationship_type TEXT NOT NULL",
		"CHECK (from_asset_id <> to_asset_id)",
		"CHECK (relationship_type IN ('derived_from', 'followup_to'))",
		"created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"CREATE INDEX IF NOT EXISTS idx_asset_relationships_from_asset_id",
		"CREATE INDEX IF NOT EXISTS idx_asset_relationships_to_asset_id",
		"CREATE INDEX IF NOT EXISTS idx_asset_relationships_type",
	}
	for _, want := range checks {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration up missing %q", want)
		}
	}
}

func TestAssetRelationshipsMigrationDownDropsTable(t *testing.T) {
	path := filepath.Join("000014_asset_relationships.down.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)
	if !strings.Contains(sql, "DROP TABLE IF EXISTS asset_relationships") {
		t.Fatalf("migration down missing DROP TABLE for asset_relationships")
	}
}

func TestStoryClustersMigrationUpContainsTableAndIndex(t *testing.T) {
	path := filepath.Join("000015_story_clusters.up.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)

	checks := []string{
		"CREATE TABLE IF NOT EXISTS story_clusters",
		"cluster_key TEXT NOT NULL UNIQUE",
		"title TEXT NOT NULL",
		"summary TEXT NOT NULL",
		"created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
	}
	for _, want := range checks {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration up missing %q", want)
		}
	}
}

func TestStoryClustersMigrationDownDropsTable(t *testing.T) {
	path := filepath.Join("000015_story_clusters.down.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)
	if !strings.Contains(sql, "DROP TABLE IF EXISTS story_clusters") {
		t.Fatalf("migration down missing DROP TABLE for story_clusters")
	}
}

func TestMonetizationHooksMigrationUpContainsTableAndIndexes(t *testing.T) {
	path := filepath.Join("000016_monetization_hooks.up.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)

	checks := []string{
		"CREATE TABLE IF NOT EXISTS monetization_hooks",
		"draft_id BIGINT NOT NULL",
		"channel_id BIGINT NOT NULL",
		"hook_type TEXT NOT NULL",
		"CHECK (hook_type IN ('affiliate_cta', 'sponsored_cta'))",
		"disclosure TEXT NOT NULL",
		"cta_text TEXT NOT NULL",
		"target_url TEXT NOT NULL",
		"created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"CREATE INDEX IF NOT EXISTS idx_monetization_hooks_draft_id",
		"CREATE INDEX IF NOT EXISTS idx_monetization_hooks_channel_id",
		"CREATE INDEX IF NOT EXISTS idx_monetization_hooks_type",
	}
	for _, want := range checks {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration up missing %q", want)
		}
	}
}

func TestMonetizationHooksMigrationDownDropsTable(t *testing.T) {
	path := filepath.Join("000016_monetization_hooks.down.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)
	if !strings.Contains(sql, "DROP TABLE IF EXISTS monetization_hooks") {
		t.Fatalf("migration down missing DROP TABLE for monetization_hooks")
	}
}

func TestClusterEventsMigrationUpContainsTableAndIndexes(t *testing.T) {
	path := filepath.Join("000017_cluster_events.up.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)

	checks := []string{
		"CREATE TABLE IF NOT EXISTS cluster_events",
		"story_cluster_id BIGINT NOT NULL",
		"raw_item_id BIGINT",
		"asset_id BIGINT",
		"event_type TEXT NOT NULL",
		"CHECK (event_type IN ('signal_added', 'asset_added'))",
		"event_time TIMESTAMPTZ NOT NULL",
		"metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb",
		"created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"CREATE INDEX IF NOT EXISTS idx_cluster_events_story_cluster_id",
		"CREATE INDEX IF NOT EXISTS idx_cluster_events_event_time",
		"CREATE INDEX IF NOT EXISTS idx_cluster_events_event_type",
	}
	for _, want := range checks {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration up missing %q", want)
		}
	}
}

func TestClusterEventsMigrationDownDropsTable(t *testing.T) {
	path := filepath.Join("000017_cluster_events.down.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)
	if !strings.Contains(sql, "DROP TABLE IF EXISTS cluster_events") {
		t.Fatalf("migration down missing DROP TABLE for cluster_events")
	}
}


func TestRankingFeaturesMigrationUpContainsTableAndIndexes(t *testing.T) {
	path := filepath.Join("000018_ranking_features.up.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)

	checks := []string{
		"CREATE TABLE IF NOT EXISTS ranking_features",
		"entity_type TEXT NOT NULL",
		"entity_id BIGINT NOT NULL",
		"feature_name TEXT NOT NULL",
		"feature_value DOUBLE PRECISION NOT NULL",
		"calculated_at TIMESTAMPTZ NOT NULL",
		"created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"CREATE INDEX IF NOT EXISTS idx_ranking_features_entity",
		"CREATE INDEX IF NOT EXISTS idx_ranking_features_feature_name",
		"CREATE INDEX IF NOT EXISTS idx_ranking_features_calculated_at",
	}
	for _, want := range checks {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration up missing %q", want)
		}
	}
}

func TestRankingFeaturesMigrationDownDropsTable(t *testing.T) {
	path := filepath.Join("000018_ranking_features.down.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)
	if !strings.Contains(sql, "DROP TABLE IF EXISTS ranking_features") {
		t.Fatalf("migration down missing DROP TABLE for ranking_features")
	}
}
