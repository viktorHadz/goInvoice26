package invoiceTx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
)

var (
	// ErrInvoiceDraftForRevision is returned when attempting to append a revision to a draft invoice.
	ErrInvoiceDraftForRevision = errors.New("invoice is draft; issue it before saving revisions")
	// ErrInvoiceVoidForRevision is returned when appending a revision to a void invoice.
	ErrInvoiceVoidForRevision = errors.New("invoice is void; revisions are not allowed")
	// ErrInvoicePaidForRevision is returned when appending a revision while status is paid (reopen to issued first).
	ErrInvoicePaidForRevision = errors.New("invoice is paid; reopen to issued before revising")
	// ErrSourceRevisionInvalid is returned when source revision is outside valid range.
	ErrSourceRevisionInvalid = errors.New("source revision is invalid for this invoice")
	// ErrPaymentStateMismatch is returned when saved payment receipts changed while the invoice was being edited.
	ErrPaymentStateMismatch = errors.New("saved payment receipts changed; refresh invoice before saving")
)

func sumPaymentsByInvoice(ctx context.Context, tx *sql.Tx, invoiceID int64) (int64, error) {
	var existing int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount_minor), 0)
		FROM payments
		WHERE invoice_id = ?
		  AND payment_type = 'payment'
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
			AND p.payment_type = 'payment'
			AND ap.revision_no <= ?
	`, invoiceID, revisionNo).Scan(&existing); err != nil {
		return 0, fmt.Errorf("sum visible payments by revision: %w", err)
	}
	return existing, nil
}

func sumPaymentsByRevision(ctx context.Context, tx *sql.Tx, revisionID int64) (int64, error) {
	var existing int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount_minor), 0)
		FROM payments
		WHERE applied_in_revision_id = ?
		  AND payment_type = 'payment'
	`, revisionID).Scan(&existing); err != nil {
		return 0, fmt.Errorf("sum payments by revision: %w", err)
	}
	return existing, nil
}

// applyAutoPaidIfSettled syncs the invoice status against the current revision snapshot.
func applyAutoPaidIfSettled(ctx context.Context, tx *sql.Tx, invoiceID int64, _ int64) error {
	return syncInvoiceStatusForCurrentRevision(ctx, tx, invoiceID)
}

func syncInvoiceStatusForCurrentRevision(ctx context.Context, tx *sql.Tx, invoiceID int64) error {
	var (
		status            string
		currentRevisionID sql.NullInt64
		totalMinor        sql.NullInt64
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT
			i.status,
			i.current_revision_id,
			r.total_minor
		FROM invoices i
		LEFT JOIN invoice_revisions r
			ON r.id = i.current_revision_id
		WHERE i.id = ?;
	`, invoiceID).Scan(&status, &currentRevisionID, &totalMinor); err != nil {
		return fmt.Errorf("load invoice status sync state: %w", err)
	}

	if status == "void" || !currentRevisionID.Valid || !totalMinor.Valid {
		return nil
	}

	paidMinor, err := sumPaymentsByRevision(ctx, tx, currentRevisionID.Int64)
	if err != nil {
		return err
	}

	switch {
	case status == "issued" && paidMinor >= totalMinor.Int64:
		if _, err := tx.ExecContext(ctx, `
			UPDATE invoices
			SET status = 'paid'
			WHERE id = ?;
		`, invoiceID); err != nil {
			return fmt.Errorf("set invoice paid: %w", err)
		}
	case status == "paid" && paidMinor < totalMinor.Int64:
		if _, err := tx.ExecContext(ctx, `
			UPDATE invoices
			SET status = 'issued'
			WHERE id = ?;
		`, invoiceID); err != nil {
			return fmt.Errorf("reopen invoice to issued: %w", err)
		}
	}

	return nil
}

// LoadInvoiceIDAndStatus loads invoice id and status for client + base number.
func LoadInvoiceIDAndStatus(ctx context.Context, tx *sql.Tx, clientID, baseNumber int64) (invoiceID int64, status string, err error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return 0, "", err
	}

	err = tx.QueryRowContext(ctx, `
		SELECT id, status
		FROM invoices
		WHERE account_id = ? AND client_id = ? AND base_number = ?
	`, accountID, clientID, baseNumber).Scan(&invoiceID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, "", ErrInvoiceNotFound
	}
	if err != nil {
		return 0, "", fmt.Errorf("load invoice: %w", err)
	}
	return invoiceID, status, nil
}

func assertClientBelongsToAccount(ctx context.Context, tx *sql.Tx, accountID, clientID int64) error {
	var exists bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM clients
			WHERE id = ?
			  AND account_id = ?
		);
	`, clientID, accountID).Scan(&exists); err != nil {
		return fmt.Errorf("verify invoice client ownership: %w", err)
	}
	if !exists {
		return ErrInvoiceNotFound
	}

	return nil
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
