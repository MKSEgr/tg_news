package editorial

import (
	"strings"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

func TestCheckAcceptsValidDraft(t *testing.T) {
	guard := NewGuard()
	result, err := guard.Check(domain.Draft{
		SourceItemID: 1,
		ChannelID:    2,
		Title:        "AI weekly update",
		Body:         "Short and informative content.",
		Status:       domain.DraftStatusPending,
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !result.Accepted {
		t.Fatalf("expected accepted, reasons: %v", result.Reasons)
	}
}

func TestCheckRejectsInvalidDraft(t *testing.T) {
	guard := NewGuard()
	result, err := guard.Check(domain.Draft{})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Accepted {
		t.Fatalf("expected rejected")
	}
	if len(result.Reasons) == 0 {
		t.Fatalf("expected rejection reasons")
	}
}

func TestCheckRejectsNonPendingDraft(t *testing.T) {
	guard := NewGuard()
	result, err := guard.Check(domain.Draft{
		SourceItemID: 1,
		ChannelID:    2,
		Title:        "AI weekly update",
		Body:         "Short and informative content.",
		Status:       domain.DraftStatusApproved,
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Accepted {
		t.Fatalf("expected rejected")
	}
	if !strings.Contains(strings.Join(result.Reasons, "|"), "draft status must be pending") {
		t.Fatalf("reasons = %v", result.Reasons)
	}
}

func TestCheckRejectsBlockedPhraseAndLongBody(t *testing.T) {
	guard := NewGuard()
	longBody := strings.Repeat("a", 2100) + " guaranteed profit"
	result, err := guard.Check(domain.Draft{
		SourceItemID: 1,
		ChannelID:    2,
		Title:        "AI post",
		Body:         longBody,
		Status:       domain.DraftStatusPending,
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if result.Accepted {
		t.Fatalf("expected rejected")
	}
	joined := strings.Join(result.Reasons, "|")
	if !strings.Contains(joined, "body is too long") || !strings.Contains(joined, "contains blocked phrase") {
		t.Fatalf("reasons = %v", result.Reasons)
	}
}

func TestCheckRejectsNilGuard(t *testing.T) {
	var guard *Guard
	if _, err := guard.Check(domain.Draft{}); err == nil {
		t.Fatalf("expected error for nil guard")
	}
}

func TestCheckWithMemoryRejectsOverusedTopic(t *testing.T) {
	guard := NewGuard()
	result, err := guard.CheckWithMemory(domain.Draft{
		SourceItemID: 1,
		ChannelID:    2,
		Title:        "AI weekly update",
		Body:         "This week: llm updates and benchmarks.",
		Status:       domain.DraftStatusPending,
	}, []domain.TopicMemory{{Topic: "llm", MentionCount: 12}})
	if err != nil {
		t.Fatalf("CheckWithMemory() error = %v", err)
	}
	if result.Accepted {
		t.Fatalf("expected rejected due to overused topic")
	}
}
