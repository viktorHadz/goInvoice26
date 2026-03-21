package invoiceTx

import (
	"context"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/app"
)

// GetSuggestedNextBaseNumber returns the next base number to use for a new invoice
// without allocating or incrementing any sequence. It is safe to call on every
// page load/mount. Returns COALESCE(MAX(base_number), 0) + 1 from invoices.
// Use this for the "new invoice" form; allocate only on successful create.
func GetSuggestedNextBaseNumber(ctx context.Context, a *app.App) (int64, error) {
	var next int64
	err := a.DB.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(base_number), 0) + 1 FROM invoices
	`).Scan(&next)
	if err != nil {
		return 0, fmt.Errorf("get suggested next base number: %w", err)
	}
	return next, nil
}
