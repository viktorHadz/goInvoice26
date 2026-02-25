// Package provides methods for retrieving clients and their details
package clientsTx

import (
	"context"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

// Checks if client with id exists in DB and returns it or an error
func GetByID(ctx context.Context, a *app.App, id int64) (models.Client, error) {
	var c models.Client
	err := a.DB.QueryRowContext(ctx, `
		SELECT
			id,
			name,
			COALESCE(company_name, '') AS companyName,
			COALESCE(address, '')      AS address,
			COALESCE(email, '')        AS email,
			created_at,
			updated_at
		FROM clients
		WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &c.CompanyName, &c.Address, &c.Email, &c.CreatedAt, &c.UpdatedAt)

	return c, err
}

// Checks the DB for a client and returns a boolean if it exists or an error if it doesnt
func Exists(ctx context.Context, a *app.App, id int64) (bool, error) {
	var exists bool
	err := a.DB.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM clients WHERE id = ?)`,
		id,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
