package clientsTx

import (
	"context"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

// Creates a new client
func Insert(ctx context.Context, a *app.App, c *models.CreateClient) (int64, error) {
	res, err := a.DB.ExecContext(ctx, `
    INSERT INTO clients (name, company_name, address, email)
    VALUES (?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''))
  `, c.Name, c.CompanyName, c.Address, c.Email)

	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}
