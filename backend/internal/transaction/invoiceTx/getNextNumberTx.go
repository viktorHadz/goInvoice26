package invoiceTx

import (
	"context"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
)

// GetSuggestedNextBaseNumber returns the next base number to use for a new invoice
// without allocating or incrementing any sequence. It is safe to call on every
// page load/mount. The allocator sequence is the primary source of truth, with
// a safety clamp against existing invoice max for legacy consistency.
func GetSuggestedNextBaseNumber(ctx context.Context, a *app.App) (int64, error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return 0, err
	}

	if _, err := a.DB.ExecContext(ctx, `
		INSERT OR IGNORE INTO invoice_number_seq (account_id, next_base_number)
		VALUES (?, 1);
	`, accountID); err != nil {
		return 0, fmt.Errorf("ensure invoice sequence row: %w", err)
	}

	var next int64
	err = a.DB.QueryRowContext(ctx, `
		SELECT MAX(
			COALESCE((SELECT next_base_number FROM invoice_number_seq WHERE account_id = ?), 1),
			COALESCE((SELECT MAX(base_number) FROM invoices WHERE account_id = ?), 0) + 1
		)
	`, accountID, accountID).Scan(&next)
	if err != nil {
		return 0, fmt.Errorf("get suggested next base number: %w", err)
	}
	return next, nil
}
