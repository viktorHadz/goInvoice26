package productsTx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/app"
)

func DeleteTx(a *app.App, ctx context.Context, productID, clientID int64) error {
	res, err := a.DB.ExecContext(ctx, `
		DELETE FROM products
		WHERE id = ? AND client_id = ?
	`, productID, clientID)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete product rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
