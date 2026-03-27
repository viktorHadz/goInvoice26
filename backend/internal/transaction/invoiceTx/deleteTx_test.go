package invoiceTx_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func newTestApp(t *testing.T) (*app.App, func()) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.sqlite")

	d, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	if err := db.Migrate(context.Background(), d); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	a := &app.App{DB: d}
	cleanup := func() {
		_ = d.Close()
		_ = os.Remove(dbPath)
	}
	return a, cleanup
}

func insertClient(t *testing.T, a *app.App) int64 {
	t.Helper()

	res, err := a.DB.Exec(`INSERT INTO clients (name) VALUES (?)`, "Test Client")
	if err != nil {
		t.Fatalf("insert client: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("client lastInsertId: %v", err)
	}

	return id
}

func insertInvoiceGraph(t *testing.T, a *app.App, clientID, baseNumber int64, status string) int64 {
	t.Helper()

	res, err := a.DB.Exec(`
		INSERT INTO invoices (client_id, base_number, status)
		VALUES (?, ?, ?)
	`, clientID, baseNumber, status)
	if err != nil {
		t.Fatalf("insert invoice: %v", err)
	}

	invoiceID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("invoice lastInsertId: %v", err)
	}

	res, err = a.DB.Exec(`
		INSERT INTO invoice_revisions (
			invoice_id,
			revision_no,
			issue_date,
			due_by_date,
			client_name,
			client_company_name,
			client_address,
			client_email,
			note,
			vat_rate,
			discount_type,
			discount_rate,
			discount_minor,
			deposit_type,
			deposit_rate,
			deposit_minor,
			subtotal_minor,
			vat_amount_minor,
			total_minor
		) VALUES (?, 1, '2026-03-27', '2026-04-10', 'Test Client', '', '', '', NULL, 0, 'none', 0, 0, 'none', 0, 0, 1000, 0, 1000)
	`, invoiceID)
	if err != nil {
		t.Fatalf("insert invoice revision: %v", err)
	}

	revisionID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("revision lastInsertId: %v", err)
	}

	if _, err := a.DB.Exec(`
		UPDATE invoices
		SET current_revision_id = ?
		WHERE id = ?
	`, revisionID, invoiceID); err != nil {
		t.Fatalf("update current revision: %v", err)
	}

	if _, err := a.DB.Exec(`
		INSERT INTO invoice_items (
			invoice_revision_id,
			name,
			line_type,
			pricing_mode,
			quantity,
			unit_price_minor,
			line_total_minor,
			minutes_worked,
			sort_order
		) VALUES (?, 'Service line', 'custom', 'flat', 1, 1000, 1000, NULL, 1)
	`, revisionID); err != nil {
		t.Fatalf("insert invoice item: %v", err)
	}

	if _, err := a.DB.Exec(`
		INSERT INTO payments (
			invoice_id,
			payment_type,
			amount_minor,
			payment_date,
			applied_in_revision_id,
			label
		) VALUES (?, 'payment', 100, '2026-03-28', ?, 'Initial payment')
	`, invoiceID, revisionID); err != nil {
		t.Fatalf("insert payment: %v", err)
	}

	return invoiceID
}

func countRows(t *testing.T, a *app.App, table string, invoiceID int64) int {
	t.Helper()

	var query string
	switch table {
	case "invoices":
		query = `SELECT COUNT(*) FROM invoices WHERE id = ?`
	case "invoice_revisions":
		query = `SELECT COUNT(*) FROM invoice_revisions WHERE invoice_id = ?`
	case "invoice_items":
		query = `
			SELECT COUNT(*)
			FROM invoice_items it
			JOIN invoice_revisions r ON r.id = it.invoice_revision_id
			WHERE r.invoice_id = ?
		`
	case "payments":
		query = `SELECT COUNT(*) FROM payments WHERE invoice_id = ?`
	default:
		t.Fatalf("unsupported table: %s", table)
	}

	var count int
	if err := a.DB.QueryRow(query, invoiceID).Scan(&count); err != nil {
		t.Fatalf("count rows in %s: %v", table, err)
	}

	return count
}

func TestDelete_RemovesInvoiceGraph(t *testing.T) {
	ctx := context.Background()
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a)
	invoiceID := insertInvoiceGraph(t, a, clientID, 101, "draft")

	if err := invoiceTx.Delete(ctx, a, clientID, 101); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	for _, table := range []string{"invoices", "invoice_revisions", "invoice_items", "payments"} {
		if count := countRows(t, a, table, invoiceID); count != 0 {
			t.Fatalf("%s count = %d, want 0", table, count)
		}
	}
}

func TestDelete_AllowsIssuedAndPaidInvoices(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		status string
	}{
		{name: "issued", status: "issued"},
		{name: "paid", status: "paid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, cleanup := newTestApp(t)
			defer cleanup()

			clientID := insertClient(t, a)
			invoiceID := insertInvoiceGraph(t, a, clientID, 202, tt.status)

			if err := invoiceTx.Delete(ctx, a, clientID, 202); err != nil {
				t.Fatalf("Delete() error = %v", err)
			}

			for _, table := range []string{"invoices", "invoice_revisions", "invoice_items", "payments"} {
				if count := countRows(t, a, table, invoiceID); count != 0 {
					t.Fatalf("%s count = %d, want 0", table, count)
				}
			}
		})
	}
}

func TestDelete_RejectsVoidInvoices(t *testing.T) {
	ctx := context.Background()
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a)
	invoiceID := insertInvoiceGraph(t, a, clientID, 303, "void")

	err := invoiceTx.Delete(ctx, a, clientID, 303)
	if !errors.Is(err, invoiceTx.ErrInvoiceDeleteVoid) {
		t.Fatalf("Delete() error = %v, want %v", err, invoiceTx.ErrInvoiceDeleteVoid)
	}

	if count := countRows(t, a, "invoices", invoiceID); count != 1 {
		t.Fatalf("invoice count = %d, want 1", count)
	}
}
