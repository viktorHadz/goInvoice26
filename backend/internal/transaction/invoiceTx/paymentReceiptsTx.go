package invoiceTx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

var (
	ErrInvoiceDraftForReceipt = errors.New("invoice revision not found")
	ErrInvoiceVoidForReceipt  = errors.New("invoice is void; payment receipts are not allowed")
	ErrInvoicePaidForReceipt  = errors.New("invoice revision is already fully paid")
	ErrPaymentReceiptNotFound = errors.New("payment receipt not found")
)

type paymentReceiptState struct {
	InvoiceID         int64
	InvoiceStatus     string
	CurrentRevisionID int64
	RevisionID        int64
	RevisionNo        int64
	TotalMinor        int64
	PaidMinor         int64
}

type PaymentReceiptRow struct {
	ID                int64
	InvoiceID         int64
	BaseNumber        int64
	ReceiptNo         int64
	PaymentDate       string
	AmountMinor       int64
	Label             sql.NullString
	AppliedRevisionID int64
	AppliedRevisionNo int64
}

func CreatePaymentReceipt(
	ctx context.Context,
	a *app.App,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
	canonical *models.PaymentReceiptCreateIn,
) (invoiceID, paymentID, receiptNo int64, err error) {
	tx, err := a.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	state, err := loadPaymentReceiptState(ctx, tx, clientID, baseNumber, revisionNo)
	if err != nil {
		return 0, 0, 0, err
	}
	if err := assertReceiptCreateAllowed(state.InvoiceStatus, state.TotalMinor, state.PaidMinor); err != nil {
		return 0, 0, 0, err
	}

	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(receipt_no), 0) + 1
		FROM payments
		WHERE applied_in_revision_id = ?;
	`, state.RevisionID).Scan(&receiptNo); err != nil {
		return 0, 0, 0, fmt.Errorf("next receipt number: %w", err)
	}

	var label any
	if canonical.Label != nil {
		label = *canonical.Label
	}

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO payments (
			invoice_id,
			receipt_no,
			payment_type,
			amount_minor,
			payment_date,
			applied_in_revision_id,
			label
		)
		VALUES (?, ?, 'payment', ?, ?, ?, ?)
		RETURNING id;
	`, state.InvoiceID, receiptNo, canonical.AmountMinor, canonical.PaymentDate, state.RevisionID, label).Scan(&paymentID); err != nil {
		return 0, 0, 0, fmt.Errorf("insert payment receipt: %w", err)
	}

	if err := syncInvoiceStatusForCurrentRevision(ctx, tx, state.InvoiceID); err != nil {
		return 0, 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, 0, fmt.Errorf("commit payment receipt: %w", err)
	}

	return state.InvoiceID, paymentID, receiptNo, nil
}

func UpdatePaymentReceiptMetadata(
	ctx context.Context,
	a *app.App,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
	receiptNo int64,
	canonical *models.PaymentReceiptUpdateIn,
) (paymentID int64, err error) {
	tx, err := a.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	state, err := loadPaymentReceiptState(ctx, tx, clientID, baseNumber, revisionNo)
	if err != nil {
		return 0, err
	}
	if err := assertReceiptMutationAllowed(state.InvoiceStatus); err != nil {
		return 0, err
	}

	var label any
	if canonical.Label != nil {
		label = *canonical.Label
	}

	err = tx.QueryRowContext(ctx, `
		UPDATE payments
		SET payment_date = ?, label = ?
		WHERE applied_in_revision_id = ?
		  AND receipt_no = ?
		  AND payment_type = 'payment'
		RETURNING id;
	`, canonical.PaymentDate, label, state.RevisionID, receiptNo).Scan(&paymentID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrPaymentReceiptNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("update payment receipt metadata: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit payment receipt metadata: %w", err)
	}

	return paymentID, nil
}

func DeletePaymentReceipt(
	ctx context.Context,
	a *app.App,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
	receiptNo int64,
) (paymentID int64, err error) {
	tx, err := a.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	state, err := loadPaymentReceiptState(ctx, tx, clientID, baseNumber, revisionNo)
	if err != nil {
		return 0, err
	}
	if err := assertReceiptMutationAllowed(state.InvoiceStatus); err != nil {
		return 0, err
	}

	err = tx.QueryRowContext(ctx, `
		DELETE FROM payments
		WHERE applied_in_revision_id = ?
		  AND receipt_no = ?
		  AND payment_type = 'payment'
		RETURNING id;
	`, state.RevisionID, receiptNo).Scan(&paymentID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrPaymentReceiptNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("delete payment receipt: %w", err)
	}

	if err := syncInvoiceStatusForCurrentRevision(ctx, tx, state.InvoiceID); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit payment receipt delete: %w", err)
	}

	return paymentID, nil
}

