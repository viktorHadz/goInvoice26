package httpx

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/clients"
	"github.com/viktorHadz/goInvoice26/internal/httpx/editor"
	"github.com/viktorHadz/goInvoice26/internal/httpx/invoice"
	"github.com/viktorHadz/goInvoice26/internal/httpx/midware"
	"github.com/viktorHadz/goInvoice26/internal/httpx/products"
	"github.com/viktorHadz/goInvoice26/internal/httpx/settings"
)

func RegisterAllRouters(r chi.Router, a *app.App) {
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	r.Route("/api/image", func(r chi.Router) {
		r.Use(midware.LimitBodyMaxSize(5 << 20))
		r.Post("/", settings.LogoUpload())
	})

	r.Route("/api/settings", func(r chi.Router) {
		r.Get("/", settings.Get(a))
		r.Put("/", settings.Put(a))
	})

	r.Route("/api/clients", func(r chi.Router) {
		r.Use(midware.LimitBodyMaxSize(2 << 20)) // 2MB
		r.Post("/", clients.Create(a))
		r.Get("/", clients.ListAll(a))

		r.Route("/{clientID}", func(r chi.Router) {
			r.Patch("/", clients.UpdateClient(a))
			r.Delete("/", clients.DeleteClient(a))

			// /api/clients/{clientID}/edits/...
			r.Route("/edits", func(r chi.Router) {
				r.Get("/", editor.HandleINVBookData(a))
				r.Get("/get/{baseNo}/{revNo}", editor.GetInvoice(a))
			})

			// /api/clients/{clientID}/products/...
			r.Route("/products", func(r chi.Router) {
				r.Get("/", products.ListItems(a))
				r.Post("/", products.CreateProduct(a))
				r.Route("/{productID}", func(r chi.Router) {
					r.Patch("/", products.UpdateProduct(a))
					r.Delete("/", products.DeleteProduct(a))
				})
			})
			// TODO: rate limit create revision manually

			// /api/clients/{clientID}/invoice/...
			r.Route("/invoice", func(r chi.Router) {
				r.Get("/", invoice.GetNextInvoiceNumber(a))
				r.Route("/{baseNumber}", func(r chi.Router) {
					r.Post("/", invoice.CreateInvoice(a))
					r.Put("/", invoice.UpdateInvoice(a))
					r.Delete("/", invoice.DeleteInvoice(a))
					r.Patch("/status", invoice.PatchInvoiceStatus(a))
					r.Post("/verify", invoice.VerifyInvoice())
					r.Post("/revisions", invoice.CreateRevision(a))
					r.Get("/{revisionNo}/pdf", invoice.GeneratePDFHandler(a))
					r.Post("/{revisionNo}/pdf/quick", invoice.QuickPDFHandler(a))
				})
			})
		})
	})
}
