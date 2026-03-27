package invoiceTx

import (
	"context"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/app"
)

// GetSuggestedNextBaseNumber returns the next base number to use for a new invoice
// without allocating or incrementing any sequence. It is safe to call on every
// page load/mount. The allocator sequence is the primary source of truth, with
// a safety clamp against existing invoice max for legacy consistency.
func GetSuggestedNextBaseNumber(ctx context.Context, a *app.App) (int64, error) {
	var next int64
	err := a.DB.QueryRowContext(ctx, `
		SELECT MAX(
			COALESCE((SELECT next_base_number FROM invoice_number_seq WHERE id = 1), 1),
			COALESCE((SELECT MAX(base_number) FROM invoices), 0) + 1
		)
	`).Scan(&next)
	if err != nil {
		return 0, fmt.Errorf("get suggested next base number: %w", err)
	}
	return next, nil
}
