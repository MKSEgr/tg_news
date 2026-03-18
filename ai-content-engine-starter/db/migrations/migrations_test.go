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
