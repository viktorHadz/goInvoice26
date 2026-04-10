package db

import (
	"context"
	"database/sql"
	"fmt"
)

func ensureClientsAccountIDColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "clients", "account_id")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE clients
			ADD COLUMN account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id);
		`); err != nil {
			return fmt.Errorf("add clients.account_id: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE clients
		SET account_id = 1
		WHERE account_id IS NULL OR account_id <= 0;
	`); err != nil {
		return fmt.Errorf("backfill clients.account_id: %w", err)
	}

	return nil
}

func ensureProductsAccountIDColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "products", "account_id")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE products
			ADD COLUMN account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id);
		`); err != nil {
			return fmt.Errorf("add products.account_id: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE products
		SET account_id = COALESCE(
			(
				SELECT c.account_id
				FROM clients c
				WHERE c.id = products.client_id
			),
			1
		)
		WHERE account_id IS NULL
		   OR account_id <= 0
		   OR account_id <> COALESCE(
				(
					SELECT c.account_id
					FROM clients c
					WHERE c.id = products.client_id
				),
				1
		   );
	`); err != nil {
		return fmt.Errorf("backfill products.account_id: %w", err)
	}

	return nil
}

func ensureInvoiceSupplyDateColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "invoice_revisions", "supply_date")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE invoice_revisions
			ADD COLUMN supply_date TEXT;
		`); err != nil {
			return fmt.Errorf("add invoice_revisions.supply_date: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE invoice_revisions
		SET supply_date = NULL
		WHERE TRIM(COALESCE(supply_date, '')) = ''
		   OR supply_date = issue_date;
	`); err != nil {
		return fmt.Errorf("normalize invoice_revisions.supply_date: %w", err)
	}

	return nil
}

func ensurePaymentReceiptNumberColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "payments", "receipt_no")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE payments
			ADD COLUMN receipt_no INTEGER NOT NULL DEFAULT 0;
		`); err != nil {
			return fmt.Errorf("add payments.receipt_no: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE payments
		SET receipt_no = (
			SELECT COUNT(*)
			FROM payments p2
			WHERE p2.invoice_id = payments.invoice_id
			  AND (
				p2.created_at < payments.created_at
				OR (p2.created_at = payments.created_at AND p2.id <= payments.id)
			  )
		)
		WHERE receipt_no IS NULL OR receipt_no <= 0;
	`); err != nil {
		return fmt.Errorf("backfill payments.receipt_no: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_invoice_receipt_no
		ON payments(invoice_id, receipt_no);
	`); err != nil {
		return fmt.Errorf("ensure payments receipt number index: %w", err)
	}

	return nil
}

func reconcileInvoiceStatusesToSavedPayments(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		WITH payment_totals AS (
			SELECT
				invoice_id,
				COALESCE(SUM(amount_minor), 0) AS paid_minor
			FROM payments
			WHERE payment_type = 'payment'
			GROUP BY invoice_id
		)
		UPDATE invoices
		SET status = CASE
			WHEN status = 'paid' AND COALESCE((SELECT paid_minor FROM payment_totals pt WHERE pt.invoice_id = invoices.id), 0) < COALESCE((
				SELECT total_minor
				FROM invoice_revisions r
				WHERE r.id = invoices.current_revision_id
			), 0) THEN 'issued'
			WHEN status = 'issued' AND COALESCE((SELECT paid_minor FROM payment_totals pt WHERE pt.invoice_id = invoices.id), 0) >= COALESCE((
				SELECT total_minor
				FROM invoice_revisions r
				WHERE r.id = invoices.current_revision_id
			), 0) THEN 'paid'
			ELSE status
		END
		WHERE status IN ('issued', 'paid');
	`); err != nil {
		return fmt.Errorf("reconcile invoices.status against saved payments: %w", err)
	}

	return nil
}
