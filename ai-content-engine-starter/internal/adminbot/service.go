package adminbot

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

const defaultPendingLimit = 10

// Service provides minimal Telegram-admin command handling.
type Service struct {
	drafts       domain.DraftRepository
	allowedChats map[int64]struct{}
}

// New creates admin bot service.
func New(drafts domain.DraftRepository, allowedChatIDs []int64) (*Service, error) {
	if drafts == nil {
		return nil, fmt.Errorf("draft repository is nil")
	}
	if len(allowedChatIDs) == 0 {
		return nil, fmt.Errorf("allowed chat ids are empty")
	}
	allowed := make(map[int64]struct{}, len(allowedChatIDs))
	for _, id := range allowedChatIDs {
		if id == 0 {
			return nil, fmt.Errorf("allowed chat id is invalid")
		}
		allowed[id] = struct{}{}
	}
	return &Service{drafts: drafts, allowedChats: allowed}, nil
}

// HandleCommand handles a minimal set of admin commands.
func (s *Service) HandleCommand(ctx context.Context, chatID int64, text string) (string, error) {
	if s == nil {
		return "", fmt.Errorf("admin bot service is nil")
	}
	if ctx == nil {
		return "", fmt.Errorf("context is nil")
	}
	if s.drafts == nil {
		return "", fmt.Errorf("draft repository is nil")
	}
	if _, ok := s.allowedChats[chatID]; !ok {
		return "", fmt.Errorf("chat is not allowed")
	}

	parts := strings.Fields(strings.TrimSpace(text))
	if len(parts) == 0 {
		return helpText(), nil
	}

	switch strings.ToLower(parts[0]) {
	case "/start", "/help":
		return helpText(), nil
	case "/pending":
		if len(parts) > 2 {
			return "", fmt.Errorf("pending limit is invalid")
		}
		limit := defaultPendingLimit
		if len(parts) > 1 {
			value, err := strconv.Atoi(parts[1])
			if err != nil || value <= 0 {
				return "", fmt.Errorf("pending limit is invalid")
			}
			limit = value
		}
		return s.listPending(ctx, limit)
	case "/approve":
		return s.updateStatus(ctx, parts, domain.DraftStatusApproved, "approved")
	case "/reject":
		return s.updateStatus(ctx, parts, domain.DraftStatusRejected, "rejected")
	default:
		return helpText(), nil
	}
}

func (s *Service) listPending(ctx context.Context, limit int) (string, error) {
	drafts, err := s.drafts.ListByStatus(ctx, domain.DraftStatusPending, limit)
	if err != nil {
		return "", fmt.Errorf("list pending drafts: %w", err)
	}
	if len(drafts) == 0 {
		return "No pending drafts.", nil
	}

	sort.Slice(drafts, func(i, j int) bool { return drafts[i].ID < drafts[j].ID })
	lines := make([]string, 0, len(drafts)+1)
	lines = append(lines, "Pending drafts:")
	for _, draft := range drafts {
		if draft.ID <= 0 {
			continue
		}
		lines = append(lines, fmt.Sprintf("- #%d [channel:%d variant:%s] %s", draft.ID, draft.ChannelID, normalizeVariant(draft.Variant), strings.TrimSpace(draft.Title)))
	}
	if len(lines) == 1 {
		return "No pending drafts.", nil
	}
	return strings.Join(lines, "\n"), nil
}

func (s *Service) updateStatus(ctx context.Context, parts []string, status domain.DraftStatus, statusLabel string) (string, error) {
	if len(parts) != 2 {
		return "", fmt.Errorf("draft id is required")
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || id <= 0 {
		return "", fmt.Errorf("draft id is invalid")
	}
	if err := s.drafts.UpdateStatus(ctx, id, status); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return "", fmt.Errorf("draft not found")
		}
		return "", fmt.Errorf("update draft status: %w", err)
	}
	return fmt.Sprintf("Draft #%d %s.", id, statusLabel), nil
}

func normalizeVariant(raw string) string {
	value := strings.ToUpper(strings.TrimSpace(raw))
	if value == "B" {
		return "B"
	}
	return "A"
}

func helpText() string {
	return strings.Join([]string{
		"Admin bot commands:",
		"/pending [limit] - list pending drafts",
		"/approve <draft_id> - approve draft",
		"/reject <draft_id> - reject draft",
	}, "\n")
}
