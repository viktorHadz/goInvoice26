package clients

import (
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
)

// Create - establishes context | validates reqest body | calls DB Transaction
func Create(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var client models.CreateClient
		if ok := res.DecodeJSON(w, r, &client); !ok {
			return
		}

		var errs []res.FieldError
		client, errs = ValidateCreate(client)
		if len(errs) > 0 {
			slog.DebugContext(r.Context(), "client validation failed", "errs", errs)
			res.Validation(w, errs...)
			return
		}

		id, err := clientsTx.Insert(r.Context(), a, &client)
		if err != nil {
			slog.Error("Database transaction failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		created, err := clientsTx.GetByID(r.Context(), a, id)
		if err != nil {
			slog.Error("Failed to retrieve created client", "id", id, "err", err)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		slog.DebugContext(r.Context(), "client created", "id", id)
		res.JSON(w, http.StatusCreated, created)
	}
}
