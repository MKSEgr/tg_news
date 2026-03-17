package seed

import (
	"context"
	"errors"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

type fakeChannelRepo struct {
	listResult []domain.Channel
	createCall int
	listErr    error
}

func (f *fakeChannelRepo) Create(_ context.Context, c domain.Channel) (domain.Channel, error) {
	f.createCall++
	return c, nil
}
func (f *fakeChannelRepo) GetByID(context.Context, int64) (domain.Channel, error) {
	return domain.Channel{}, domain.ErrNotFound
}
func (f *fakeChannelRepo) List(context.Context) ([]domain.Channel, error) {
	return f.listResult, f.listErr
}

type fakeSourceRepo struct {
	listResult []domain.Source
	createCall int
	listErr    error
}

func (f *fakeSourceRepo) Create(_ context.Context, s domain.Source) (domain.Source, error) {
	f.createCall++
	return s, nil
}
func (f *fakeSourceRepo) GetByID(context.Context, int64) (domain.Source, error) {
	return domain.Source{}, domain.ErrNotFound
}
func (f *fakeSourceRepo) List(context.Context) ([]domain.Source, error) {
	return f.listResult, f.listErr
}
func (f *fakeSourceRepo) ListEnabled(context.Context) ([]domain.Source, error) {
	return f.listResult, f.listErr
}

func TestSeedCreatesDefaultsWhenEmpty(t *testing.T) {
	channels := &fakeChannelRepo{}
	sources := &fakeSourceRepo{}
	seeder := New(channels, sources)

	if err := seeder.Seed(context.Background()); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}
	if channels.createCall != len(DefaultChannels) {
		t.Fatalf("channel create calls = %d, want %d", channels.createCall, len(DefaultChannels))
	}
	if sources.createCall != len(DefaultSources) {
		t.Fatalf("source create calls = %d, want %d", sources.createCall, len(DefaultSources))
	}
}

func TestSeedSkipsWhenDataExists(t *testing.T) {
	channels := &fakeChannelRepo{listResult: []domain.Channel{{ID: 1, Slug: "ai-news", Name: "AI News"}}}
	sources := &fakeSourceRepo{listResult: []domain.Source{{ID: 1, Name: "AI News RSS", Enabled: true}}}
	seeder := New(channels, sources)

	if err := seeder.Seed(context.Background()); err != nil {
		t.Fatalf("Seed() error = %v", err)
	}
	if channels.createCall != 0 {
		t.Fatalf("channel create calls = %d, want 0", channels.createCall)
	}
	if sources.createCall != 0 {
		t.Fatalf("source create calls = %d, want 0", sources.createCall)
	}
}

func TestSeedReturnsChannelListError(t *testing.T) {
	channels := &fakeChannelRepo{listErr: errors.New("db down")}
	sources := &fakeSourceRepo{}
	seeder := New(channels, sources)

	if err := seeder.Seed(context.Background()); err == nil {
		t.Fatalf("Seed() expected error")
	}
}

func TestSeedReturnsSourceListError(t *testing.T) {
	channels := &fakeChannelRepo{listResult: []domain.Channel{{ID: 1, Slug: "ai-news", Name: "AI News"}}}
	sources := &fakeSourceRepo{listErr: errors.New("db down")}
	seeder := New(channels, sources)

	if err := seeder.Seed(context.Background()); err == nil {
		t.Fatalf("Seed() expected error")
	}
}

func TestSeedRejectsNilRepositories(t *testing.T) {
	if err := New(nil, &fakeSourceRepo{}).Seed(context.Background()); err == nil {
		t.Fatalf("Seed() expected error for nil channel repository")
	}
	if err := New(&fakeChannelRepo{}, nil).Seed(context.Background()); err == nil {
		t.Fatalf("Seed() expected error for nil source repository")
	}
}
