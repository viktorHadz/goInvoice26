package products

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/productsTx"
)

func DeleteProduct(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}

		productID, ok := params.ValidateParam(w, r, "productID")
		if !ok {
			return
		}

		if err := clientsTx.VerifyClientID(r.Context(), a, clientID); err != nil {
			if errors.Is(err, clientsTx.ErrClientNotFound) {
				res.NotFound(w, "client not found")
				return
			}

			slog.ErrorContext(r.Context(), "verify client failed", "client_id", clientID, "err", err)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		err := productsTx.DeleteTx(a, r.Context(), productID, clientID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				res.NotFound(w, "product not found")
				return
			}

			slog.ErrorContext(r.Context(),
				"delete product failed",
				"client_id", clientID,
				"product_id", productID,
				"err", err,
			)

			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		res.NoContent(w)
	}
}
