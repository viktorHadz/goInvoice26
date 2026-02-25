package clients

import (
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
)

func updateClient(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.IDParam(w, r, "clientID")
		if !ok {
			return
		}

		var client models.UpdateClient
		if ok := res.DecodeJSON(w, r, &client); !ok {
			return
		}

		client, errs := ValidateUpdate(client)
		if len(errs) > 0 {
			res.Error(w, res.Validation(errs...))
			return
		}

		affected, err := clientsTx.UpdateClient(r.Context(), a, clientID, client)
		if err != nil {
			res.Error(w, res.Database(err))
			return
		}
		if affected == 0 {
			res.Error(w, res.NotFound("client not found"))
			return
		}

		updated, err := clientsTx.GetByID(r.Context(), a, clientID)
		if err != nil {
			res.Error(w, res.Database(err))
			return
		}
		res.JSON(w, http.StatusOK, updated)
	}
}
