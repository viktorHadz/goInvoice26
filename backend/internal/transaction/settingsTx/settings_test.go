package settingsTx_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/settingsTx"
)

func newSettingsDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "settings.sqlite")

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := db.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		_ = os.Remove(dbPath)
	}
	return conn, cleanup
}

func baseSettings() models.Settings {
	return models.Settings{
		CompanyName:                  "Acme Co",
		Email:                        "hello@example.com",
		Phone:                        "+44 1234",
		CompanyAddress:               "1 Example Street",
		InvoicePrefix:                "INV-",
		Currency:                     "GBP",
		DateFormat:                   "dd/mm/yyyy",
		PaymentTerms:                 "Pay in 14 days",
		PaymentDetails:               "Bank details",
		NotesFooter:                  "Thanks",
		LogoURL:                      "",
		ShowItemTypeHeaders:          true,
		StartingInvoiceNumber:        100,
		CanEditStartingInvoiceNumber: true,
	}
}

func TestUpsert_AllowsStartingInvoiceNumberWhenNoInvoicesExist(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newSettingsDB(t)
	defer cleanup()

	if err := settingsTx.Upsert(ctx, conn, baseSettings()); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := settingsTx.Get(ctx, conn)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.StartingInvoiceNumber != 100 {
		t.Fatalf("starting invoice number = %d, want 100", got.StartingInvoiceNumber)
	}
	if !got.CanEditStartingInvoiceNumber {
		t.Fatalf("canEditStartingInvoiceNumber = false, want true")
	}
}

func TestUpsert_RejectsStartingInvoiceNumberChangeWhenInvoicesExist(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newSettingsDB(t)
	defer cleanup()

	if err := settingsTx.Upsert(ctx, conn, baseSettings()); err != nil {
		t.Fatalf("initial Upsert: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO clients (name) VALUES ('Client');
	`); err != nil {
		t.Fatalf("insert client: %v", err)
	}
	if _, err := conn.Exec(`
		INSERT INTO invoices (client_id, base_number, status) VALUES (1, 100, 'draft');
	`); err != nil {
		t.Fatalf("insert invoice: %v", err)
	}

	locked := baseSettings()
	locked.StartingInvoiceNumber = 250

	err := settingsTx.Upsert(ctx, conn, locked)
	if !errors.Is(err, settingsTx.ErrStartingInvoiceNumberLocked) {
		t.Fatalf("Upsert() error = %v, want %v", err, settingsTx.ErrStartingInvoiceNumberLocked)
	}
}

func TestGet_AllowsEditingAgainAfterAllInvoicesDeleted(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newSettingsDB(t)
	defer cleanup()

	if _, err := conn.Exec(`INSERT INTO clients (name) VALUES ('Client');`); err != nil {
		t.Fatalf("insert client: %v", err)
	}
	if _, err := conn.Exec(`INSERT INTO invoices (client_id, base_number, status) VALUES (1, 9, 'draft');`); err != nil {
		t.Fatalf("insert invoice: %v", err)
	}

	got, err := settingsTx.Get(ctx, conn)
	if err != nil {
		t.Fatalf("Get locked: %v", err)
	}
	if got.CanEditStartingInvoiceNumber {
		t.Fatalf("canEditStartingInvoiceNumber = true, want false")
	}

	if _, err := conn.Exec(`DELETE FROM invoices;`); err != nil {
		t.Fatalf("delete invoices: %v", err)
	}

	got, err = settingsTx.Get(ctx, conn)
	if err != nil {
		t.Fatalf("Get unlocked: %v", err)
	}
	if !got.CanEditStartingInvoiceNumber {
		t.Fatalf("canEditStartingInvoiceNumber = false, want true")
	}
}
