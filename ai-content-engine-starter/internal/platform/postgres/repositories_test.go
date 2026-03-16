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
	var _ domain.TopicMemoryRepository = (*TopicMemoryRepository)(nil)
	var _ domain.ContentRuleRepository = (*ContentRuleRepository)(nil)
	var _ domain.PerformanceFeedbackRepository = (*PerformanceFeedbackRepository)(nil)
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
