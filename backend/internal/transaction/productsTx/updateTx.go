package productsTx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

func UpdateTx(a *app.App, ctx context.Context, productID int64, in models.ProductCreate) (models.Product, error) {
	const q = `
		UPDATE products
		SET
			product_type = ?,
			pricing_mode = ?,
			name = ?,
			flat_price_minor = ?,
			hourly_rate_minor = ?,
			default_minutes_worked = ?,
			updated_at = (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		WHERE id = ? AND client_id = ?
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

	// nullable inputs
	var flat sql.NullInt64
	var hourly sql.NullInt64
	var minutes sql.NullInt64
	if in.FlatPriceMinor != nil {
		flat = sql.NullInt64{Int64: *in.FlatPriceMinor, Valid: true}
	}
	if in.HourlyRateMinor != nil {
		hourly = sql.NullInt64{Int64: *in.HourlyRateMinor, Valid: true}
	}
	if in.MinutesWorked != nil {
		minutes = sql.NullInt64{Int64: *in.MinutesWorked, Valid: true}
	}

	// updated_at nullable output
	var updated sql.NullString

	err := a.DB.QueryRowContext(
		ctx,
		q,
		in.ProductType,
		in.PricingMode,
		in.ProductName,
		flat,
		hourly,
		minutes,
		productID,
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
		// sql.ErrNoRows - not found for that client
		return models.Product{}, fmt.Errorf("update product: %w", err)
	}

	if updated.Valid {
		out.UpdatedAt = &updated.String
	}

	return out, nil
}
