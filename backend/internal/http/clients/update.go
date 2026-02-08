package clients

import (
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
)

func update(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// DB call here

	}
}
