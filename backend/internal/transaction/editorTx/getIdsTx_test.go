package editorTx_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/transaction/editorTx"
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

func insertInvoiceBookInvoice(
	t *testing.T,
	a *app.App,
	clientID int64,
	baseNumber int64,
	status string,
	latestRevisionNo int,
	issueDate string,
	totalMinor int64,
	depositMinor int64,
	paidMinor int64,
) int64 {
	t.Helper()

	var accountID int64
	if err := a.DB.QueryRow(`SELECT account_id FROM clients WHERE id = ?`, clientID).Scan(&accountID); err != nil {
		t.Fatalf("load client account id: %v", err)
	}

	res, err := a.DB.Exec(`
		INSERT INTO invoices (account_id, client_id, base_number, status)
		VALUES (?, ?, ?, ?)
	`, accountID, clientID, baseNumber, status)
	if err != nil {
		t.Fatalf("insert invoice: %v", err)
	}

	invoiceID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("invoice lastInsertId: %v", err)
	}

	var currentRevisionID int64
	for revisionNo := 1; revisionNo <= latestRevisionNo; revisionNo++ {
		revisionIssueDate := issueDate
		if revisionNo != latestRevisionNo {
			revisionIssueDate = "2026-03-01"
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
			) VALUES (?, ?, ?, '2026-04-10', 'Test Client', '', '', '', NULL, 0, 'none', 0, 0, 'fixed', 0, ?, ?, 0, ?)
		`, invoiceID, revisionNo, revisionIssueDate, depositMinor, totalMinor, totalMinor)
		if err != nil {
			t.Fatalf("insert invoice revision %d: %v", revisionNo, err)
		}

		currentRevisionID, err = res.LastInsertId()
		if err != nil {
			t.Fatalf("revision lastInsertId: %v", err)
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
			) VALUES (?, 'Service line', 'custom', 'flat', 1, ?, ?, NULL, 1)
		`, currentRevisionID, totalMinor, totalMinor); err != nil {
			t.Fatalf("insert invoice item: %v", err)
		}
	}

	if _, err := a.DB.Exec(`
		UPDATE invoices
		SET current_revision_id = ?
		WHERE id = ?
	`, currentRevisionID, invoiceID); err != nil {
		t.Fatalf("update current revision: %v", err)
	}

	if paidMinor > 0 {
		if _, err := a.DB.Exec(`
			INSERT INTO payments (
				invoice_id,
				payment_type,
				amount_minor,
				payment_date,
				applied_in_revision_id,
				label
			) VALUES (?, 'payment', ?, '2026-03-28', ?, 'Recorded payment')
		`, invoiceID, paidMinor, currentRevisionID); err != nil {
			t.Fatalf("insert payment: %v", err)
		}
	}

	return invoiceID
}

func TestQueryInvoiceBookPage_FiltersUnpaidAndSortsByOutstanding(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a)
	insertInvoiceBookInvoice(t, a, clientID, 101, "issued", 2, "2026-03-22", 10000, 0, 2000)
	insertInvoiceBookInvoice(t, a, clientID, 102, "paid", 1, "2026-03-18", 7000, 0, 7000)
	insertInvoiceBookInvoice(t, a, clientID, 103, "issued", 1, "2026-03-20", 4000, 1000, 1000)

	got, err := editorTx.QueryInvoiceBookPage(a, ctx, clientID, 10, 0, editorTx.InvoiceBookPageFilters{
		SortBy:       "balance",
		PaymentState: "unpaid",
	})
	if err != nil {
		t.Fatalf("QueryInvoiceBookPage: %v", err)
	}

	if got.Total != 2 {
		t.Fatalf("total = %d, want 2", got.Total)
	}
	if len(got.Items) != 2 {
		t.Fatalf("items len = %d, want 2", len(got.Items))
	}

	if got.Items[0].BaseNo != 101 || got.Items[0].BalanceDueMinor != 8000 {
		t.Fatalf("first item = %+v, want base 101 with balance 8000", got.Items[0])
	}
	if got.Items[0].LatestRevisionNo != 2 {
		t.Fatalf("latest revision = %d, want 2", got.Items[0].LatestRevisionNo)
	}
	if got.Items[1].BaseNo != 103 || got.Items[1].BalanceDueMinor != 2000 {
		t.Fatalf("second item = %+v, want base 103 with balance 2000", got.Items[1])
	}
}

