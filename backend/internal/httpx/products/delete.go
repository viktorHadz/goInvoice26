package products

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/productsTx"
)

func deleteProduct(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.IDParam(w, r, "clientID")
		if !ok {
			return
		}

		productID, ok := params.IDParam(w, r, "productID")
		if !ok {
			return
		}

		if err := clientsTx.VerifyClientID(r.Context(), a, clientID, "client not found"); err != nil {
			res.Error(w, err)
			return
		}

		err := productsTx.DeleteTx(a, r.Context(), productID, clientID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				res.Error(w, res.NotFound("product not found"))
				return
			}
			res.Error(w, res.Database(err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
