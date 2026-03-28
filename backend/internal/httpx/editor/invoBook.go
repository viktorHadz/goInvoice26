package editor

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/editorTx"
)

func optionalPositiveInt64(raw string) (int64, bool, error) {
	if raw == "" {
		return 0, false, nil
	}

	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v < 1 {
		return 0, false, fmt.Errorf("invalid positive int: %q", raw)
	}

	return v, true, nil
}

func HandleINVBookData(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := 10
		offset := 0
		clientID := int64(0)

		routeClientID, hasRouteClientID, err := optionalPositiveInt64(chi.URLParam(r, "clientID"))
		if hasRouteClientID {
			clientID = routeClientID
		} else if err != nil {
			res.Validation(w, res.Invalid("clientID", "invalid route parameter"))
			return
		}

		queryClientID, hasQueryClientID, err := optionalPositiveInt64(r.URL.Query().Get("clientId"))
		if hasQueryClientID {
			clientID = queryClientID
		} else if err != nil {
			res.Error(w, http.StatusBadRequest, "BAD_QUERY", "Invalid clientId")
			return
		}

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

		filters := editorTx.InvoiceBookPageFilters{
			SortBy:        r.URL.Query().Get("sortBy"),
			SortDirection: r.URL.Query().Get("sortDirection"),
			PaymentState:  r.URL.Query().Get("paymentState"),
		}

		IBData, err := editorTx.QueryInvoiceBookPage(
			a,
			r.Context(),
			clientID,
			limit,
			offset,
			filters,
		)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "DB_ERROR - error while getting invoice book data",
				"err", err,
				"clientID", clientID,
				"limit", limit,
				"offset", offset,
				"sortBy", filters.SortBy,
				"sortDirection", filters.SortDirection,
				"paymentState", filters.PaymentState,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		res.JSON(w, http.StatusOK, IBData)
	}
}
