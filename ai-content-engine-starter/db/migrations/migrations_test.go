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

func TestSponsorsMigrationUpContainsTableAndIndexes(t *testing.T) {
	path := filepath.Join("000021_sponsors.up.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)

	checks := []string{
		"CREATE TABLE IF NOT EXISTS sponsors",
		"name TEXT NOT NULL",
		"status TEXT NOT NULL",
		"CHECK (status IN ('active', 'inactive'))",
		"contact_info TEXT NOT NULL",
		"created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"CREATE INDEX IF NOT EXISTS idx_sponsors_status",
		"CREATE INDEX IF NOT EXISTS idx_sponsors_name",
	}
	for _, want := range checks {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration up missing %q", want)
		}
	}
}

func TestSponsorsMigrationDownDropsTable(t *testing.T) {
	path := filepath.Join("000021_sponsors.down.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)
	if !strings.Contains(sql, "DROP TABLE IF EXISTS sponsors") {
		t.Fatalf("migration down missing DROP TABLE for sponsors")
	}
}

func TestAdCampaignsMigrationUpContainsTableAndIndexes(t *testing.T) {
	path := filepath.Join("000022_ad_campaigns.up.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)

	checks := []string{
		"CREATE TABLE IF NOT EXISTS ad_campaigns",
		"sponsor_id BIGINT NOT NULL",
		"campaign_name TEXT NOT NULL",
		"campaign_type TEXT NOT NULL",
		"CHECK (campaign_type IN ('sponsored_post', 'branding'))",
		"status TEXT NOT NULL",
		"CHECK (status IN ('draft', 'active', 'paused', 'ended'))",
		"start_at TIMESTAMPTZ NOT NULL",
		"end_at TIMESTAMPTZ NOT NULL",
		"CREATE INDEX IF NOT EXISTS idx_ad_campaigns_sponsor_id",
		"CREATE INDEX IF NOT EXISTS idx_ad_campaigns_status",
	}
	for _, want := range checks {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration up missing %q", want)
		}
	}
}

func TestAdCampaignsMigrationDownDropsTable(t *testing.T) {
	path := filepath.Join("000022_ad_campaigns.down.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)
	if !strings.Contains(sql, "DROP TABLE IF EXISTS ad_campaigns") {
		t.Fatalf("migration down missing DROP TABLE for ad_campaigns")
	}
}

func TestAdSlotsMigrationUpContainsTableAndIndexes(t *testing.T) {
	path := filepath.Join("000023_ad_slots.up.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)

	checks := []string{
		"CREATE TABLE IF NOT EXISTS ad_slots",
		"channel_id BIGINT NOT NULL",
		"scheduled_at TIMESTAMPTZ NOT NULL",
		"slot_type TEXT NOT NULL",
		"CHECK (slot_type IN ('sponsored_post', 'branding'))",
		"campaign_id BIGINT NOT NULL",
		"status TEXT NOT NULL",
		"CHECK (status IN ('scheduled', 'cancelled'))",
		"CREATE INDEX IF NOT EXISTS idx_ad_slots_channel_id",
		"CREATE INDEX IF NOT EXISTS idx_ad_slots_scheduled_at",
		"CREATE INDEX IF NOT EXISTS idx_ad_slots_campaign_id",
	}
	for _, want := range checks {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration up missing %q", want)
		}
	}
}

func TestAdSlotsMigrationDownDropsTable(t *testing.T) {
	path := filepath.Join("000023_ad_slots.down.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)
	if !strings.Contains(sql, "DROP TABLE IF EXISTS ad_slots") {
		t.Fatalf("migration down missing DROP TABLE for ad_slots")
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

func TestChannelRelationshipsMigrationUpContainsTableAndIndexes(t *testing.T) {
	path := filepath.Join("000019_channel_relationships.up.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)

	checks := []string{
		"CREATE TABLE IF NOT EXISTS channel_relationships",
		"channel_id BIGINT NOT NULL",
		"related_channel_id BIGINT NOT NULL",
		"relationship_type TEXT NOT NULL",
		"CHECK (relationship_type IN ('parent', 'sibling', 'promotion_target'))",
		"strength DOUBLE PRECISION NOT NULL DEFAULT 0",
		"CHECK (strength >= 0 AND strength <= 1)",
		"metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb",
		"CHECK (channel_id <> related_channel_id)",
		"CREATE INDEX IF NOT EXISTS idx_channel_relationships_channel_id",
		"CREATE INDEX IF NOT EXISTS idx_channel_relationships_related_channel_id",
		"CREATE INDEX IF NOT EXISTS idx_channel_relationships_type",
	}
	for _, want := range checks {
		if !strings.Contains(sql, want) {
			t.Fatalf("migration up missing %q", want)
		}
	}
}

func TestChannelRelationshipsMigrationDownDropsTable(t *testing.T) {
	path := filepath.Join("000019_channel_relationships.down.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)
	if !strings.Contains(sql, "DROP TABLE IF EXISTS channel_relationships") {
		t.Fatalf("migration down missing DROP TABLE for channel_relationships")
	}
}

func TestChannelRelationshipsUniquenessMigrationUpContainsUniqueIndex(t *testing.T) {
	path := filepath.Join("000020_channel_relationships_uniqueness.up.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)
	if !strings.Contains(sql, "CREATE UNIQUE INDEX IF NOT EXISTS idx_channel_relationships_unique_link") {
		t.Fatalf("migration up missing unique index declaration")
	}
	if !strings.Contains(sql, "ON channel_relationships(channel_id, related_channel_id, relationship_type)") {
		t.Fatalf("migration up missing unique index target columns")
	}
}

func TestChannelRelationshipsUniquenessMigrationDownDropsUniqueIndex(t *testing.T) {
	path := filepath.Join("000020_channel_relationships_uniqueness.down.sql")
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	sql := string(body)
	if !strings.Contains(sql, "DROP INDEX IF EXISTS idx_channel_relationships_unique_link") {
		t.Fatalf("migration down missing DROP INDEX for channel relationship uniqueness")
	}
}
