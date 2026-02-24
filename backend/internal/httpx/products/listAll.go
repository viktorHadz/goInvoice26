package products

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/products"
)

func listItems(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		idStr := chi.URLParam(r, "clientId")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			slog.ErrorContext(r.Context(), "Bad ID in path", "id", id, "idStr", idStr)
			res.Error(w, res.Internal(err))
			return
		}
		fmt.Printf("ID: %v", id)

		// need to give listall an id, db & ctx minimum
		products, err := products.ListAll(a, r.Context(), id)
		if err != nil {
			slog.ErrorContext(r.Context(), "db failed", "err", err)
			res.Error(w, res.Database(err))
			return
		}

		res.JSON(w, http.StatusOK, products)
	}
}