func TestQueryInvoiceBookPage_FiltersPaidInvoices(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a)
	insertInvoiceBookInvoice(t, a, clientID, 201, "issued", 1, "2026-03-12", 5000, 0, 1000)
	insertInvoiceBookInvoice(t, a, clientID, 202, "paid", 1, "2026-03-11", 6000, 0, 6000)
	insertInvoiceBookInvoice(t, a, clientID, 203, "void", 1, "2026-03-10", 3000, 0, 0)

	got, err := editorTx.QueryInvoiceBookPage(a, ctx, clientID, 10, 0, editorTx.InvoiceBookPageFilters{
		SortBy:        "date",
		SortDirection: "asc",
		PaymentState:  "paid",
	})
	if err != nil {
		t.Fatalf("QueryInvoiceBookPage: %v", err)
	}

	if got.Total != 1 {
		t.Fatalf("total = %d, want 1", got.Total)
	}
	if len(got.Items) != 1 {
		t.Fatalf("items len = %d, want 1", len(got.Items))
	}

	item := got.Items[0]
	if item.BaseNo != 202 {
		t.Fatalf("base number = %d, want 202", item.BaseNo)
	}
	if item.PaidMinor != 6000 {
		t.Fatalf("paid minor = %d, want 6000", item.PaidMinor)
	}
	if item.BalanceDueMinor != 0 {
		t.Fatalf("balance due = %d, want 0", item.BalanceDueMinor)
	}
}

func TestQueryInvoiceBookPage_LoadsAllClientsByDefaultAndCanScopeToOneClient(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientOneID := insertClient(t, a)
	clientTwoID := insertClient(t, a)

	insertInvoiceBookInvoice(t, a, clientOneID, 301, "issued", 1, "2026-03-15", 5000, 0, 0)
	insertInvoiceBookInvoice(t, a, clientTwoID, 401, "issued", 1, "2026-03-16", 6000, 0, 0)

	allClients, err := editorTx.QueryInvoiceBookPage(a, ctx, 0, 10, 0, editorTx.InvoiceBookPageFilters{})
	if err != nil {
		t.Fatalf("QueryInvoiceBookPage all clients: %v", err)
	}
	if allClients.Total != 2 {
		t.Fatalf("all-clients total = %d, want 2", allClients.Total)
	}

	scoped, err := editorTx.QueryInvoiceBookPage(
		a,
		ctx,
		clientOneID,
		10,
		0,
		editorTx.InvoiceBookPageFilters{},
	)
	if err != nil {
		t.Fatalf("QueryInvoiceBookPage scoped: %v", err)
	}
	if scoped.Total != 1 {
		t.Fatalf("scoped total = %d, want 1", scoped.Total)
	}
	if len(scoped.Items) != 1 || scoped.Items[0].ClientID != clientOneID {
		t.Fatalf("scoped items = %+v, want only client %d", scoped.Items, clientOneID)
	}
}

func TestQueryInvoiceBookPage_IsScopedPerAccount(t *testing.T) {
	defaultCtx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	secondCtx := accountscope.WithAccountID(context.Background(), 2)
	a, cleanup := newTestApp(t)
	defer cleanup()

	if _, err := a.DB.Exec(`INSERT INTO accounts (id, name) VALUES (2, 'Second account')`); err != nil {
		t.Fatalf("insert second account: %v", err)
	}

	defaultClientID := insertClient(t, a)
	res, err := a.DB.Exec(`INSERT INTO clients (account_id, name) VALUES (2, 'Second Account Client')`)
	if err != nil {
		t.Fatalf("insert second account client: %v", err)
	}
	secondClientID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("second client lastInsertId: %v", err)
	}

	insertInvoiceBookInvoice(t, a, defaultClientID, 501, "issued", 1, "2026-03-21", 5000, 0, 0)
	insertInvoiceBookInvoice(t, a, secondClientID, 601, "issued", 1, "2026-03-22", 6000, 0, 0)

	defaultPage, err := editorTx.QueryInvoiceBookPage(a, defaultCtx, 0, 10, 0, editorTx.InvoiceBookPageFilters{})
	if err != nil {
		t.Fatalf("QueryInvoiceBookPage default account: %v", err)
	}
	if defaultPage.Total != 1 || len(defaultPage.Items) != 1 || defaultPage.Items[0].BaseNo != 501 {
		t.Fatalf("default account page = %+v, want only base 501", defaultPage)
	}

	secondPage, err := editorTx.QueryInvoiceBookPage(a, secondCtx, 0, 10, 0, editorTx.InvoiceBookPageFilters{})
	if err != nil {
		t.Fatalf("QueryInvoiceBookPage second account: %v", err)
	}
	if secondPage.Total != 1 || len(secondPage.Items) != 1 || secondPage.Items[0].BaseNo != 601 {
		t.Fatalf("second account page = %+v, want only base 601", secondPage)
	}
}
