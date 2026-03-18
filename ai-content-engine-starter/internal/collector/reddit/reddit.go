package reddit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ai-content-engine-starter/internal/domain"
)

const (
	kind             = "reddit"
	baseURL          = "https://www.reddit.com"
	defaultUserAgent = "ai-content-engine-starter/1.0"
)

// Collector collects source items from Reddit JSON endpoints.
type Collector struct {
	httpClient *http.Client
	userAgent  string
}

// New creates a Reddit collector.
func New(httpClient *http.Client) *Collector {
	return NewWithUserAgent(httpClient, "")
}

// NewWithUserAgent creates a Reddit collector with an optional custom User-Agent.
func NewWithUserAgent(httpClient *http.Client, userAgent string) *Collector {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	userAgent = strings.TrimSpace(userAgent)
	if userAgent == "" {
		userAgent = defaultUserAgent
	}

	return &Collector{httpClient: httpClient, userAgent: userAgent}
}

// Kind returns supported source kind.
func (c *Collector) Kind() string { return kind }

// Collect fetches and maps posts from a Reddit listing endpoint.
func (c *Collector) Collect(ctx context.Context, source domain.Source) ([]domain.SourceItem, error) {
	endpoint := strings.TrimSpace(source.Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("source endpoint is empty")
	}
	if _, err := url.ParseRequestURI(endpoint); err != nil {
		return nil, fmt.Errorf("invalid source endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch reddit: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read reddit body: %w", err)
	}

	listing, err := parseListing(body)
	if err != nil {
		return nil, fmt.Errorf("parse reddit payload: %w", err)
	}

	items := make([]domain.SourceItem, 0, len(listing.Data.Children))
	for _, child := range listing.Data.Children {
		post := child.Data
		title := strings.TrimSpace(post.Title)
		if title == "" {
			continue
		}

		externalID := parseExternalID(post.Name, post.ID, post.Permalink, post.URL)
		url := parseURL(post.URL, post.Permalink)
		if externalID == "" || url == "" {
			continue
		}

		var bodyPtr *string
		if text := strings.TrimSpace(post.SelfText); text != "" {
			value := text
			bodyPtr = &value
		}

		item := domain.SourceItem{
			ExternalID: externalID,
			URL:        url,
			Title:      title,
			Body:       bodyPtr,
		}
		if publishedAt := parseCreatedAt(post.CreatedUTC); publishedAt != nil {
			item.PublishedAt = publishedAt
		}

		items = append(items, item)
	}

	return items, nil
}

type listing struct {
	Data listingData `json:"data"`
}

type listingData struct {
	Children []listingChild `json:"children"`
}

type listingChild struct {
	Data postData `json:"data"`
}

type postData struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Permalink  string  `json:"permalink"`
	URL        string  `json:"url"`
	Title      string  `json:"title"`
	SelfText   string  `json:"selftext"`
	CreatedUTC float64 `json:"created_utc"`
}

func parseListing(payload []byte) (listing, error) {
	var value listing
	if err := json.Unmarshal(payload, &value); err != nil {
		return listing{}, err
	}
	return value, nil
}

func parseExternalID(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func parseURL(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
			return value
		}
		if strings.HasPrefix(value, "/") {
			return baseURL + value
		}
	}
	return ""
}

func parseCreatedAt(createdUTC float64) *time.Time {
	if createdUTC <= 0 {
		return nil
	}
	seconds := int64(createdUTC)
	publishedAt := time.Unix(seconds, 0).UTC()
	return &publishedAt
}
