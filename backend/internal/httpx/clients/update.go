package clients

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
)

func updateClient(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			res.Error(w, res.Validation(res.Invalid("id", "invalid route param")))
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

		affected, err := clientsTx.UpdateClient(r.Context(), a, id, client)
		if err != nil {
			slog.ErrorContext(r.Context(), "update client failed", "id", id, "err", err)
			res.Error(w, res.Database(err))
			return
		}
		if affected == 0 {
			res.Error(w, res.NotFound("client not found"))
			return
		}

		updated, err := clientsTx.GetByID(r.Context(), a, id)
		if err != nil {
			slog.ErrorContext(r.Context(), "fetch updated client failed", "id", id, "err", err)
			res.Error(w, res.Database(err))
			return
		}
		res.JSON(w, http.StatusOK, updated)
	}
}
