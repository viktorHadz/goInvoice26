package clients

import (
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
)

func deleteClient(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.IDParam(w, r, "clientID")
		if !ok {
			return
		}

		affected, err := clientsTx.DeleteClient(a, r.Context(), clientID)
		if err != nil {
			res.Error(w, res.Database(err))
			return
		}

		if affected == 0 {
			res.Error(w, res.NotFound("client not found"))
			return
		}

		res.NoContent(w)
	}
}
