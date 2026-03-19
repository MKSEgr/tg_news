package assetgen

import (
	"context"
	"fmt"
	"strings"

	"ai-content-engine-starter/internal/domain"
)

const defaultListLimit = 50

// Service performs a minimal 1:1 transformation from publish intent to content asset.
type Service struct {
	items  domain.SourceItemRepository
	assets domain.ContentAssetRepository
}

// New creates an asset generation service.
func New(items domain.SourceItemRepository, assets domain.ContentAssetRepository) (*Service, error) {
	if items == nil {
		return nil, fmt.Errorf("source item repository is nil")
	}
	if assets == nil {
		return nil, fmt.Errorf("content asset repository is nil")
	}
	return &Service{items: items, assets: assets}, nil
}

// GenerateFromIntent creates one content asset from one publish intent.
func (s *Service) GenerateFromIntent(ctx context.Context, intent domain.PublishIntent) (domain.ContentAsset, error) {
	if s == nil {
		return domain.ContentAsset{}, fmt.Errorf("asset generation service is nil")
	}
	if ctx == nil {
		return domain.ContentAsset{}, fmt.Errorf("context is nil")
	}
	if intent.RawItemID <= 0 {
		return domain.ContentAsset{}, fmt.Errorf("raw item id is invalid")
	}
	if intent.ChannelID <= 0 {
		return domain.ContentAsset{}, fmt.Errorf("channel id is invalid")
	}
	status := domain.PublishIntentStatus(strings.TrimSpace(string(intent.Status)))
	if status != "" && status != domain.PublishIntentStatusPlanned {
		return domain.ContentAsset{}, nil
	}

	assetType := strings.ToLower(strings.TrimSpace(intent.Format))
	if assetType == "" {
		return domain.ContentAsset{}, fmt.Errorf("intent format is empty")
	}

	existing, err := s.assets.ListByRawItemID(ctx, intent.RawItemID, defaultListLimit)
	if err != nil {
		return domain.ContentAsset{}, fmt.Errorf("list assets by raw item id: %w", err)
	}
	for _, asset := range existing {
		if asset.ChannelID == intent.ChannelID && asset.AssetType == assetType {
			return asset, nil
		}
	}

	item, err := s.items.GetByID(ctx, intent.RawItemID)
	if err != nil {
		return domain.ContentAsset{}, fmt.Errorf("get raw item by id: %w", err)
	}

	body := strings.TrimSpace(item.Title)
	if item.Body != nil && strings.TrimSpace(*item.Body) != "" {
		body = strings.TrimSpace(*item.Body)
	}

	asset := domain.ContentAsset{
		RawItemID: intent.RawItemID,
		ChannelID: intent.ChannelID,
		AssetType: assetType,
		Title:     strings.TrimSpace(item.Title),
		Body:      body,
		Status:    domain.ContentAssetStatusPending,
	}
	created, err := s.assets.Create(ctx, asset)
	if err != nil {
		return domain.ContentAsset{}, fmt.Errorf("create content asset: %w", err)
	}
	return created, nil
}
