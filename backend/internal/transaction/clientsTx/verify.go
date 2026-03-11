package clientsTx

import (
	"context"
	"errors"

	"github.com/viktorHadz/goInvoice26/internal/app"
)

var ErrClientNotFound = errors.New("client not found")

func VerifyClientID(ctx context.Context, a *app.App, id int64) error {
	exists, err := Exists(ctx, a, id)
	if err != nil {
		return err
	}
	if !exists {
		return ErrClientNotFound
	}
	return nil
}
