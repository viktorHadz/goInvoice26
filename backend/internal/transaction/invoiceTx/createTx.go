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
	"errors"
	"fmt"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

// Create inserts a new invoice with revision 1 and all line items.
func Create(ctx context.Context, a *app.App, canonical *models.FEInvoiceIn) (invoiceID, revisionID int64, err error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return 0, 0, err
	}

	tx, err := a.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, err
	}
	defer tx.Rollback()

	ov := &canonical.Overview

	if err := assertClientBelongsToAccount(ctx, tx, accountID, ov.ClientID); err != nil {
		return 0, 0, err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT OR IGNORE INTO invoice_number_seq (account_id, next_base_number)
		VALUES (?, 1);
	`, accountID); err != nil {
		return 0, 0, fmt.Errorf("ensure invoice sequence row: %w", err)
	}

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO invoices (account_id, client_id, base_number, status)
		VALUES (?, ?, ?, 'draft')
		RETURNING id;
	`, accountID, ov.ClientID, ov.BaseNumber).Scan(&invoiceID); err != nil {
		if isUniqueViolation(err) {
			return 0, 0, fmt.Errorf("invoice base_number %d already exists: %w", ov.BaseNumber, err)
		}
		return 0, 0, fmt.Errorf("insert invoice: %w", err)
	}

	revisionID, err = insertRevisionWithItems(ctx, tx, invoiceID, 1, canonical)
	if err != nil {
		return 0, 0, err
	}

	if err := applyAutoPaidIfSettled(ctx, tx, invoiceID, canonical.Totals.TotalMinor); err != nil {
		return 0, 0, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE invoice_number_seq
		SET next_base_number = MAX(next_base_number, ?)
		WHERE account_id = ?;
	`, ov.BaseNumber+1, accountID); err != nil {
		return 0, 0, fmt.Errorf("sync invoice_number_seq: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("commit: %w", err)
	}
	return invoiceID, revisionID, nil
}

var ErrInvoiceNotFound = errors.New("invoice not found")

// CreateRevision appends a new latest revision to an existing invoice.
// It never mutates older revisions in place.
func CreateRevision(ctx context.Context, a *app.App, canonical *models.FEInvoiceIn) (invoiceID, revisionID, revisionNo int64, err error) {
	tx, err := a.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, 0, 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	ov := &canonical.Overview
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return 0, 0, 0, err
	}

	if err := assertClientBelongsToAccount(ctx, tx, accountID, ov.ClientID); err != nil {
		return 0, 0, 0, err
	}

	invoiceID, invStatus, err := LoadInvoiceIDAndStatus(ctx, tx, ov.ClientID, ov.BaseNumber)
	if err != nil {
		if errors.Is(err, ErrInvoiceNotFound) {
			return 0, 0, 0, ErrInvoiceNotFound
		}
		return 0, 0, 0, err
	}
	if err := assertRevisionAllowed(invStatus); err != nil {
		return 0, 0, 0, err
	}

	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(revision_no), 0) + 1
		FROM invoice_revisions
		WHERE invoice_id = ?;
	`, invoiceID).Scan(&revisionNo); err != nil {
		return 0, 0, 0, fmt.Errorf("get next revision number: %w", err)
	}

	revisionID, err = insertRevisionWithItems(ctx, tx, invoiceID, revisionNo, canonical)
	if err != nil {
		return 0, 0, 0, err
	}

	sourceRevisionNo := revisionNo - 1
	if canonical.Overview.SourceRevisionNo != nil {
		sourceRevisionNo = *canonical.Overview.SourceRevisionNo
	}
	if sourceRevisionNo < 1 || sourceRevisionNo > revisionNo-1 {
		return 0, 0, 0, ErrSourceRevisionInvalid
	}

	if err := applyAutoPaidIfSettled(ctx, tx, invoiceID, canonical.Totals.TotalMinor); err != nil {
		return 0, 0, 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, 0, fmt.Errorf("commit: %w", err)
	}

	return invoiceID, revisionID, revisionNo, nil
}

// isUniqueViolation returns true if the error is a SQLite unique constraint violation.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE") || strings.Contains(msg, "unique")
}
