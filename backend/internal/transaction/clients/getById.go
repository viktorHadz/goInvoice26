package clients

import (
	"context"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

// Check if client with id exists in DB and return it
func GetByID(ctx context.Context, a *app.App, id int64) (models.Client, error) {
	var c models.Client
	err := a.DB.QueryRowContext(ctx, `
		SELECT id, name, company_name, address, email, created_at, updated_at
		FROM clients
		WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &c.CompanyName, &c.Address, &c.Email, &c.CreatedAt, &c.UpdatedAt)

	return c, err
}
