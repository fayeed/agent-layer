package migrations

import (
	"strings"
	"testing"
)

func TestV0CoreSQLIsEmbedded(t *testing.T) {
	sql := V0CoreSQL()
	if sql == "" {
		t.Fatal("expected embedded migration sql")
	}
	if !strings.Contains(sql, "CREATE TABLE organizations") {
		t.Fatalf("expected organizations table in embedded migration, got %q", sql[:64])
	}
}
