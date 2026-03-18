package contentrules

import (
	"context"
	"fmt"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

// Decision describes explainable rule matching outcome.
type Decision struct {
	Allowed bool
	Reason  string
	Rule    *domain.ContentRule
}

// Service evaluates blacklist/whitelist rules deterministically.
type Service struct {
	repo domain.ContentRuleRepository
}

// New creates a content rules service.
func New(repo domain.ContentRuleRepository) (*Service, error) {
	if repo == nil {
		return nil, fmt.Errorf("content rule repository is nil")
	}
	return &Service{repo: repo}, nil
}

// Evaluate checks text against enabled rules for a channel.
func (s *Service) Evaluate(ctx context.Context, channelID int64, text string) (Decision, error) {
	if s == nil {
		return Decision{}, fmt.Errorf("content rules service is nil")
	}
	if s.repo == nil {
		return Decision{}, fmt.Errorf("content rule repository is nil")
	}
	if ctx == nil {
		return Decision{}, fmt.Errorf("context is nil")
	}
	if channelID <= 0 {
		return Decision{}, fmt.Errorf("channel id is invalid")
	}
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return Decision{}, fmt.Errorf("text is empty")
	}

	cid := channelID
	rules, err := s.repo.ListEnabled(ctx, &cid)
	if err != nil {
		return Decision{}, fmt.Errorf("list enabled content rules: %w", err)
	}

	// deterministic ordering already provided by repository (kind, pattern asc)
	hasWhitelist := false
	for i := range rules {
		r := rules[i]
		if r.Kind == domain.ContentRuleKindWhitelist {
			hasWhitelist = true
		}
		pattern := strings.ToLower(strings.TrimSpace(r.Pattern))
		if pattern == "" || !strings.Contains(text, pattern) {
			continue
		}
		if r.Kind == domain.ContentRuleKindBlacklist {
			return Decision{Allowed: false, Reason: "matched blacklist rule", Rule: &r}, nil
		}
		if r.Kind == domain.ContentRuleKindWhitelist {
			copy := r
			return Decision{Allowed: true, Reason: "matched whitelist rule", Rule: &copy}, nil
		}
	}
	if hasWhitelist {
		return Decision{Allowed: false, Reason: "no whitelist rule matched"}, nil
	}

	return Decision{Allowed: true, Reason: "no blocking rule matched"}, nil
}

// AddRule validates and stores a new rule.
func (s *Service) AddRule(ctx context.Context, rule domain.ContentRule) (domain.ContentRule, error) {
	if s == nil {
		return domain.ContentRule{}, fmt.Errorf("content rules service is nil")
	}
	if s.repo == nil {
		return domain.ContentRule{}, fmt.Errorf("content rule repository is nil")
	}
	if ctx == nil {
		return domain.ContentRule{}, fmt.Errorf("context is nil")
	}
	rule.Pattern = strings.TrimSpace(strings.ToLower(rule.Pattern))
	if rule.Pattern == "" {
		return domain.ContentRule{}, fmt.Errorf("rule pattern is empty")
	}
	if rule.Kind != domain.ContentRuleKindBlacklist && rule.Kind != domain.ContentRuleKindWhitelist {
		return domain.ContentRule{}, fmt.Errorf("rule kind is invalid")
	}
	if rule.ChannelID != nil && *rule.ChannelID <= 0 {
		return domain.ContentRule{}, fmt.Errorf("channel id is invalid")
	}
	rule.Enabled = true
	return s.repo.Create(ctx, rule)
}

// EvaluateAllowed is a minimal adapter for pipeline integration without leaking Decision type.
func (s *Service) EvaluateAllowed(ctx context.Context, channelID int64, text string) (bool, error) {
	decision, err := s.Evaluate(ctx, channelID, text)
	if err != nil {
		return false, err
	}
	return decision.Allowed, nil
}
