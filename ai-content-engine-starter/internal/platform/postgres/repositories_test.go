package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"ai-content-engine-starter/internal/domain"
)

func TestRepositoriesImplementDomainInterfaces(t *testing.T) {
	var _ domain.ChannelRepository = (*ChannelRepository)(nil)
	var _ domain.SourceRepository = (*SourceRepository)(nil)
	var _ domain.SourceItemRepository = (*SourceItemRepository)(nil)
	var _ domain.DraftRepository = (*DraftRepository)(nil)
	var _ domain.PublishIntentRepository = (*PublishIntentRepository)(nil)
	var _ domain.ContentAssetRepository = (*ContentAssetRepository)(nil)
	var _ domain.TopicMemoryRepository = (*TopicMemoryRepository)(nil)
	var _ domain.ContentRuleRepository = (*ContentRuleRepository)(nil)
	var _ domain.PerformanceFeedbackRepository = (*PerformanceFeedbackRepository)(nil)
}

func TestPublishIntentRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewPublishIntentRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.PublishIntent{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	if _, err := repo.Create(context.Background(), domain.PublishIntent{RawItemID: 1, ChannelID: 1, Format: "text", Priority: 1, Status: "invalid"}); err == nil {
		t.Fatalf("Create expected status validation error")
	}
	if _, err := repo.ListByRawItemID(context.Background(), 0, 10); err == nil {
		t.Fatalf("ListByRawItemID expected validation error for raw item id")
	}
	if _, err := repo.ListByRawItemID(context.Background(), 1, 0); err == nil {
		t.Fatalf("ListByRawItemID expected validation error for limit")
	}
	if err := repo.UpdateStatus(context.Background(), 0, domain.PublishIntentStatusPlanned); err == nil {
		t.Fatalf("UpdateStatus expected id validation error")
	}
	if err := repo.UpdateStatus(context.Background(), 1, "invalid"); err == nil {
		t.Fatalf("UpdateStatus expected status validation error")
	}
}

func TestListBySourceIDRejectsInvalidLimit(t *testing.T) {
	repo := NewSourceItemRepository(&sql.DB{})
	_, err := repo.ListBySourceID(context.Background(), 1, 0)
	if err == nil {
		t.Fatalf("ListBySourceID expected error for invalid limit")
	}
}

func TestListByStatusRejectsInvalidLimit(t *testing.T) {
	repo := NewDraftRepository(&sql.DB{})
	_, err := repo.ListByStatus(context.Background(), domain.DraftStatusPending, 0)
	if err == nil {
		t.Fatalf("ListByStatus expected error for invalid limit")
	}
}

func TestRepositoryRejectsNilDB(t *testing.T) {
	repo := NewChannelRepository(nil)
	_, err := repo.GetByID(context.Background(), 1)
	if err == nil {
		t.Fatalf("GetByID expected error when db is nil")
	}
}

func TestListTopByChannelRejectsInvalidInput(t *testing.T) {
	repo := NewTopicMemoryRepository(&sql.DB{})
	if _, err := repo.ListTopByChannel(context.Background(), 0, 10); err == nil {
		t.Fatalf("ListTopByChannel expected error for invalid channel id")
	}
	if _, err := repo.ListTopByChannel(context.Background(), 1, 0); err == nil {
		t.Fatalf("ListTopByChannel expected error for invalid limit")
	}
}

func TestUpsertMentionRejectsInvalidInput(t *testing.T) {
	repo := NewTopicMemoryRepository(&sql.DB{})
	base := domain.TopicMemory{ChannelID: 1, Topic: "ai", MentionCount: 1, LastSeenAt: time.Now().UTC()}

	invalid := []domain.TopicMemory{
		{Topic: base.Topic, MentionCount: base.MentionCount, LastSeenAt: base.LastSeenAt},
		{ChannelID: base.ChannelID, MentionCount: base.MentionCount, LastSeenAt: base.LastSeenAt},
		{ChannelID: base.ChannelID, Topic: base.Topic, LastSeenAt: base.LastSeenAt},
		{ChannelID: base.ChannelID, Topic: base.Topic, MentionCount: base.MentionCount},
	}

	for _, item := range invalid {
		if _, err := repo.UpsertMention(context.Background(), item); err == nil {
			t.Fatalf("UpsertMention expected validation error")
		}
	}
}

func TestContentRuleRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewContentRuleRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.ContentRule{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	invalidChannel := int64(0)
	if _, err := repo.ListEnabled(context.Background(), &invalidChannel); err == nil {
		t.Fatalf("ListEnabled expected channel validation error")
	}
}

func TestPerformanceFeedbackRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewPerformanceFeedbackRepository(&sql.DB{})
	if _, err := repo.Upsert(context.Background(), domain.PerformanceFeedback{}); err == nil {
		t.Fatalf("Upsert expected validation error")
	}
	if _, err := repo.Upsert(context.Background(), domain.PerformanceFeedback{DraftID: 1, ChannelID: 1, Variant: "C"}); err == nil {
		t.Fatalf("Upsert expected variant validation error")
	}
	if _, err := repo.GetByDraftID(context.Background(), 0); err == nil {
		t.Fatalf("GetByDraftID expected validation error")
	}
}

func TestNormalizeFeedbackVariant(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "spaces", in: "   ", want: ""},
		{name: "lower a", in: "a", want: "A"},
		{name: "mixed b", in: " b ", want: "B"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeFeedbackVariant(tc.in); got != tc.want {
				t.Fatalf("normalizeFeedbackVariant(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestContentAssetRepositoryRejectsInvalidInput(t *testing.T) {
	repo := NewContentAssetRepository(&sql.DB{})
	if _, err := repo.Create(context.Background(), domain.ContentAsset{}); err == nil {
		t.Fatalf("Create expected validation error")
	}
	if _, err := repo.Create(context.Background(), domain.ContentAsset{RawItemID: 1, ChannelID: 1, AssetType: "text", Status: "invalid"}); err == nil {
		t.Fatalf("Create expected status validation error")
	}
	if _, err := repo.GetByID(context.Background(), 0); err == nil {
		t.Fatalf("GetByID expected id validation error")
	}
	if _, err := repo.ListByRawItemID(context.Background(), 0, 10); err == nil {
		t.Fatalf("ListByRawItemID expected raw item id validation error")
	}
	if _, err := repo.ListByRawItemID(context.Background(), 1, 0); err == nil {
		t.Fatalf("ListByRawItemID expected limit validation error")
	}
}
