package clientsTx

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

// Partial update PATCH
func UpdateClient(ctx context.Context, a *app.App, id int64, input models.UpdateClient) (int64, error) {
	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return 0, err
	}

	setParts := make([]string, 0, 5)
	args := make([]any, 0, 6)

	if input.Name != nil {
		setParts = append(setParts, "name = ?")
		args = append(args, *input.Name)
	}
	if input.CompanyName != nil {
		setParts = append(setParts, "company_name = NULLIF(?, '')")
		args = append(args, *input.CompanyName)
	}
	if input.Address != nil {
		setParts = append(setParts, "address = NULLIF(?, '')")
		args = append(args, *input.Address)
	}
	if input.Email != nil {
		setParts = append(setParts, "email = NULLIF(?, '')")
		args = append(args, *input.Email)
	}

	if len(setParts) == 0 {
		return 0, errors.New("no fields to update")
	}

	// update timestamp
	setParts = append(setParts, "updated_at = CURRENT_TIMESTAMP")

	// WHERE id/account_id go at the end
	args = append(args, id)
	args = append(args, accountID)

	query := fmt.Sprintf(`
		UPDATE clients
		SET %s
		WHERE id = ?
		  AND account_id = ?
	`, strings.Join(setParts, ", "))

	result, err := a.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
