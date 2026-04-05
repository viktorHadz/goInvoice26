package productsTx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

func BulkInsertTx(a *app.App, ctx context.Context, rows []models.ProductCreate) (int, error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}

	tx, err := a.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin product import tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO products (
			account_id,
			product_type,
			pricing_mode,
			name,
			flat_price_minor,
			hourly_rate_minor,
			default_minutes_worked,
			client_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return 0, fmt.Errorf("prepare product import insert: %w", err)
	}
	defer stmt.Close()

	for _, row := range rows {
		var flat sql.NullInt64
		var hourly sql.NullInt64
		var minutes sql.NullInt64

		if row.FlatPriceMinor != nil {
			flat = sql.NullInt64{Int64: *row.FlatPriceMinor, Valid: true}
		}
		if row.HourlyRateMinor != nil {
			hourly = sql.NullInt64{Int64: *row.HourlyRateMinor, Valid: true}
		}
		if row.MinutesWorked != nil {
			minutes = sql.NullInt64{Int64: *row.MinutesWorked, Valid: true}
		}

		if _, err := stmt.ExecContext(
			ctx,
			accountID,
			row.ProductType,
			row.PricingMode,
			row.ProductName,
			flat,
			hourly,
			minutes,
			row.ClientID,
		); err != nil {
			return 0, fmt.Errorf("insert imported product: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit product import tx: %w", err)
	}

	return len(rows), nil
}
