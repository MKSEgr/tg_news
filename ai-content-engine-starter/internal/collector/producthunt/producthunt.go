package producthunt

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

const kind = "producthunt"

// Collector collects source items from Product Hunt API-compatible endpoints.
type Collector struct {
	httpClient *http.Client
}

// New creates a Product Hunt collector.
func New(httpClient *http.Client) *Collector {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Collector{httpClient: httpClient}
}

// Kind returns supported source kind.
func (c *Collector) Kind() string { return kind }

// Collect fetches and maps posts from a Product Hunt endpoint.
func (c *Collector) Collect(ctx context.Context, source domain.Source) ([]domain.SourceItem, error) {
	endpoint := strings.TrimSpace(source.Endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("source endpoint is empty")
	}

	parsedURL, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid source endpoint: %w", err)
	}
	requestURL := sanitizedRequestURL(parsedURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	applyAuthFromEndpoint(req, parsedURL)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch product hunt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("unexpected status code: %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read product hunt body: %w", err)
	}

	posts, err := parsePosts(body)
	if err != nil {
		return nil, fmt.Errorf("parse product hunt payload: %w", err)
	}

	items := make([]domain.SourceItem, 0, len(posts))
	for _, post := range posts {
		externalID := parseID(post.ID)
		url := strings.TrimSpace(post.URL)
		title := strings.TrimSpace(post.Name)
		if title == "" {
			title = strings.TrimSpace(post.Title)
		}
		if externalID == "" || url == "" || title == "" {
			continue
		}

		var bodyPtr *string
		if body := parseBody(post.Tagline, post.Body); body != "" {
			value := body
			bodyPtr = &value
		}

		item := domain.SourceItem{
			ExternalID: externalID,
			URL:        url,
			Title:      title,
			Body:       bodyPtr,
		}
		if publishedAt := parsePublishedAt(post.PublishedAt, post.CreatedAt); publishedAt != nil {
			item.PublishedAt = publishedAt
		}

		items = append(items, item)
	}

	return items, nil
}

type postsEnvelope struct {
	Posts []postItem `json:"posts"`
	Data  struct {
		Posts struct {
			Edges []struct {
				Node postItem `json:"node"`
			} `json:"edges"`
		} `json:"posts"`
	} `json:"data"`
}

type postItem struct {
	ID          json.RawMessage `json:"id"`
	URL         string          `json:"url"`
	Name        string          `json:"name"`
	Title       string          `json:"title"`
	Tagline     string          `json:"tagline"`
	Body        string          `json:"body"`
	CreatedAt   string          `json:"createdAt"`
	PublishedAt string          `json:"publishedAt"`
}

func parsePosts(payload []byte) ([]postItem, error) {
	var direct []postItem
	if err := json.Unmarshal(payload, &direct); err == nil {
		return direct, nil
	}

	var envelope postsEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return nil, err
	}
	if len(envelope.Posts) > 0 {
		return envelope.Posts, nil
	}

	items := make([]postItem, 0, len(envelope.Data.Posts.Edges))
	for _, edge := range envelope.Data.Posts.Edges {
		items = append(items, edge.Node)
	}
	return items, nil
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

func parseBody(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func parsePublishedAt(values ...string) *time.Time {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339, value)
		if err == nil {
			return &parsed
		}
	}
	return nil
}

func applyAuthFromEndpoint(req *http.Request, parsedURL *url.URL) {
	query := parsedURL.Query()

	authHeader := strings.TrimSpace(query.Get("auth_header"))
	authValue := strings.TrimSpace(query.Get("auth_value"))
	if authHeader != "" && authValue != "" {
		req.Header.Set(authHeader, authValue)
	}

	if strings.TrimSpace(req.Header.Get("Authorization")) != "" {
		return
	}

	token := strings.TrimSpace(query.Get("access_token"))
	if token == "" {
		token = strings.TrimSpace(query.Get("token"))
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func sanitizedRequestURL(parsedURL *url.URL) *url.URL {
	copyURL := *parsedURL
	query := copyURL.Query()
	query.Del("auth_header")
	query.Del("auth_value")
	query.Del("access_token")
	query.Del("token")
	copyURL.RawQuery = query.Encode()
	return &copyURL
}