func QueryPaymentReceiptByNumber(
	ctx context.Context,
	db *sql.DB,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
	receiptNo int64,
) (*PaymentReceiptRow, error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return nil, err
	}

	var out PaymentReceiptRow
	err = db.QueryRowContext(ctx, `
		SELECT
			p.id,
			i.id,
			i.base_number,
			p.receipt_no,
			p.payment_date,
			p.amount_minor,
			p.label,
			p.applied_in_revision_id,
			r.revision_no
		FROM invoices i
		JOIN invoice_revisions r
			ON r.invoice_id = i.id
		   AND r.revision_no = ?
		JOIN payments p
			ON p.applied_in_revision_id = r.id
		WHERE i.account_id = ?
		  AND i.client_id = ?
		  AND i.base_number = ?
		  AND p.receipt_no = ?
		  AND p.payment_type = 'payment'
	`, revisionNo, accountID, clientID, baseNumber, receiptNo).Scan(
		&out.ID,
		&out.InvoiceID,
		&out.BaseNumber,
		&out.ReceiptNo,
		&out.PaymentDate,
		&out.AmountMinor,
		&out.Label,
		&out.AppliedRevisionID,
		&out.AppliedRevisionNo,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPaymentReceiptNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("query payment receipt: %w", err)
	}

	return &out, nil
}

func loadPaymentReceiptState(
	ctx context.Context,
	tx *sql.Tx,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
) (paymentReceiptState, error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return paymentReceiptState{}, err
	}

	var state paymentReceiptState
	err = tx.QueryRowContext(ctx, `
		SELECT
			i.id,
			i.status,
			i.current_revision_id,
			r.id,
			r.revision_no,
			r.total_minor,
			COALESCE(SUM(p.amount_minor), 0) AS paid_minor
		FROM invoices i
		JOIN invoice_revisions r
			ON r.invoice_id = i.id
		   AND r.revision_no = ?
		LEFT JOIN payments p
			ON p.applied_in_revision_id = r.id
		   AND p.payment_type = 'payment'
		WHERE i.account_id = ?
		  AND i.client_id = ?
		  AND i.base_number = ?
		GROUP BY i.id, i.status, i.current_revision_id, r.id, r.revision_no, r.total_minor;
	`, revisionNo, accountID, clientID, baseNumber).Scan(
		&state.InvoiceID,
		&state.InvoiceStatus,
		&state.CurrentRevisionID,
		&state.RevisionID,
		&state.RevisionNo,
		&state.TotalMinor,
		&state.PaidMinor,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return paymentReceiptState{}, ErrInvoiceNotFound
	}
	if err != nil {
		return paymentReceiptState{}, fmt.Errorf("load invoice receipt state: %w", err)
	}

	return state, nil
}

func assertReceiptCreateAllowed(status string, totalMinor int64, paidMinor int64) error {
	if status == "void" {
		return ErrInvoiceVoidForReceipt
	}

	if paidMinor >= totalMinor && totalMinor > 0 {
		return ErrInvoicePaidForReceipt
	}

	return nil
}

func assertReceiptMutationAllowed(status string) error {
	if status == "void" {
		return ErrInvoiceVoidForReceipt
	}

	return nil
}

func cloneReceiptSnapshot(
	ctx context.Context,
	tx *sql.Tx,
	invoiceID int64,
	sourceRevisionID int64,
	targetRevisionID int64,
) error {
	if sourceRevisionID < 1 || targetRevisionID < 1 {
		return nil
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO payments (
			invoice_id,
			receipt_no,
			payment_type,
			amount_minor,
			payment_date,
			applied_in_revision_id,
			label
		)
		SELECT
			?,
			p.receipt_no,
			p.payment_type,
			p.amount_minor,
			p.payment_date,
			?,
			p.label
		FROM payments p
		WHERE p.applied_in_revision_id = ?
		  AND p.payment_type = 'payment'
		ORDER BY p.receipt_no ASC, p.id ASC;
	`, invoiceID, targetRevisionID, sourceRevisionID); err != nil {
		return fmt.Errorf("clone payment receipt snapshot: %w", err)
	}

	return nil
}
