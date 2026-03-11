package clients

import (
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
)

func ListAll(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		allClients, err := clientsTx.ListClients(a, r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(),
				"list clients failed ",
				"All Clients struct", allClients,
				"err", err,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}
		res.JSON(w, http.StatusOK, allClients)
	}
}
