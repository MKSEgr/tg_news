package generator

import (
	"context"
	"fmt"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

// AIClient abstracts text generation backend.
type AIClient interface {
	GenerateText(ctx context.Context, prompt string) (string, error)
}

// Service generates channel-targeted drafts from source items.
type Service struct {
	ai AIClient
}

// New creates content generator service.
func New(ai AIClient) (*Service, error) {
	if ai == nil {
		return nil, fmt.Errorf("ai client is nil")
	}
	return &Service{ai: ai}, nil
}

// GenerateDraft builds prompt, calls AI, and returns pending draft.
func (s *Service) GenerateDraft(ctx context.Context, item domain.SourceItem, channel domain.Channel) (domain.Draft, error) {
	if s == nil {
		return domain.Draft{}, fmt.Errorf("generator service is nil")
	}
	if s.ai == nil {
		return domain.Draft{}, fmt.Errorf("ai client is nil")
	}
	if ctx == nil {
		return domain.Draft{}, fmt.Errorf("context is nil")
	}
	if item.ID <= 0 {
		return domain.Draft{}, fmt.Errorf("source item id is invalid")
	}
	if strings.TrimSpace(item.Title) == "" {
		return domain.Draft{}, fmt.Errorf("source item title is empty")
	}
	if channel.ID <= 0 {
		return domain.Draft{}, fmt.Errorf("channel id is invalid")
	}
	if strings.TrimSpace(channel.Name) == "" {
		return domain.Draft{}, fmt.Errorf("channel name is empty")
	}
	if strings.TrimSpace(channel.Slug) == "" {
		return domain.Draft{}, fmt.Errorf("channel slug is empty")
	}

	prompt := buildPrompt(item, channel, nil)
	generated, err := s.ai.GenerateText(ctx, prompt)
	if err != nil {
		return domain.Draft{}, fmt.Errorf("generate content: %w", err)
	}
	generated = strings.TrimSpace(generated)
	if generated == "" {
		return domain.Draft{}, fmt.Errorf("generated body is empty")
	}

	title := strings.TrimSpace(item.Title)
	if len([]rune(title)) > 120 {
		title = string([]rune(title)[:120])
	}

	return domain.Draft{
		SourceItemID: item.ID,
		ChannelID:    channel.ID,
		Title:        title,
		Body:         generated,
		Variant:      "A",
		Status:       domain.DraftStatusPending,
	}, nil
}

// GenerateDraftWithFeedback adds lightweight deterministic feedback context for the target channel.
func (s *Service) GenerateDraftWithFeedback(ctx context.Context, item domain.SourceItem, channel domain.Channel, channelFeedback float64) (domain.Draft, error) {
	return s.generateDraft(ctx, item, channel, channelFeedback, nil, false)
}

// GenerateDraftVariants generates deterministic A/B draft variants for one source item and channel.
func (s *Service) GenerateDraftVariants(ctx context.Context, item domain.SourceItem, channel domain.Channel, channelFeedback float64) ([]domain.Draft, error) {
	return s.generateDraftVariants(ctx, item, channel, channelFeedback, nil)
}

func (s *Service) GenerateDraftWithCluster(ctx context.Context, item domain.SourceItem, channel domain.Channel, cluster domain.StoryCluster) (domain.Draft, error) {
	return s.generateDraft(ctx, item, channel, 0, &cluster, false)
}

// GenerateDraftWithFeedbackAndCluster adds both feedback and optional cluster context.
func (s *Service) GenerateDraftWithFeedbackAndCluster(ctx context.Context, item domain.SourceItem, channel domain.Channel, channelFeedback float64, cluster domain.StoryCluster) (domain.Draft, error) {
	return s.generateDraft(ctx, item, channel, channelFeedback, &cluster, false)
}

// GenerateDraftVariantsWithCluster generates deterministic A/B variants with optional cluster context.
func (s *Service) GenerateDraftVariantsWithCluster(ctx context.Context, item domain.SourceItem, channel domain.Channel, channelFeedback float64, cluster domain.StoryCluster) ([]domain.Draft, error) {
	return s.generateDraftVariants(ctx, item, channel, channelFeedback, &cluster)
}

func (s *Service) generateDraft(ctx context.Context, item domain.SourceItem, channel domain.Channel, channelFeedback float64, cluster *domain.StoryCluster, variantB bool) (domain.Draft, error) {
	if s == nil {
		return domain.Draft{}, fmt.Errorf("generator service is nil")
	}
	if s.ai == nil {
		return domain.Draft{}, fmt.Errorf("ai client is nil")
	}
	if ctx == nil {
		return domain.Draft{}, fmt.Errorf("context is nil")
	}
	if item.ID <= 0 {
		return domain.Draft{}, fmt.Errorf("source item id is invalid")
	}
	if strings.TrimSpace(item.Title) == "" {
		return domain.Draft{}, fmt.Errorf("source item title is empty")
	}
	if channel.ID <= 0 {
		return domain.Draft{}, fmt.Errorf("channel id is invalid")
	}
	if strings.TrimSpace(channel.Name) == "" {
		return domain.Draft{}, fmt.Errorf("channel name is empty")
	}
	if strings.TrimSpace(channel.Slug) == "" {
		return domain.Draft{}, fmt.Errorf("channel slug is empty")
	}

	prompt := buildPrompt(item, channel, cluster)
	if channelFeedback > 0 {
		prompt += "\nТон: деловой и полезный; опирайся на практическую ценность, так как канал показывает высокий отклик."
	}
	if variantB {
		prompt += "\nСделай альтернативный вариант B: более провокационный заголовок, но без кликбейта и без потери фактов."
	}
	generated, err := s.ai.GenerateText(ctx, prompt)
	if err != nil {
		return domain.Draft{}, fmt.Errorf("generate content: %w", err)
	}
	generated = strings.TrimSpace(generated)
	if generated == "" {
		return domain.Draft{}, fmt.Errorf("generated body is empty")
	}

	title := strings.TrimSpace(item.Title)
	if len([]rune(title)) > 120 {
		title = string([]rune(title)[:120])
	}

	variant := "A"
	if variantB {
		variant = "B"
	}
	return domain.Draft{SourceItemID: item.ID, ChannelID: channel.ID, Variant: variant, Title: title, Body: generated, Status: domain.DraftStatusPending}, nil
}

func (s *Service) generateDraftVariants(ctx context.Context, item domain.SourceItem, channel domain.Channel, channelFeedback float64, cluster *domain.StoryCluster) ([]domain.Draft, error) {
	variantA, err := s.generateDraft(ctx, item, channel, channelFeedback, cluster, false)
	if err != nil {
		return nil, err
	}
	variantB, err := s.generateDraft(ctx, item, channel, channelFeedback, cluster, true)
	if err != nil {
		return nil, err
	}
	return []domain.Draft{variantA, variantB}, nil
}

func buildPrompt(item domain.SourceItem, channel domain.Channel, cluster *domain.StoryCluster) string {
	var b strings.Builder
	b.WriteString("Сгенерируй короткий пост для Telegram-канала.\n")
	b.WriteString("Канал: ")
	b.WriteString(strings.TrimSpace(channel.Name))
	b.WriteString(" (slug: ")
	b.WriteString(strings.TrimSpace(channel.Slug))
	b.WriteString(")\n")
	b.WriteString("Заголовок источника: ")
	b.WriteString(strings.TrimSpace(item.Title))
	b.WriteString("\n")
	if item.Body != nil && strings.TrimSpace(*item.Body) != "" {
		b.WriteString("Текст источника: ")
		b.WriteString(strings.TrimSpace(*item.Body))
		b.WriteString("\n")
	}
	b.WriteString("URL: ")
	b.WriteString(strings.TrimSpace(item.URL))
	b.WriteString("\n")
	if cluster != nil && cluster.ID > 0 {
		b.WriteString("Контекст кластера: id=")
		b.WriteString(fmt.Sprintf("%d", cluster.ID))
		if title := strings.TrimSpace(cluster.Title); title != "" {
			b.WriteString(", title=")
			b.WriteString(title)
		}
		if summary := strings.TrimSpace(cluster.Summary); summary != "" {
			b.WriteString("\nСводка кластера: ")
			b.WriteString(summary)
			b.WriteString("\n")
		} else {
			b.WriteString("\n")
		}
	}
	b.WriteString("Требования: информативно, без кликбейта, 1-3 абзаца.")
	return b.String()
}
