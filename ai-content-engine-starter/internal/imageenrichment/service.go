package imageenrichment

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

var (
	markdownImagePattern = regexp.MustCompile(`!\[[^\]]*\]\(([^)\s]+)`) // ![alt](url)
	htmlImagePattern     = regexp.MustCompile(`(?i)<img[^>]+src=["']([^"']+)["']`)
	urlPattern           = regexp.MustCompile(`https?://[^\s"'<>]+`)
)

// Service selects an image URL from source-item fields using deterministic heuristics.
type Service struct{}

// New creates image enrichment service.
func New() *Service {
	return &Service{}
}

// Enrich adds ImageURL when a valid image URL can be derived.
func (s *Service) Enrich(item domain.SourceItem) (domain.SourceItem, error) {
	if s == nil {
		return domain.SourceItem{}, fmt.Errorf("image enrichment service is nil")
	}
	if item.ID <= 0 {
		return domain.SourceItem{}, fmt.Errorf("source item id is invalid")
	}

	if imageURL := firstImageURL(item); imageURL != "" {
		item.ImageURL = &imageURL
	}
	return item, nil
}

func firstImageURL(item domain.SourceItem) string {
	if direct := normalizeImageURL(item.URL); direct != "" {
		return direct
	}
	if item.Body == nil {
		return ""
	}
	body := strings.TrimSpace(*item.Body)
	if body == "" {
		return ""
	}

	if match := markdownImagePattern.FindStringSubmatch(body); len(match) > 1 {
		if image := normalizeImageURL(match[1]); image != "" {
			return image
		}
	}
	if match := htmlImagePattern.FindStringSubmatch(body); len(match) > 1 {
		if image := normalizeImageURL(match[1]); image != "" {
			return image
		}
	}

	candidates := urlPattern.FindAllString(body, -1)
	for _, candidate := range candidates {
		if image := normalizeImageURL(candidate); image != "" {
			return image
		}
	}
	return ""
}

func normalizeImageURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return ""
	}
	ext := strings.ToLower(path.Ext(parsed.Path))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		return parsed.String()
	default:
		return ""
	}
}
