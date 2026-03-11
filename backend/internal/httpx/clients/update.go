package clients

import (
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
)

func UpdateClient(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}

		var client models.UpdateClient
		if ok := res.DecodeJSON(w, r, &client); !ok {
			return
		}

		client, errs := ValidateUpdate(client)
		if len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}

		affected, err := clientsTx.UpdateClient(r.Context(), a, clientID, client)
		if err != nil {
			slog.ErrorContext(r.Context(),
				"update client failed",
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

		updated, err := clientsTx.GetByID(r.Context(), a, clientID)
		if err != nil {
			slog.ErrorContext(r.Context(),
				"get updated client failed",
				"client_id", clientID,
				"err", err,
			)

			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}
		res.JSON(w, http.StatusOK, updated)
	}
}
