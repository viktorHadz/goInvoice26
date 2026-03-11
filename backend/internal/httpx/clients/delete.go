package clients

import (
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
)

func DeleteClient(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}

		affected, err := clientsTx.DeleteClient(a, r.Context(), clientID)
		if err != nil {
			slog.ErrorContext(r.Context(),
				"delete client failed",
				"client_id", clientID,
				"err", err,
			)

			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		if affected == 0 {
			res.NotFound(w, "client not found")
			return
		}

		res.NoContent(w)
	}
}
