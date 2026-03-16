package adminbot

import (
	"context"
	"errors"
	"strings"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

type draftRepoStub struct {
	byStatus     map[domain.DraftStatus][]domain.Draft
	listErr      error
	updateErr    error
	lastUpdateID int64
	lastStatus   domain.DraftStatus
}

func (s *draftRepoStub) Create(context.Context, domain.Draft) (domain.Draft, error) {
	return domain.Draft{}, nil
}
func (s *draftRepoStub) GetByID(context.Context, int64) (domain.Draft, error) {
	return domain.Draft{}, nil
}
func (s *draftRepoStub) ListByStatus(_ context.Context, status domain.DraftStatus, _ int) ([]domain.Draft, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.byStatus[status], nil
}
func (s *draftRepoStub) UpdateStatus(_ context.Context, id int64, status domain.DraftStatus) error {
	if s.updateErr != nil {
		return s.updateErr
	}
	s.lastUpdateID = id
	s.lastStatus = status
	return nil
}

func TestNewValidation(t *testing.T) {
	if _, err := New(nil, []int64{1}); err == nil {
		t.Fatalf("expected nil drafts error")
	}
	if _, err := New(&draftRepoStub{}, nil); err == nil {
		t.Fatalf("expected empty allowed chats error")
	}
	if _, err := New(&draftRepoStub{}, []int64{0}); err == nil {
		t.Fatalf("expected invalid chat id error")
	}
}

func TestHandleCommandPendingApproveReject(t *testing.T) {
	repo := &draftRepoStub{byStatus: map[domain.DraftStatus][]domain.Draft{
		domain.DraftStatusPending: {
			{ID: 7, ChannelID: 1, Variant: "A", Title: "A title"},
			{ID: 9, ChannelID: 2, Variant: "B", Title: "B title"},
		},
	}}
	svc, err := New(repo, []int64{42})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	msg, err := svc.HandleCommand(context.Background(), 42, "/pending 5")
	if err != nil {
		t.Fatalf("/pending error = %v", err)
	}
	if !strings.Contains(msg, "#7") || !strings.Contains(msg, "#9") {
		t.Fatalf("pending output = %q", msg)
	}

	msg, err = svc.HandleCommand(context.Background(), 42, "/approve 7")
	if err != nil {
		t.Fatalf("/approve error = %v", err)
	}
	if repo.lastUpdateID != 7 || repo.lastStatus != domain.DraftStatusApproved {
		t.Fatalf("approve args id=%d status=%q", repo.lastUpdateID, repo.lastStatus)
	}
	if msg != "Draft #7 approved." {
		t.Fatalf("approve message = %q", msg)
	}

	msg, err = svc.HandleCommand(context.Background(), 42, "/reject 9")
	if err != nil {
		t.Fatalf("/reject error = %v", err)
	}
	if repo.lastUpdateID != 9 || repo.lastStatus != domain.DraftStatusRejected {
		t.Fatalf("reject args id=%d status=%q", repo.lastUpdateID, repo.lastStatus)
	}
	if msg != "Draft #9 rejected." {
		t.Fatalf("reject message = %q", msg)
	}
}

func TestHandleCommandValidationAndErrors(t *testing.T) {
	svc, _ := New(&draftRepoStub{}, []int64{42})
	if _, err := svc.HandleCommand(nil, 42, "/pending"); err == nil {
		t.Fatalf("expected nil context error")
	}
	if _, err := svc.HandleCommand(context.Background(), 43, "/pending"); err == nil {
		t.Fatalf("expected unauthorized chat error")
	}
	if _, err := svc.HandleCommand(context.Background(), 42, "/pending x"); err == nil {
		t.Fatalf("expected pending limit parse error")
	}
	if _, err := svc.HandleCommand(context.Background(), 42, "/pending 5 extra"); err == nil {
		t.Fatalf("expected pending limit extra args error")
	}
	if _, err := svc.HandleCommand(context.Background(), 42, "/approve"); err == nil {
		t.Fatalf("expected missing id error")
	}
	if _, err := svc.HandleCommand(context.Background(), 42, "/approve x"); err == nil {
		t.Fatalf("expected invalid id error")
	}

	repo := &draftRepoStub{updateErr: domain.ErrNotFound}
	svc, _ = New(repo, []int64{42})
	if _, err := svc.HandleCommand(context.Background(), 42, "/approve 1"); err == nil {
		t.Fatalf("expected not found error")
	}

	repo = &draftRepoStub{updateErr: errors.New("boom")}
	svc, _ = New(repo, []int64{42})
	if _, err := svc.HandleCommand(context.Background(), 42, "/reject 1"); err == nil {
		t.Fatalf("expected update error")
	}

	repo = &draftRepoStub{listErr: errors.New("boom")}
	svc, _ = New(repo, []int64{42})
	if _, err := svc.HandleCommand(context.Background(), 42, "/pending"); err == nil {
		t.Fatalf("expected list error")
	}
}

func TestHandleCommandHelpAndUnknown(t *testing.T) {
	svc, _ := New(&draftRepoStub{}, []int64{42})
	msg, err := svc.HandleCommand(context.Background(), 42, "")
	if err != nil {
		t.Fatalf("empty command error = %v", err)
	}
	if !strings.Contains(msg, "Admin bot commands") {
		t.Fatalf("help message = %q", msg)
	}

	msg, err = svc.HandleCommand(context.Background(), 42, "/unknown")
	if err != nil {
		t.Fatalf("unknown command error = %v", err)
	}
	if !strings.Contains(msg, "/pending") {
		t.Fatalf("unknown command fallback = %q", msg)
	}
}
