package clients

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clients"
)

// Create establishes context | validates reqest body | calls the clients service layer
func create(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var client models.CreateClient
		if err := json.NewDecoder(r.Body).Decode(&client); err != nil {
			res.Error(w, res.Validation(
				res.Invalid("body", "invalid JSON"),
			))
			return
		}

		errs := []res.FieldError{}
		if client.Name == "" {
			errs = append(errs, res.Required("name"))
		} else if len(client.Name) > 50 {
			errs = append(errs, res.MaxLen("name", 50))
		}

		if len(errs) > 0 {
			res.Error(w, res.Validation(errs...))
			return
		}

		id, err := clients.Insert(r.Context(), a, &client)
		if err != nil {
			slog.ErrorContext(r.Context(), "insert client failed", "err", err)
			res.Error(w, res.Internal())
			return
		}

		slog.InfoContext(r.Context(), "client created", "id", id)
		res.JSON(w, http.StatusCreated, map[string]any{
			"message": "client created",
			"id":      id,
		})
	}
}
