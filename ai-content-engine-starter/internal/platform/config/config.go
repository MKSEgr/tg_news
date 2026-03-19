package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAppEnv       = "development"
	defaultHTTPPort     = 8080
	defaultLoopInterval = time.Minute
)

// Config contains application runtime settings loaded from environment.
type Config struct {
	AppEnv             string
	HTTPPort           int
	PostgresDSN        string
	RedisAddr          string
	EnablePipeline     bool
	EnablePublisher    bool
	LoopInterval       time.Duration
	YandexAIAPIKey     string
	YandexAIModelURI   string
	TelegramBotToken   string
	ChannelChatMap     map[string]string
	RecentItemsLimit   int
	DraftScanLimit     int
	PublisherBatchSize int
	Features           FeatureFlags
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
		AppEnv:             getEnvOrDefault("APP_ENV", defaultAppEnv),
		HTTPPort:           defaultHTTPPort,
		PostgresDSN:        strings.TrimSpace(os.Getenv("POSTGRES_DSN")),
		RedisAddr:          strings.TrimSpace(os.Getenv("REDIS_ADDR")),
		EnablePipeline:     getEnvBool("ENABLE_PIPELINE", false),
		EnablePublisher:    getEnvBool("ENABLE_PUBLISHER", false),
		LoopInterval:       defaultLoopInterval,
		YandexAIAPIKey:     strings.TrimSpace(os.Getenv("YANDEX_AI_API_KEY")),
		YandexAIModelURI:   strings.TrimSpace(os.Getenv("YANDEX_AI_MODEL_URI")),
		TelegramBotToken:   strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		ChannelChatMap:     parseChannelChatMap(os.Getenv("CHANNEL_CHAT_MAP")),
		RecentItemsLimit:   50,
		DraftScanLimit:     200,
		PublisherBatchSize: 50,
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

	if rawInterval := strings.TrimSpace(os.Getenv("LOOP_INTERVAL")); rawInterval != "" {
		interval, err := time.ParseDuration(rawInterval)
		if err != nil {
			return Config{}, fmt.Errorf("parse LOOP_INTERVAL: %w", err)
		}
		if interval <= 0 {
			return Config{}, fmt.Errorf("LOOP_INTERVAL must be greater than zero")
		}
		cfg.LoopInterval = interval
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
	if rawLimit := strings.TrimSpace(os.Getenv("RECENT_ITEMS_LIMIT")); rawLimit != "" {
		value, err := strconv.Atoi(rawLimit)
		if err != nil {
			return Config{}, fmt.Errorf("parse RECENT_ITEMS_LIMIT: %w", err)
		}
		if value <= 0 {
			return Config{}, fmt.Errorf("RECENT_ITEMS_LIMIT must be greater than zero")
		}
		cfg.RecentItemsLimit = value
	}
	if rawLimit := strings.TrimSpace(os.Getenv("DRAFT_SCAN_LIMIT")); rawLimit != "" {
		value, err := strconv.Atoi(rawLimit)
		if err != nil {
			return Config{}, fmt.Errorf("parse DRAFT_SCAN_LIMIT: %w", err)
		}
		if value <= 0 {
			return Config{}, fmt.Errorf("DRAFT_SCAN_LIMIT must be greater than zero")
		}
		cfg.DraftScanLimit = value
	}
	if rawLimit := strings.TrimSpace(os.Getenv("PUBLISHER_BATCH_SIZE")); rawLimit != "" {
		value, err := strconv.Atoi(rawLimit)
		if err != nil {
			return Config{}, fmt.Errorf("parse PUBLISHER_BATCH_SIZE: %w", err)
		}
		if value <= 0 {
			return Config{}, fmt.Errorf("PUBLISHER_BATCH_SIZE must be greater than zero")
		}
		cfg.PublisherBatchSize = value
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

	if cfg.EnablePipeline {
		if cfg.YandexAIAPIKey == "" {
			return Config{}, fmt.Errorf("YANDEX_AI_API_KEY is required when ENABLE_PIPELINE=true")
		}
		if cfg.YandexAIModelURI == "" {
			return Config{}, fmt.Errorf("YANDEX_AI_MODEL_URI is required when ENABLE_PIPELINE=true")
		}
	}
	if cfg.EnablePublisher && cfg.TelegramBotToken == "" {
		return Config{}, fmt.Errorf("TELEGRAM_BOT_TOKEN is required when ENABLE_PUBLISHER=true")
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

func parseChannelChatMap(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]string{}
	}
	out := map[string]string{}
	for _, pair := range strings.Split(raw, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			continue
		}
		slug := strings.TrimSpace(parts[0])
		chatID := strings.TrimSpace(parts[1])
		if slug == "" || chatID == "" {
			continue
		}
		out[slug] = chatID
	}
	return out
}
