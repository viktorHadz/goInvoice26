package products

import (
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/clients"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/productsTx"
)

func listItems(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := params.IDParam(w, r, "clientID")
		if !ok {
			return
		}

		if err := clients.VerifyClientID(r.Context(), a, id, "client not found"); err != nil {
			res.Error(w, err)
			return
		}

		products, err := productsTx.ListAll(a, r.Context(), id)
		if err != nil {
			slog.ErrorContext(r.Context(), "DB ERRORED", "err", err)
			res.Error(w, res.Database(err))
			return
		}

		res.JSON(w, http.StatusOK, products)
	}
}
