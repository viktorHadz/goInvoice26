package clients

import (
	"context"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
)

// Checks that a client exists by ID.
// Returns an error to consume as res.Error(w, err).
func VerifyClientID(ctx context.Context, a *app.App, id int64, notFoundMsg string) error {
	exists, err := clientsTx.Exists(ctx, a, id)
	if err != nil {
		return res.Database(err)
	}
	if !exists {
		if notFoundMsg == "" {
			notFoundMsg = "client not found"
		}
		return res.NotFound(notFoundMsg)
	}
	return nil
}
