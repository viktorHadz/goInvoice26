package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/viktorHadz/goInvoice26/internal/service/storage"
)

func TestStageAccountDirRemoval_RollbackRestoresDirectory(t *testing.T) {
	store := storage.NewLocalStore(t.TempDir())
	accountDir := store.AccountDir(7)
	logoPath := filepath.Join(accountDir, "logos", "logo.png")
	if err := os.MkdirAll(filepath.Dir(logoPath), 0o755); err != nil {
		t.Fatalf("mkdir account dir: %v", err)
	}
	if err := os.WriteFile(logoPath, []byte("logo"), 0o644); err != nil {
		t.Fatalf("write account file: %v", err)
	}

	staged, ok, err := store.StageAccountDirRemoval(7)
	if err != nil {
		t.Fatalf("StageAccountDirRemoval: %v", err)
	}
	if !ok {
		t.Fatal("StageAccountDirRemoval ok = false, want true")
	}
	if _, err := os.Stat(accountDir); !os.IsNotExist(err) {
		t.Fatalf("account dir still present after staging, err=%v", err)
	}

	if err := staged.Rollback(); err != nil {
		t.Fatalf("Rollback: %v", err)
	}
	if _, err := os.Stat(logoPath); err != nil {
		t.Fatalf("restored account file missing: %v", err)
	}
}

func TestStageAccountDirRemoval_CommitDeletesDirectory(t *testing.T) {
	store := storage.NewLocalStore(t.TempDir())
	accountDir := store.AccountDir(11)
	logoPath := filepath.Join(accountDir, "logos", "logo.png")
	if err := os.MkdirAll(filepath.Dir(logoPath), 0o755); err != nil {
		t.Fatalf("mkdir account dir: %v", err)
	}
	if err := os.WriteFile(logoPath, []byte("logo"), 0o644); err != nil {
		t.Fatalf("write account file: %v", err)
	}

	staged, ok, err := store.StageAccountDirRemoval(11)
	if err != nil {
		t.Fatalf("StageAccountDirRemoval: %v", err)
	}
	if !ok {
		t.Fatal("StageAccountDirRemoval ok = false, want true")
	}

	if err := staged.Commit(); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	if _, err := os.Stat(accountDir); !os.IsNotExist(err) {
		t.Fatalf("account dir still present after commit, err=%v", err)
	}
}
