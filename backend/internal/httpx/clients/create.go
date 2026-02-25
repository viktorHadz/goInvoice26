package clients

import (
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	clients "github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
)

// Create - establishes context | validates reqest body | calls DB Transaction
func create(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var client models.CreateClient
		if ok := res.DecodeJSON(w, r, &client); !ok {
			return
		}

		var errs []res.FieldError
		client, errs = ValidateCreate(client)
		if len(errs) > 0 {
			res.Error(w, res.Validation(errs...))
			return
		}

		id, err := clients.Insert(r.Context(), a, &client)
		if err != nil {
			slog.ErrorContext(r.Context(), "Insert client failed", "err", err)
			res.Error(w, res.Database(err))
			return
		}

		created, err := clients.GetByID(r.Context(), a, id)
		if err != nil {
			slog.ErrorContext(r.Context(), "Fetch created client failed", "id", id, "err", err)
			res.Error(w, res.Database(err))
			return
		}

		slog.InfoContext(r.Context(), "client created", "id", id)
		res.JSON(w, http.StatusCreated, created)
	}
}
