package clients

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clients"
)

func updateClient(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse id
		idStr := chi.URLParam(r, "id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			res.Error(w, res.Validation(res.Invalid("id", "invalid route param")))
			return
		}

		// Decode PATCH payload
		var patch models.UpdateClient
		if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
			res.Error(w, res.Validation(res.Invalid("body", "invalid JSON payload")))
			return
		}

		// Validate and sanitize
		patch, err = ValidateUpdate(patch)
		if err != nil {
			res.Error(w, err)
			return
		}

		// Update
		affected, err := clients.UpdateClient(r.Context(), a, id, patch)
		if err != nil {
			res.Error(w, res.Database())
			return
		}
		if affected == 0 {
			res.Error(w, res.NotFound("client"))
			return
		}

		// Respond
		// Needs to be refetched via res.JSON()
		// if multiple users access the same record concurently
		res.NoContent(w)
	}
}
