package app

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	dbmigrations "ai-content-engine-starter/db/migrations"
)

func applyStartupMigrations(ctx context.Context, db *sql.DB) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	if _, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`); err != nil {
		return fmt.Errorf("create schema migrations table: %w", err)
	}

	appliedRows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("list applied migrations: %w", err)
	}
	defer appliedRows.Close()

	applied := make(map[string]struct{})
	for appliedRows.Next() {
		var version string
		if err := appliedRows.Scan(&version); err != nil {
			return fmt.Errorf("scan applied migration: %w", err)
		}
		applied[strings.TrimSpace(version)] = struct{}{}
	}
	if err := appliedRows.Err(); err != nil {
		return fmt.Errorf("iterate applied migrations: %w", err)
	}

	files, err := dbmigrations.UpFileNames()
	if err != nil {
		return fmt.Errorf("list embedded migrations: %w", err)
	}

	for _, name := range files {
		if _, ok := applied[name]; ok {
			continue
		}
		body, err := dbmigrations.Files.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}

	return nil
}
