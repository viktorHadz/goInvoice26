package invoice

import (
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

type verifyResponse struct {
	Invoice models.FEInvoiceIn `json:"invoice"`
}

// Ensures frontend calculations are consistent. Called on FE for optimistic invoice update
func VerifyInvoice() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}
		baseNumber, ok := params.ValidateParam(w, r, "baseNumber")
		if !ok {
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
			res.Validation(w, routeErrs...)
			return
		}

		validInvoice, errs := ValidateInvoiceCreate(invoice)
		if len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}

		canonical := RecalcInvoice(validInvoice)
		slog.DebugContext(r.Context(), "invoice verified", "client_id", clientID, "base_number", baseNumber)

		res.JSON(w, http.StatusOK, verifyResponse{Invoice: canonical})
	}
}
