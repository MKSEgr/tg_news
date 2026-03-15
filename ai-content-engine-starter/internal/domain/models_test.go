package domain

import "testing"

func TestDraftStatusesAreStable(t *testing.T) {
	if DraftStatusPending != "pending" {
		t.Fatalf("DraftStatusPending = %q, want %q", DraftStatusPending, "pending")
	}
	if DraftStatusApproved != "approved" {
		t.Fatalf("DraftStatusApproved = %q, want %q", DraftStatusApproved, "approved")
	}
	if DraftStatusRejected != "rejected" {
		t.Fatalf("DraftStatusRejected = %q, want %q", DraftStatusRejected, "rejected")
	}
	if DraftStatusPosted != "posted" {
		t.Fatalf("DraftStatusPosted = %q, want %q", DraftStatusPosted, "posted")
	}
}

func TestSourceItemAllowsNullableBody(t *testing.T) {
	item := SourceItem{}
	if item.Body != nil {
		t.Fatalf("SourceItem.Body = %v, want nil", item.Body)
	}
}

func TestContentRuleKindsAreStable(t *testing.T) {
	if ContentRuleKindBlacklist != "blacklist" {
		t.Fatalf("ContentRuleKindBlacklist = %q, want %q", ContentRuleKindBlacklist, "blacklist")
	}
	if ContentRuleKindWhitelist != "whitelist" {
		t.Fatalf("ContentRuleKindWhitelist = %q, want %q", ContentRuleKindWhitelist, "whitelist")
	}
}

func TestPerformanceFeedbackDefaults(t *testing.T) {
	feedback := PerformanceFeedback{}
	if feedback.Score != 0 {
		t.Fatalf("PerformanceFeedback.Score = %f, want 0", feedback.Score)
	}
}
