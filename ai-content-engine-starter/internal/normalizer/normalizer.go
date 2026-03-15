package normalizer

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

// Service normalizes collected source items before downstream processing.
type Service struct{}

// New creates a normalization service.
func New() *Service { return &Service{} }

// Normalize applies minimal deterministic normalization to a source item.
func (s *Service) Normalize(item domain.SourceItem) (domain.SourceItem, error) {
	item.ExternalID = strings.TrimSpace(item.ExternalID)
	item.Title = normalizeWhitespace(item.Title)

	normalizedURL, err := normalizeURL(item.URL)
	if err != nil {
		return domain.SourceItem{}, fmt.Errorf("normalize url: %w", err)
	}
	item.URL = normalizedURL

	if item.Body != nil {
		body := normalizeWhitespace(*item.Body)
		if body == "" {
			item.Body = nil
		} else {
			item.Body = &body
		}
	}

	if item.ExternalID == "" {
		return domain.SourceItem{}, fmt.Errorf("external id is empty")
	}
	if item.Title == "" {
		return domain.SourceItem{}, fmt.Errorf("title is empty")
	}
	if item.URL == "" {
		return domain.SourceItem{}, fmt.Errorf("url is empty")
	}

	return item, nil
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func normalizeURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("url must include scheme and host")
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)
	if port := parsed.Port(); (parsed.Scheme == "http" && port == "80") || (parsed.Scheme == "https" && port == "443") {
		parsed.Host = strings.TrimSuffix(parsed.Host, ":"+port)
	}
	parsed.Fragment = ""

	if parsed.Path != "" {
		cleaned := path.Clean(parsed.Path)
		if cleaned == "." {
			cleaned = ""
		}
		parsed.Path = cleaned
	}
	if parsed.RawQuery != "" {
		parsed.RawQuery = parsed.Query().Encode()
	}

	return parsed.String(), nil
}
