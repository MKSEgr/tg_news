package publisher

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

func TestPublishDraftHappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/sendMessage") {
			t.Fatalf("path = %s", r.URL.Path)
		}

		var req sendMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.ChatID != "@ai_news" {
			t.Fatalf("chat_id = %q", req.ChatID)
		}
		if req.Text != "hello" {
			t.Fatalf("text = %q", req.Text)
		}

		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":42}}`))
	}))
	defer server.Close()

	client, err := New(server.Client(), "token")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	client.baseURL = server.URL

	messageID, err := client.PublishDraft(context.Background(), domain.Draft{ID: 1, Body: "hello", Status: domain.DraftStatusApproved}, "@ai_news")
	if err != nil {
		t.Fatalf("PublishDraft() error = %v", err)
	}
	if messageID != 42 {
		t.Fatalf("messageID = %d", messageID)
	}
}

func TestPublishDraftValidationAndErrors(t *testing.T) {
	if _, err := New(nil, ""); err == nil {
		t.Fatalf("New() expected bot token error")
	}

	client, err := New(nil, "token")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if _, err := client.PublishDraft(nil, domain.Draft{ID: 1, Body: "x", Status: domain.DraftStatusApproved}, "@ai_news"); err == nil {
		t.Fatalf("expected nil context error")
	}
	if _, err := client.PublishDraft(context.Background(), domain.Draft{}, "@ai_news"); err == nil {
		t.Fatalf("expected invalid draft id error")
	}
	if _, err := client.PublishDraft(context.Background(), domain.Draft{ID: 1, Body: "x", Status: domain.DraftStatusPending}, "@ai_news"); err == nil {
		t.Fatalf("expected status error")
	}
	if _, err := client.PublishDraft(context.Background(), domain.Draft{ID: 1, Body: " ", Status: domain.DraftStatusApproved}, "@ai_news"); err == nil {
		t.Fatalf("expected empty body error")
	}
	if _, err := client.PublishDraft(context.Background(), domain.Draft{ID: 1, Body: "x", Status: domain.DraftStatusApproved}, ""); err == nil {
		t.Fatalf("expected empty chat id error")
	}

	var nilClient *Client
	if _, err := nilClient.PublishDraft(context.Background(), domain.Draft{ID: 1, Body: "x", Status: domain.DraftStatusApproved}, "@ai_news"); err == nil {
		t.Fatalf("expected nil client error")
	}
}

func TestPublishDraftReturnsErrorsOnBadResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	client, err := New(server.Client(), "token")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	client.baseURL = server.URL

	if _, err := client.PublishDraft(context.Background(), domain.Draft{ID: 1, Body: "hello", Status: domain.DraftStatusApproved}, "@ai_news"); err == nil {
		t.Fatalf("expected status error")
	}
}

func TestPublishDraftReturnsTelegramAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":false,"description":"chat not found"}`))
	}))
	defer server.Close()

	client, err := New(server.Client(), "token")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	client.baseURL = server.URL

	if _, err := client.PublishDraft(context.Background(), domain.Draft{ID: 1, Body: "hello", Status: domain.DraftStatusApproved}, "@ai_news"); err == nil {
		t.Fatalf("expected telegram api error")
	}
}

func TestPublishDraftTrimsManualClientTokenAndBaseURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/bottoken/sendMessage") {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":7}}`))
	}))
	defer server.Close()

	client := &Client{httpClient: server.Client(), botToken: " token ", baseURL: " " + server.URL + " "}
	messageID, err := client.PublishDraft(context.Background(), domain.Draft{ID: 1, Body: "hello", Status: domain.DraftStatusApproved}, "@ai_news")
	if err != nil {
		t.Fatalf("PublishDraft() error = %v", err)
	}
	if messageID != 7 {
		t.Fatalf("messageID = %d", messageID)
	}
}
