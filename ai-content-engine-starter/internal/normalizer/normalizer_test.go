package normalizer

import (
	"testing"

	"ai-content-engine-starter/internal/domain"
)

func TestNormalizeTrimsAndCanonicalizesFields(t *testing.T) {
	svc := New()
	body := "  first\n\tsecond   "

	item, err := svc.Normalize(domain.SourceItem{
		ExternalID: "  ext-1  ",
		URL:        " HTTPS://Example.COM/path/../news/?b=2&a=1#part ",
		Title:      "  new   ai   model ",
		Body:       &body,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	if item.ExternalID != "ext-1" {
		t.Fatalf("ExternalID = %q", item.ExternalID)
	}
	if item.Title != "new ai model" {
		t.Fatalf("Title = %q", item.Title)
	}
	if item.URL != "https://example.com/news?a=1&b=2" {
		t.Fatalf("URL = %q", item.URL)
	}
	if item.Body == nil || *item.Body != "first second" {
		t.Fatalf("Body = %v", item.Body)
	}
}

func TestNormalizeRejectsMissingRequiredFields(t *testing.T) {
	svc := New()

	if _, err := svc.Normalize(domain.SourceItem{URL: "https://example.com", Title: "ok"}); err == nil {
		t.Fatalf("expected error for missing external id")
	}
	if _, err := svc.Normalize(domain.SourceItem{ExternalID: "id", URL: "https://example.com"}); err == nil {
		t.Fatalf("expected error for missing title")
	}
	if _, err := svc.Normalize(domain.SourceItem{ExternalID: "id", Title: "ok"}); err == nil {
		t.Fatalf("expected error for missing url")
	}
}

func TestNormalizeRejectsInvalidURL(t *testing.T) {
	svc := New()

	_, err := svc.Normalize(domain.SourceItem{
		ExternalID: "id",
		Title:      "ok",
		URL:        "not a url",
	})
	if err == nil {
		t.Fatalf("expected error for invalid url")
	}
}

func TestNormalizeEmptyBodyBecomesNil(t *testing.T) {
	svc := New()
	body := " \n\t "

	item, err := svc.Normalize(domain.SourceItem{
		ExternalID: "id",
		Title:      "ok",
		URL:        "https://example.com",
		Body:       &body,
	})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if item.Body != nil {
		t.Fatalf("Body expected nil, got %v", *item.Body)
	}
}

func TestNormalizeStripsDefaultPorts(t *testing.T) {
	svc := New()

	httpItem, err := svc.Normalize(domain.SourceItem{ExternalID: "id-1", Title: "title", URL: "http://example.com:80/path"})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if httpItem.URL != "http://example.com/path" {
		t.Fatalf("http URL = %q", httpItem.URL)
	}

	httpsItem, err := svc.Normalize(domain.SourceItem{ExternalID: "id-2", Title: "title", URL: "https://example.com:443/path"})
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if httpsItem.URL != "https://example.com/path" {
		t.Fatalf("https URL = %q", httpsItem.URL)
	}
}
