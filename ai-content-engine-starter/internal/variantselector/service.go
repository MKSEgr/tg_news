package variantselector

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"ai-content-engine-starter/internal/domain"
)

const defaultFallbackVariant = "A"

// Candidate represents one selectable draft variant from the existing V2 A/B flow.
type Candidate struct {
	DraftID int64
	Variant string
}

// PerformanceSignal is a lightweight contextual performance summary for one variant.
type PerformanceSignal struct {
	Variant   string
	ChannelID int64
	Topic     string
	Hour      int
	Score     float64
}

// SelectionInput contains the deterministic context used for variant selection.
type SelectionInput struct {
	ChannelID       int64
	Topic           string
	At              time.Time
	Candidates      []Candidate
	PastPerformance []PerformanceSignal
	FallbackVariant string
}

// Service selects the best available variant using simple contextual rules.
type Service struct{}

// New creates a contextual variant-selection service.
func New() *Service { return &Service{} }

// Select returns the selected draft ID using channel/topic/time context plus past performance.
func (s *Service) Select(input SelectionInput) (int64, error) {
	if s == nil {
		return 0, fmt.Errorf("variant selector service is nil")
	}
	if input.ChannelID <= 0 {
		return 0, fmt.Errorf("channel id is invalid")
	}
	if input.At.IsZero() {
		return 0, fmt.Errorf("selection time is required")
	}
	if len(input.Candidates) == 0 {
		return 0, fmt.Errorf("at least one candidate is required")
	}

	fallbackVariant := normalizeVariant(input.FallbackVariant)
	if fallbackVariant == "" {
		fallbackVariant = defaultFallbackVariant
	}

	normalizedTopic := normalizeTopic(input.Topic)
	hour := input.At.UTC().Hour()

	type scoredCandidate struct {
		Candidate
		score         float64
		hasContextual bool
	}

	scored := make([]scoredCandidate, 0, len(input.Candidates))
	seenIDs := make(map[int64]struct{}, len(input.Candidates))
	for _, candidate := range input.Candidates {
		if candidate.DraftID <= 0 {
			return 0, fmt.Errorf("candidate draft id is invalid")
		}
		if _, exists := seenIDs[candidate.DraftID]; exists {
			return 0, fmt.Errorf("candidate draft id %d is duplicated", candidate.DraftID)
		}
		seenIDs[candidate.DraftID] = struct{}{}

		variant := normalizeVariant(candidate.Variant)
		if variant == "" {
			variant = defaultFallbackVariant
		}
		if variant != "A" && variant != "B" {
			return 0, fmt.Errorf("candidate variant is invalid")
		}

		score, hasContext := contextualScore(variant, input.ChannelID, normalizedTopic, hour, input.PastPerformance)
		scored = append(scored, scoredCandidate{
			Candidate:     Candidate{DraftID: candidate.DraftID, Variant: variant},
			score:         score,
			hasContextual: hasContext,
		})
	}

	sort.SliceStable(scored, func(i, j int) bool {
		left := scored[i]
		right := scored[j]
		if left.hasContextual != right.hasContextual {
			return left.hasContextual
		}
		if left.hasContextual && right.hasContextual && left.score != right.score {
			return left.score > right.score
		}
		if left.Variant == fallbackVariant && right.Variant != fallbackVariant {
			return true
		}
		if right.Variant == fallbackVariant && left.Variant != fallbackVariant {
			return false
		}
		return left.DraftID < right.DraftID
	})

	return scored[0].DraftID, nil
}

func contextualScore(variant string, channelID int64, topic string, hour int, history []PerformanceSignal) (float64, bool) {
	total := 0.0
	weightTotal := 0.0
	hasContext := false

	for _, signal := range history {
		if normalizeVariant(signal.Variant) != variant {
			continue
		}
		if math.IsNaN(signal.Score) || math.IsInf(signal.Score, 0) {
			continue
		}

		weight := 1.0
		if signal.ChannelID > 0 && signal.ChannelID == channelID {
			weight += 2
			hasContext = true
		}
		if topic != "" && normalizeTopic(signal.Topic) == topic {
			weight += 2
			hasContext = true
		}
		if signal.Hour >= 0 && signal.Hour <= 23 && signal.Hour == hour {
			weight += 1
			hasContext = true
		}

		total += signal.Score * weight
		weightTotal += weight
	}

	if weightTotal == 0 {
		return 0, false
	}
	return total / weightTotal, hasContext
}

func normalizeVariant(raw string) string {
	value := strings.ToUpper(strings.TrimSpace(raw))
	if value == "" {
		return ""
	}
	return value
}

func normalizeTopic(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

// CandidatesFromDrafts converts stored drafts into selection candidates.
func CandidatesFromDrafts(drafts []domain.Draft) []Candidate {
	out := make([]Candidate, 0, len(drafts))
	for _, draft := range drafts {
		if draft.ID <= 0 {
			continue
		}
		out = append(out, Candidate{
			DraftID: draft.ID,
			Variant: normalizeVariant(draft.Variant),
		})
	}
	return out
}
