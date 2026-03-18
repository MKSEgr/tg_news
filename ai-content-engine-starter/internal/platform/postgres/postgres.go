package postgres

import (
	"fmt"
	"net/url"
)

// ValidateDSN checks whether PostgreSQL DSN has required shape.
func ValidateDSN(dsn string) error {
	parsed, err := url.Parse(dsn)
	if err != nil {
		return fmt.Errorf("parse postgres dsn: %w", err)
	}

	if parsed.Scheme != "postgres" && parsed.Scheme != "postgresql" {
		return fmt.Errorf("unsupported postgres scheme: %q", parsed.Scheme)
	}

	if parsed.Hostname() == "" {
		return fmt.Errorf("postgres host is required")
	}

	if parsed.Path == "" || parsed.Path == "/" {
		return fmt.Errorf("postgres database name is required")
	}

	return nil
}
