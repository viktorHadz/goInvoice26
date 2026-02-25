// validates route parameters
package params

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
)

// IDParam reads a chi URL param and validates it as a positive int64.
// It writes a 400 validation response and returns ok=false on failure.
func IDParam(w http.ResponseWriter, r *http.Request, name string) (id int64, ok bool) {
	s := chi.URLParam(r, name)

	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil || id <= 0 {
		res.Error(w, res.Validation(res.FieldError{Field: name, Code: "INVALID_ID"}))
		return 0, false
	}

	return id, true
}
