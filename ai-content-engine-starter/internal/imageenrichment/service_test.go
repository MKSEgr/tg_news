package imageenrichment

import (
	"strings"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

func TestEnrichValidation(t *testing.T) {
	service := New()
	if _, err := service.Enrich(domain.SourceItem{}); err == nil {
		t.Fatalf("expected invalid item id error")
	}

	var nilService *Service
	if _, err := nilService.Enrich(domain.SourceItem{ID: 1}); err == nil {
		t.Fatalf("expected nil service error")
	}
}

func TestEnrichFindsImageURL(t *testing.T) {
	service := New()

	tests := []struct {
		name string
		item domain.SourceItem
		want string
	}{
		{
			name: "direct url",
			item: domain.SourceItem{ID: 1, URL: "https://example.com/image.png"},
			want: "https://example.com/image.png",
		},
		{
			name: "markdown image in body",
			item: domain.SourceItem{ID: 1, URL: "https://example.com/post", Body: ptr("Hello ![preview](https://cdn.example.com/preview.webp)")},
			want: "https://cdn.example.com/preview.webp",
		},
		{
			name: "html img in body",
			item: domain.SourceItem{ID: 1, URL: "https://example.com/post", Body: ptr("<img alt='x' src=\"https://cdn.example.com/hero.jpg\">")},
			want: "https://cdn.example.com/hero.jpg",
		},
		{
			name: "raw image url in body",
			item: domain.SourceItem{ID: 1, URL: "https://example.com/post", Body: ptr("see https://cdn.example.com/photo.jpeg now")},
			want: "https://cdn.example.com/photo.jpeg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			enriched, err := service.Enrich(tc.item)
			if err != nil {
				t.Fatalf("Enrich() error = %v", err)
			}
			if enriched.ImageURL == nil {
				t.Fatalf("ImageURL is nil")
			}
			if *enriched.ImageURL != tc.want {
				t.Fatalf("ImageURL = %q, want %q", *enriched.ImageURL, tc.want)
			}
		})
	}
}

func TestEnrichSkipsNonImageURLs(t *testing.T) {
	service := New()
	item := domain.SourceItem{ID: 1, URL: "https://example.com/post", Body: ptr(strings.Repeat("text ", 3) + "https://example.com/page")}
	enriched, err := service.Enrich(item)
	if err != nil {
		t.Fatalf("Enrich() error = %v", err)
	}
	if enriched.ImageURL != nil {
		t.Fatalf("ImageURL expected nil, got %q", *enriched.ImageURL)
	}
}

func ptr(v string) *string { return &v }
