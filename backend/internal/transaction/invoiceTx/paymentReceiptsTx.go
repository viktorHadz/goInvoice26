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
	ErrInvoiceDraftForReceipt = errors.New("invoice is draft; issue it before recording payment receipts")
	ErrInvoiceVoidForReceipt  = errors.New("invoice is void; payment receipts are not allowed")
	ErrInvoicePaidForReceipt  = errors.New("invoice is already paid; no further payment receipts can be recorded")
	ErrPaymentReceiptNotFound = errors.New("payment receipt not found")
)

type paymentReceiptState struct {
	InvoiceID         int64
	Status            string
	CurrentRevisionID int64
	TotalMinor        int64
}

type PaymentReceiptRow struct {
	ID                int64
	InvoiceID         int64
	BaseNumber        int64
	ReceiptNo         int64
	PaymentDate       string
	AmountMinor       int64
	Label             sql.NullString
	CreatedAt         string
	AppliedRevisionID int64
	AppliedRevisionNo int64
}

func CreatePaymentReceipt(
	ctx context.Context,
	a *app.App,
	clientID int64,
	baseNumber int64,
	canonical *models.PaymentReceiptCreateIn,
) (invoiceID, paymentID, receiptNo int64, err error) {
	tx, err := a.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	state, err := loadPaymentReceiptState(ctx, tx, clientID, baseNumber)
	if err != nil {
		return 0, 0, 0, err
	}
	if err := assertReceiptCreateAllowed(state.Status); err != nil {
		return 0, 0, 0, err
	}

	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(receipt_no), 0) + 1
		FROM payments
		WHERE invoice_id = ?;
	`, state.InvoiceID).Scan(&receiptNo); err != nil {
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
	`, state.InvoiceID, receiptNo, canonical.AmountMinor, canonical.PaymentDate, state.CurrentRevisionID, label).Scan(&paymentID); err != nil {
		return 0, 0, 0, fmt.Errorf("insert payment receipt: %w", err)
	}

	if err := applyAutoPaidIfSettled(ctx, tx, state.InvoiceID, state.TotalMinor); err != nil {
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
	receiptNo int64,
	canonical *models.PaymentReceiptUpdateIn,
) (paymentID int64, err error) {
	tx, err := a.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	state, err := loadPaymentReceiptState(ctx, tx, clientID, baseNumber)
	if err != nil {
		return 0, err
	}
	if err := assertReceiptMetadataUpdateAllowed(state.Status); err != nil {
		return 0, err
	}

	var label any
	if canonical.Label != nil {
		label = *canonical.Label
	}

	err = tx.QueryRowContext(ctx, `
		UPDATE payments
		SET payment_date = ?, label = ?
		WHERE invoice_id = ?
		  AND receipt_no = ?
		  AND payment_type = 'payment'
		RETURNING id;
	`, canonical.PaymentDate, label, state.InvoiceID, receiptNo).Scan(&paymentID)
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

func QueryPaymentReceiptByNumber(
	ctx context.Context,
	db *sql.DB,
	clientID int64,
	baseNumber int64,
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
			p.created_at,
			p.applied_in_revision_id,
			ap.revision_no
		FROM invoices i
		JOIN payments p
			ON p.invoice_id = i.id
		JOIN invoice_revisions ap
			ON ap.id = p.applied_in_revision_id
		WHERE i.account_id = ?
		  AND i.client_id = ?
		  AND i.base_number = ?
		  AND p.receipt_no = ?
		  AND p.payment_type = 'payment'
	`, accountID, clientID, baseNumber, receiptNo).Scan(
		&out.ID,
		&out.InvoiceID,
		&out.BaseNumber,
		&out.ReceiptNo,
		&out.PaymentDate,
		&out.AmountMinor,
		&out.Label,
		&out.CreatedAt,
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
			cur.total_minor
		FROM invoices i
		JOIN invoice_revisions cur
			ON cur.id = i.current_revision_id
		WHERE i.account_id = ?
		  AND i.client_id = ?
		  AND i.base_number = ?;
	`, accountID, clientID, baseNumber).Scan(
		&state.InvoiceID,
		&state.Status,
		&state.CurrentRevisionID,
		&state.TotalMinor,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return paymentReceiptState{}, ErrInvoiceNotFound
	}
	if err != nil {
		return paymentReceiptState{}, fmt.Errorf("load invoice receipt state: %w", err)
	}

	return state, nil
}

func assertReceiptCreateAllowed(status string) error {
	switch status {
	case "draft":
		return ErrInvoiceDraftForReceipt
	case "void":
		return ErrInvoiceVoidForReceipt
	case "paid":
		return ErrInvoicePaidForReceipt
	case "issued":
		return nil
	default:
		return fmt.Errorf("unexpected invoice status: %s", status)
	}
}

func assertReceiptMetadataUpdateAllowed(status string) error {
	switch status {
	case "draft":
		return ErrInvoiceDraftForReceipt
	case "void":
		return ErrInvoiceVoidForReceipt
	case "issued", "paid":
		return nil
	default:
		return fmt.Errorf("unexpected invoice status: %s", status)
	}
}
