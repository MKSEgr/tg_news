package variantselector

import (
	"math"
	"testing"
	"time"

	"ai-content-engine-starter/internal/domain"
)

func TestSelectPrefersBestContextualVariant(t *testing.T) {
	svc := New()

	selectedID, err := svc.Select(SelectionInput{
		ChannelID: 7,
		Topic:     "ai agents",
		At:        time.Date(2026, 3, 18, 9, 0, 0, 0, time.UTC),
		Candidates: []Candidate{
			{DraftID: 101, Variant: "A"},
			{DraftID: 102, Variant: "B"},
		},
		PastPerformance: []PerformanceSignal{
			{Variant: "A", ChannelID: 7, Topic: "ai agents", Hour: 9, Score: 1.2},
			{Variant: "B", ChannelID: 7, Topic: "ai agents", Hour: 9, Score: 2.5},
			{Variant: "B", ChannelID: 7, Topic: "ai agents", Hour: 12, Score: 2.0},
		},
	})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if selectedID != 102 {
		t.Fatalf("selected draft id = %d, want 102", selectedID)
	}
}

func TestSelectFallsBackToConfiguredVariantWithoutContext(t *testing.T) {
	svc := New()

	selectedID, err := svc.Select(SelectionInput{
		ChannelID: 7,
		Topic:     "ai agents",
		At:        time.Date(2026, 3, 18, 9, 0, 0, 0, time.UTC),
		Candidates: []Candidate{
			{DraftID: 101, Variant: "A"},
			{DraftID: 102, Variant: "B"},
		},
		FallbackVariant: "B",
	})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if selectedID != 102 {
		t.Fatalf("selected draft id = %d, want 102", selectedID)
	}
}

func TestSelectUsesFallbackInsteadOfGlobalHistoryWithoutContext(t *testing.T) {
	svc := New()

	selectedID, err := svc.Select(SelectionInput{
		ChannelID: 7,
		Topic:     "ai agents",
		At:        time.Date(2026, 3, 18, 9, 0, 0, 0, time.UTC),
		Candidates: []Candidate{
			{DraftID: 101, Variant: "A"},
			{DraftID: 102, Variant: "B"},
		},
		FallbackVariant: "B",
		PastPerformance: []PerformanceSignal{
			{Variant: "A", ChannelID: 99, Topic: "other", Hour: 12, Score: 5.0},
			{Variant: "B", ChannelID: 99, Topic: "other", Hour: 12, Score: 1.0},
		},
	})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if selectedID != 102 {
		t.Fatalf("selected draft id = %d, want 102", selectedID)
	}
}

func TestSelectFallsBackToVariantAThenLowestID(t *testing.T) {
	svc := New()

	selectedID, err := svc.Select(SelectionInput{
		ChannelID: 7,
		Topic:     "ai agents",
		At:        time.Date(2026, 3, 18, 9, 0, 0, 0, time.UTC),
		Candidates: []Candidate{
			{DraftID: 103, Variant: "B"},
			{DraftID: 102, Variant: "A"},
		},
	})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if selectedID != 102 {
		t.Fatalf("selected draft id = %d, want 102", selectedID)
	}
}

func TestSelectIgnoresInvalidPerformanceValues(t *testing.T) {
	svc := New()

	selectedID, err := svc.Select(SelectionInput{
		ChannelID: 7,
		Topic:     "ai agents",
		At:        time.Date(2026, 3, 18, 9, 0, 0, 0, time.UTC),
		Candidates: []Candidate{
			{DraftID: 101, Variant: "A"},
			{DraftID: 102, Variant: "B"},
		},
		PastPerformance: []PerformanceSignal{
			{Variant: "A", ChannelID: 7, Topic: "ai agents", Hour: 9, Score: 1.0},
			{Variant: "B", ChannelID: 7, Topic: "ai agents", Hour: 9, Score: math.Inf(1)},
		},
	})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if selectedID != 101 {
		t.Fatalf("selected draft id = %d, want 101", selectedID)
	}
}

func TestCandidatesFromDrafts(t *testing.T) {
	candidates := CandidatesFromDrafts([]domain.Draft{
		{ID: 11, Variant: "a"},
		{ID: 0, Variant: "B"},
		{ID: 12, Variant: "B"},
	})
	if len(candidates) != 2 {
		t.Fatalf("candidate count = %d, want 2", len(candidates))
	}
	if candidates[0].Variant != "A" || candidates[1].Variant != "B" {
		t.Fatalf("candidates = %+v", candidates)
	}
}
