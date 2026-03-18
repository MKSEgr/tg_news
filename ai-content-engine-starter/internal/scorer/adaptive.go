package scorer

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"sort"
	"strings"
	"unicode"

	"ai-content-engine-starter/internal/domain"
)

const (
	adaptiveFeatureLimit     = 5
	maxAdaptiveChannelPoints = 10
	maxAdaptiveTopicPoints   = 8
	maxAdaptiveFormatPoints  = 6
	maxAdaptiveTotalPoints   = 20
	defaultDraftFormat       = "text"
)

const (
	rankingEntityTypeChannel = "channel"
	rankingEntityTypeTopic   = "topic"
	rankingEntityTypeFormat  = "format"

	rankingFeatureChannelPerformance = "channel_performance"
	rankingFeatureTopicPerformance   = "topic_performance"
	rankingFeatureFormatSuccess      = "format_success"
)

const (
	// RankingEntityTypeChannel identifies per-channel adaptive features.
	RankingEntityTypeChannel = rankingEntityTypeChannel
	// RankingEntityTypeTopic identifies per-topic adaptive features.
	RankingEntityTypeTopic = rankingEntityTypeTopic
	// RankingEntityTypeFormat identifies per-format adaptive features.
	RankingEntityTypeFormat = rankingEntityTypeFormat

	// RankingFeatureNameChannelPerformance stores bounded per-channel performance priors.
	RankingFeatureNameChannelPerformance = rankingFeatureChannelPerformance
	// RankingFeatureNameTopicPerformance stores bounded per-topic performance priors.
	RankingFeatureNameTopicPerformance = rankingFeatureTopicPerformance
	// RankingFeatureNameFormatSuccess stores bounded per-format success priors.
	RankingFeatureNameFormatSuccess = rankingFeatureFormatSuccess
)

type rankingFeatureReader interface {
	ListByEntity(ctx context.Context, entityType string, entityID int64, limit int) ([]domain.RankingFeature, error)
}

type baseScorer interface {
	Score(item domain.SourceItem) int
}

// AdaptiveBreakdown exposes deterministic and bounded score modifiers.
type AdaptiveBreakdown struct {
	BaseScore         int
	ChannelAdjustment int
	TopicAdjustment   int
	FormatAdjustment  int
	TotalAdjustment   int
	FinalScore        int
}

// AdaptiveService adds a bounded explainable adjustment layer on top of the base scorer.
type AdaptiveService struct {
	base     baseScorer
	features rankingFeatureReader
	limit    int
}

// NewAdaptive creates an adaptive scoring layer.
func NewAdaptive(base baseScorer, features rankingFeatureReader) (*AdaptiveService, error) {
	if base == nil {
		return nil, fmt.Errorf("base scorer is nil")
	}
	if features == nil {
		return nil, fmt.Errorf("ranking feature reader is nil")
	}
	return &AdaptiveService{base: base, features: features, limit: adaptiveFeatureLimit}, nil
}

// Score preserves the base scorer behavior for existing integrations.
func (s *AdaptiveService) Score(item domain.SourceItem) int {
	if s == nil || s.base == nil {
		return 0
	}
	return s.base.Score(item)
}

// ScoreAdaptive returns only the final bounded score for per-channel pipeline use.
func (s *AdaptiveService) ScoreAdaptive(ctx context.Context, item domain.SourceItem, channelID int64, format string) (int, error) {
	breakdown, err := s.ExplainAdaptive(ctx, item, channelID, format)
	if err != nil {
		return 0, err
	}
	return breakdown.FinalScore, nil
}

