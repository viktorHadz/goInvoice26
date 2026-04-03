package invoiceTx_test

import (
	"context"
	"testing"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func TestGetSuggestedNextBaseNumber_UsesSequenceWithSafetyClamp(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	if _, err := a.DB.Exec(`
		UPDATE invoice_number_seq
		SET next_base_number = 100
		WHERE account_id = 1;
	`); err != nil {
		t.Fatalf("seed invoice sequence: %v", err)
	}

	got, err := invoiceTx.GetSuggestedNextBaseNumber(ctx, a)
	if err != nil {
		t.Fatalf("GetSuggestedNextBaseNumber: %v", err)
	}
	if got != 100 {
		t.Fatalf("next number = %d, want 100", got)
	}

	clientID := insertClient(t, a)
	insertInvoiceGraph(t, a, clientID, 150, "draft")

	got, err = invoiceTx.GetSuggestedNextBaseNumber(ctx, a)
	if err != nil {
		t.Fatalf("GetSuggestedNextBaseNumber after invoice: %v", err)
	}
	if got != 151 {
		t.Fatalf("next number = %d, want 151", got)
	}
}

func TestGetSuggestedNextBaseNumber_IsScopedPerAccount(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	if _, err := a.DB.Exec(`
		INSERT INTO accounts (id, name) VALUES (2, 'Second account')
		ON CONFLICT(id) DO NOTHING;
	`); err != nil {
		t.Fatalf("insert second account: %v", err)
	}
	if _, err := a.DB.Exec(`
		INSERT INTO invoice_number_seq (account_id, next_base_number)
		VALUES (2, 7)
		ON CONFLICT(account_id) DO UPDATE SET next_base_number = excluded.next_base_number;
	`); err != nil {
		t.Fatalf("seed second account invoice sequence: %v", err)
	}
	if _, err := a.DB.Exec(`
		INSERT INTO clients (account_id, name)
		VALUES (2, 'Second Account Client');
	`); err != nil {
		t.Fatalf("insert second account client: %v", err)
	}

	secondCtx := accountscope.WithAccountID(ctx, 2)
	got, err := invoiceTx.GetSuggestedNextBaseNumber(secondCtx, a)
	if err != nil {
		t.Fatalf("GetSuggestedNextBaseNumber second account: %v", err)
	}
	if got != 7 {
		t.Fatalf("second account next number = %d, want 7", got)
	}
}
