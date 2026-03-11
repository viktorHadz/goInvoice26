package products

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/productsTx"
)

func UpdateProduct(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}

		productID, ok := params.ValidateParam(w, r, "productID")
		if !ok {
			return
		}

		// Verify client exists
		if err := clientsTx.VerifyClientID(r.Context(), a, clientID); err != nil {
			if errors.Is(err, clientsTx.ErrClientNotFound) {
				res.NotFound(w, "client not found")
				return
			}

			slog.ErrorContext(r.Context(),
				"verify client failed",
				"client_id", clientID,
				"err", err,
			)

			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		// Decode incoming
		var in models.ProductCreateIn
		if ok := res.DecodeJSON(w, r, &in); !ok {
			return
		}

		// Validate and normalize into DB payload
		cmd, errs := ValidateCreate(in, clientID)
		if len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}

		updated, err := productsTx.UpdateTx(a, r.Context(), productID, cmd)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				res.NotFound(w, "product not found")
				return
			}

			slog.ErrorContext(r.Context(),
				"update product failed",
				"client_id", clientID,
				"product_id", productID,
				"err", err,
			)

			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		res.JSON(w, http.StatusOK, updated)
	}
}
