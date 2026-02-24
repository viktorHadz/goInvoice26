package products

import (
	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
)

func Router(r chi.Router, a *app.App) {
	r.Route("/api/products/{clientId}", func(r chi.Router) {
		r.Get("/", listItems(a)) // READ    GET  /api/products/{clientID}
	})

}
