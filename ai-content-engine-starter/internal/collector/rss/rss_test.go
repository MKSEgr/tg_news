package rss

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

func TestCollectParsesRSS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <item>
      <title>New release</title>
      <link>https://example.com/release</link>
      <guid>rel-1</guid>
      <description>Release details</description>
      <pubDate>Mon, 02 Jan 2006 15:04:05 -0700</pubDate>
    </item>
  </channel>
</rss>`))
	}))
	defer server.Close()

	collector := New(server.Client())
	items, err := collector.Collect(context.Background(), domain.Source{Endpoint: server.URL})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	if items[0].ExternalID != "rel-1" {
		t.Fatalf("ExternalID = %q, want %q", items[0].ExternalID, "rel-1")
	}
	if items[0].PublishedAt == nil {
		t.Fatalf("PublishedAt expected non-nil")
	}
}

func TestCollectFallbackExternalIDToLink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<?xml version="1.0"?><rss><channel><item><title>T</title><link>https://example.com/t</link></item></channel></rss>`))
	}))
	defer server.Close()

	collector := New(server.Client())
	items, err := collector.Collect(context.Background(), domain.Source{Endpoint: server.URL})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if items[0].ExternalID != "https://example.com/t" {
		t.Fatalf("ExternalID = %q, want fallback link", items[0].ExternalID)
	}
}

func TestCollectReturnsErrorOnNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	collector := New(server.Client())
	_, err := collector.Collect(context.Background(), domain.Source{Endpoint: server.URL})
	if err == nil {
		t.Fatalf("Collect() expected error")
	}
}

func TestCollectRejectsEmptyEndpoint(t *testing.T) {
	collector := New(nil)
	_, err := collector.Collect(context.Background(), domain.Source{})
	if err == nil {
		t.Fatalf("Collect() expected error")
	}
}
