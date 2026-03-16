package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-content-engine-starter/internal/domain"
)

const kind = "rss"

// Collector collects source items from RSS feeds.
type Collector struct {
	httpClient *http.Client
}

// New creates an RSS collector.
func New(httpClient *http.Client) *Collector {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Collector{httpClient: httpClient}
}

// Kind returns supported source kind.
func (c *Collector) Kind() string { return kind }

// Collect fetches and parses items from an RSS endpoint.
func (c *Collector) Collect(ctx context.Context, source domain.Source) ([]domain.SourceItem, error) {
	if strings.TrimSpace(source.Endpoint) == "" {
		return nil, fmt.Errorf("source endpoint is empty")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.Endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch rss: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read rss body: %w", err)
	}

	feed, err := parseRSS(body)
	if err != nil {
		return nil, fmt.Errorf("parse rss: %w", err)
	}

	items := make([]domain.SourceItem, 0, len(feed.Channel.Items))
	for _, rssItem := range feed.Channel.Items {
		title := strings.TrimSpace(rssItem.Title)
		link := strings.TrimSpace(rssItem.Link)
		if title == "" || link == "" {
			continue
		}

		description := strings.TrimSpace(rssItem.Description)
		var bodyPtr *string
		if description != "" {
			value := description
			bodyPtr = &value
		}

		externalID := strings.TrimSpace(rssItem.GUID)
		if externalID == "" {
			externalID = link
		}

		item := domain.SourceItem{
			ExternalID: externalID,
			URL:        link,
			Title:      title,
			Body:       bodyPtr,
		}
		if publishedAt := parsePubDate(rssItem.PubDate); publishedAt != nil {
			item.PublishedAt = publishedAt
		}

		items = append(items, item)
	}

	return items, nil
}

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Items []rssItem `xml:"item"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	GUID        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
}

func parseRSS(payload []byte) (rssFeed, error) {
	var feed rssFeed
	if err := xml.Unmarshal(payload, &feed); err != nil {
		return rssFeed{}, err
	}
	return feed, nil
}

func parsePubDate(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	layouts := []string{time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return &parsed
		}
	}

	return nil
}