// ExplainAdaptive returns the full bounded adjustment breakdown.
func (s *AdaptiveService) ExplainAdaptive(ctx context.Context, item domain.SourceItem, channelID int64, format string) (AdaptiveBreakdown, error) {
	if s == nil {
		return AdaptiveBreakdown{}, fmt.Errorf("adaptive scorer is nil")
	}
	if s.base == nil {
		return AdaptiveBreakdown{}, fmt.Errorf("base scorer is nil")
	}
	if s.features == nil {
		return AdaptiveBreakdown{}, fmt.Errorf("ranking feature reader is nil")
	}
	if ctx == nil {
		return AdaptiveBreakdown{}, fmt.Errorf("context is nil")
	}
	if channelID <= 0 {
		return AdaptiveBreakdown{}, fmt.Errorf("channel id is invalid")
	}
	if s.limit <= 0 {
		s.limit = adaptiveFeatureLimit
	}

	baseScore := s.base.Score(item)
	channelAdj, err := s.entityAdjustment(ctx, rankingEntityTypeChannel, channelID, rankingFeatureChannelPerformance, maxAdaptiveChannelPoints)
	if err != nil {
		return AdaptiveBreakdown{}, fmt.Errorf("channel adjustment: %w", err)
	}

	topicAdj := 0
	for _, topic := range extractAdaptiveTopics(item.Title, item.Body) {
		adj, err := s.entityAdjustment(ctx, rankingEntityTypeTopic, topicEntityID(topic), rankingFeatureTopicPerformance, maxAdaptiveTopicPoints)
		if err != nil {
			return AdaptiveBreakdown{}, fmt.Errorf("topic adjustment %q: %w", topic, err)
		}
		topicAdj += adj
		if topicAdj > maxAdaptiveTopicPoints {
			topicAdj = maxAdaptiveTopicPoints
			break
		}
		if topicAdj < -maxAdaptiveTopicPoints {
			topicAdj = -maxAdaptiveTopicPoints
			break
		}
	}

	format = strings.TrimSpace(strings.ToLower(format))
	if format == "" {
		format = defaultDraftFormat
	}
	formatAdj, err := s.entityAdjustment(ctx, rankingEntityTypeFormat, formatEntityID(format), rankingFeatureFormatSuccess, maxAdaptiveFormatPoints)
	if err != nil {
		return AdaptiveBreakdown{}, fmt.Errorf("format adjustment: %w", err)
	}

	totalAdj := channelAdj + topicAdj + formatAdj
	if totalAdj > maxAdaptiveTotalPoints {
		totalAdj = maxAdaptiveTotalPoints
	}
	if totalAdj < -maxAdaptiveTotalPoints {
		totalAdj = -maxAdaptiveTotalPoints
	}
	finalScore := baseScore + totalAdj
	if finalScore < 0 {
		finalScore = 0
	}
	if finalScore > maxScore {
		finalScore = maxScore
	}
	return AdaptiveBreakdown{
		BaseScore:         baseScore,
		ChannelAdjustment: channelAdj,
		TopicAdjustment:   topicAdj,
		FormatAdjustment:  formatAdj,
		TotalAdjustment:   totalAdj,
		FinalScore:        finalScore,
	}, nil
}

func (s *AdaptiveService) entityAdjustment(ctx context.Context, entityType string, entityID int64, featureName string, maxAbs int) (int, error) {
	features, err := s.features.ListByEntity(ctx, entityType, entityID, s.limit)
	if err != nil {
		return 0, err
	}
	values := make([]float64, 0, len(features))
	for _, feature := range features {
		if strings.TrimSpace(feature.FeatureName) != featureName {
			continue
		}
		if math.IsNaN(feature.FeatureValue) || math.IsInf(feature.FeatureValue, 0) {
			continue
		}
		values = append(values, feature.FeatureValue)
	}
	if len(values) == 0 {
		return 0, nil
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	points := int(math.Round((total / float64(len(values))) * 4))
	if points > maxAbs {
		return maxAbs, nil
	}
	if points < -maxAbs {
		return -maxAbs, nil
	}
	return points, nil
}

func extractAdaptiveTopics(title string, body *string) []string {
	text := strings.ToLower(strings.TrimSpace(title))
	if body != nil {
		text += " " + strings.ToLower(strings.TrimSpace(*body))
	}
	if strings.TrimSpace(text) == "" {
		return nil
	}
	stopWords := map[string]struct{}{"the": {}, "and": {}, "for": {}, "with": {}, "this": {}, "that": {}, "from": {}, "into": {}, "after": {}, "over": {}, "what": {}, "when": {}, "where": {}}
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	seen := make(map[string]struct{}, len(parts))
	topics := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len([]rune(part)) < 4 {
			continue
		}
		if _, blocked := stopWords[part]; blocked {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		topics = append(topics, part)
	}
	sort.Strings(topics)
	if len(topics) > 3 {
		topics = topics[:3]
	}
	return topics
}

func topicEntityID(topic string) int64 {
	return stableEntityID("topic:" + strings.TrimSpace(strings.ToLower(topic)))
}
func formatEntityID(format string) int64 {
	return stableEntityID("format:" + strings.TrimSpace(strings.ToLower(format)))
}

// TopicEntityID returns the deterministic ranking_features entity ID for a topic.
func TopicEntityID(topic string) int64 {
	return topicEntityID(topic)
}

// FormatEntityID returns the deterministic ranking_features entity ID for a draft format.
func FormatEntityID(format string) int64 {
	return formatEntityID(format)
}

func stableEntityID(value string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(value))
	return int64(h.Sum64() & math.MaxInt64)
}
