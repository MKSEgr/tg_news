package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterValidation(t *testing.T) {
	if err := Register(nil); err == nil {
		t.Fatalf("expected nil mux error")
	}
}

func TestIndexPage(t *testing.T) {
	mux := http.NewServeMux()
	if err := Register(mux); err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Content-Type"); !strings.Contains(got, "text/html") {
		t.Fatalf("content type = %q, want text/html", got)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "AI Content Engine") || !strings.Contains(body, "/admin/drafts?status=pending") {
		t.Fatalf("body = %q", body)
	}
}
