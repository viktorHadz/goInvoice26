package res

import (
	"encoding/json"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
)

// Client expects: JSON || No Content (204) || Error
//
// See also WriteError
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Client expects: JSON || No Content (204) || Error
//
// See also WriteJSON
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, app.ErrorResponse{
		Error: msg,
	})
}
