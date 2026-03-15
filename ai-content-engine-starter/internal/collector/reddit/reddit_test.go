package reddit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

func TestCollectParsesRedditListing(t *testing.T) {
	const customUserAgent = "collector-tests/1.0"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != customUserAgent {
			t.Fatalf("User-Agent = %q", got)
		}
		_, _ = w.Write([]byte(`{"data":{"children":[{"data":{"name":"t3_abc","url":"https://example.com/article","title":"AI breakthrough","selftext":"extra details","created_utc":1704164645}}]}}`))
	}))
	defer server.Close()

	collector := NewWithUserAgent(server.Client(), customUserAgent)
	items, err := collector.Collect(context.Background(), domain.Source{Endpoint: server.URL})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	if items[0].ExternalID != "t3_abc" {
		t.Fatalf("ExternalID = %q", items[0].ExternalID)
	}
	if items[0].PublishedAt == nil {
		t.Fatalf("PublishedAt expected non-nil")
	}
}

func TestCollectBuildsURLFromPermalink(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"data":{"children":[{"data":{"id":"abc","permalink":"/r/artificial/comments/abc/post","title":"Post"}}]}}`))
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
	if items[0].URL != "https://www.reddit.com/r/artificial/comments/abc/post" {
		t.Fatalf("URL = %q", items[0].URL)
	}
}

func TestCollectReturnsErrorOnNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("upstream error"))
	}))
	defer server.Close()

	collector := New(server.Client())
	_, err := collector.Collect(context.Background(), domain.Source{Endpoint: server.URL})
	if err == nil {
		t.Fatalf("Collect() expected error")
	}
	if !strings.Contains(err.Error(), "502") || !strings.Contains(err.Error(), "upstream error") {
		t.Fatalf("error = %q, expected status and response snippet", err)
	}
}

func TestCollectRejectsEmptyEndpoint(t *testing.T) {
	collector := New(nil)
	_, err := collector.Collect(context.Background(), domain.Source{})
	if err == nil {
		t.Fatalf("Collect() expected error")
	}
}

func TestCollectRejectsInvalidEndpoint(t *testing.T) {
	collector := New(nil)
	_, err := collector.Collect(context.Background(), domain.Source{Endpoint: "http://[::1"})
	if err == nil {
		t.Fatalf("Collect() expected error")
	}
}

func TestNewUsesDefaultUserAgentWhenBlank(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != defaultUserAgent {
			t.Fatalf("User-Agent = %q", got)
		}
		_, _ = w.Write([]byte(`{"data":{"children":[{"data":{"name":"t3_abc","url":"https://example.com/article","title":"AI breakthrough"}}]}}`))
	}))
	defer server.Close()

	collector := NewWithUserAgent(server.Client(), "   ")
	if _, err := collector.Collect(context.Background(), domain.Source{Endpoint: server.URL}); err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
}
