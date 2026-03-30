package clientsTx

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mattn/go-sqlite3"
	"github.com/viktorHadz/goInvoice26/internal/app"
)

var ErrClientHasInvoices = errors.New("client has saved invoices")

func DeleteClient(a *app.App, ctx context.Context, id int64) (int64, error) {
	res, err := a.DB.ExecContext(ctx, `
		DELETE FROM clients 
		WHERE id = ?
	`, id)

	if err != nil {
		if isForeignKeyConstraint(err) {
			return 0, fmt.Errorf("%w: %v", ErrClientHasInvoices, err)
		}
		return 0, err
	}

	return res.RowsAffected()

}

func isForeignKeyConstraint(err error) bool {
	var sqliteErr sqlite3.Error

	if errors.As(err, &sqliteErr) &&
		sqliteErr.Code == sqlite3.ErrConstraint &&
		sqliteErr.ExtendedCode == sqlite3.ErrConstraintForeignKey {
		return true
	}

	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "foreign key constraint failed")
}
