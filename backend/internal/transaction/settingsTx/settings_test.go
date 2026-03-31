package settingsTx_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
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

	if err := settingsTx.Upsert(ctx, conn, accountscope.DefaultAccountID, baseSettings()); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	got, err := settingsTx.Get(ctx, conn, accountscope.DefaultAccountID)
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

	if err := settingsTx.Upsert(ctx, conn, accountscope.DefaultAccountID, baseSettings()); err != nil {
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

	err := settingsTx.Upsert(ctx, conn, accountscope.DefaultAccountID, locked)
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

	got, err := settingsTx.Get(ctx, conn, accountscope.DefaultAccountID)
	if err != nil {
		t.Fatalf("Get locked: %v", err)
	}
	if got.CanEditStartingInvoiceNumber {
		t.Fatalf("canEditStartingInvoiceNumber = true, want false")
	}

	if _, err := conn.Exec(`DELETE FROM invoices;`); err != nil {
		t.Fatalf("delete invoices: %v", err)
	}

	got, err = settingsTx.Get(ctx, conn, accountscope.DefaultAccountID)
	if err != nil {
		t.Fatalf("Get unlocked: %v", err)
	}
	if !got.CanEditStartingInvoiceNumber {
		t.Fatalf("canEditStartingInvoiceNumber = false, want true")
	}
}

func TestReplaceLogo_SwapsCurrentStoredAsset(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newSettingsDB(t)
	defer cleanup()

	first, prev, err := settingsTx.ReplaceLogo(ctx, conn, accountscope.DefaultAccountID, "accounts/1/logos/first.png", "image/png")
	if err != nil {
		t.Fatalf("first ReplaceLogo: %v", err)
	}
	if prev != nil {
		t.Fatalf("first previous logo = %#v, want nil", prev)
	}

	second, prev, err := settingsTx.ReplaceLogo(ctx, conn, accountscope.DefaultAccountID, "accounts/1/logos/second.png", "image/png")
	if err != nil {
		t.Fatalf("second ReplaceLogo: %v", err)
	}
	if prev == nil {
		t.Fatal("second previous logo = nil, want first asset")
	}
	if prev.ID != first.ID || prev.StorageKey != first.StorageKey {
		t.Fatalf("previous logo = %#v, want %#v", prev, first)
	}

	current, ok, err := settingsTx.GetLogoFile(ctx, conn, accountscope.DefaultAccountID)
	if err != nil {
		t.Fatalf("GetLogoFile: %v", err)
	}
	if !ok {
		t.Fatal("current logo not found")
	}
	if current.ID != second.ID || current.StorageKey != second.StorageKey {
		t.Fatalf("current logo = %#v, want %#v", current, second)
	}
}

func TestRemoveLogo_ClearsCurrentStoredAsset(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newSettingsDB(t)
	defer cleanup()

	inserted, prev, err := settingsTx.ReplaceLogo(ctx, conn, accountscope.DefaultAccountID, "accounts/1/logos/logo.png", "image/png")
	if err != nil {
		t.Fatalf("ReplaceLogo: %v", err)
	}
	if prev != nil {
		t.Fatalf("previous logo = %#v, want nil", prev)
	}

	removed, err := settingsTx.RemoveLogo(ctx, conn, accountscope.DefaultAccountID)
	if err != nil {
		t.Fatalf("RemoveLogo: %v", err)
	}
	if removed == nil {
		t.Fatal("removed logo = nil, want stored asset")
	}
	if removed.ID != inserted.ID {
		t.Fatalf("removed logo id = %d, want %d", removed.ID, inserted.ID)
	}

	_, ok, err := settingsTx.GetLogoFile(ctx, conn, accountscope.DefaultAccountID)
	if err != nil {
		t.Fatalf("GetLogoFile after remove: %v", err)
	}
	if ok {
		t.Fatal("expected no current logo after remove")
	}
}

func TestGet_DerivesStableLogoURLFromCurrentAsset(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newSettingsDB(t)
	defer cleanup()

	inserted, _, err := settingsTx.ReplaceLogo(ctx, conn, accountscope.DefaultAccountID, "accounts/1/logos/logo.png", "image/png")
	if err != nil {
		t.Fatalf("ReplaceLogo: %v", err)
	}

	got, err := settingsTx.Get(ctx, conn, accountscope.DefaultAccountID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	want := "/api/settings/logo?v=" + fmt.Sprint(inserted.ID)
	if got.LogoURL != want {
		t.Fatalf("logoUrl = %q, want %q", got.LogoURL, want)
	}
}

func TestMigrate_LegacyUserSettingsConsolidatesIntoAccountSettings(t *testing.T) {
	ctx := context.Background()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "legacy-settings.sqlite")

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	if _, err := conn.Exec(`
		CREATE TABLE user_settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			company_name TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			phone TEXT NOT NULL DEFAULT '',
			company_address TEXT NOT NULL DEFAULT '',
			invoice_prefix TEXT NOT NULL DEFAULT 'INV-',
			currency TEXT NOT NULL DEFAULT 'GBP',
			date_format TEXT NOT NULL DEFAULT 'dd/mm/yyyy',
			payment_terms TEXT NOT NULL DEFAULT 'Please make payment within 14 days.',
			payment_details TEXT NOT NULL DEFAULT '',
			notes_footer TEXT NOT NULL DEFAULT '',
			logo_url TEXT NOT NULL DEFAULT ''
		);
	`); err != nil {
		t.Fatalf("create legacy user_settings: %v", err)
	}

	if _, err := conn.Exec(`
		INSERT INTO user_settings (
			id,
			company_name,
			email,
			phone,
			company_address,
			invoice_prefix,
			currency,
			date_format,
			payment_terms,
			payment_details,
			notes_footer,
			logo_url
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`,
		1,
		"Legacy Co",
		"legacy@example.com",
		"+44 9999",
		"99 Legacy Street",
		"LEG-",
		"USD",
		"mm/dd/yyyy",
		"Pay within 7 days",
		"Legacy bank details",
		"Legacy footer",
		"/uploads/legacy/logo.png",
	); err != nil {
		t.Fatalf("insert legacy user_settings: %v", err)
	}

	if err := db.Migrate(ctx, conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	got, err := settingsTx.Get(ctx, conn, accountscope.DefaultAccountID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got.CompanyName != "Legacy Co" {
		t.Fatalf("companyName = %q, want %q", got.CompanyName, "Legacy Co")
	}
	if got.InvoicePrefix != "LEG-" {
		t.Fatalf("invoicePrefix = %q, want %q", got.InvoicePrefix, "LEG-")
	}
	if !got.ShowItemTypeHeaders {
		t.Fatal("showItemTypeHeaders = false, want true backfilled from legacy table")
	}

	var legacyLogoURL string
	if err := conn.QueryRow(`
		SELECT legacy_logo_url
		FROM account_settings
		WHERE account_id = 1;
	`).Scan(&legacyLogoURL); err != nil {
		t.Fatalf("load migrated legacy_logo_url: %v", err)
	}
	if legacyLogoURL != "/uploads/legacy/logo.png" {
		t.Fatalf("legacy_logo_url = %q, want %q", legacyLogoURL, "/uploads/legacy/logo.png")
	}

	var legacyTableCount int
	if err := conn.QueryRow(`
		SELECT COUNT(*)
		FROM sqlite_master
		WHERE type = 'table' AND name = 'user_settings';
	`).Scan(&legacyTableCount); err != nil {
		t.Fatalf("count legacy user_settings tables: %v", err)
	}
	if legacyTableCount != 0 {
		t.Fatalf("legacy user_settings table count = %d, want 0", legacyTableCount)
	}
}
