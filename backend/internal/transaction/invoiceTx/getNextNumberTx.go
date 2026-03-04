package invoiceTx

import (
	"context"

	"github.com/viktorHadz/goInvoice26/internal/app"
)

// Checks invoices base number for client returns current or null
func GetNextBaseNumber(ctx context.Context, a *app.App) (int64, error) {
	const sql = `
		SELECT COALESCE(MAX(base_number), 0)
		FROM invoices
	`

	var max int64
	if err := a.DB.QueryRowContext(ctx, sql).Scan(&max); err != nil {
		return 0, err
	}

	return max + 1, nil
}
