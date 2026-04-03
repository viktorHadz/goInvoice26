package invoiceTx_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func draftUpdatePayload(clientID, baseNumber int64, totalMinor int64, paidMinor int64, lineName string) *models.FEInvoiceIn {
	dueBy := "2026-04-15"
	note := "Updated draft note"

	return &models.FEInvoiceIn{
		Overview: models.InvoiceCreateIn{
			ClientID:          clientID,
			BaseNumber:        baseNumber,
			IssueDate:         "2026-03-30",
			DueByDate:         &dueBy,
			ClientName:        "Updated Client",
			ClientCompanyName: "Acme Co",
			ClientAddress:     "2 Updated Street",
			ClientEmail:       "updated@example.com",
			Note:              &note,
		},
		Lines: []models.LineCreateIn{
			{
				Name:           lineName,
				LineType:       "custom",
				PricingMode:    "flat",
				Quantity:       1,
				UnitPriceMinor: totalMinor,
				LineTotalMinor: totalMinor,
				SortOrder:      1,
			},
		},
		Totals: models.TotalsCreateIn{
			VATRate:           0,
			VatAmountMinor:    0,
			DepositType:       "none",
			DepositRate:       0,
			DepositMinor:      0,
			DiscountType:      "none",
			DiscountRate:      0,
			DiscountMinor:     0,
			PaidMinor:         paidMinor,
			SubtotalAfterDisc: totalMinor,
			SubtotalMinor:     totalMinor,
			TotalMinor:        totalMinor,
			BalanceDue:        totalMinor - paidMinor,
		},
		Payments: []models.PaymentCreateIn{
			{
				AmountMinor: 50,
				PaymentDate: "2026-03-31",
			},
		},
	}
}

func TestUpdateDraft_ReplacesBaseRevisionInPlace(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a)
	invoiceID := insertInvoiceGraph(t, a, clientID, 101, "draft")

	var beforeRevisionID int64
	if err := a.DB.QueryRow(`SELECT current_revision_id FROM invoices WHERE id = ?`, invoiceID).Scan(&beforeRevisionID); err != nil {
		t.Fatalf("load current revision id: %v", err)
	}

	gotInvoiceID, gotRevisionID, err := invoiceTx.UpdateDraft(ctx, a, draftUpdatePayload(clientID, 101, 150, 150, "Updated service line"))
	if err != nil {
		t.Fatalf("UpdateDraft: %v", err)
	}
	if gotInvoiceID != invoiceID {
		t.Fatalf("invoice id = %d, want %d", gotInvoiceID, invoiceID)
	}
	if gotRevisionID != beforeRevisionID {
		t.Fatalf("revision id = %d, want %d", gotRevisionID, beforeRevisionID)
	}

	if count := countRows(t, a, "invoice_revisions", invoiceID); count != 1 {
		t.Fatalf("revision count = %d, want 1", count)
	}

	var (
		issueDate string
		note      sql.NullString
	)
	if err := a.DB.QueryRow(`
		SELECT issue_date, note
		FROM invoice_revisions
		WHERE id = ?
	`, gotRevisionID).Scan(&issueDate, &note); err != nil {
		t.Fatalf("load updated revision: %v", err)
	}
	if issueDate != "2026-03-30" {
		t.Fatalf("issue date = %q, want 2026-03-30", issueDate)
	}
	if !note.Valid || note.String != "Updated draft note" {
		t.Fatalf("note = %#v, want Updated draft note", note)
	}

	var itemName string
	if err := a.DB.QueryRow(`
		SELECT name
		FROM invoice_items
		WHERE invoice_revision_id = ?
	`, gotRevisionID).Scan(&itemName); err != nil {
		t.Fatalf("load updated item: %v", err)
	}
	if itemName != "Updated service line" {
		t.Fatalf("item name = %q, want Updated service line", itemName)
	}

	if count := countRows(t, a, "payments", invoiceID); count != 2 {
		t.Fatalf("payment count = %d, want 2", count)
	}

	var status string
	if err := a.DB.QueryRow(`SELECT status FROM invoices WHERE id = ?`, invoiceID).Scan(&status); err != nil {
		t.Fatalf("load invoice status: %v", err)
	}
	if status != "draft" {
		t.Fatalf("status = %q, want draft", status)
	}
}

func TestUpdateDraft_RejectsDraftWithMultipleRevisions(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a)
	invoiceID := insertInvoiceGraph(t, a, clientID, 102, "draft")

	res, err := a.DB.Exec(`
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
		) VALUES (?, 2, '2026-04-01', '2026-04-15', 'Test Client', '', '', '', NULL, 0, 'none', 0, 0, 'none', 0, 0, 1000, 0, 1000)
	`, invoiceID)
	if err != nil {
		t.Fatalf("insert second revision: %v", err)
	}
	secondRevisionID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("second revision lastInsertId: %v", err)
	}
	if _, err := a.DB.Exec(`
		UPDATE invoices
		SET current_revision_id = ?
		WHERE id = ?
	`, secondRevisionID, invoiceID); err != nil {
		t.Fatalf("update current revision: %v", err)
	}

	_, _, err = invoiceTx.UpdateDraft(ctx, a, draftUpdatePayload(clientID, 102, 1200, 150, "Ignored"))
	if !errors.Is(err, invoiceTx.ErrDraftInvoiceHasRevisions) {
		t.Fatalf("UpdateDraft() error = %v, want %v", err, invoiceTx.ErrDraftInvoiceHasRevisions)
	}
}

func TestCreateRevision_RejectsNonIssuedInvoices(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)

	tests := []struct {
		name    string
		status  string
		wantErr error
	}{
		{name: "draft", status: "draft", wantErr: invoiceTx.ErrInvoiceDraftForRevision},
		{name: "paid", status: "paid", wantErr: invoiceTx.ErrInvoicePaidForRevision},
		{name: "void", status: "void", wantErr: invoiceTx.ErrInvoiceVoidForRevision},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, cleanup := newTestApp(t)
			defer cleanup()

			clientID := insertClient(t, a)
			insertInvoiceGraph(t, a, clientID, 150, tt.status)

			_, _, _, err := invoiceTx.CreateRevision(ctx, a, draftUpdatePayload(clientID, 150, 1200, 100, "Revision line"))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateRevision() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateRevision_AutoMarksIssuedInvoicePaidWhenSettled(t *testing.T) {
	ctx := accountscope.WithAccountID(context.Background(), accountscope.DefaultAccountID)
	a, cleanup := newTestApp(t)
	defer cleanup()

	clientID := insertClient(t, a)
	invoiceID := insertInvoiceGraph(t, a, clientID, 160, "issued")

	_, _, _, err := invoiceTx.CreateRevision(ctx, a, draftUpdatePayload(clientID, 160, 150, 150, "Issued revision line"))
	if err != nil {
		t.Fatalf("CreateRevision: %v", err)
	}

	var status string
	if err := a.DB.QueryRow(`SELECT status FROM invoices WHERE id = ?`, invoiceID).Scan(&status); err != nil {
		t.Fatalf("load invoice status: %v", err)
	}
	if status != "paid" {
		t.Fatalf("status = %q, want paid", status)
	}
}
