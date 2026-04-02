package app

import (
	"context"

	"github.com/agentlayer/agentlayer/internal/domain"
)

type ThreadGetter interface {
	GetByID(ctx context.Context, threadID string) (domain.Thread, error)
}

type ThreadMessagesGetter interface {
	ListByThreadID(ctx context.Context, threadID string, limit int) ([]domain.Message, error)
}

type ContactGetter interface {
	GetByID(ctx context.Context, contactID string) (domain.Contact, error)
}

type ThreadReadService struct {
	repository ThreadGetter
}

func NewThreadReadService(repository ThreadGetter) ThreadReadService {
	return ThreadReadService{repository: repository}
}

func (s ThreadReadService) GetThread(ctx context.Context, threadID string) (domain.Thread, error) {
	return s.repository.GetByID(ctx, threadID)
}

type ThreadMessagesReadService struct {
	repository ThreadMessagesGetter
	limit      int
}

func NewThreadMessagesReadService(repository ThreadMessagesGetter, limit int) ThreadMessagesReadService {
	if limit <= 0 {
		limit = 20
	}
	return ThreadMessagesReadService{
		repository: repository,
		limit:      limit,
	}
}

func (s ThreadMessagesReadService) ListThreadMessages(ctx context.Context, threadID string) ([]domain.Message, error) {
	return s.repository.ListByThreadID(ctx, threadID, s.limit)
}

type ContactReadService struct {
	repository ContactGetter
}

func NewContactReadService(repository ContactGetter) ContactReadService {
	return ContactReadService{repository: repository}
}

func (s ContactReadService) GetContact(ctx context.Context, contactID string) (domain.Contact, error) {
	return s.repository.GetByID(ctx, contactID)
}
