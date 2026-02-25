package clientsTx

import (
	"context"

	"github.com/viktorHadz/goInvoice26/internal/app"
)

func DeleteClient(a *app.App, ctx context.Context, id int64) (int64, error) {
	res, err := a.DB.ExecContext(ctx, `
		DELETE FROM clients 
		WHERE id = ?
	`, id)

	if err != nil {
		return 0, err
	}

	return res.RowsAffected()

}
