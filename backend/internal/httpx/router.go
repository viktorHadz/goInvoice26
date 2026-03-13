package httpx

import (
	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/clients"
	"github.com/viktorHadz/goInvoice26/internal/httpx/invoice"
	"github.com/viktorHadz/goInvoice26/internal/httpx/midware"
	"github.com/viktorHadz/goInvoice26/internal/httpx/products"
)

func RegisterAllRouters(r chi.Router, a *app.App) {
	r.Use(midware.LimitBodyMaxSize(2 << 20))

	r.Route("/api/clients", func(r chi.Router) {
		r.Post("/", clients.Create(a))
		r.Get("/", clients.ListAll(a))

		r.Route("/{clientID}", func(r chi.Router) {
			r.Patch("/", clients.UpdateClient(a))
			r.Delete("/", clients.DeleteClient(a))

			// /api/clients/{clientID}/products/...
			r.Route("/products", func(r chi.Router) {
				r.Get("/", products.ListItems(a))
				r.Post("/", products.CreateProduct(a))
				r.Route("/{productID}", func(r chi.Router) {
					r.Patch("/", products.UpdateProduct(a))
					r.Delete("/", products.DeleteProduct(a))
				})
			})

			// /api/clients/{clientID}/invoice/...
			r.Route("/invoice", func(r chi.Router) {
				r.Get("/", invoice.GetNextInvoiceNumber(a))
				r.Route("/{baseNumber}", func(r chi.Router) {
					r.Post("/", invoice.CreateInvoice(a))
					r.Post("/verify", invoice.VerifyInvoice())
					// /api/clients/{clientID}/invoice/{baseNumber}/{revisionNO}pdf
					r.Get("/{revisionNo}/pdf", invoice.GeneratePDFHandler(a))
					r.Post("/{revisionNo}/pdf/quick", invoice.QuickPDFHandler(a))
				})
			})
		})
	})
}
