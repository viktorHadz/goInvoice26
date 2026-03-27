package invoiceTx_test

import (
	"context"
	"testing"

	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func TestGetSuggestedNextBaseNumber_UsesSequenceWithSafetyClamp(t *testing.T) {
	ctx := context.Background()
	a, cleanup := newTestApp(t)
	defer cleanup()

	if _, err := a.DB.Exec(`
		UPDATE invoice_number_seq
		SET next_base_number = 100
		WHERE id = 1;
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
