package editor

import (
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/editorTx"
)

func HandleINVBookData(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}

		// Set query LIMIT and OFFSET here
		IBData, err := editorTx.QueryInvoiceBookPage(a, r.Context(), id, 10, 0)
		if err != nil {
			slog.ErrorContext(r.Context(), "DB_ERROR",
				"Error while getting invo book data", err,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		res.JSON(w, http.StatusOK, IBData)
	}
}
