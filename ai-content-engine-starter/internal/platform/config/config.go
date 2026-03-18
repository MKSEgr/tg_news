package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	defaultAppEnv   = "development"
	defaultHTTPPort = 8080
)

// Config contains application runtime settings loaded from environment.
type Config struct {
	AppEnv      string
	HTTPPort    int
	PostgresDSN string
	RedisAddr   string
	Features    FeatureFlags
}

// FeatureFlags controls rollout of V2 features.
// Flags are disabled by default unless explicitly enabled via environment variables.
type FeatureFlags struct {
	V2Enabled         bool
	TopicMemory       bool
	ContentRules      bool
	PerformanceFeedbk bool
	ABVariants        bool
	AutoRepost        bool
	Analytics         bool
	ImageEnrichment   bool
	SourceDiscovery   bool
	AdminBot          bool
	WebUI             bool
}

// Load reads configuration from environment variables and applies defaults.
func Load() (Config, error) {
	cfg := Config{
		AppEnv:      getEnvOrDefault("APP_ENV", defaultAppEnv),
		HTTPPort:    defaultHTTPPort,
		PostgresDSN: strings.TrimSpace(os.Getenv("POSTGRES_DSN")),
		RedisAddr:   strings.TrimSpace(os.Getenv("REDIS_ADDR")),
		Features: FeatureFlags{
			V2Enabled:         getEnvBool("FEATURE_V2_ENABLED", false),
			TopicMemory:       getEnvBool("FEATURE_TOPIC_MEMORY", false),
			ContentRules:      getEnvBool("FEATURE_CONTENT_RULES", false),
			PerformanceFeedbk: getEnvBool("FEATURE_PERFORMANCE_FEEDBACK", false),
			ABVariants:        getEnvBool("FEATURE_AB_VARIANTS", false),
			AutoRepost:        getEnvBool("FEATURE_AUTO_REPOST", false),
			Analytics:         getEnvBool("FEATURE_ANALYTICS", false),
			ImageEnrichment:   getEnvBool("FEATURE_IMAGE_ENRICHMENT", false),
			SourceDiscovery:   getEnvBool("FEATURE_SOURCE_DISCOVERY", false),
			AdminBot:          getEnvBool("FEATURE_ADMIN_BOT", false),
			WebUI:             getEnvBool("FEATURE_WEB_UI", false),
		},
	}

	if rawPort := os.Getenv("HTTP_PORT"); rawPort != "" {
		port, err := strconv.Atoi(rawPort)
		if err != nil {
			return Config{}, fmt.Errorf("parse HTTP_PORT: %w", err)
		}
		if port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("HTTP_PORT out of range: %d", port)
		}
		cfg.HTTPPort = port
	}

	if cfg.PostgresDSN == "" {
		return Config{}, fmt.Errorf("POSTGRES_DSN is required")
	}
	if cfg.RedisAddr == "" {
		return Config{}, fmt.Errorf("REDIS_ADDR is required")
	}

	if err := cfg.Features.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate ensures feature toggles are configured coherently.
func (f FeatureFlags) Validate() error {
	if !f.V2Enabled {
		if f.TopicMemory || f.ContentRules || f.PerformanceFeedbk || f.ABVariants || f.AutoRepost || f.Analytics || f.ImageEnrichment || f.SourceDiscovery || f.AdminBot || f.WebUI {
			return fmt.Errorf("FEATURE_V2_ENABLED must be true when any V2 feature flag is enabled")
		}
	}
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return defaultValue
	}
	return value == "1" || value == "true" || value == "yes" || value == "on"
}
