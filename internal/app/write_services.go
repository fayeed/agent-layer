package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/agentlayer/agentlayer/internal/api"
	"github.com/agentlayer/agentlayer/internal/domain"
)

type ThreadSaver interface {
	Save(ctx context.Context, thread domain.Thread) (domain.Thread, error)
}

type ContactMemoryWriter interface {
	Create(ctx context.Context, entry domain.ContactMemoryEntry) (domain.ContactMemoryEntry, error)
}

type ThreadEscalationService struct {
	threads ThreadSaver
	now     func() time.Time
}

func NewThreadEscalationService(threads ThreadSaver, now func() time.Time) ThreadEscalationService {
	if now == nil {
		now = time.Now
	}
	return ThreadEscalationService{
		threads: threads,
		now:     now,
	}
}

func (s ThreadEscalationService) EscalateThread(ctx context.Context, threadID, _ string) (domain.Thread, error) {
	thread := domain.Thread{
		ID:        threadID,
		State:     domain.ThreadStateEscalated,
		UpdatedAt: s.now().UTC(),
	}
	return s.threads.Save(ctx, thread)
}

type ContactMemoryService struct {
	writer ContactMemoryWriter
	now    func() time.Time
}

func NewContactMemoryService(writer ContactMemoryWriter, now func() time.Time) ContactMemoryService {
	if now == nil {
		now = time.Now
	}
	return ContactMemoryService{
		writer: writer,
		now:    now,
	}
}

func (s ContactMemoryService) CreateContactMemory(ctx context.Context, contactID string, input api.CreateContactMemoryInput) (domain.ContactMemoryEntry, error) {
	entry := domain.ContactMemoryEntry{
		ID:        newContactMemoryID(),
		ContactID: contactID,
		ThreadID:  input.ThreadID,
		Note:      input.Note,
		Tags:      input.Tags,
		CreatedAt: s.now().UTC(),
	}
	return s.writer.Create(ctx, entry)
}

func newContactMemoryID() string {
	var buf [8]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "memory-generated"
	}
	return "memory-" + hex.EncodeToString(buf[:])
}
