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
		slog.DebugContext(r.Context(), "Invoice received from FE", "inv", &invoice)
		// validate.invoice
	}
}
