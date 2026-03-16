package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

func TestCollectParsesSearchItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
			t.Fatalf("Accept header = %q", got)
		}
		_, _ = w.Write([]byte(`{"items":[{"id":1,"html_url":"https://github.com/o/r/pull/1","title":"Add feature","body":"Details","updated_at":"2024-01-02T03:04:05Z"}]}`))
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
	if items[0].ExternalID != "1" {
		t.Fatalf("ExternalID = %q, want %q", items[0].ExternalID, "1")
	}
	if items[0].Title != "Add feature" {
		t.Fatalf("Title = %q", items[0].Title)
	}
	if items[0].PublishedAt == nil {
		t.Fatalf("PublishedAt expected non-nil")
	}
}

func TestCollectParsesDirectArrayAndFallsBackToName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`[{"id":"2","html_url":"https://github.com/o/r","name":"repo-name","created_at":"2024-01-02T03:04:05Z"}]`))
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
	if items[0].Title != "repo-name" {
		t.Fatalf("Title = %q, want fallback name", items[0].Title)
	}
}

func TestCollectReturnsErrorOnNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
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
