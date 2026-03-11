package products

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/productsTx"
)

func ListItems(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}

		if err := clientsTx.VerifyClientID(r.Context(), a, id); err != nil {
			if errors.Is(err, clientsTx.ErrClientNotFound) {
				res.NotFound(w, "client not found")
				return
			}

			slog.ErrorContext(r.Context(),
				"verify client failed",
				"client_id", id,
				"err", err,
			)

			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		products, err := productsTx.ListAll(a, r.Context(), id)
		if err != nil {
			slog.ErrorContext(r.Context(),
				"list products failed",
				"client_id", id,
				"err", err,
			)

			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		res.JSON(w, http.StatusOK, products)
	}
}
