package invoiceTx

import (
	"context"
	"database/sql"
	"fmt"
)

// DB row for overview + totals, used to assemble models.InvoicePDFData model.
type InvoiceOverviewTotals struct {
	BaseNumber        int64
	RevisionNo        int64
	IssueDate         string
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

// Retrieves overview and totals for a specific invoice revision in a single query.
// PaidMinor is aggregated from the payments table.
func GetInvoiceOverviewTotals(
	ctx context.Context,
	db *sql.DB,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
) (*InvoiceOverviewTotals, error) {
	query := `
		SELECT
			i.base_number,
			r.revision_no,
			r.issue_date,
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
				(SELECT SUM(p.amount_minor) FROM payments p WHERE p.invoice_id = i.id), 0
			) AS paid_minor
		FROM invoices i
		JOIN invoice_revisions r
			ON r.invoice_id = i.id AND r.revision_no = ?
		WHERE i.base_number = ? AND i.client_id = ?
	`

	var o InvoiceOverviewTotals
	err := db.QueryRowContext(ctx, query, revisionNo, baseNumber, clientID).Scan(
		&o.BaseNumber, &o.RevisionNo,
		&o.IssueDate, &o.DueByDate,
		&o.ClientName, &o.ClientCompanyName, &o.ClientAddress, &o.ClientEmail,
		&o.Note,
		&o.VATRate, &o.VATAmountMin,
		&o.DiscountType, &o.DiscountRate, &o.DiscountMinor,
		&o.DepositType, &o.DepositRate, &o.DepositMinor,
		&o.SubtotalMinor, &o.TotalMinor,
		&o.PaidMinor,
	)
	if err != nil {
		return nil, fmt.Errorf("get overview+totals: %w", err)
	}
	return &o, nil
}

type ItemLine struct {
	Name         string
	LineType     string
	Quantity     int64
	UnitPriceMin int64
	LineTotalMin int64
	SortOrder    int64
}

// Retrieves all revision items for an invoice
func GetInvoiceItems(
	ctx context.Context,
	db *sql.DB,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
) ([]ItemLine, error) {
	query := `
		SELECT
			it.sort_order,
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
		WHERE i.base_number = ? AND i.client_id = ?
		ORDER BY it.sort_order ASC
	`

	rows, err := db.QueryContext(ctx, query, revisionNo, baseNumber, clientID)
	if err != nil {
		return nil, fmt.Errorf("query items: %w", err)
	}
	defer rows.Close()

	var items []ItemLine
	for rows.Next() {
		var item ItemLine

		if err := rows.Scan(
			&item.SortOrder,
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
