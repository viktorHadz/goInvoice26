package clients

import (
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
)

func listAll(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allClients, err := clientsTx.ListClients(a, r.Context())
		if err != nil {
			res.Error(w, res.Database(err))
			return
		}
		res.JSON(w, http.StatusOK, allClients)
	}
}
