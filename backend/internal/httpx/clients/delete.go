package clients

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clients"
)

func deleteClient(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		idStr := chi.URLParam(r, "id")

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			res.Error(w, res.Validation(res.Invalid("id", "invalid id")))
			return
		}

		affected, err := clients.DeleteClient(a, r.Context(), id)
		if err != nil {
			res.Error(w, res.Database())
			return
		}

		if affected == 0 {
			res.Error(w, res.Database())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
