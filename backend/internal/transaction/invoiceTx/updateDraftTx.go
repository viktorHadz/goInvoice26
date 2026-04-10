package invoiceTx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

var (
	ErrInvoiceIssuedForDraftUpdate = errors.New("issued invoices must be saved as revisions")
	ErrInvoicePaidForDraftUpdate   = errors.New("paid invoices cannot be edited")
	ErrInvoiceVoidForDraftUpdate   = errors.New("void invoices cannot be edited")
	ErrDraftInvoiceHasRevisions    = errors.New("draft invoice has existing revisions")
)

func UpdateDraft(ctx context.Context, a *app.App, canonical *models.FEInvoiceIn) (invoiceID, revisionID int64, err error) {
	tx, err := a.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	ov := &canonical.Overview

	invoiceID, status, err := LoadInvoiceIDAndStatus(ctx, tx, ov.ClientID, ov.BaseNumber)
	if err != nil {
		if errors.Is(err, ErrInvoiceNotFound) {
			return 0, 0, ErrInvoiceNotFound
		}
		return 0, 0, err
	}

	switch status {
	case "draft":
	case "issued":
		return 0, 0, ErrInvoiceIssuedForDraftUpdate
	case "paid":
		return 0, 0, ErrInvoicePaidForDraftUpdate
	case "void":
		return 0, 0, ErrInvoiceVoidForDraftUpdate
	default:
		return 0, 0, fmt.Errorf("unexpected invoice status: %s", status)
	}

	var revisionCount int64
	if err := tx.QueryRowContext(ctx, `
		SELECT
			COALESCE(MAX(CASE WHEN revision_no = 1 THEN id END), 0),
			COUNT(*)
		FROM invoice_revisions
		WHERE invoice_id = ?;
	`, invoiceID).Scan(&revisionID, &revisionCount); err != nil {
		return 0, 0, fmt.Errorf("load draft revisions: %w", err)
	}
	if revisionID < 1 {
		return 0, 0, fmt.Errorf("draft invoice missing base revision: invoice_id=%d", invoiceID)
	}
	if revisionCount != 1 {
		return 0, 0, ErrDraftInvoiceHasRevisions
	}

	existingPaid, err := sumPaymentsByInvoice(ctx, tx, invoiceID)
	if err != nil {
		return 0, 0, err
	}
	if canonical.Totals.PaidMinor != existingPaid {
		return 0, 0, ErrPaymentStateMismatch
	}

	var dueBy any
	if ov.DueByDate != nil {
		dueBy = *ov.DueByDate
	}

	var note any
	if ov.Note != nil {
		note = *ov.Note
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE invoice_revisions
		SET
			issue_date = ?,
			supply_date = ?,
			due_by_date = ?,
			updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now'),
			client_name = ?,
			client_company_name = ?,
			client_address = ?,
			client_email = ?,
			note = ?,
			vat_rate = ?,
			discount_type = ?,
			discount_rate = ?,
			discount_minor = ?,
			deposit_type = ?,
			deposit_rate = ?,
			deposit_minor = ?,
			subtotal_minor = ?,
			vat_amount_minor = ?,
			total_minor = ?
		WHERE id = ?;
	`,
		ov.IssueDate,
		normalizedOptionalString(ov.SupplyDate),
		dueBy,
		ov.ClientName,
		ov.ClientCompanyName,
		ov.ClientAddress,
		ov.ClientEmail,
		note,
		canonical.Totals.VATRate,
		canonical.Totals.DiscountType,
		canonical.Totals.DiscountRate,
		canonical.Totals.DiscountMinor,
		canonical.Totals.DepositType,
		canonical.Totals.DepositRate,
		canonical.Totals.DepositMinor,
		canonical.Totals.SubtotalMinor,
		canonical.Totals.VatAmountMinor,
		canonical.Totals.TotalMinor,
		revisionID,
	); err != nil {
		return 0, 0, fmt.Errorf("update draft revision: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM invoice_items
		WHERE invoice_revision_id = ?;
	`, revisionID); err != nil {
		return 0, 0, fmt.Errorf("delete draft invoice items: %w", err)
	}

	if err := insertRevisionItems(ctx, tx, revisionID, canonical); err != nil {
		return 0, 0, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE invoices
		SET current_revision_id = ?
		WHERE id = ?;
	`, revisionID, invoiceID); err != nil {
		return 0, 0, fmt.Errorf("update current draft revision: %w", err)
	}

	if err := applyAutoPaidIfSettled(ctx, tx, invoiceID, canonical.Totals.TotalMinor); err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("commit update draft: %w", err)
	}

	return invoiceID, revisionID, nil
}

func normalizedOptionalString(value *string) any {
	if value == nil {
		return nil
	}

	return *value
}
