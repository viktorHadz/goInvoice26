package invoice

import (
	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
)

func Router(r chi.Router, a *app.App) {
	r.Route("/invoice", func(r chi.Router) {
		r.Get("/", getNextInvoiceNumber(a))
		r.Route("/{baseNumber}", func(r chi.Router) {
			r.Post("/", createInvoice(a)) // Create new invoice

		})
	})
}
