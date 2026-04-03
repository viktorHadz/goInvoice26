package clientsTx

import (
	"context"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

func ListClients(a *app.App, ctx context.Context) ([]models.Client, error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := a.DB.QueryContext(ctx, `
		SELECT
			id,
			name,
			COALESCE(company_name, '') AS companyName,
			COALESCE(address, '')      AS address,
			COALESCE(email, '')        AS email,
			created_at,
			updated_at
		FROM clients
		WHERE account_id = ?
		ORDER BY id DESC
	`, accountID)
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
