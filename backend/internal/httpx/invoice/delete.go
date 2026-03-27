package invoice

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func DeleteInvoice(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}
		baseNumber, ok := params.ValidateParam(w, r, "baseNumber")
		if !ok {
			return
		}

		err := invoiceTx.Delete(r.Context(), a, clientID, baseNumber)
		if err != nil {
			switch {
			case errors.Is(err, invoiceTx.ErrInvoiceNotFound):
				res.NotFound(w, "Invoice not found")
				return
			case errors.Is(err, invoiceTx.ErrInvoiceDeleteVoid):
				res.Error(w, http.StatusConflict, "INVOICE_VOID", "Void invoices are final records and cannot be deleted")
				return
			}

			slog.ErrorContext(r.Context(),
				"delete invoice failed",
				"client_id", clientID,
				"base_number", baseNumber,
				"err", err,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
