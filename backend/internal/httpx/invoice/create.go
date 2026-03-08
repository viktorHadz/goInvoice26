package invoice

import (
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

func createInvoice(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var invoice models.FEInvoiceIn
		if ok := res.DecodeJSON(w, r, &invoice); !ok {
			slog.WarnContext(r.Context(), "Received bad JSON for invoice",
				"json", &invoice,
			)
			return
		}

		slog.Debug("Invoice received from FE", "inv", &invoice)

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
