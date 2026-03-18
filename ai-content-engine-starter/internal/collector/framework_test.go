package collector

import (
	"context"
	"errors"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

type fakeSourceRepo struct {
	enabled []domain.Source
	err     error
}

func (f *fakeSourceRepo) Create(context.Context, domain.Source) (domain.Source, error) {
	return domain.Source{}, nil
}
func (f *fakeSourceRepo) GetByID(context.Context, int64) (domain.Source, error) {
	return domain.Source{}, domain.ErrNotFound
}
func (f *fakeSourceRepo) List(context.Context) ([]domain.Source, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.enabled, nil
}
func (f *fakeSourceRepo) ListEnabled(context.Context) ([]domain.Source, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.enabled, nil
}

type fakeItemRepo struct {
	created []domain.SourceItem
	err     error
}

func (f *fakeItemRepo) Create(_ context.Context, item domain.SourceItem) (domain.SourceItem, error) {
	if f.err != nil {
		return domain.SourceItem{}, f.err
	}
	f.created = append(f.created, item)
	return item, nil
}
func (f *fakeItemRepo) GetByID(context.Context, int64) (domain.SourceItem, error) {
	return domain.SourceItem{}, domain.ErrNotFound
}
func (f *fakeItemRepo) ListBySourceID(context.Context, int64, int) ([]domain.SourceItem, error) {
	return nil, nil
}

type fakeCollector struct {
	kind  string
	items []domain.SourceItem
	err   error
}

func (f *fakeCollector) Kind() string { return f.kind }
func (f *fakeCollector) Collect(context.Context, domain.Source) ([]domain.SourceItem, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.items, nil
}

func TestRunOnceCollectsAndStoresItems(t *testing.T) {
	sources := &fakeSourceRepo{enabled: []domain.Source{{ID: 10, Kind: "rss", Enabled: true}}}
	items := &fakeItemRepo{}
	collector := &fakeCollector{kind: "rss", items: []domain.SourceItem{{ExternalID: "1", Title: "T", URL: "https://e"}}}

	framework, err := New(sources, items, collector)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := framework.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce() error = %v", err)
	}

	if len(items.created) != 1 {
		t.Fatalf("created items = %d, want 1", len(items.created))
	}
	if items.created[0].SourceID != 10 {
		t.Fatalf("stored item source_id = %d, want 10", items.created[0].SourceID)
	}
}

func TestRunOnceReturnsErrorWhenCollectorMissing(t *testing.T) {
	sources := &fakeSourceRepo{enabled: []domain.Source{{ID: 1, Kind: "rss", Enabled: true}}}
	items := &fakeItemRepo{}

	framework, err := New(sources, items)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := framework.RunOnce(context.Background()); err == nil {
		t.Fatalf("RunOnce() expected error for missing collector")
	}
}

func TestNewRejectsDuplicateCollectorKinds(t *testing.T) {
	sources := &fakeSourceRepo{}
	items := &fakeItemRepo{}

	_, err := New(sources, items, &fakeCollector{kind: "rss"}, &fakeCollector{kind: "rss"})
	if err == nil {
		t.Fatalf("New() expected duplicate collector kind error")
	}
}

func TestRunOncePropagatesSourceListError(t *testing.T) {
	sources := &fakeSourceRepo{err: errors.New("db down")}
	items := &fakeItemRepo{}
	framework, err := New(sources, items, &fakeCollector{kind: "rss"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if err := framework.RunOnce(context.Background()); err == nil {
		t.Fatalf("RunOnce() expected error")
	}
}

func TestRunOnceRejectsNilFramework(t *testing.T) {
	var framework *Framework
	if err := framework.RunOnce(context.Background()); err == nil {
		t.Fatalf("RunOnce() expected error for nil framework")
	}
}

func TestRunOnceRejectsNilContext(t *testing.T) {
	sources := &fakeSourceRepo{}
	items := &fakeItemRepo{}
	framework, err := New(sources, items, &fakeCollector{kind: "rss"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	if err := framework.RunOnce(nil); err == nil {
		t.Fatalf("RunOnce() expected error for nil context")
	}
}
