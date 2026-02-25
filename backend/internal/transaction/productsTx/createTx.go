package productsTx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

func InsertTx(a *app.App, ctx context.Context, in models.ProductCreate) (models.Product, error) {
	const q = `
		INSERT INTO products (
			product_type,
			pricing_mode,
			name,
			flat_price_minor,
			hourly_rate_minor,
			default_minutes_worked,
			client_id
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING
			id,
			product_type,
			pricing_mode,
			name,
			flat_price_minor,
			hourly_rate_minor,
			default_minutes_worked,
			client_id,
			created_at,
			updated_at;
	`

	var out models.Product

	// sql.Null* for nullable columns
	var flat sql.NullInt64
	var hourly sql.NullInt64
	var minutes sql.NullInt64
	var updated sql.NullString

	if in.FlatPriceMinor != nil {
		flat = sql.NullInt64{Int64: *in.FlatPriceMinor, Valid: true}
	}
	if in.HourlyRateMinor != nil {
		hourly = sql.NullInt64{Int64: *in.HourlyRateMinor, Valid: true}
	}
	if in.MinutesWorked != nil {
		minutes = sql.NullInt64{Int64: *in.MinutesWorked, Valid: true}
	}

	err := a.DB.QueryRowContext(
		ctx,
		q,
		in.ProductType,
		in.PricingMode,
		in.ProductName,
		flat,
		hourly,
		minutes,
		in.ClientID,
	).Scan(
		&out.ID,
		&out.ProductType,
		&out.PricingMode,
		&out.ProductName,
		&out.FlatPriceMinor,
		&out.HourlyRateMinor,
		&out.MinutesWorked,
		&out.ClientID,
		&out.CreatedAt,
		&updated,
	)
	if err != nil {
		return models.Product{}, fmt.Errorf("insert product: %w", err)
	}

	if updated.Valid {
		out.UpdatedAt = &updated.String
	}

	return out, nil
}
