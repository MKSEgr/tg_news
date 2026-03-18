package generator

import (
	"context"
	"errors"
	"strings"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

type fakeAIClient struct {
	response   string
	err        error
	lastPrompt string
}

func (f *fakeAIClient) GenerateText(_ context.Context, prompt string) (string, error) {
	f.lastPrompt = prompt
	if f.err != nil {
		return "", f.err
	}
	return f.response, nil
}

func TestGenerateDraftHappyPath(t *testing.T) {
	ai := &fakeAIClient{response: "Готовый текст поста"}
	svc, err := New(ai)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	body := "Подробности новости"
	draft, err := svc.GenerateDraft(
		context.Background(),
		domain.SourceItem{ID: 11, Title: "AI release", Body: &body, URL: "https://example.com"},
		domain.Channel{ID: 7, Name: "AI News", Slug: "ai-news"},
	)
	if err != nil {
		t.Fatalf("GenerateDraft() error = %v", err)
	}
	if draft.SourceItemID != 11 || draft.ChannelID != 7 {
		t.Fatalf("draft ids = %+v", draft)
	}
	if draft.Status != domain.DraftStatusPending {
		t.Fatalf("status = %q", draft.Status)
	}
	if draft.Variant != "A" {
		t.Fatalf("variant = %q, want A", draft.Variant)
	}
	if draft.Body != "Готовый текст поста" {
		t.Fatalf("body = %q", draft.Body)
	}
	if !strings.Contains(ai.lastPrompt, "Канал: AI News") {
		t.Fatalf("prompt = %q", ai.lastPrompt)
	}
}

func TestGenerateDraftValidationAndErrors(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatalf("New() expected error for nil ai")
	}

	ai := &fakeAIClient{response: "ok"}
	svc, err := New(ai)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if _, err := svc.GenerateDraft(nil, domain.SourceItem{ID: 1, Title: "t"}, domain.Channel{ID: 1, Name: "n", Slug: "ai-news"}); err == nil {
		t.Fatalf("expected error for nil context")
	}
	if _, err := svc.GenerateDraft(context.Background(), domain.SourceItem{}, domain.Channel{ID: 1, Name: "n"}); err == nil {
		t.Fatalf("expected error for invalid source item")
	}
	if _, err := svc.GenerateDraft(context.Background(), domain.SourceItem{ID: 1, Title: "t"}, domain.Channel{}); err == nil {
		t.Fatalf("expected error for invalid channel")
	}
	if _, err := svc.GenerateDraft(context.Background(), domain.SourceItem{ID: 1, Title: "t"}, domain.Channel{ID: 1, Name: "n"}); err == nil {
		t.Fatalf("expected error for missing channel slug")
	}

	ai.err = errors.New("upstream")
	if _, err := svc.GenerateDraft(context.Background(), domain.SourceItem{ID: 1, Title: "t"}, domain.Channel{ID: 1, Name: "n", Slug: "ai-news"}); err == nil {
		t.Fatalf("expected generation error")
	}

	ai.err = nil
	ai.response = "   "
	if _, err := svc.GenerateDraft(context.Background(), domain.SourceItem{ID: 1, Title: "t"}, domain.Channel{ID: 1, Name: "n", Slug: "ai-news"}); err == nil {
		t.Fatalf("expected empty body error")
	}
}

func TestGenerateDraftNilServiceSafety(t *testing.T) {
	var svc *Service
	if _, err := svc.GenerateDraft(context.Background(), domain.SourceItem{ID: 1, Title: "t"}, domain.Channel{ID: 1, Name: "n", Slug: "ai-news"}); err == nil {
		t.Fatalf("expected error for nil service")
	}
}

func TestGenerateDraftWithFeedbackAddsPromptHint(t *testing.T) {
	ai := &fakeAIClient{response: "Готовый текст поста"}
	svc, err := New(ai)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = svc.GenerateDraftWithFeedback(
		context.Background(),
		domain.SourceItem{ID: 11, Title: "AI release", URL: "https://example.com"},
		domain.Channel{ID: 7, Name: "AI News", Slug: "ai-news"},
		1.5,
	)
	if err != nil {
		t.Fatalf("GenerateDraftWithFeedback() error = %v", err)
	}
	if !strings.Contains(ai.lastPrompt, "высокий отклик") {
		t.Fatalf("prompt = %q", ai.lastPrompt)
	}
}

func TestGenerateDraftWithClusterAddsPromptContext(t *testing.T) {
	ai := &fakeAIClient{response: "Готовый текст поста"}
	svc, err := New(ai)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	_, err = svc.GenerateDraftWithCluster(
		context.Background(),
		domain.SourceItem{ID: 11, Title: "AI release", URL: "https://example.com"},
		domain.Channel{ID: 7, Name: "AI News", Slug: "ai-news"},
		domain.StoryCluster{ID: 99, Title: "OpenAI launches", Summary: "Release cluster"},
	)
	if err != nil {
		t.Fatalf("GenerateDraftWithCluster() error = %v", err)
	}
	if !strings.Contains(ai.lastPrompt, "Контекст кластера: id=99") {
		t.Fatalf("prompt = %q", ai.lastPrompt)
	}
}

func TestGenerateDraftVariantsHappyPath(t *testing.T) {
	ai := &fakeAIClient{response: "Готовый текст поста"}
	svc, err := New(ai)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	body := "Подробности новости"
	variants, err := svc.GenerateDraftVariants(
		context.Background(),
		domain.SourceItem{ID: 11, Title: "AI release", Body: &body, URL: "https://example.com"},
		domain.Channel{ID: 7, Name: "AI News", Slug: "ai-news"},
		0,
	)
	if err != nil {
		t.Fatalf("GenerateDraftVariants() error = %v", err)
	}
	if len(variants) != 2 || variants[0].Variant != "A" || variants[1].Variant != "B" {
		t.Fatalf("variants = %+v", variants)
	}
}
