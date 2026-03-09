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

func createInvoice(a *app.App) http.HandlerFunc {
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
			slog.WarnContext(r.Context(), "Received bad JSON for invoice",
				"json", &invoice,
			)
			return
		}

		slog.Debug("Invoice received from FE", "inv", &invoice)

		// Route param consistency (prevents mismatched path/body invoices)
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

		// validate received invoice
		validInvoice, errs := ValidateInvoiceCreate(invoice)
		if len(errs) > 0 {
			res.Error(w, res.Validation(errs...))
			return
		}

		slog.Debug("Validated Invoice", "------", "-----", "inv", validInvoice)
		// recalculate totals and check they match those written by the frontend totals match

	}
}
