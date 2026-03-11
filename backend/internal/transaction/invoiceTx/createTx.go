/*
The invoiceTx package exposes methods for:
  - Invoice creation

And allows retrieval of:
  - next invoice number,
  - totals,
  - line items,
  - client details
*/
package invoiceTx

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

// Create inserts a new invoice with one revision (revision_no 1) and all line items in a single transaction.
//
// The canonical invoice must already be validated and recalculated (use RecalcInvoice output).
//
// Returns (invoiceID, revisionID, error). On success, invoices.current_revision_id is set to the new revision.
func Create(ctx context.Context, a *app.App, canonical *models.FEInvoiceIn) (invoiceID, revisionID int64, err error) {
	tx, err := a.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	ov := &canonical.Overview
	tot := &canonical.Totals

	// 1. Insert invoice row (current_revision_id set after revision insert)
	var invID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO invoices (client_id, base_number, status)
		VALUES (?, ?, 'draft')
		RETURNING id;
	`, ov.ClientID, ov.BaseNumber).Scan(&invID); err != nil {
		if isUniqueViolation(err) {
			return 0, 0, fmt.Errorf("invoice base_number %d already exists: %w", ov.BaseNumber, err)
		}
		return 0, 0, fmt.Errorf("insert invoice: %w", err)
	}

	// 2. Insert revision (revision_no = 1)
	var dueBy, note any
	if ov.DueByDate != nil {
		dueBy = *ov.DueByDate
	} else {
		dueBy = nil
	}
	if ov.Note != nil {
		note = *ov.Note
	} else {
		note = nil
	}

	var revID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO invoice_revisions (
			invoice_id, revision_no,
			issue_date, due_by_date,
			client_name, client_company_name, client_address, client_email, note,
			vat_rate,
			discount_type, discount_rate, discount_minor,
			deposit_type, deposit_rate, deposit_minor,
			subtotal_minor, vat_amount_minor, total_minor
		) VALUES (?, 1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id;
	`,
		invID,
		ov.IssueDate, dueBy,
		ov.ClientName, ov.ClientCompanyName, ov.ClientAddress, ov.ClientEmail, note,
		tot.VATRate,
		tot.DiscountType, tot.DiscountRate, tot.DiscountMinor,
		tot.DepositType, tot.DepositRate, tot.DepositMinor,
		tot.SubtotalMinor, tot.VatAmountMinor, tot.TotalMinor,
	).Scan(&revID); err != nil {
		return 0, 0, fmt.Errorf("insert invoice_revision: %w", err)
	}

	// 3. Insert line items
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO invoice_items (
			invoice_revision_id, product_id, name, line_type, pricing_mode,
			quantity, unit_price_minor, line_total_minor, minutes_worked, sort_order
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("prepare invoice_items: %w", err)
	}
	defer stmt.Close()

	for _, ln := range canonical.Lines {
		var productID, minutesWorked interface{}
		if ln.ProductID != nil {
			productID = *ln.ProductID
		} else {
			productID = nil
		}
		if ln.MinutesWorked != nil {
			minutesWorked = *ln.MinutesWorked
		} else {
			minutesWorked = nil
		}
		_, err := stmt.ExecContext(ctx,
			revID, productID, ln.Name, ln.LineType, ln.PricingMode,
			ln.Quantity, ln.UnitPriceMinor, ln.LineTotalMinor, minutesWorked, ln.SortOrder,
		)
		if err != nil {
			return 0, 0, fmt.Errorf("insert invoice_item: %w", err)
		}
	}

	// 4. Set current revision on invoice
	if _, err := tx.ExecContext(ctx, `
		UPDATE invoices SET current_revision_id = ? WHERE id = ?
	`, revID, invID); err != nil {
		return 0, 0, fmt.Errorf("update invoices.current_revision_id: %w", err)
	}

	// 5. Keep sequence in sync so suggested next number stays correct (no allocation on GET).
	if _, err := tx.ExecContext(ctx, `
		UPDATE invoice_number_seq
		SET next_base_number = MAX(next_base_number, ?)
		WHERE id = 1;
	`, ov.BaseNumber+1); err != nil {
		return 0, 0, fmt.Errorf("sync invoice_number_seq: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("commit: %w", err)
	}
	return invID, revID, nil
}

// isUniqueViolation returns true if the error is a SQLite unique constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE") || strings.Contains(msg, "unique")
}
