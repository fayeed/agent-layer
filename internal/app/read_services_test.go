package app

import (
	"context"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestThreadReadServiceGetsThread(t *testing.T) {
	repository := &threadRepositoryStub{
		thread: domain.Thread{
			ID:             "thread-123",
			OrganizationID: "org-123",
		},
	}
	service := NewThreadReadService(repository)

	thread, err := service.GetThread(context.Background(), "thread-123")
	if err != nil {
		t.Fatalf("expected get thread to succeed, got error: %v", err)
	}

	if repository.threadID != "thread-123" {
		t.Fatalf("expected thread lookup by id, got %q", repository.threadID)
	}

	if thread.ID != "thread-123" {
		t.Fatalf("expected returned thread, got %#v", thread)
	}
}

func TestThreadMessagesReadServiceListsMessages(t *testing.T) {
	repository := &threadMessagesRepositoryStub{
		messages: []domain.Message{
			{ID: "message-100", ThreadID: "thread-123"},
			{ID: "message-101", ThreadID: "thread-123"},
		},
	}
	service := NewThreadMessagesReadService(repository, 25)

	messages, err := service.ListThreadMessages(context.Background(), "thread-123", 0)
	if err != nil {
		t.Fatalf("expected list thread messages to succeed, got error: %v", err)
	}

	if repository.threadID != "thread-123" {
		t.Fatalf("expected thread message lookup by id, got %q", repository.threadID)
	}

	if repository.limit != 25 {
		t.Fatalf("expected configured message limit, got %d", repository.limit)
	}

	if len(messages) != 2 {
		t.Fatalf("expected returned messages, got %#v", messages)
	}
}

func TestThreadMessagesReadServiceUsesExplicitLimitOverride(t *testing.T) {
	repository := &threadMessagesRepositoryStub{}
	service := NewThreadMessagesReadService(repository, 25)

	_, err := service.ListThreadMessages(context.Background(), "thread-123", 2)
	if err != nil {
		t.Fatalf("expected list thread messages to succeed, got error: %v", err)
	}

	if repository.limit != 2 {
		t.Fatalf("expected explicit message limit override, got %d", repository.limit)
	}
}

func TestContactReadServiceGetsContact(t *testing.T) {
	repository := &contactRepositoryStub{
		contact: domain.Contact{
			ID:           "contact-123",
			EmailAddress: "sender@example.com",
		},
	}
	service := NewContactReadService(repository)

	contact, err := service.GetContact(context.Background(), "contact-123")
	if err != nil {
		t.Fatalf("expected get contact to succeed, got error: %v", err)
	}

	if repository.contactID != "contact-123" {
		t.Fatalf("expected contact lookup by id, got %q", repository.contactID)
	}

	if contact.ID != "contact-123" {
		t.Fatalf("expected returned contact, got %#v", contact)
	}
}

type threadRepositoryStub struct {
	threadID string
	thread   domain.Thread
	err      error
}

func (s *threadRepositoryStub) GetByID(_ context.Context, threadID string) (domain.Thread, error) {
	s.threadID = threadID
	return s.thread, s.err
}

type threadMessagesRepositoryStub struct {
	threadID string
	limit    int
	messages []domain.Message
	err      error
}

func (s *threadMessagesRepositoryStub) ListByThreadID(_ context.Context, threadID string, limit int) ([]domain.Message, error) {
	s.threadID = threadID
	s.limit = limit
	return s.messages, s.err
}

type contactRepositoryStub struct {
	contactID string
	contact   domain.Contact
	err       error
}

func (s *contactRepositoryStub) GetByID(_ context.Context, contactID string) (domain.Contact, error) {
	s.contactID = contactID
	return s.contact, s.err
}
