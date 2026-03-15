package postgres

import (
	"context"
	"database/sql"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

func TestRepositoriesImplementDomainInterfaces(t *testing.T) {
	var _ domain.ChannelRepository = (*ChannelRepository)(nil)
	var _ domain.SourceRepository = (*SourceRepository)(nil)
	var _ domain.SourceItemRepository = (*SourceItemRepository)(nil)
	var _ domain.DraftRepository = (*DraftRepository)(nil)
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
