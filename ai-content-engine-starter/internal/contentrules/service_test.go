package contentrules

import (
	"context"
	"errors"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

type repoStub struct {
	rules   []domain.ContentRule
	created domain.ContentRule
	err     error
}

func (r *repoStub) Create(_ context.Context, rule domain.ContentRule) (domain.ContentRule, error) {
	if r.err != nil {
		return domain.ContentRule{}, r.err
	}
	r.created = rule
	return rule, nil
}

func (r *repoStub) ListEnabled(_ context.Context, _ *int64) ([]domain.ContentRule, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.rules, nil
}

func TestEvaluateHonorsBlacklist(t *testing.T) {
	repo := &repoStub{rules: []domain.ContentRule{{Kind: domain.ContentRuleKindBlacklist, Pattern: "spam"}}}
	svc, _ := New(repo)
	d, err := svc.Evaluate(context.Background(), 1, "this is spam content")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if d.Allowed {
		t.Fatalf("expected blocked decision")
	}
}

func TestEvaluateAllowsWhenNoMatch(t *testing.T) {
	repo := &repoStub{rules: []domain.ContentRule{{Kind: domain.ContentRuleKindBlacklist, Pattern: "spam"}}}
	svc, _ := New(repo)
	d, err := svc.Evaluate(context.Background(), 1, "legit update")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !d.Allowed {
		t.Fatalf("expected allowed decision")
	}
}

func TestAddRuleValidation(t *testing.T) {
	repo := &repoStub{}
	svc, _ := New(repo)
	if _, err := svc.AddRule(context.Background(), domain.ContentRule{}); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestEvaluateValidationAndErrors(t *testing.T) {
	repo := &repoStub{}
	svc, _ := New(repo)
	if _, err := svc.Evaluate(nil, 1, "a"); err == nil {
		t.Fatalf("expected nil context error")
	}
	if _, err := svc.Evaluate(context.Background(), 0, "a"); err == nil {
		t.Fatalf("expected invalid channel id error")
	}
	repo.err = errors.New("db")
	if _, err := svc.Evaluate(context.Background(), 1, "a"); err == nil {
		t.Fatalf("expected repo error")
	}
}

func TestNewValidation(t *testing.T) {
	if _, err := New(nil); err == nil {
		t.Fatalf("expected error for nil repo")
	}
}

func TestEvaluateAllowsOnWhitelistMatch(t *testing.T) {
	repo := &repoStub{rules: []domain.ContentRule{{Kind: domain.ContentRuleKindWhitelist, Pattern: "trusted"}}}
	svc, _ := New(repo)
	d, err := svc.Evaluate(context.Background(), 1, "trusted source update")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if !d.Allowed {
		t.Fatalf("expected allowed decision")
	}
	if d.Reason != "matched whitelist rule" {
		t.Fatalf("reason = %q, want matched whitelist rule", d.Reason)
	}
}

func TestEvaluateBlocksWhenWhitelistExistsAndNoMatch(t *testing.T) {
	repo := &repoStub{rules: []domain.ContentRule{{Kind: domain.ContentRuleKindWhitelist, Pattern: "trusted"}}}
	svc, _ := New(repo)
	d, err := svc.Evaluate(context.Background(), 1, "general source update")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}
	if d.Allowed {
		t.Fatalf("expected blocked decision")
	}
	if d.Reason != "no whitelist rule matched" {
		t.Fatalf("reason = %q, want no whitelist rule matched", d.Reason)
	}
}
