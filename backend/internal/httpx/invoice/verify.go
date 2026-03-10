package invoice

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

type verifyResponse struct {
	Invoice models.FEInvoiceIn `json:"invoice"`
}

// Ensures frontend calculations are consistent. Called on FE for optimistic invoice update
func verifyInvoice(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIDParam := chi.URLParam(r, "clientID")
		baseNumberParam := chi.URLParam(r, "baseNumber")

		clientID, err := strconv.ParseInt(clientIDParam, 10, 64)
		if err != nil || clientID < 1 {
			res.Error(w, res.Validation(res.Invalid("clientId", "invalid route param")))
			return
		}

		baseNumber, err := strconv.ParseInt(baseNumberParam, 10, 64)
		if err != nil || baseNumber < 1 {
			res.Error(w, res.Validation(res.Invalid("baseNumber", "invalid route param")))
			return
		}

		var invoice models.FEInvoiceIn
		if ok := res.DecodeJSON(w, r, &invoice); !ok {
			return
		}

		// Route/body consistency
		var routeErrs []res.FieldError
		if invoice.Overview.ClientID != clientID {
			routeErrs = append(routeErrs, res.Invalid("clientId", "does not match route param"))
		}
		if invoice.Overview.BaseNumber != baseNumber {
			routeErrs = append(routeErrs, res.Invalid("baseNumber", "does not match route param"))
		}
		if len(routeErrs) > 0 {
			res.Error(w, res.Validation(routeErrs...))
			return
		}

		validInvoice, errs := ValidateInvoiceCreate(invoice)
		if len(errs) > 0 {
			res.Error(w, res.Validation(errs...))
			return
		}

		canonical := RecalcInvoice(validInvoice)
		slog.Debug("invoice verified", "clientID", clientID, "baseNumber", baseNumber)

		res.JSON(w, http.StatusOK, verifyResponse{Invoice: canonical})
	}
}
