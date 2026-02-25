package productsTx

import (
	"context"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

func ListAll(a *app.App, ctx context.Context, clientID int64) ([]models.Product, error) {
	rows, err := a.DB.QueryContext(ctx, `
		SELECT
			id,
			product_type,
			pricing_mode,
			name,
			flat_price_minor,
			hourly_rate_minor,
			default_minutes_worked,
			client_id,
			created_at,
			updated_at
		FROM products
		WHERE client_id = ?
`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// rows scan
	var out []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(
			&p.ID,
			&p.ProductType,
			&p.PricingMode,
			&p.ProductName,
			&p.FlatPriceMinor,
			&p.HourlyRateMinor,
			&p.MinutesWorked,
			&p.ClientID,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, p)
	}

	return out, nil
}
