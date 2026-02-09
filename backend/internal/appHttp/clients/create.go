package clients

import (
	"encoding/json"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clients"
)

// Create is a handler func satisfiying the http.Handler interface
//
// Create establishes context | validates reqest body | calls the clients service layer
func create(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode request payload
		var client models.CreateClient
		err := json.NewDecoder(r.Body).Decode(&client)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if client.Name == "" || len(client.Name) > 50 {
			http.Error(w, "Name must be less than 50 characters and cannot be empty", http.StatusBadRequest)
			return
		}
		// Perform write transaction
		id, err := clients.Insert(r.Context(), a, &client)
		if err != nil {
			http.Error(w, "Server error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"id": id})

	}
}
