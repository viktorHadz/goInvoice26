package invoiceTx_test

import (
	"context"
	"testing"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func insertClientForAccount(t *testing.T, a *app.App, accountID int64, name string) int64 {
	t.Helper()

	if _, err := a.DB.Exec(`
		INSERT INTO accounts (id, name)
		VALUES (?, ?)
		ON CONFLICT(id) DO NOTHING;
	`, accountID, name+" account"); err != nil {
		t.Fatalf("insert account %d: %v", accountID, err)
	}

	res, err := a.DB.Exec(`
		INSERT INTO clients (account_id, name)
		VALUES (?, ?);
	`, accountID, name)
	if err != nil {
		t.Fatalf("insert client for account %d: %v", accountID, err)
	}

	clientID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("client lastInsertId for account %d: %v", accountID, err)
	}

	return clientID
}

func TestCreate_AllowsSameBaseNumberAcrossAccounts(t *testing.T) {
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientOneID := insertClient(t, a)
	clientTwoID := insertClientForAccount(t, a, 2, "Second Account Client")

	ctxOne := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	ctxTwo := accountscope.WithAccountID(context.Background(), 2)

	if _, _, err := invoiceTx.Create(ctxOne, a, draftUpdatePayload(clientOneID, 1, 1000, 50, "Account one line")); err != nil {
		t.Fatalf("Create account one invoice: %v", err)
	}

	if _, _, err := invoiceTx.Create(ctxTwo, a, draftUpdatePayload(clientTwoID, 1, 1000, 50, "Account two line")); err != nil {
		t.Fatalf("Create account two invoice: %v", err)
	}

	var count int
	if err := a.DB.QueryRow(`
		SELECT COUNT(*)
		FROM invoices
		WHERE base_number = 1;
	`).Scan(&count); err != nil {
		t.Fatalf("count invoices: %v", err)
	}
	if count != 2 {
		t.Fatalf("invoice count for base number 1 = %d, want 2", count)
	}

	for _, accountID := range []int64{1, 2} {
		var nextBase int64
		if err := a.DB.QueryRow(`
			SELECT next_base_number
			FROM invoice_number_seq
			WHERE account_id = ?;
		`, accountID).Scan(&nextBase); err != nil {
			t.Fatalf("load next base number for account %d: %v", accountID, err)
		}
		if nextBase != 2 {
			t.Fatalf("account %d next base number = %d, want 2", accountID, nextBase)
		}
	}
}
