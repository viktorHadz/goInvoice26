package clients

import (
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clients"
)

func listAll(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allClients, err := clients.ListClients(a, r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "list clients failed", "err", err)
			res.Error(w, res.Database(err))
			return
		}
		res.JSON(w, http.StatusOK, allClients)
	}
}
