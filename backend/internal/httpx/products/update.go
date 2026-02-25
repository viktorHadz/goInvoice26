package products

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/productsTx"
)

func updateProduct(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.IDParam(w, r, "clientID")
		if !ok {
			return
		}

		productID, ok := params.IDParam(w, r, "productID")
		if !ok {
			return
		}

		// Verify client exists
		if err := clientsTx.VerifyClientID(r.Context(), a, clientID, "client not found"); err != nil {
			res.Error(w, err)
			return
		}

		// Decode incoming
		var in models.ProductCreateIn
		if ok := res.DecodeJSON(w, r, &in); !ok {
			return
		}

		// Validate and normalize into DB payload
		cmd, errs := ValidateCreate(in, clientID)
		if len(errs) > 0 {
			res.Error(w, res.Validation(errs...))
			return
		}

		updated, err := productsTx.UpdateTx(a, r.Context(), productID, cmd)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				res.Error(w, res.NotFound("product not found"))
				return
			}
			res.Error(w, res.Database(err))
			return
		}

		res.JSON(w, http.StatusOK, updated)
	}
}
