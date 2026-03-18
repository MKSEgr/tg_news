package network

import (
	"context"
	"errors"
	"testing"
	"time"

	"ai-content-engine-starter/internal/domain"
)

type stubRelationshipRepository struct {
	relsByChannel map[int64][]domain.ChannelRelationship
	err           error
}

func (r stubRelationshipRepository) Create(context.Context, domain.ChannelRelationship) (domain.ChannelRelationship, error) {
	return domain.ChannelRelationship{}, errors.New("not implemented")
}

func (r stubRelationshipRepository) ListByChannel(_ context.Context, channelID int64, limit int) ([]domain.ChannelRelationship, error) {
	if r.err != nil {
		return nil, r.err
	}
	if limit <= 0 {
		return nil, errors.New("limit must be greater than zero")
	}
	rels := r.relsByChannel[channelID]
	if len(rels) > limit {
		rels = rels[:limit]
	}
	return rels, nil
}

func TestNewValidation(t *testing.T) {
	if _, err := New(nil, Config{}); err == nil {
		t.Fatalf("expected nil repository validation error")
	}
	if _, err := New(stubRelationshipRepository{}, Config{RelationshipLimit: -1}); err == nil {
		t.Fatalf("expected relationship limit validation error")
	}
	if _, err := New(stubRelationshipRepository{}, Config{BaseDelay: -time.Second}); err == nil {
		t.Fatalf("expected base delay validation error")
	}
}

func TestBuildRanksChannelsByRelationshipScore(t *testing.T) {
	svc, err := New(stubRelationshipRepository{relsByChannel: map[int64][]domain.ChannelRelationship{
		1: {{ChannelID: 1, RelatedChannelID: 2, RelationshipType: domain.ChannelRelationshipTypeSibling, Strength: 0.5}},
		2: {{ChannelID: 2, RelatedChannelID: 3, RelationshipType: domain.ChannelRelationshipTypePromotionTarget, Strength: 1}},
		3: {{ChannelID: 3, RelatedChannelID: 1, RelationshipType: domain.ChannelRelationshipTypeParent, Strength: 0.5}},
	}}, Config{BaseDelay: time.Minute})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	entries, err := svc.Build(context.Background(), []domain.Channel{{ID: 1}, {ID: 2}, {ID: 3}})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("len(entries) = %d, want 3", len(entries))
	}
	if entries[0].ChannelID != 2 || entries[1].ChannelID != 3 || entries[2].ChannelID != 1 {
		t.Fatalf("unexpected ordering: %#v", entries)
	}
	if entries[0].Delay != 0 || entries[1].Delay != time.Minute || entries[2].Delay != 2*time.Minute {
		t.Fatalf("unexpected delays: %#v", entries)
	}
}

func TestBuildUsesFallbackReasonWithoutRelationships(t *testing.T) {
	svc, err := New(stubRelationshipRepository{}, Config{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	entries, err := svc.Build(context.Background(), []domain.Channel{{ID: 7}})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if got := entries[0].Reasons[0]; got != "no_relationships" {
		t.Fatalf("reason = %q, want no_relationships", got)
	}
}

func TestBuildValidation(t *testing.T) {
	svc, err := New(stubRelationshipRepository{}, Config{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := svc.Build(nil, []domain.Channel{{ID: 1}}); err == nil {
		t.Fatalf("expected nil context error")
	}
	if _, err := svc.Build(context.Background(), []domain.Channel{{ID: 0}}); err == nil {
		t.Fatalf("expected invalid channel id error")
	}

	var nilSvc *Service
	if _, err := nilSvc.Build(context.Background(), []domain.Channel{{ID: 1}}); err == nil {
		t.Fatalf("expected nil service error")
	}
}

func TestBuildPropagatesRepositoryError(t *testing.T) {
	svc, err := New(stubRelationshipRepository{err: errors.New("boom")}, Config{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := svc.Build(context.Background(), []domain.Channel{{ID: 1}}); err == nil {
		t.Fatalf("expected repository error")
	}
}
