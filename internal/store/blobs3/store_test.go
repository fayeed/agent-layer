package blobs3

import (
	"context"
	"testing"
)

func TestNewStoreValidatesRequiredConfig(t *testing.T) {
	_, err := NewStore(context.Background(), Config{})
	if err == nil {
		t.Fatal("expected missing config to fail")
	}

	_, err = NewStore(context.Background(), Config{
		Region: "us-east-1",
	})
	if err == nil {
		t.Fatal("expected missing bucket to fail")
	}
}
