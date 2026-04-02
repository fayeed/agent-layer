package app

import (
	"context"
	"testing"
	"time"

	"github.com/agentlayer/agentlayer/internal/api"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestThreadEscalationServiceEscalatesThread(t *testing.T) {
	repository := &threadSaverStub{
		thread: domain.Thread{
			ID:    "thread-123",
			State: domain.ThreadStateEscalated,
		},
	}
	service := NewThreadEscalationService(repository, func() time.Time {
		return time.Date(2026, 4, 3, 18, 0, 0, 0, time.UTC)
	})

	thread, err := service.EscalateThread(context.Background(), "thread-123", "needs human review")
	if err != nil {
		t.Fatalf("expected escalate thread to succeed, got error: %v", err)
	}

	if repository.thread.ID != "thread-123" {
		t.Fatalf("expected thread save by id, got %#v", repository.thread)
	}

	if repository.thread.State != domain.ThreadStateEscalated {
		t.Fatalf("expected escalated thread state, got %#v", repository.thread)
	}

	if thread.State != domain.ThreadStateEscalated {
		t.Fatalf("expected returned thread state, got %#v", thread)
	}
}

func TestContactMemoryServiceCreatesMemoryEntry(t *testing.T) {
	writer := &contactMemoryWriterStub{
		entry: domain.ContactMemoryEntry{
			ID:        "memory-123",
			ContactID: "contact-123",
			ThreadID:  "thread-123",
			Note:      "Prefers email follow-up.",
			Tags:      []string{"preference"},
		},
	}
	service := NewContactMemoryService(writer, func() time.Time {
		return time.Date(2026, 4, 3, 18, 5, 0, 0, time.UTC)
	})

	entry, err := service.CreateContactMemory(context.Background(), "contact-123", api.CreateContactMemoryInput{
		ThreadID: "thread-123",
		Note:     "Prefers email follow-up.",
		Tags:     []string{"preference"},
	})
	if err != nil {
		t.Fatalf("expected create contact memory to succeed, got error: %v", err)
	}

	if writer.entry.ContactID != "contact-123" {
		t.Fatalf("expected contact id on saved entry, got %#v", writer.entry)
	}

	if writer.entry.ThreadID != "thread-123" {
		t.Fatalf("expected thread id on saved entry, got %#v", writer.entry)
	}

	if entry.ContactID != "contact-123" {
		t.Fatalf("expected returned entry, got %#v", entry)
	}

	if entry.ID == "" {
		t.Fatalf("expected generated memory id, got %#v", entry)
	}
}

type threadSaverStub struct {
	thread domain.Thread
	err    error
}

func (s *threadSaverStub) Save(_ context.Context, thread domain.Thread) (domain.Thread, error) {
	s.thread = thread
	if s.err != nil {
		return domain.Thread{}, s.err
	}
	return thread, nil
}

type contactMemoryWriterStub struct {
	entry domain.ContactMemoryEntry
	err   error
}

func (s *contactMemoryWriterStub) Create(_ context.Context, entry domain.ContactMemoryEntry) (domain.ContactMemoryEntry, error) {
	s.entry = entry
	if s.err != nil {
		return domain.ContactMemoryEntry{}, s.err
	}
	return entry, nil
}
