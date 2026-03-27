package invoiceTx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/models"
)

var (
	// ErrInvoiceDraftForRevision is returned when attempting to append a revision to a draft invoice.
	ErrInvoiceDraftForRevision = errors.New("invoice is draft; issue it before saving revisions")
	// ErrInvoiceVoidForRevision is returned when appending a revision to a void invoice.
	ErrInvoiceVoidForRevision = errors.New("invoice is void; revisions are not allowed")
	// ErrInvoicePaidForRevision is returned when appending a revision while status is paid (reopen to issued first).
	ErrInvoicePaidForRevision = errors.New("invoice is paid; reopen to issued before revising")
	// ErrPaymentTotalsMismatch is returned when totals.paidMinor does not match visible + staged payments.
	ErrPaymentTotalsMismatch = errors.New("paid total does not match staged payments for this revision")
	// ErrSourceRevisionInvalid is returned when source revision is outside valid range.
	ErrSourceRevisionInvalid = errors.New("source revision is invalid for this invoice")
)

func sumPaymentsByInvoice(ctx context.Context, tx *sql.Tx, invoiceID int64) (int64, error) {
	var existing int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount_minor), 0) FROM payments WHERE invoice_id = ?
	`, invoiceID).Scan(&existing); err != nil {
		return 0, fmt.Errorf("sum payments: %w", err)
	}
	return existing, nil
}

func sumPaymentsVisibleUpToRevision(
	ctx context.Context,
	tx *sql.Tx,
	invoiceID int64,
	revisionNo int64,
) (int64, error) {
	var existing int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(p.amount_minor), 0)
		FROM payments p
		JOIN invoice_revisions ap ON ap.id = p.applied_in_revision_id
		WHERE p.invoice_id = ?
			AND ap.revision_no <= ?
	`, invoiceID, revisionNo).Scan(&existing); err != nil {
		return 0, fmt.Errorf("sum visible payments by revision: %w", err)
	}
	return existing, nil
}

func sumStagedPayments(payments []models.PaymentCreateIn) int64 {
	var out int64
	for _, p := range payments {
		out += p.AmountMinor
	}
	return out
}

func insertRevisionPayments(
	ctx context.Context,
	tx *sql.Tx,
	invoiceID int64,
	revisionID int64,
	payments []models.PaymentCreateIn,
) error {
	if len(payments) == 0 {
		return nil
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO payments (invoice_id, payment_type, amount_minor, payment_date, applied_in_revision_id, label)
		VALUES (?, 'payment', ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare payments insert: %w", err)
	}
	defer stmt.Close()

	for _, p := range payments {
		var label any
		if p.Label != nil {
			label = *p.Label
		}
		if _, err := stmt.ExecContext(
			ctx,
			invoiceID,
			p.AmountMinor,
			p.PaymentDate,
			revisionID,
			label,
		); err != nil {
			return fmt.Errorf("insert payment row: %w", err)
		}
	}

	return nil
}

// applyAutoPaidIfSettled sets status to paid when balance due is zero and invoice is issued.
func applyAutoPaidIfSettled(ctx context.Context, tx *sql.Tx, invoiceID int64, totalMinor, depositMinor int64) error {
	var paidSum int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount_minor), 0) FROM payments WHERE invoice_id = ?
	`, invoiceID).Scan(&paidSum); err != nil {
		return fmt.Errorf("sum payments for auto-paid: %w", err)
	}

	balance := totalMinor - depositMinor - paidSum
	if balance < 0 {
		balance = 0
	}
	if balance > 0 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE invoices
		SET status = 'paid'
		WHERE id = ? AND status = 'issued'
	`, invoiceID); err != nil {
		return fmt.Errorf("auto-paid status: %w", err)
	}
	return nil
}

// LoadInvoiceIDAndStatus loads invoice id and status for client + base number.
func LoadInvoiceIDAndStatus(ctx context.Context, tx *sql.Tx, clientID, baseNumber int64) (invoiceID int64, status string, err error) {
	err = tx.QueryRowContext(ctx, `
		SELECT id, status FROM invoices WHERE client_id = ? AND base_number = ?
	`, clientID, baseNumber).Scan(&invoiceID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", ErrInvoiceNotFound
	}
	if err != nil {
		return 0, "", fmt.Errorf("load invoice: %w", err)
	}
	return invoiceID, status, nil
}

func assertRevisionAllowed(status string) error {
	switch status {
	case "draft":
		return ErrInvoiceDraftForRevision
	case "void":
		return ErrInvoiceVoidForRevision
	case "paid":
		return ErrInvoicePaidForRevision
	default:
		return nil
	}
}
