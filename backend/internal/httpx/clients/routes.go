package clients

import (
	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/midware"
	"github.com/viktorHadz/goInvoice26/internal/httpx/products"
)

// Register "/api/clients" mux
func Router(r chi.Router, a *app.App) {
	r.Use(midware.LimitBodyMaxSize(2 << 20))

	r.Route("/api/clients", func(r chi.Router) {
		r.Post("/", create(a)) // CREATE  POST /api/clients
		r.Get("/", listAll(a)) // READ    GET  /api/clients

		r.Route("/{clientID}", func(r chi.Router) {
			r.Patch("/", updateClient(a))  // UPDATE  PATCH /api/clients/{id}
			r.Delete("/", deleteClient(a)) // DELETE  DELETE /api/clients/{id}
			products.Router(r, a)
		})
	})
}
