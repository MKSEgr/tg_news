package yandexai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateTextHappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if got := r.Header.Get("Authorization"); got != "Api-Key test-key" {
			t.Fatalf("Authorization = %q", got)
		}

		var reqBody map[string]any
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if reqBody["modelUri"] != "gpt://folder/model" {
			t.Fatalf("modelUri = %v", reqBody["modelUri"])
		}

		_, _ = w.Write([]byte(`{"result":{"alternatives":[{"message":{"text":"Generated draft"}}]}}`))
	}))
	defer server.Close()

	client, err := New(server.Client(), Config{Endpoint: server.URL, APIKey: "test-key", ModelURI: "gpt://folder/model"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	text, err := client.GenerateText(context.Background(), "Write short post")
	if err != nil {
		t.Fatalf("GenerateText() error = %v", err)
	}
	if text != "Generated draft" {
		t.Fatalf("text = %q", text)
	}
}

func TestGenerateTextReturnsErrorOnNon200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("invalid api key"))
	}))
	defer server.Close()

	client, err := New(server.Client(), Config{Endpoint: server.URL, APIKey: "test-key", ModelURI: "gpt://folder/model"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = client.GenerateText(context.Background(), "prompt")
	if err == nil {
		t.Fatalf("GenerateText() expected error")
	}
	if !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "invalid api key") {
		t.Fatalf("error = %q", err)
	}
}

func TestValidation(t *testing.T) {
	if _, err := New(nil, Config{APIKey: "", ModelURI: "m"}); err == nil {
		t.Fatalf("New() expected error for empty api key")
	}
	if _, err := New(nil, Config{APIKey: "k", ModelURI: ""}); err == nil {
		t.Fatalf("New() expected error for empty model uri")
	}

	client, err := New(nil, Config{APIKey: "k", ModelURI: "m"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := client.GenerateText(context.Background(), "   "); err == nil {
		t.Fatalf("GenerateText() expected error for empty prompt")
	}
	if _, err := client.GenerateText(nil, "prompt"); err == nil {
		t.Fatalf("GenerateText() expected error for nil context")
	}

	var nilClient *Client
	if _, err := nilClient.GenerateText(context.Background(), "prompt"); err == nil {
		t.Fatalf("GenerateText() expected error for nil client")
	}
}

func TestGenerateTextHandlesZeroValueClientSafely(t *testing.T) {
	client := &Client{}
	if _, err := client.GenerateText(context.Background(), "prompt"); err == nil {
		t.Fatalf("GenerateText() expected validation error for empty config")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"result":{"alternatives":[{"message":{"text":"ok"}}]}}`))
	}))
	defer server.Close()

	client = &Client{endpoint: server.URL, apiKey: "k", modelURI: "m"}
	text, err := client.GenerateText(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("GenerateText() error = %v", err)
	}
	if text != "ok" {
		t.Fatalf("text = %q", text)
	}
}
