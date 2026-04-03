package db_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/viktorHadz/goInvoice26/internal/db"
)

func newTenantIntegrityDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "tenant-integrity.sqlite")

	conn, err := db.OpenDB(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	cleanup := func() {
		_ = conn.Close()
		_ = os.Remove(dbPath)
	}

	return conn, cleanup
}

func insertAccount(t *testing.T, conn *sql.DB, id int64, name string) {
	t.Helper()

	if _, err := conn.Exec(`INSERT INTO accounts (id, name) VALUES (?, ?)`, id, name); err != nil {
		t.Fatalf("insert account %d: %v", id, err)
	}
}

func insertClient(t *testing.T, conn *sql.DB, accountID int64, name string) int64 {
	t.Helper()

	res, err := conn.Exec(`INSERT INTO clients (account_id, name) VALUES (?, ?)`, accountID, name)
	if err != nil {
		t.Fatalf("insert client for account %d: %v", accountID, err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("client lastInsertId: %v", err)
	}
	return id
}

func insertInvoice(t *testing.T, conn *sql.DB, accountID, clientID, baseNumber int64) int64 {
	t.Helper()

	res, err := conn.Exec(`
		INSERT INTO invoices (account_id, client_id, base_number, status)
		VALUES (?, ?, ?, 'draft')
	`, accountID, clientID, baseNumber)
	if err != nil {
		t.Fatalf("insert invoice: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("invoice lastInsertId: %v", err)
	}
	return id
}

func insertRevision(t *testing.T, conn *sql.DB, invoiceID, revisionNo int64) int64 {
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
		) VALUES (?, ?, '2026-03-31', '2026-04-14', 'Test Client', '', '', '', NULL, 0, 'none', 0, 0, 'none', 0, 0, 1000, 0, 1000)
	`, invoiceID, revisionNo)
	if err != nil {
		t.Fatalf("insert invoice revision: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("invoice revision lastInsertId: %v", err)
	}
	return id
}

func insertProduct(t *testing.T, conn *sql.DB, accountID, clientID int64, name string) int64 {
	t.Helper()

	res, err := conn.Exec(`
		INSERT INTO products (
			account_id,
			product_type,
			pricing_mode,
			name,
			flat_price_minor,
			client_id
		) VALUES (?, 'style', 'flat', ?, 1000, ?)
	`, accountID, name, clientID)
	if err != nil {
		t.Fatalf("insert product: %v", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("product lastInsertId: %v", err)
	}
	return id
}

func requireErrorContains(t *testing.T, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatalf("expected error containing %q, got nil", want)
	}
	if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(want)) {
		t.Fatalf("error = %v, want substring %q", err, want)
	}
}

func TestProductsRejectClientFromDifferentAccount(t *testing.T) {
	conn, cleanup := newTenantIntegrityDB(t)
	defer cleanup()

	insertAccount(t, conn, 2, "Second account")
	otherClientID := insertClient(t, conn, 2, "Other Client")

	_, err := conn.Exec(`
		INSERT INTO products (
			account_id,
			product_type,
			pricing_mode,
			name,
			flat_price_minor,
			client_id
		) VALUES (1, 'style', 'flat', 'Bad Product', 1000, ?)
	`, otherClientID)
	requireErrorContains(t, err, "foreign key")
}

func TestInvoicesRejectClientFromDifferentAccount(t *testing.T) {
	conn, cleanup := newTenantIntegrityDB(t)
	defer cleanup()

	insertAccount(t, conn, 2, "Second account")
	otherClientID := insertClient(t, conn, 2, "Other Client")

	_, err := conn.Exec(`
		INSERT INTO invoices (account_id, client_id, base_number, status)
		VALUES (1, ?, 100, 'draft')
	`, otherClientID)
	requireErrorContains(t, err, "foreign key")
}

func TestInvoicesRejectCurrentRevisionFromDifferentInvoice(t *testing.T) {
	conn, cleanup := newTenantIntegrityDB(t)
	defer cleanup()

	clientID := insertClient(t, conn, 1, "Client")
	firstInvoiceID := insertInvoice(t, conn, 1, clientID, 100)
	secondInvoiceID := insertInvoice(t, conn, 1, clientID, 101)
	_ = insertRevision(t, conn, firstInvoiceID, 1)
	secondRevisionID := insertRevision(t, conn, secondInvoiceID, 1)

	_, err := conn.Exec(`
		UPDATE invoices
		SET current_revision_id = ?
		WHERE id = ?
	`, secondRevisionID, firstInvoiceID)
	requireErrorContains(t, err, "invoice current revision must belong to same invoice")
}

func TestClientsRejectOwnershipReparenting(t *testing.T) {
	conn, cleanup := newTenantIntegrityDB(t)
	defer cleanup()

	insertAccount(t, conn, 2, "Second account")
	clientID := insertClient(t, conn, 1, "Client")

	_, err := conn.Exec(`UPDATE clients SET account_id = 2 WHERE id = ?`, clientID)
	requireErrorContains(t, err, "client ownership is immutable")
}

func TestProductsRejectOwnershipReparenting(t *testing.T) {
	conn, cleanup := newTenantIntegrityDB(t)
	defer cleanup()

	insertAccount(t, conn, 2, "Second account")
	firstClientID := insertClient(t, conn, 1, "Client")
	secondClientID := insertClient(t, conn, 2, "Other Client")
	productID := insertProduct(t, conn, 1, firstClientID, "Product")

	_, err := conn.Exec(`
		UPDATE products
		SET account_id = 2, client_id = ?
		WHERE id = ?
	`, secondClientID, productID)
	requireErrorContains(t, err, "product ownership is immutable")
}

func TestUsersRejectOwnershipReparenting(t *testing.T) {
	conn, cleanup := newTenantIntegrityDB(t)
	defer cleanup()

	insertAccount(t, conn, 2, "Second account")
	res, err := conn.Exec(`
		INSERT INTO users (name, email, password_hash, account_id, role)
		VALUES ('Owner', 'owner@example.com', 'hash', 1, 'owner')
	`)
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	userID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("user lastInsertId: %v", err)
	}

	_, err = conn.Exec(`UPDATE users SET account_id = 2 WHERE id = ?`, userID)
	requireErrorContains(t, err, "user ownership is immutable")
}

func TestPaymentsRejectAppliedRevisionFromDifferentInvoice(t *testing.T) {
	conn, cleanup := newTenantIntegrityDB(t)
	defer cleanup()

	clientID := insertClient(t, conn, 1, "Client")
	firstInvoiceID := insertInvoice(t, conn, 1, clientID, 100)
	secondInvoiceID := insertInvoice(t, conn, 1, clientID, 101)
	_ = insertRevision(t, conn, firstInvoiceID, 1)
	secondRevisionID := insertRevision(t, conn, secondInvoiceID, 1)

	_, err := conn.Exec(`
		INSERT INTO payments (
			invoice_id,
			payment_type,
			amount_minor,
			payment_date,
			applied_in_revision_id,
			label
		) VALUES (?, 'payment', 500, '2026-03-31', ?, 'Bad payment')
	`, firstInvoiceID, secondRevisionID)
	requireErrorContains(t, err, "payment applied revision must belong to same invoice")
}

func TestInvoiceItemsRejectProductFromDifferentClient(t *testing.T) {
	conn, cleanup := newTenantIntegrityDB(t)
	defer cleanup()

	firstClientID := insertClient(t, conn, 1, "First Client")
	secondClientID := insertClient(t, conn, 1, "Second Client")
	invoiceID := insertInvoice(t, conn, 1, firstClientID, 100)
	revisionID := insertRevision(t, conn, invoiceID, 1)
	productID := insertProduct(t, conn, 1, secondClientID, "Wrong Client Product")

	_, err := conn.Exec(`
		INSERT INTO invoice_items (
			invoice_revision_id,
			product_id,
			name,
			line_type,
			pricing_mode,
			quantity,
			unit_price_minor,
			line_total_minor,
			minutes_worked,
			sort_order
		) VALUES (?, ?, 'Bad line', 'style', 'flat', 1, 1000, 1000, NULL, 1)
	`, revisionID, productID)
	requireErrorContains(t, err, "invoice item product must belong to same account and client")
}
