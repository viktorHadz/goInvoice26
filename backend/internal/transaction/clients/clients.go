package clients

import (
	"context"

	"github.com/viktorHadz/goInvoice26/internal/app"
)

// Performs database transaction inserting a new client
func Insert(ctx context.Context, a *app.App, c *app.CreateClient) (int64, error) {
	res, err := a.DB.ExecContext(ctx, `
    INSERT INTO clients (name, company_name, address, email)
    VALUES (?, ?, ?, ?)
  `, c.Name, c.CompanyName, c.Address, c.Email)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}
