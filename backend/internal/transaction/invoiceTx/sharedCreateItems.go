package invoiceTx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/models"
)

func insertRevisionWithItems(
	ctx context.Context,
	tx *sql.Tx,
	invoiceID int64,
	revisionNo int64,
	canonical *models.FEInvoiceIn,
) (revisionID int64, err error) {
	ov := &canonical.Overview
	tot := &canonical.Totals

	var dueBy interface{}
	if ov.DueByDate != nil {
		dueBy = *ov.DueByDate
	}

	var supplyDate interface{}
	if ov.SupplyDate != nil {
		supplyDate = *ov.SupplyDate
	}

	var note interface{}
	if ov.Note != nil {
		note = *ov.Note
	}

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO invoice_revisions (
			invoice_id, revision_no,
			issue_date, supply_date, due_by_date,
			client_name, client_company_name, client_address, client_email, note,
			vat_rate,
			discount_type, discount_rate, discount_minor,
			deposit_type, deposit_rate, deposit_minor,
			subtotal_minor, vat_amount_minor, total_minor
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id;
	`,
		invoiceID, revisionNo,
		ov.IssueDate, supplyDate, dueBy,
		ov.ClientName, ov.ClientCompanyName, ov.ClientAddress, ov.ClientEmail, note,
		tot.VATRate,
		tot.DiscountType, tot.DiscountRate, tot.DiscountMinor,
		tot.DepositType, tot.DepositRate, tot.DepositMinor,
		tot.SubtotalMinor, tot.VatAmountMinor, tot.TotalMinor,
	).Scan(&revisionID); err != nil {
		return 0, fmt.Errorf("insert invoice_revision: %w", err)
	}

	if err := insertRevisionItems(ctx, tx, revisionID, canonical); err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE invoices
		SET current_revision_id = ?
		WHERE id = ?
	`, revisionID, invoiceID); err != nil {
		return 0, fmt.Errorf("update invoices.current_revision_id: %w", err)
	}

	return revisionID, nil
}

func insertRevisionItems(
	ctx context.Context,
	tx *sql.Tx,
	revisionID int64,
	canonical *models.FEInvoiceIn,
) error {
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO invoice_items (
			invoice_revision_id, product_id, name, line_type, pricing_mode,
			quantity, unit_price_minor, line_total_minor, minutes_worked, sort_order
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare invoice_items: %w", err)
	}
	defer stmt.Close()

	for _, ln := range canonical.Lines {
		var productID interface{}
		if ln.ProductID != nil {
			productID = *ln.ProductID
		}

		var minutesWorked interface{}
		if ln.MinutesWorked != nil {
			minutesWorked = *ln.MinutesWorked
		}

		_, err := stmt.ExecContext(ctx,
			revisionID,
			productID,
			ln.Name,
			ln.LineType,
			ln.PricingMode,
			ln.Quantity,
			ln.UnitPriceMinor,
			ln.LineTotalMinor,
			minutesWorked,
			ln.SortOrder,
		)
		if err != nil {
			return fmt.Errorf("insert invoice_item: %w", err)
		}
	}

	return nil
}
