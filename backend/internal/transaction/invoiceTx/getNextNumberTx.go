package invoiceTx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/app"
)

// Allocates the next global invoice base number in a concurrency-safe way.
//
// Implementation notes (SQLite):
// - We keep a single-row sequence table (`invoice_number_seq`).
// - Allocation is performed inside a transaction with an atomic UPDATE.
func GetNextBaseNumber(ctx context.Context, a *app.App) (int64, error) {
	tx, err := a.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Ensure the sequence never lags behind existing invoices.
	// If this is the first allocation after importing data, this bumps the
	// starting point to MAX(base_number)+1.
	if _, err := tx.ExecContext(ctx, `
		UPDATE invoice_number_seq
		SET next_base_number = MAX(
			next_base_number,
			(SELECT COALESCE(MAX(base_number), 0) + 1 FROM invoices)
		)
		WHERE id = 1;
	`); err != nil {
		return 0, err
	}

	var allocated int64
	// Allocate current and increment for next caller.
	// Requires SQLite with RETURNING support.
	if err := tx.QueryRowContext(ctx, `
		UPDATE invoice_number_seq
		SET next_base_number = next_base_number + 1
		WHERE id = 1
		RETURNING next_base_number - 1;
	`).Scan(&allocated); err != nil {
		return 0, fmt.Errorf("allocate next_base_number: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return allocated, nil
}
