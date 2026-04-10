package app

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type SuppressionGetter interface {
	GetSuppressionByID(ctx context.Context, suppressionID string) (domain.SuppressedAddress, error)
}

type SuppressionLister interface {
	ListSuppressions(ctx context.Context, limit int) ([]domain.SuppressedAddress, error)
}

type SuppressionReadService struct {
	repository SuppressionGetter
}

func NewSuppressionReadService(repository SuppressionGetter) SuppressionReadService {
	return SuppressionReadService{repository: repository}
}

func (s SuppressionReadService) GetSuppression(ctx context.Context, suppressionID string) (domain.SuppressedAddress, error) {
	return s.repository.GetSuppressionByID(ctx, suppressionID)
}

type SuppressionListService struct {
	repository SuppressionLister
	limit      int
}

func NewSuppressionListService(repository SuppressionLister, limit int) SuppressionListService {
	if limit <= 0 {
		limit = 20
	}
	return SuppressionListService{
		repository: repository,
		limit:      limit,
	}
}

func (s SuppressionListService) ListSuppressions(ctx context.Context, limit int) ([]domain.SuppressedAddress, error) {
	if limit <= 0 {
		limit = s.limit
	}
	return s.repository.ListSuppressions(ctx, limit)
}
