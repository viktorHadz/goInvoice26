package invoice

import (
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func GetNextInvoiceNumber(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Return suggested next number (no allocation); number is "used" only on successful create.
		maxNum, err := invoiceTx.GetSuggestedNextBaseNumber(r.Context(), a)
		if err != nil {
			slog.ErrorContext(r.Context(),
				"get next invoice number failed",
				"err", err,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}
		res.JSON(w, http.StatusOK, maxNum)
	}
}
