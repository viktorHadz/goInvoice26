package clients

import (
	"encoding/json"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clients"
)

func listAll(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var allClients []models.Client
		allClients, err := clients.ListClients(a, r.Context())
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(allClients)

	}
}
