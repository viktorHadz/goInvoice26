package invoiceTx

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
)

type InvoiceHistoryRow struct {
	ID          int64
	InvoiceID   int64
	Type        string
	CreatedAt   string
	RevisionNo  sql.NullInt64
	ReceiptNo   sql.NullInt64
	IssueDate   sql.NullString
	DueByDate   sql.NullString
	PaymentDate sql.NullString
	AmountMinor sql.NullInt64
	Label       sql.NullString
}

func QueryInvoiceHistory(
	ctx context.Context,
	db *sql.DB,
	clientID int64,
	baseNumber int64,
) ([]InvoiceHistoryRow, error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, `
		WITH target_invoice AS (
			SELECT id
			FROM invoices
			WHERE account_id = ? AND client_id = ? AND base_number = ?
		)
		SELECT
			entry_id,
			invoice_id,
			entry_type,
			created_at,
			revision_no,
			receipt_no,
			issue_date,
			due_by_date,
			payment_date,
			amount_minor,
			label
		FROM (
			SELECT
				r.id AS entry_id,
				r.invoice_id,
				'revision' AS entry_type,
				r.created_at,
				r.revision_no,
				NULL AS receipt_no,
				r.issue_date,
				r.due_by_date,
				NULL AS payment_date,
				NULL AS amount_minor,
				NULL AS label
			FROM invoice_revisions r
			JOIN target_invoice ti
				ON ti.id = r.invoice_id

			UNION ALL

			SELECT
				p.id AS entry_id,
				p.invoice_id,
				'payment_receipt' AS entry_type,
				p.created_at,
				NULL AS revision_no,
				p.receipt_no,
				NULL AS issue_date,
				NULL AS due_by_date,
				p.payment_date,
				p.amount_minor,
				p.label
			FROM payments p
			JOIN target_invoice ti
				ON ti.id = p.invoice_id
			WHERE p.payment_type = 'payment'
		)
		ORDER BY created_at ASC, entry_type ASC, entry_id ASC;
	`, accountID, clientID, baseNumber)
	if err != nil {
		return nil, fmt.Errorf("query invoice history: %w", err)
	}
	defer rows.Close()

	return scanInvoiceHistoryRows(rows)
}

func QueryInvoiceHistoryForInvoices(
	ctx context.Context,
	db *sql.DB,
	invoiceIDs []int64,
) ([]InvoiceHistoryRow, error) {
	if len(invoiceIDs) == 0 {
		return []InvoiceHistoryRow{}, nil
	}

	placeholders := make([]string, 0, len(invoiceIDs))
	args := make([]any, 0, len(invoiceIDs))
	for _, id := range invoiceIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	rows, err := db.QueryContext(ctx, fmt.Sprintf(`
		SELECT
			entry_id,
			invoice_id,
			entry_type,
			created_at,
			revision_no,
			receipt_no,
			issue_date,
			due_by_date,
			payment_date,
			amount_minor,
			label
		FROM (
			SELECT
				r.id AS entry_id,
				r.invoice_id,
				'revision' AS entry_type,
				r.created_at,
				r.revision_no,
				NULL AS receipt_no,
				r.issue_date,
				r.due_by_date,
				NULL AS payment_date,
				NULL AS amount_minor,
				NULL AS label
			FROM invoice_revisions r
			WHERE r.invoice_id IN (%s)

			UNION ALL

			SELECT
				p.id AS entry_id,
				p.invoice_id,
				'payment_receipt' AS entry_type,
				p.created_at,
				NULL AS revision_no,
				p.receipt_no,
				NULL AS issue_date,
				NULL AS due_by_date,
				p.payment_date,
				p.amount_minor,
				p.label
			FROM payments p
			WHERE p.invoice_id IN (%s)
			  AND p.payment_type = 'payment'
		)
		ORDER BY invoice_id DESC, created_at ASC, entry_type ASC, entry_id ASC;
	`, strings.Join(placeholders, ","), strings.Join(placeholders, ",")), append(args, args...)...)
	if err != nil {
		return nil, fmt.Errorf("query invoice history for page: %w", err)
	}
	defer rows.Close()

	return scanInvoiceHistoryRows(rows)
}

func scanInvoiceHistoryRows(rows *sql.Rows) ([]InvoiceHistoryRow, error) {
	items := make([]InvoiceHistoryRow, 0)
	for rows.Next() {
		var item InvoiceHistoryRow
		if err := rows.Scan(
			&item.ID,
			&item.InvoiceID,
			&item.Type,
			&item.CreatedAt,
			&item.RevisionNo,
			&item.ReceiptNo,
			&item.IssueDate,
			&item.DueByDate,
			&item.PaymentDate,
			&item.AmountMinor,
			&item.Label,
		); err != nil {
			return nil, fmt.Errorf("scan invoice history row: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate invoice history rows: %w", err)
	}

	return items, nil
}
