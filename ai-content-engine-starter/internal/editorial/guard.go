package editorial

import (
	"fmt"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

// Guard validates generated drafts before scheduling/publishing.
type Guard struct {
	maxTitleRunes  int
	maxBodyRunes   int
	blockedPhrases []string
}

// Result describes guard decision details.
type Result struct {
	Accepted bool
	Reasons  []string
}

// NewGuard creates editorial guard with minimal MVP rules.
func NewGuard() *Guard {
	return &Guard{
		maxTitleRunes: 120,
		maxBodyRunes:  2000,
		blockedPhrases: []string{
			"guaranteed profit",
			"100% sure",
			"click here now",
		},
	}
}

// Check validates a draft and returns acceptance result.
func (g *Guard) Check(draft domain.Draft) (Result, error) {
	if g == nil {
		return Result{}, fmt.Errorf("editorial guard is nil")
	}

	reasons := make([]string, 0)
	title := strings.TrimSpace(draft.Title)
	body := strings.TrimSpace(draft.Body)

	if draft.SourceItemID <= 0 {
		reasons = append(reasons, "source item id is invalid")
	}
	if draft.ChannelID <= 0 {
		reasons = append(reasons, "channel id is invalid")
	}
	if draft.Status != domain.DraftStatusPending {
		reasons = append(reasons, "draft status must be pending")
	}
	if title == "" {
		reasons = append(reasons, "title is empty")
	}
	if body == "" {
		reasons = append(reasons, "body is empty")
	}
	if len([]rune(title)) > g.maxTitleRunes {
		reasons = append(reasons, "title is too long")
	}
	if len([]rune(body)) > g.maxBodyRunes {
		reasons = append(reasons, "body is too long")
	}

	lowerBody := strings.ToLower(body)
	for _, phrase := range g.blockedPhrases {
		if strings.Contains(lowerBody, phrase) {
			reasons = append(reasons, "contains blocked phrase")
			break
		}
	}

	return Result{Accepted: len(reasons) == 0, Reasons: reasons}, nil
}
