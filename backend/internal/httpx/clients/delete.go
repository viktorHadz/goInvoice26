package clients

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	clients "github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
)

func deleteClient(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			res.Error(w, res.Validation(res.Invalid("id", "invalid route param")))
			return
		}

		affected, err := clients.DeleteClient(a, r.Context(), id)
		if err != nil {
			slog.ErrorContext(r.Context(), "delete client failed", "id", id, "err", err)
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
