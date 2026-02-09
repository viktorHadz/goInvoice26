package clients

import (
	"context"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

func ListClients(a *app.App, ctx context.Context) ([]models.Client, error) {
	rows, err := a.DB.QueryContext(ctx, `
		SELECT id, name, company_name, address, email, created_at, updated_at
		FROM clients
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.Client
	for rows.Next() {
		var c models.Client
		if err := rows.Scan(
			&c.ID, &c.Name, &c.CompanyName, &c.Address, &c.Email, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
