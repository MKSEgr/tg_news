package sourcediscovery

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

const defaultMaxCandidates = 5
const minDiscoveryScore = 0.2

// Service derives deterministic source candidates from collected item URLs.
type Service struct {
	sources       domain.SourceRepository
	maxCandidates int
	analytics     analyticsSummaryBuilder
	rules         channelRuleEvaluator
}

// ChannelMetrics contains minimal analytics fields required by discovery integration.
type ChannelMetrics struct {
	ChannelID      int64
	FeedbackDrafts int
	AvgScore       float64
}

type analyticsSummaryBuilder interface {
	BuildByChannelMetrics(ctx context.Context) ([]ChannelMetrics, error)
}

type channelRuleEvaluator interface {
	EvaluateAllowed(ctx context.Context, channelID int64, text string) (bool, error)
}

// New creates source discovery service.
func New(sources domain.SourceRepository) (*Service, error) {
	if sources == nil {
		return nil, fmt.Errorf("source repository is nil")
	}
	return &Service{sources: sources, maxCandidates: defaultMaxCandidates}, nil
}

// WithAnalytics attaches optional analytics dependency for channel-aware discovery gating.
func (s *Service) WithAnalytics(builder analyticsSummaryBuilder) *Service {
	if s == nil {
		return nil
	}
	s.analytics = builder
	return s
}

// WithRules attaches optional rules dependency for filtering discovered candidates.
func (s *Service) WithRules(rules channelRuleEvaluator) *Service {
	if s == nil {
		return nil
	}
	s.rules = rules
	return s
}

// Discover returns at most maxCandidates newly discovered source candidates.
// It does not persist anything yet; integration is done in follow-up tasks.
func (s *Service) Discover(ctx context.Context, items []domain.SourceItem) ([]domain.Source, error) {
	if s == nil {
		return nil, fmt.Errorf("source discovery service is nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	existing, err := s.sources.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	existingEndpoints := make(map[string]struct{}, len(existing))
	existingHosts := make(map[string]struct{}, len(existing))
	for _, src := range existing {
		u, err := url.Parse(strings.TrimSpace(src.Endpoint))
		if err == nil && strings.TrimSpace(u.Host) != "" {
			existingHosts[strings.ToLower(strings.TrimSpace(u.Host))] = struct{}{}
		}
		existingEndpoints[normalizeEndpoint(src.Endpoint)] = struct{}{}
	}

	candidatesByEndpoint := make(map[string]domain.Source)
	discoveredHosts := make(map[string]struct{})
	for _, item := range items {
		endpoint, host, ok := candidateEndpoint(item.URL)
		if !ok {
			continue
		}
		if _, ok := existingEndpoints[endpoint]; ok {
			continue
		}
		if _, ok := existingHosts[host]; ok {
			continue
		}
		if _, ok := discoveredHosts[host]; ok {
			continue
		}
		if _, ok := candidatesByEndpoint[endpoint]; ok {
			continue
		}

		candidatesByEndpoint[endpoint] = domain.Source{
			Kind:     "rss",
			Name:     fmt.Sprintf("Discovered %s", host),
			Endpoint: endpoint,
			Enabled:  false,
		}
		discoveredHosts[host] = struct{}{}
	}

	endpoints := make([]string, 0, len(candidatesByEndpoint))
	for endpoint := range candidatesByEndpoint {
		endpoints = append(endpoints, endpoint)
	}
	sort.Strings(endpoints)

	limit := s.maxCandidates
	if limit <= 0 {
		limit = defaultMaxCandidates
	}
	if len(endpoints) < limit {
		limit = len(endpoints)
	}

	out := make([]domain.Source, 0, limit)
	for i := 0; i < limit; i++ {
		out = append(out, candidatesByEndpoint[endpoints[i]])
	}
	return out, nil
}

// DiscoverForChannel runs discovery and applies optional analytics/rules for one channel.
func (s *Service) DiscoverForChannel(ctx context.Context, channelID int64, items []domain.SourceItem) ([]domain.Source, error) {
	if s == nil {
		return nil, fmt.Errorf("source discovery service is nil")
	}
	if ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}
	if channelID <= 0 {
		return nil, fmt.Errorf("channel id is invalid")
	}

	if s.analytics != nil {
		summaries, err := s.analytics.BuildByChannelMetrics(ctx)
		if err != nil {
			return nil, fmt.Errorf("build analytics summaries: %w", err)
		}
		for _, summary := range summaries {
			if summary.ChannelID != channelID {
				continue
			}
			if summary.FeedbackDrafts > 0 && summary.AvgScore < minDiscoveryScore {
				return []domain.Source{}, nil
			}
			break
		}
	}

	candidates, err := s.Discover(ctx, items)
	if err != nil {
		return nil, err
	}
	if s.rules == nil {
		return candidates, nil
	}

	filtered := make([]domain.Source, 0, len(candidates))
	for _, candidate := range candidates {
		allowed, err := s.rules.EvaluateAllowed(ctx, channelID, strings.TrimSpace(candidate.Name+" "+candidate.Endpoint))
		if err != nil {
			return nil, fmt.Errorf("evaluate rules for discovered source %s: %w", candidate.Endpoint, err)
		}
		if allowed {
			filtered = append(filtered, candidate)
		}
	}
	return filtered, nil
}

func candidateEndpoint(raw string) (endpoint string, host string, ok bool) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return "", "", false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", "", false
	}
	host = strings.ToLower(strings.TrimSpace(u.Host))
	if host == "" {
		return "", "", false
	}

	cleanPath := strings.ToLower(strings.TrimSpace(path.Clean(u.Path)))
	if cleanPath == "." || cleanPath == "/" {
		return normalizeEndpoint(fmt.Sprintf("%s://%s/feed", u.Scheme, host)), host, true
	}
	if strings.Contains(cleanPath, "rss") || strings.Contains(cleanPath, "feed") || strings.HasSuffix(cleanPath, ".xml") {
		return normalizeEndpoint(fmt.Sprintf("%s://%s%s", u.Scheme, host, cleanPath)), host, true
	}
	return normalizeEndpoint(fmt.Sprintf("%s://%s/feed", u.Scheme, host)), host, true
}

func normalizeEndpoint(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return strings.TrimSpace(strings.ToLower(raw))
	}
	u.Scheme = strings.ToLower(strings.TrimSpace(u.Scheme))
	u.Host = strings.ToLower(strings.TrimSpace(u.Host))
	u.RawQuery = ""
	u.Fragment = ""
	return strings.TrimRight(u.String(), "/")
}
