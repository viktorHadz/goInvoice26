package editor

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/editorTx"
)

func HandleINVBookData(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}

		limit := 10
		offset := 0

		if raw := r.URL.Query().Get("limit"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v < 1 {
				res.Error(w, http.StatusBadRequest, "BAD_QUERY", "Invalid limit")
				return
			}
			limit = v
		}

		if raw := r.URL.Query().Get("offset"); raw != "" {
			v, err := strconv.Atoi(raw)
			if err != nil || v < 0 {
				res.Error(w, http.StatusBadRequest, "BAD_QUERY", "Invalid offset")
				return
			}
			offset = v
		}

		IBData, err := editorTx.QueryInvoiceBookPage(a, r.Context(), clientID, limit, offset)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "DB_ERROR - error while getting invoice book data",
				"err", err,
				"clientID", clientID,
				"limit", limit,
				"offset", offset,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		res.JSON(w, http.StatusOK, IBData)
	}
}
