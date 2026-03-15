package producthunt

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

func TestCollectHappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.RawQuery; got != "" {
			t.Fatalf("RawQuery = %q, want empty", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer secret-token" {
			t.Fatalf("Authorization = %q", got)
		}
		_, _ = w.Write([]byte(`{"posts":[{"id":"ph_1","url":"https://www.producthunt.com/posts/tool","name":"AI Tool","tagline":"Ship faster with AI","publishedAt":"2026-03-15T10:00:00Z"}]}`))
	}))
	defer server.Close()

	collector := New(server.Client())
	endpoint := server.URL + "?token=secret-token"
	items, err := collector.Collect(context.Background(), domain.Source{Endpoint: endpoint})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	if items[0].ExternalID != "ph_1" {
		t.Fatalf("ExternalID = %q", items[0].ExternalID)
	}
	if items[0].Body == nil || *items[0].Body != "Ship faster with AI" {
		t.Fatalf("Body = %v", items[0].Body)
	}
	if items[0].PublishedAt == nil {
		t.Fatalf("PublishedAt expected non-nil")
	}
}

func TestCollectGraphQLEndpointUsesPostWithQueryPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Method = %s, want POST", r.Method)
		}
		if got := r.Header.Get("Content-Type"); !strings.Contains(got, "application/json") {
			t.Fatalf("Content-Type = %q", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		if !strings.Contains(string(body), `"query"`) || !strings.Contains(string(body), `posts`) {
			t.Fatalf("body = %q, expected graphql query payload", string(body))
		}
		_, _ = w.Write([]byte(`{"data":{"posts":{"edges":[{"node":{"id":"ph_gql_1","url":"https://www.producthunt.com/posts/tool","name":"AI Tool","tagline":"Ship faster with AI","publishedAt":"2026-03-15T10:00:00Z"}}]}}}`))
	}))
	defer server.Close()

	collector := New(server.Client())
	items, err := collector.Collect(context.Background(), domain.Source{Endpoint: server.URL + "/v2/api/graphql"})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
}

func TestCollectMappingFromGraphQLEdges(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.RawQuery; got != "" {
			t.Fatalf("RawQuery = %q, want empty", got)
		}
		if got := r.Header.Get("X-PH-Token"); got != "abc123" {
			t.Fatalf("X-PH-Token = %q", got)
		}
		_, _ = w.Write([]byte(`{"data":{"posts":{"edges":[{"node":{"id":42,"url":"https://www.producthunt.com/posts/agent","title":"Agent","body":"Automation helper","createdAt":"2026-03-14T09:00:00Z"}}]}}}`))
	}))
	defer server.Close()

	collector := New(server.Client())
	endpoint := server.URL + "?auth_header=X-PH-Token&auth_value=abc123"
	items, err := collector.Collect(context.Background(), domain.Source{Endpoint: endpoint})
	if err != nil {
		t.Fatalf("Collect() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items len = %d, want 1", len(items))
	}
	if items[0].ExternalID != "42" {
		t.Fatalf("ExternalID = %q", items[0].ExternalID)
	}
	if items[0].Title != "Agent" {
		t.Fatalf("Title = %q", items[0].Title)
	}
	if items[0].Body == nil || *items[0].Body != "Automation helper" {
		t.Fatalf("Body = %v", items[0].Body)
	}
}

func TestCollectReturnsErrorOnNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("missing token"))
	}))
	defer server.Close()

	collector := New(server.Client())
	_, err := collector.Collect(context.Background(), domain.Source{Endpoint: server.URL})
	if err == nil {
		t.Fatalf("Collect() expected error")
	}
	if !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "missing token") {
		t.Fatalf("error = %q, expected status and body snippet", err)
	}
}

func TestCollectRejectsInvalidEndpointConfig(t *testing.T) {
	collector := New(nil)

	if _, err := collector.Collect(context.Background(), domain.Source{}); err == nil {
		t.Fatalf("Collect() expected error for empty endpoint")
	}

	if _, err := collector.Collect(context.Background(), domain.Source{Endpoint: "http://[::1"}); err == nil {
		t.Fatalf("Collect() expected error for invalid endpoint")
	}
}
