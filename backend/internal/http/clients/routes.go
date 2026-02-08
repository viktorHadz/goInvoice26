package clients

import (
	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
)

// Registers a "/api/clients" router and its sub routers,
// with POST, GET, PATCH, DELETE functionality
func Router(r chi.Router, a *app.App) {
	r.Route("/api/clients", func(r chi.Router) {
		r.Post("/", create(a)) // CREATE  POST /clients
		r.Get("/", listAll(a)) // READ    GET  /clients

		r.Route("/{id}", func(r chi.Router) {
			r.Patch("/", update(a))  // UPDATE  PATCH /clients/{id}
			r.Delete("/", delete(a)) // DELETE DELETE /clients/{id}
		})
	})
}
