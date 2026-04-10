package invoiceTx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
)

// ItemLine is a DB/query row for invoice items.
//
// It is returned by [QueryInvoiceLines].
// Do not return it directly in HTTP responses.
// Map it to an API response model in the handler layer.
type ItemLine struct {
	ProductID     *int64
	PricingMode   *string
	Name          string
	LineType      string
	Quantity      int64
	MinutesWorked *int64
	UnitPriceMin  int64
	LineTotalMin  int64
	SortOrder     int64
}

// InvoiceOverviewTotals is a DB/query row for invoice overview and totals.
//
// It is returned by [QueryInvoiceSummary].
// Do not send it directly in JSON responses.
// Map it to an API response model in the handler layer.
type InvoiceOverviewTotals struct {
	Status            string
	BaseNumber        int64
	RevisionNo        int64
	IssueDate         string
	SupplyDate        sql.NullString
	DueByDate         sql.NullString
	ClientName        string
	ClientCompanyName string
	ClientAddress     string
	ClientEmail       string
	Note              sql.NullString

	VATRate       int64
	VATAmountMin  int64
	DiscountType  string
	DiscountRate  int64
	DiscountMinor int64
	DepositType   string
	DepositRate   int64
	DepositMinor  int64
	SubtotalMinor int64
	TotalMinor    int64
	PaidMinor     int64
}

type PaymentRow struct {
	ID          int64
	AmountMinor int64
	PaymentDate string
	PaymentType string
	Label       sql.NullString
}

// QueryInvoiceSummary returns the DB/query row for one invoice revision.
//
// The returned value is an internal backend shape.
// Handlers should map it to an API response model before sending via res.JSON.
//
// PaidMinor is aggregated from the payments table.
func QueryInvoiceSummary(
	ctx context.Context,
	db *sql.DB,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
) (*InvoiceOverviewTotals, error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			i.status,
			i.base_number,
			r.revision_no,
			r.issue_date,
			r.supply_date,
			r.due_by_date,
			r.client_name,
			r.client_company_name,
			r.client_address,
			r.client_email,
			r.note,
			r.vat_rate,
			r.vat_amount_minor,
			r.discount_type,
			r.discount_rate,
			r.discount_minor,
			r.deposit_type,
			r.deposit_rate,
			r.deposit_minor,
			r.subtotal_minor,
			r.total_minor,
			COALESCE(
				(
					SELECT SUM(p.amount_minor)
					FROM payments p
					JOIN invoice_revisions ap ON ap.id = p.applied_in_revision_id
					WHERE p.invoice_id = i.id
						AND p.payment_type = 'payment'
						AND ap.revision_no <= r.revision_no
				), 0
			) AS paid_minor
		FROM invoices i
		JOIN invoice_revisions r
			ON r.invoice_id = i.id AND r.revision_no = ?
		WHERE i.account_id = ? AND i.base_number = ? AND i.client_id = ?
	`

	var o InvoiceOverviewTotals
	err = db.QueryRowContext(ctx, query, revisionNo, accountID, baseNumber, clientID).Scan(
		&o.Status,
		&o.BaseNumber, &o.RevisionNo,
		&o.IssueDate, &o.SupplyDate, &o.DueByDate,
		&o.ClientName, &o.ClientCompanyName, &o.ClientAddress, &o.ClientEmail,
		&o.Note,
		&o.VATRate, &o.VATAmountMin,
		&o.DiscountType, &o.DiscountRate, &o.DiscountMinor,
		&o.DepositType, &o.DepositRate, &o.DepositMinor,
		&o.SubtotalMinor, &o.TotalMinor,
		&o.PaidMinor,
	)
	if err != nil {
		return nil, fmt.Errorf("GetInvoiceSummary() => %w,\nrevisionNumber: %v,\nbaseNumber: %v,\nclientID: %v", err, revisionNo, baseNumber, clientID)
	}
	return &o, nil
}

func QueryInvoicePaymentsForRevision(
	ctx context.Context,
	db *sql.DB,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
) ([]PaymentRow, error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			p.id,
			p.amount_minor,
			p.payment_date,
			p.payment_type,
			p.label
		FROM invoices i
		JOIN invoice_revisions r
			ON r.invoice_id = i.id AND r.revision_no = ?
		JOIN payments p
			ON p.invoice_id = i.id
		JOIN invoice_revisions ap
			ON ap.id = p.applied_in_revision_id
		WHERE i.base_number = ? AND i.client_id = ?
			AND i.account_id = ?
			AND p.payment_type = 'payment'
			AND ap.revision_no <= r.revision_no
		ORDER BY p.payment_date ASC, p.id ASC
	`

	rows, err := db.QueryContext(ctx, query, revisionNo, baseNumber, clientID, accountID)
	if err != nil {
		return nil, fmt.Errorf(
			"QueryInvoicePaymentsForRevision()=> %w\nrevisionNumber: %v,\nbaseNumber: %v,\nclientID: %v,",
			err,
			revisionNo,
			baseNumber,
			clientID,
		)
	}
	defer rows.Close()

	payments := make([]PaymentRow, 0)
	for rows.Next() {
		var p PaymentRow
		if err := rows.Scan(
			&p.ID,
			&p.AmountMinor,
			&p.PaymentDate,
			&p.PaymentType,
			&p.Label,
		); err != nil {
			return nil, fmt.Errorf("scan payment row: %w", err)
		}
		payments = append(payments, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("payment rows error: %w", err)
	}

	return payments, nil
}

// QueryInvoiceLines returns the DB/query rows for one invoice revision.
//
// The returned lines are internal backend shapes.
// Handlers should map them to API response models before sending via res.JSON.
func QueryInvoiceLines(
	ctx context.Context,
	db *sql.DB,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
) ([]ItemLine, error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT
			it.sort_order,
			it.product_id,
			it.pricing_mode,
			it.minutes_worked,
			it.name,
			it.line_type,
			it.quantity,
			it.unit_price_minor,
			it.line_total_minor
		FROM invoices i
		JOIN invoice_revisions r
			ON r.invoice_id = i.id AND r.revision_no = ?
		JOIN invoice_items it
			ON it.invoice_revision_id = r.id
		WHERE i.account_id = ? AND i.base_number = ? AND i.client_id = ?
		ORDER BY it.sort_order ASC
	`

	rows, err := db.QueryContext(ctx, query, revisionNo, accountID, baseNumber, clientID)
	if err != nil {
		return nil, fmt.Errorf("QueryInvoiceItems()=> %w\nctx: %v,\nquery: %v,\nrevisionNumber: %v,\nbaseNumber: %v,\nclientID: %v,", err, ctx, query, revisionNo, baseNumber, clientID)
	}
	defer rows.Close()

	var items []ItemLine
	for rows.Next() {
		var item ItemLine

		if err := rows.Scan(
			&item.SortOrder,
			&item.ProductID,
			&item.PricingMode,
			&item.MinutesWorked,
			&item.Name,
			&item.LineType,
			&item.Quantity,
			&item.UnitPriceMin,
			&item.LineTotalMin,
		); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return items, nil
}
