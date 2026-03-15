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
