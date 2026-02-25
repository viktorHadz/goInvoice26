package products

import (
	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
)

func Router(r chi.Router, a *app.App) {
	r.Route("/products", func(r chi.Router) {
		r.Get("/", listItems(a))      // READ    GET  /api/clients/{clientID}/products/...
		r.Post("/", createProduct(a)) // CREATE POST /api/clients/{clientID}/products/...
		r.Route("/{productID}", func(r chi.Router) {
			r.Patch("/", updateProduct(a))
			r.Delete("/", deleteProduct(a))
		})
	})
}
