package clientsTx_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
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

func insertClient(t *testing.T, a *app.App, name string) int64 {
	t.Helper()

	res, err := a.DB.Exec(`INSERT INTO clients (name) VALUES (?)`, name)
	if err != nil {
		t.Fatalf("insert client: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("client lastInsertId: %v", err)
	}

	return id
}

func insertInvoice(t *testing.T, a *app.App, clientID, baseNumber int64) {
	t.Helper()

	var accountID int64
	if err := a.DB.QueryRow(`SELECT account_id FROM clients WHERE id = ?`, clientID).Scan(&accountID); err != nil {
		t.Fatalf("load client account id: %v", err)
	}

	if _, err := a.DB.Exec(
		`INSERT INTO invoices (account_id, client_id, base_number, status) VALUES (?, ?, ?, 'draft')`,
		accountID,
		clientID,
		baseNumber,
	); err != nil {
		t.Fatalf("insert invoice: %v", err)
	}
}

func TestDeleteClient_RemovesClientWithoutInvoices(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a, "Delete Me")

	affected, err := clientsTx.DeleteClient(a, ctx, clientID)
	if err != nil {
		t.Fatalf("DeleteClient: %v", err)
	}
	if affected != 1 {
		t.Fatalf("affected rows: got %d want 1", affected)
	}

	var count int
	if err := a.DB.QueryRow(`SELECT COUNT(*) FROM clients WHERE id = ?`, clientID).Scan(&count); err != nil {
		t.Fatalf("count client: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected client to be deleted, count=%d", count)
	}
}

func TestDeleteClient_ReturnsFriendlyErrorWhenInvoicesExist(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a, "In Use")
	insertInvoice(t, a, clientID, 1001)

	affected, err := clientsTx.DeleteClient(a, ctx, clientID)
	if affected != 0 {
		t.Fatalf("affected rows: got %d want 0", affected)
	}
	if !errors.Is(err, clientsTx.ErrClientHasInvoices) {
		t.Fatalf("expected ErrClientHasInvoices, got %v", err)
	}

	var count int
	if err := a.DB.QueryRow(`SELECT COUNT(*) FROM clients WHERE id = ?`, clientID).Scan(&count); err != nil {
		t.Fatalf("count client: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected client to remain, count=%d", count)
	}
}

func TestDeleteClient_DoesNotDeleteAnotherAccountsClient(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	if _, err := a.DB.Exec(`INSERT INTO accounts (id, name) VALUES (2, 'Second account')`); err != nil {
		t.Fatalf("insert second account: %v", err)
	}

	res, err := a.DB.Exec(`INSERT INTO clients (account_id, name) VALUES (2, 'Other Account Client')`)
	if err != nil {
		t.Fatalf("insert second account client: %v", err)
	}
	clientID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("client lastInsertId: %v", err)
	}

	affected, err := clientsTx.DeleteClient(a, ctx, clientID)
	if err != nil {
		t.Fatalf("DeleteClient other account: %v", err)
	}
	if affected != 0 {
		t.Fatalf("affected rows: got %d want 0", affected)
	}

	var count int
	if err := a.DB.QueryRow(`SELECT COUNT(*) FROM clients WHERE id = ?`, clientID).Scan(&count); err != nil {
		t.Fatalf("count client: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected other account client to remain, count=%d", count)
	}
}

func TestDeleteClient_RequiresAccountScope(t *testing.T) {
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a, "Missing Scope")

	_, err := clientsTx.DeleteClient(a, context.Background(), clientID)
	if !errors.Is(err, accountscope.ErrMissing) {
		t.Fatalf("DeleteClient missing scope error = %v, want ErrMissing", err)
	}
}
