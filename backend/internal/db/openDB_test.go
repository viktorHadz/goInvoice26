package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenDBCreatesParentDirectory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "nested", "app.sqlite")

	conn, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("OpenDB returned error: %v", err)
	}
	defer conn.Close()

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("expected sqlite file at %s: %v", dbPath, err)
	}
}
