package invoiceTx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var (
	// ErrInvoiceVoidForRevision is returned when appending a revision to a void invoice.
	ErrInvoiceVoidForRevision = errors.New("invoice is void; revisions are not allowed")
	// ErrInvoicePaidForRevision is returned when appending a revision while status is paid (reopen to issued first).
	ErrInvoicePaidForRevision = errors.New("invoice is paid; reopen to issued before revising")
	// ErrPaymentDeltaNegative is returned when desired paid total is less than recorded payments sum.
	ErrPaymentDeltaNegative = errors.New("paid total cannot be less than already recorded payments")
)

// appendPaymentDelta inserts payment rows so SUM(payments) matches desiredPaidMinor (append-only).
func appendPaymentDelta(ctx context.Context, tx *sql.Tx, invoiceID int64, desiredPaidMinor int64) error {
	var existing int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount_minor), 0) FROM payments WHERE invoice_id = ?
	`, invoiceID).Scan(&existing); err != nil {
		return fmt.Errorf("sum payments: %w", err)
	}

	delta := desiredPaidMinor - existing
	if delta < 0 {
		return ErrPaymentDeltaNegative
	}
	if delta == 0 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO payments (invoice_id, payment_type, amount_minor)
		VALUES (?, 'payment', ?)
	`, invoiceID, delta); err != nil {
		return fmt.Errorf("insert payment: %w", err)
	}
	return nil
}

// applyAutoPaidIfSettled sets status to paid when balance due is zero and invoice is draft or issued.
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
		WHERE id = ? AND status IN ('draft', 'issued')
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
	case "void":
		return ErrInvoiceVoidForRevision
	case "paid":
		return ErrInvoicePaidForRevision
	default:
		return nil
	}
}
