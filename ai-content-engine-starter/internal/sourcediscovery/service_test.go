package sourcediscovery

import (
	"context"
	"errors"
	"strings"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

type sourceRepoStub struct {
	sources []domain.Source
	err     error
}

type analyticsStub struct {
	summaries []analyticsSummary
	err       error
}

func (s *analyticsStub) BuildByChannelMetrics(context.Context) ([]analyticsSummary, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.summaries, nil
}

type rulesStub struct {
	denySubstring string
	err           error
}

func (r *rulesStub) EvaluateAllowed(_ context.Context, _ int64, text string) (bool, error) {
	if r.err != nil {
		return false, r.err
	}
	if r.denySubstring != "" && strings.Contains(strings.ToLower(text), strings.ToLower(r.denySubstring)) {
		return false, nil
	}
	return true, nil
}

func (s *sourceRepoStub) Create(context.Context, domain.Source) (domain.Source, error) {
	return domain.Source{}, nil
}
func (s *sourceRepoStub) GetByID(context.Context, int64) (domain.Source, error) {
	return domain.Source{}, nil
}
func (s *sourceRepoStub) List(context.Context) ([]domain.Source, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.sources, nil
}
func (s *sourceRepoStub) ListEnabled(context.Context) ([]domain.Source, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.sources, nil
}

func TestNewValidation(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatalf("expected nil repository error")
	}
}

func TestDiscoverValidation(t *testing.T) {
	svc, err := New(&sourceRepoStub{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := svc.Discover(nil, nil); err == nil {
		t.Fatalf("expected nil context error")
	}

	var nilSvc *Service
	if _, err := nilSvc.Discover(context.Background(), nil); err == nil {
		t.Fatalf("expected nil service error")
	}
}

func TestDiscoverReturnsDeterministicCandidates(t *testing.T) {
	svc, err := New(&sourceRepoStub{sources: []domain.Source{{Endpoint: "https://known.example/feed", Enabled: true}}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	items := []domain.SourceItem{
		{URL: "https://known.example/news/1"},
		{URL: "https://zeta.example/articles/1"},
		{URL: "https://alpha.example/rss.xml?from=utm"},
		{URL: "https://alpha.example/news/2"},
		{URL: "not-a-url"},
	}

	got, err := svc.Discover(context.Background(), items)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("candidates = %d, want 2", len(got))
	}
	if got[0].Endpoint != "https://alpha.example/rss.xml" {
		t.Fatalf("first endpoint = %q", got[0].Endpoint)
	}
	if got[1].Endpoint != "https://zeta.example/feed" {
		t.Fatalf("second endpoint = %q", got[1].Endpoint)
	}
	for _, source := range got {
		if source.Kind != "rss" {
			t.Fatalf("kind = %q, want rss", source.Kind)
		}
		if source.Enabled {
			t.Fatalf("enabled = true, want false for discovered candidates")
		}
	}
}

func TestDiscoverListEnabledError(t *testing.T) {
	svc, err := New(&sourceRepoStub{err: errors.New("boom")})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := svc.Discover(context.Background(), []domain.SourceItem{{URL: "https://a.example/news"}}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDiscoverForChannelFiltersWithRules(t *testing.T) {
	svc, err := New(&sourceRepoStub{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	svc.WithRules(&rulesStub{denySubstring: "blocked.example"})

	items := []domain.SourceItem{{URL: "https://blocked.example/feed.xml"}, {URL: "https://open.example/feed.xml"}}
	got, err := svc.DiscoverForChannel(context.Background(), 1, items)
	if err != nil {
		t.Fatalf("DiscoverForChannel() error = %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("candidates = %d, want 1", len(got))
	}
	if got[0].Endpoint != "https://open.example/feed.xml" {
		t.Fatalf("endpoint = %q", got[0].Endpoint)
	}
}

func TestDiscoverForChannelSkipsWhenAnalyticsScoreIsBelowThreshold(t *testing.T) {
	svc, err := New(&sourceRepoStub{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	svc.WithAnalytics(&analyticsStub{summaries: []analyticsSummary{{ChannelID: 9, FeedbackDrafts: 2, AvgScore: 0.1}}})

	got, err := svc.DiscoverForChannel(context.Background(), 9, []domain.SourceItem{{URL: "https://open.example/feed.xml"}})
	if err != nil {
		t.Fatalf("DiscoverForChannel() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("candidates = %d, want 0", len(got))
	}
}

func TestDiscoverForChannelValidationAndErrors(t *testing.T) {
	svc, err := New(&sourceRepoStub{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := svc.DiscoverForChannel(context.Background(), 0, nil); err == nil {
		t.Fatalf("expected channel id validation error")
	}

	svc.WithAnalytics(&analyticsStub{err: errors.New("boom")})
	if _, err := svc.DiscoverForChannel(context.Background(), 1, nil); err == nil {
		t.Fatalf("expected analytics error")
	}

	svc = svc.WithAnalytics(nil).WithRules(&rulesStub{err: errors.New("boom")})
	if _, err := svc.DiscoverForChannel(context.Background(), 1, []domain.SourceItem{{URL: "https://open.example/feed.xml"}}); err == nil {
		t.Fatalf("expected rules error")
	}
}
