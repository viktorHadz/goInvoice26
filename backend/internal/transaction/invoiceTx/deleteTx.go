package invoiceTx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/app"
)

var (
	ErrInvoiceDeleteVoid = errors.New("void invoices cannot be deleted")
)

func Delete(ctx context.Context, a *app.App, clientID, baseNumber int64) error {
	tx, err := a.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	invoiceID, status, err := LoadInvoiceIDAndStatus(ctx, tx, clientID, baseNumber)
	if err != nil {
		return err
	}

	switch status {
	case "void":
		return ErrInvoiceDeleteVoid
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM invoices
		WHERE id = ?
	`, invoiceID)
	if err != nil {
		return fmt.Errorf("delete invoice: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete invoice affected rows: %w", err)
	}
	if affected == 0 {
		return ErrInvoiceNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete invoice: %w", err)
	}

	return nil
}
