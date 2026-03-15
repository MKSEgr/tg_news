package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("POSTGRES_DSN", "postgres://localhost:5432/app")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AppEnv != defaultAppEnv {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, defaultAppEnv)
	}
	if cfg.HTTPPort != defaultHTTPPort {
		t.Fatalf("HTTPPort = %d, want %d", cfg.HTTPPort, defaultHTTPPort)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("HTTP_PORT", "9090")
	t.Setenv("POSTGRES_DSN", "postgres://localhost:5432/app")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AppEnv != "production" {
		t.Fatalf("AppEnv = %q, want %q", cfg.AppEnv, "production")
	}
	if cfg.HTTPPort != 9090 {
		t.Fatalf("HTTPPort = %d, want %d", cfg.HTTPPort, 9090)
	}
	if cfg.PostgresDSN != "postgres://localhost:5432/app" {
		t.Fatalf("PostgresDSN = %q, want %q", cfg.PostgresDSN, "postgres://localhost:5432/app")
	}
}

func TestLoadInvalidPort(t *testing.T) {
	t.Setenv("HTTP_PORT", "not-a-number")
	t.Setenv("POSTGRES_DSN", "postgres://localhost:5432/app")

	_, err := Load()
	if err == nil {
		t.Fatalf("Load() expected error for invalid HTTP_PORT")
	}
}

func TestLoadOutOfRangePort(t *testing.T) {
	t.Setenv("HTTP_PORT", "70000")
	t.Setenv("POSTGRES_DSN", "postgres://localhost:5432/app")

	_, err := Load()
	if err == nil {
		t.Fatalf("Load() expected error for out-of-range HTTP_PORT")
	}
}

func TestLoadMissingPostgresDSN(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "")

	_, err := Load()
	if err == nil {
		t.Fatalf("Load() expected error when POSTGRES_DSN is empty")
	}
}

func TestLoadWhitespacePostgresDSN(t *testing.T) {
	t.Setenv("POSTGRES_DSN", "   ")

	_, err := Load()
	if err == nil {
		t.Fatalf("Load() expected error when POSTGRES_DSN is whitespace")
	}
}
