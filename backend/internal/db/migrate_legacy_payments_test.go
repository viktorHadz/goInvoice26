package db

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const currentPaymentsTableSQL = `CREATE TABLE IF NOT EXISTS payments (
  id INTEGER PRIMARY KEY,
  invoice_id INTEGER NOT NULL,
  receipt_no INTEGER NOT NULL DEFAULT 0 CHECK (receipt_no >= 0),
  payment_type TEXT NOT NULL DEFAULT 'payment'
    CHECK (payment_type IN ('deposit','payment')),
  amount_minor INTEGER NOT NULL CHECK (amount_minor > 0),
  payment_date TEXT NOT NULL,
  applied_in_revision_id INTEGER
    REFERENCES invoice_revisions(id)
    ON DELETE SET NULL
    DEFERRABLE INITIALLY DEFERRED,
  label TEXT,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE
);`

const legacyPaymentsTableSQL = `CREATE TABLE IF NOT EXISTS payments (
  id INTEGER PRIMARY KEY,
  invoice_id INTEGER NOT NULL,
  payment_type TEXT NOT NULL DEFAULT 'payment'
    CHECK (payment_type IN ('deposit','payment')),
  amount_minor INTEGER NOT NULL CHECK (amount_minor > 0),
  payment_date TEXT NOT NULL,
  applied_in_revision_id INTEGER
    REFERENCES invoice_revisions(id)
    ON DELETE SET NULL
    DEFERRABLE INITIALLY DEFERRED,
  label TEXT,
  created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
  FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE
);`

const paymentsRevisionReceiptIndexSQL = `CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_revision_receipt_no ON payments(applied_in_revision_id, receipt_no);`

func TestMigrateAddsReceiptNumbersToLegacyPaymentsTable(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "legacy-payments.sqlite")

	conn, err := OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = conn.Close()
		_ = os.Remove(dbPath)
	}()

	legacySchema := strings.Replace(baseSchemaSQL, currentPaymentsTableSQL, legacyPaymentsTableSQL, 1)
	if legacySchema == baseSchemaSQL {
		t.Fatal("failed to build legacy schema fixture for payments table")
	}
	legacySchema = strings.Replace(legacySchema, paymentsRevisionReceiptIndexSQL, "", 1)

	if _, err := conn.ExecContext(ctx, legacySchema); err != nil {
		t.Fatalf("seed legacy schema: %v", err)
	}

	invoiceID := insertLegacyPaymentsInvoiceFixture(t, conn)
	revisionID := insertLegacyPaymentsRevisionFixture(t, conn, invoiceID)
	insertLegacyPaymentsFixture(t, conn, invoiceID, revisionID)

	if err := Migrate(ctx, conn); err != nil {
		t.Fatalf("migrate legacy db: %v", err)
	}

	hasColumn, err := dbTableHasColumn(ctx, conn, "payments", "receipt_no")
	if err != nil {
		t.Fatalf("inspect payments columns: %v", err)
	}
	if !hasColumn {
		t.Fatal("expected payments.receipt_no to be added during migration")
	}

	rows, err := conn.QueryContext(ctx, `
		SELECT receipt_no
		FROM payments
		ORDER BY created_at ASC, id ASC;
	`)
	if err != nil {
		t.Fatalf("query migrated payments: %v", err)
	}
	defer rows.Close()

	var receiptNos []int
	for rows.Next() {
		var receiptNo int
		if err := rows.Scan(&receiptNo); err != nil {
			t.Fatalf("scan receipt_no: %v", err)
		}
		receiptNos = append(receiptNos, receiptNo)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate migrated payments: %v", err)
	}

	if got, want := len(receiptNos), 2; got != want {
		t.Fatalf("payment row count = %d, want %d", got, want)
	}
	if receiptNos[0] != 1 || receiptNos[1] != 2 {
		t.Fatalf("receipt numbers = %v, want [1 2]", receiptNos)
	}

	if !hasIndex(t, conn, "idx_payments_revision_receipt_no") {
		t.Fatal("expected idx_payments_revision_receipt_no to exist after migration")
	}
}

func insertLegacyPaymentsInvoiceFixture(t *testing.T, conn *sql.DB) int64 {
	t.Helper()

	clientRes, err := conn.Exec(`
		INSERT INTO clients (account_id, name)
		VALUES (1, 'Legacy Client');
	`)
	if err != nil {
		t.Fatalf("insert client: %v", err)
	}
	clientID, err := clientRes.LastInsertId()
	if err != nil {
		t.Fatalf("client last insert id: %v", err)
	}

	invoiceRes, err := conn.Exec(`
		INSERT INTO invoices (account_id, client_id, base_number, status, created_at)
		VALUES (1, ?, 1001, 'issued', '2026-04-10T19:00:00.000Z');
	`, clientID)
	if err != nil {
		t.Fatalf("insert invoice: %v", err)
	}

	invoiceID, err := invoiceRes.LastInsertId()
	if err != nil {
		t.Fatalf("invoice last insert id: %v", err)
	}

	return invoiceID
}

func insertLegacyPaymentsRevisionFixture(t *testing.T, conn *sql.DB, invoiceID int64) int64 {
	t.Helper()

	res, err := conn.Exec(`
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
		) VALUES (?, 1, '2026-04-10', '2026-04-24', 'Legacy Client', '', '', '', NULL, 0, 'none', 0, 0, 'none', 0, 0, 1000, 0, 1000);
	`, invoiceID)
	if err != nil {
		t.Fatalf("insert invoice revision: %v", err)
	}

	revisionID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("invoice revision last insert id: %v", err)
	}

	if _, err := conn.Exec(`
		UPDATE invoices
		SET current_revision_id = ?
		WHERE id = ?;
	`, revisionID, invoiceID); err != nil {
		t.Fatalf("link invoice revision: %v", err)
	}

	return revisionID
}

func insertLegacyPaymentsFixture(t *testing.T, conn *sql.DB, invoiceID, revisionID int64) {
	t.Helper()

	if _, err := conn.Exec(`
		INSERT INTO payments (
			invoice_id,
			payment_type,
			amount_minor,
			payment_date,
			applied_in_revision_id,
			label,
			created_at
		) VALUES
			(?, 'payment', 400, '2026-04-10', ?, 'First payment', '2026-04-10T19:10:00.000Z'),
			(?, 'payment', 600, '2026-04-10', ?, 'Second payment', '2026-04-10T19:20:00.000Z');
	`, invoiceID, revisionID, invoiceID, revisionID); err != nil {
		t.Fatalf("insert legacy payments: %v", err)
	}
}

func hasIndex(t *testing.T, conn *sql.DB, indexName string) bool {
	t.Helper()

	var count int
	if err := conn.QueryRow(`
		SELECT COUNT(*)
		FROM sqlite_master
		WHERE type = 'index'
		  AND name = ?;
	`, indexName).Scan(&count); err != nil {
		t.Fatalf("check index %s: %v", indexName, err)
	}

	return count == 1
}
