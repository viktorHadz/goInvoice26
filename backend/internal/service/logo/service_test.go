package logo_test

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/service/logo"
	"github.com/viktorHadz/goInvoice26/internal/service/storage"
	"github.com/viktorHadz/goInvoice26/internal/transaction/settingsTx"
)

func newLogoService(t *testing.T) (*sql.DB, *storage.LocalStore, *logo.Service, func()) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "logo.sqlite")
	uploadRoot := filepath.Join(dir, "uploads")

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	store := storage.NewLocalStore(uploadRoot)
	service := logo.NewService(conn, store)

	cleanup := func() {
		_ = conn.Close()
	}
	return conn, store, service, cleanup
}

func TestReplaceLogo_CreatesCurrentFileAndStableURL(t *testing.T) {
	ctx := context.Background()
	conn, store, service, cleanup := newLogoService(t)
	defer cleanup()

	settings, err := service.Replace(ctx, accountscope.DefaultAccountID, bytes.NewReader([]byte("first")), ".png", "image/png")
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	if settings.LogoURL == "" {
		t.Fatal("logoUrl = empty, want stable settings logo URL")
	}

	file, ok, err := settingsTx.GetLogoFile(ctx, conn, accountscope.DefaultAccountID)
	if err != nil {
		t.Fatalf("GetLogoFile: %v", err)
	}
	if !ok {
		t.Fatal("expected current logo file row")
	}
	if _, err := os.Stat(store.Path(file.StorageKey)); err != nil {
		t.Fatalf("current logo file missing: %v", err)
	}
}

func TestReplaceLogo_DeletesPreviousDiskFile(t *testing.T) {
	ctx := context.Background()
	conn, store, service, cleanup := newLogoService(t)
	defer cleanup()

	if _, err := service.Replace(ctx, accountscope.DefaultAccountID, bytes.NewReader([]byte("first")), ".png", "image/png"); err != nil {
		t.Fatalf("first Replace: %v", err)
	}
	first, ok, err := settingsTx.GetLogoFile(ctx, conn, accountscope.DefaultAccountID)
	if err != nil {
		t.Fatalf("GetLogoFile first: %v", err)
	}
	if !ok {
		t.Fatal("expected first current logo")
	}

	if _, err := service.Replace(ctx, accountscope.DefaultAccountID, bytes.NewReader([]byte("second")), ".png", "image/png"); err != nil {
		t.Fatalf("second Replace: %v", err)
	}
	if _, err := os.Stat(store.Path(first.StorageKey)); !os.IsNotExist(err) {
		t.Fatalf("old logo file still exists, err = %v", err)
	}

	second, ok, err := settingsTx.GetLogoFile(ctx, conn, accountscope.DefaultAccountID)
	if err != nil {
		t.Fatalf("GetLogoFile second: %v", err)
	}
	if !ok {
		t.Fatal("expected second current logo")
	}
	if _, err := os.Stat(store.Path(second.StorageKey)); err != nil {
		t.Fatalf("new logo file missing: %v", err)
	}
}

func TestRemoveLogo_ClearsReferenceAndDeletesDiskFile(t *testing.T) {
	ctx := context.Background()
	conn, store, service, cleanup := newLogoService(t)
	defer cleanup()

	if _, err := service.Replace(ctx, accountscope.DefaultAccountID, bytes.NewReader([]byte("first")), ".png", "image/png"); err != nil {
		t.Fatalf("Replace: %v", err)
	}
	current, ok, err := settingsTx.GetLogoFile(ctx, conn, accountscope.DefaultAccountID)
	if err != nil {
		t.Fatalf("GetLogoFile before remove: %v", err)
	}
	if !ok {
		t.Fatal("expected current logo before remove")
	}

	settings, err := service.Remove(ctx, accountscope.DefaultAccountID)
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if settings.LogoURL != "" {
		t.Fatalf("logoUrl = %q, want empty after remove", settings.LogoURL)
	}
	if _, err := os.Stat(store.Path(current.StorageKey)); !os.IsNotExist(err) {
		t.Fatalf("removed logo file still exists, err = %v", err)
	}

	_, ok, err = settingsTx.GetLogoFile(ctx, conn, accountscope.DefaultAccountID)
	if err != nil {
		t.Fatalf("GetLogoFile after remove: %v", err)
	}
	if ok {
		t.Fatal("expected no current logo after remove")
	}
}
