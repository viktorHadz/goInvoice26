package products

import (
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

func CreateProduct(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}

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

		var in models.ProductCreateIn
		if ok := res.DecodeJSON(w, r, &in); !ok {
			return
		}

		product, errs := ValidateCreate(in, clientID)
		if len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}

		created, err := productsTx.InsertTx(a, r.Context(), product)
		if err != nil {
			slog.ErrorContext(r.Context(),
				"create product failed",
				"client_id", clientID,
				"err", err,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		res.JSON(w, http.StatusCreated, created)
	}
}
