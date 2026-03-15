package domain

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type stubChannelRepo struct{}

func (stubChannelRepo) Create(context.Context, Channel) (Channel, error) { return Channel{}, nil }
func (stubChannelRepo) GetByID(context.Context, int64) (Channel, error)  { return Channel{}, nil }
func (stubChannelRepo) List(context.Context) ([]Channel, error)          { return nil, nil }

type stubSourceRepo struct{}

func (stubSourceRepo) Create(context.Context, Source) (Source, error) { return Source{}, nil }
func (stubSourceRepo) GetByID(context.Context, int64) (Source, error) { return Source{}, nil }
func (stubSourceRepo) ListEnabled(context.Context) ([]Source, error)  { return nil, nil }

type stubSourceItemRepo struct{}

func (stubSourceItemRepo) Create(context.Context, SourceItem) (SourceItem, error) {
	return SourceItem{}, nil
}
func (stubSourceItemRepo) GetByID(context.Context, int64) (SourceItem, error) {
	return SourceItem{}, nil
}
func (stubSourceItemRepo) ListBySourceID(context.Context, int64, int) ([]SourceItem, error) {
	return nil, nil
}

type stubDraftRepo struct{}

func (stubDraftRepo) Create(context.Context, Draft) (Draft, error)  { return Draft{}, nil }
func (stubDraftRepo) GetByID(context.Context, int64) (Draft, error) { return Draft{}, nil }
func (stubDraftRepo) ListByStatus(context.Context, DraftStatus, int) ([]Draft, error) {
	return nil, nil
}
func (stubDraftRepo) UpdateStatus(context.Context, int64, DraftStatus) error { return nil }

func TestRepositoryInterfacesImplementedByStubs(t *testing.T) {
	var _ ChannelRepository = stubChannelRepo{}
	var _ SourceRepository = stubSourceRepo{}
	var _ SourceItemRepository = stubSourceItemRepo{}
	var _ DraftRepository = stubDraftRepo{}
}

func TestErrNotFoundIsWrappable(t *testing.T) {
	err := fmt.Errorf("wrapped: %w", ErrNotFound)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected wrapped error to match ErrNotFound")
	}
}
