package blobfs

import (
	"context"
	"errors"
	"testing"

	"github.com/agentlayer/agentlayer/internal/domain"
)

func TestStoreWritesAndReadsObjects(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	if err := store.Put(context.Background(), "raw/2026/04/09/message.eml", []byte("hello")); err != nil {
		t.Fatalf("expected put to succeed, got error: %v", err)
	}

	data, err := store.Get(context.Background(), "raw/2026/04/09/message.eml")
	if err != nil {
		t.Fatalf("expected get to succeed, got error: %v", err)
	}

	if string(data) != "hello" {
		t.Fatalf("expected stored data, got %q", data)
	}
}

func TestStoreRejectsTraversalAndMapsNotFound(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)

	if err := store.Put(context.Background(), "../escape", []byte("bad")); err == nil {
		t.Fatal("expected traversal put to fail")
	}

	_, err := store.Get(context.Background(), "raw/missing.eml")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found mapping, got %v", err)
	}
}
