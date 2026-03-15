package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-content-engine-starter/internal/domain"
)

const kind = "github"

// Collector collects source items from GitHub API endpoints.
type Collector struct {
	httpClient *http.Client
}

// New creates a GitHub collector.
func New(httpClient *http.Client) *Collector {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Collector{httpClient: httpClient}
}

// Kind returns supported source kind.
func (c *Collector) Kind() string { return kind }

// Collect fetches and maps items from a GitHub API endpoint.
func (c *Collector) Collect(ctx context.Context, source domain.Source) ([]domain.SourceItem, error) {
	endpoint := strings.TrimSpace(source.Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("source endpoint is empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch github: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read github body: %w", err)
	}

	apiItems, err := parseItems(body)
	if err != nil {
		return nil, fmt.Errorf("parse github payload: %w", err)
	}

	items := make([]domain.SourceItem, 0, len(apiItems))
	for _, apiItem := range apiItems {
		externalID := parseID(apiItem.ID)
		url := strings.TrimSpace(apiItem.HTMLURL)
		title := strings.TrimSpace(apiItem.Title)
		if title == "" {
			title = strings.TrimSpace(apiItem.Name)
		}
		if externalID == "" || url == "" || title == "" {
			continue
		}

		var bodyPtr *string
		if text := strings.TrimSpace(apiItem.Body); text != "" {
			value := text
			bodyPtr = &value
		}

		item := domain.SourceItem{
			ExternalID: externalID,
			URL:        url,
			Title:      title,
			Body:       bodyPtr,
		}
		if publishedAt := parsePublishedAt(apiItem.UpdatedAt, apiItem.CreatedAt); publishedAt != nil {
			item.PublishedAt = publishedAt
		}

		items = append(items, item)
	}

	return items, nil
}

type githubEnvelope struct {
	Items []githubItem `json:"items"`
}

type githubItem struct {
	ID        json.RawMessage `json:"id"`
	HTMLURL   string          `json:"html_url"`
	Title     string          `json:"title"`
	Name      string          `json:"name"`
	Body      string          `json:"body"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

func parseItems(payload []byte) ([]githubItem, error) {
	var direct []githubItem
	if err := json.Unmarshal(payload, &direct); err == nil {
		return direct, nil
	}

	var envelope githubEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, err
	}
	return envelope.Items, nil
}

func parsePublishedAt(values ...string) *time.Time {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			return &parsed
		}
	}
	return nil
}

func parseID(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return strings.TrimSpace(str)
	}

	var num int64
	if err := json.Unmarshal(raw, &num); err == nil {
		return fmt.Sprintf("%d", num)
	}

	return ""
}
