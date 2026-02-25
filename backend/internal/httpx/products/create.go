package products

import (
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/clients"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/productsTx"
)

func createProduct(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.IDParam(w, r, "clientID") // validate route param
		if !ok {
			return
		}

		// verify client exists
		if err := clients.VerifyClientID(r.Context(), a, clientID, "client not found"); err != nil {
			res.Error(w, err)
			return
		}

		// Unpack r body
		var in models.ProductCreateIn
		if ok := res.DecodeJSON(w, r, &in); !ok {
			return
		}

		// validate and normalize DB struct into models.productCreate
		product, errs := ValidateCreate(in, clientID)
		if len(errs) > 0 {
			res.Error(w, res.Validation(errs...))
			return
		}

		// insert in DB
		created, err := productsTx.InsertTx(a, r.Context(), product)
		if err != nil {
			res.Error(w, err)
			return
		}

		res.JSON(w, http.StatusCreated, created) // respond to FE
	}
}
