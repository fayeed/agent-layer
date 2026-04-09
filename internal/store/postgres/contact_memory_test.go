package postgres

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestContactMemoryStoreCreatesAndListsEntries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("expected sqlmock db, got error: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 4, 9, 20, 0, 0, 0, time.UTC)
	entry := domain.ContactMemoryEntry{
		ID:        "memory-123",
		ContactID: "contact-123",
		ThreadID:  "thread-123",
		Note:      "Prefers email.",
		Tags:      []string{"preference", "email"},
		CreatedAt: now,
	}
	tagsJSON, _ := json.Marshal(entry.Tags)

	mock.ExpectExec(regexp.QuoteMeta(`
		INSERT INTO contact_memory (id, organization_id, contact_id, thread_id, note, tags, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)).
		WithArgs(entry.ID, "org-123", entry.ContactID, entry.ThreadID, entry.Note, stringArrayValue(entry.Tags), entry.CreatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectQuery(regexp.QuoteMeta(`
		SELECT id, organization_id, contact_id, thread_id, note, tags, created_at
		FROM contact_memory
		WHERE contact_id = $1
		ORDER BY created_at DESC, id DESC
		LIMIT $2
	`)).
		WithArgs(entry.ContactID, 3).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "organization_id", "contact_id", "thread_id", "note", "tags", "created_at",
		}).AddRow(
			entry.ID, "org-123", entry.ContactID, entry.ThreadID, entry.Note, tagsJSON, entry.CreatedAt,
		))

	store := NewContactMemoryStore(db)
	if _, err := store.CreateMemory(context.Background(), entry, "org-123"); err != nil {
		t.Fatalf("expected contact memory create to succeed, got error: %v", err)
	}

	list, err := store.ListMemoryByContactID(context.Background(), entry.ContactID, 3)
	if err != nil {
		t.Fatalf("expected contact memory list to succeed, got error: %v", err)
	}
	if len(list) != 1 || len(list[0].Tags) != 2 {
		t.Fatalf("expected contact memory list result, got %#v", list)
	}
}
