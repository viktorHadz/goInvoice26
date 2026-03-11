// validates route parameters
package params

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
)

// ValidateParam reads a chi URL param and validates it as an int64.
// It writes a 400 validation response and returns ok=false on failure.
func ValidateParam(w http.ResponseWriter, r *http.Request, param string) (valid int64, ok bool) {
	s := chi.URLParam(r, param)
	valid, err := strconv.ParseInt(s, 10, 64)
	if err != nil || valid < 1 {
		res.Validation(w, res.Invalid(param, "invalid route param"))
		return 0, false
	}

	return valid, true
}
