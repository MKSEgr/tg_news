package assetgen

import (
	"context"
	"errors"
	"testing"

	"ai-content-engine-starter/internal/domain"
)

type itemRepoStub struct {
	item domain.SourceItem
	err  error
}

func (r itemRepoStub) Create(context.Context, domain.SourceItem) (domain.SourceItem, error) {
	return domain.SourceItem{}, nil
}
func (r itemRepoStub) GetByID(context.Context, int64) (domain.SourceItem, error) {
	if r.err != nil {
		return domain.SourceItem{}, r.err
	}
	return r.item, nil
}
func (r itemRepoStub) ListBySourceID(context.Context, int64, int) ([]domain.SourceItem, error) {
	return nil, nil
}

type assetRepoStub struct {
	created  []domain.ContentAsset
	listByID map[int64][]domain.ContentAsset
	err      error
	listErr  error
}

func (r *assetRepoStub) Create(_ context.Context, asset domain.ContentAsset) (domain.ContentAsset, error) {
	if r.err != nil {
		return domain.ContentAsset{}, r.err
	}
	asset.ID = int64(len(r.created) + 1)
	r.created = append(r.created, asset)
	return asset, nil
}
func (r *assetRepoStub) GetByID(context.Context, int64) (domain.ContentAsset, error) {
	return domain.ContentAsset{}, nil
}
func (r *assetRepoStub) ListByRawItemID(_ context.Context, rawItemID int64, _ int) ([]domain.ContentAsset, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	if r.listByID == nil {
		return nil, nil
	}
	return r.listByID[rawItemID], nil
}

func TestGenerateFromIntentMapsFields(t *testing.T) {
	body := "body text"
	repo := &assetRepoStub{}
	svc, err := New(itemRepoStub{item: domain.SourceItem{ID: 7, Title: "Raw title", Body: &body}}, repo)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	asset, err := svc.GenerateFromIntent(context.Background(), domain.PublishIntent{RawItemID: 7, ChannelID: 3, Format: "TEXT"})
	if err != nil {
		t.Fatalf("GenerateFromIntent() error = %v", err)
	}
	if asset.RawItemID != 7 || asset.ChannelID != 3 {
		t.Fatalf("asset ids = (%d,%d), want (7,3)", asset.RawItemID, asset.ChannelID)
	}
	if asset.AssetType != "text" {
		t.Fatalf("AssetType = %q, want text", asset.AssetType)
	}
	if asset.Title != "Raw title" {
		t.Fatalf("Title = %q, want Raw title", asset.Title)
	}
	if asset.Body != "body text" {
		t.Fatalf("Body = %q, want body text", asset.Body)
	}
}

func TestGenerateFromIntentSkipsDuplicateAsset(t *testing.T) {
	repo := &assetRepoStub{listByID: map[int64][]domain.ContentAsset{7: {{ID: 11, RawItemID: 7, ChannelID: 3, AssetType: "text"}}}}
	svc, err := New(itemRepoStub{item: domain.SourceItem{ID: 7, Title: "Raw title"}}, repo)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	asset, err := svc.GenerateFromIntent(context.Background(), domain.PublishIntent{RawItemID: 7, ChannelID: 3, Format: "text"})
	if err != nil {
		t.Fatalf("GenerateFromIntent() error = %v", err)
	}
	if asset.ID != 11 {
		t.Fatalf("asset.ID = %d, want 11", asset.ID)
	}
	if len(repo.created) != 0 {
		t.Fatalf("created assets = %d, want 0", len(repo.created))
	}
}

func TestGenerateFromIntentFallsBackToTitleBody(t *testing.T) {
	repo := &assetRepoStub{}
	svc, err := New(itemRepoStub{item: domain.SourceItem{ID: 7, Title: "Raw title"}}, repo)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	asset, err := svc.GenerateFromIntent(context.Background(), domain.PublishIntent{RawItemID: 7, ChannelID: 3, Format: "text"})
	if err != nil {
		t.Fatalf("GenerateFromIntent() error = %v", err)
	}
	if asset.Body != "Raw title" {
		t.Fatalf("Body = %q, want Raw title", asset.Body)
	}
}

func TestGenerateFromIntentPropagatesLookupError(t *testing.T) {
	svc, err := New(itemRepoStub{err: errors.New("boom")}, &assetRepoStub{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := svc.GenerateFromIntent(context.Background(), domain.PublishIntent{RawItemID: 7, ChannelID: 3, Format: "text"}); err == nil {
		t.Fatalf("GenerateFromIntent() expected error")
	}
}

func TestGenerateFromIntentSkipsNonPlannedIntent(t *testing.T) {
	repo := &assetRepoStub{}
	svc, err := New(itemRepoStub{item: domain.SourceItem{ID: 7, Title: "Raw title"}}, repo)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	asset, err := svc.GenerateFromIntent(context.Background(), domain.PublishIntent{RawItemID: 7, ChannelID: 3, Format: "text", Status: domain.PublishIntentStatusSkipped})
	if err != nil {
		t.Fatalf("GenerateFromIntent() error = %v", err)
	}
	if asset != (domain.ContentAsset{}) {
		t.Fatalf("asset = %#v, want zero value", asset)
	}
	if len(repo.created) != 0 {
		t.Fatalf("created assets = %d, want 0", len(repo.created))
	}
}
