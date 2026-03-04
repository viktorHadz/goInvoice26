package invoice

import (
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func getNextInvoiceNumber(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		maxNum, err := invoiceTx.GetNextBaseNumber(r.Context(), a)
		if err != nil {
			res.Error(w, res.Database(err))
			return
		}
		res.JSON(w, http.StatusOK, maxNum)
	}
}
