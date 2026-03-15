package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_PORT", "")

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
}

func TestLoadInvalidPort(t *testing.T) {
	t.Setenv("HTTP_PORT", "not-a-number")

	_, err := Load()
	if err == nil {
		t.Fatalf("Load() expected error for invalid HTTP_PORT")
	}
}

func TestLoadOutOfRangePort(t *testing.T) {
	t.Setenv("HTTP_PORT", "70000")

	_, err := Load()
	if err == nil {
		t.Fatalf("Load() expected error for out-of-range HTTP_PORT")
	}
}
