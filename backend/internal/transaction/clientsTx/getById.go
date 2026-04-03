// Package provides methods for retrieving clients and their details
package clientsTx

import (
	"context"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

// Checks if client with id exists in DB and returns all for client or an error
func GetByID(ctx context.Context, a *app.App, id int64) (models.Client, error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return models.Client{}, err
	}

	var c models.Client
	err = a.DB.QueryRowContext(ctx, `
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
		  AND account_id = ?
	`, id, accountID).Scan(&c.ID, &c.Name, &c.CompanyName, &c.Address, &c.Email, &c.CreatedAt, &c.UpdatedAt)

	return c, err
}

// Checks the DB for a client and returns a boolean if it exists or an error if it doesnt
func Exists(ctx context.Context, a *app.App, id int64) (bool, error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return false, err
	}

	var exists bool
	err = a.DB.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM clients WHERE id = ? AND account_id = ?)`,
		id,
		accountID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
