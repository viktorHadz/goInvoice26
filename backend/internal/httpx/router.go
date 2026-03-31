package httpx

import (
	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	authhttp "github.com/viktorHadz/goInvoice26/internal/httpx/auth"
	billinghttp "github.com/viktorHadz/goInvoice26/internal/httpx/billing"
	"github.com/viktorHadz/goInvoice26/internal/httpx/clients"
	"github.com/viktorHadz/goInvoice26/internal/httpx/editor"
	"github.com/viktorHadz/goInvoice26/internal/httpx/invoice"
	"github.com/viktorHadz/goInvoice26/internal/httpx/midware"
	"github.com/viktorHadz/goInvoice26/internal/httpx/products"
	"github.com/viktorHadz/goInvoice26/internal/httpx/settings"
	"github.com/viktorHadz/goInvoice26/internal/httpx/team"
)

func RegisterAllRouters(r chi.Router, a *app.App) {
	r.Route("/api/auth", func(r chi.Router) {
		r.Get("/me", authhttp.Me(a))
		r.Post("/logout", authhttp.Logout(a))
		r.Get("/google/start", authhttp.GoogleStart(a))
		r.Get("/google/callback", authhttp.GoogleCallback(a))
	})

	r.Post("/api/billing/stripe/webhook", billinghttp.StripeWebhook(a))

	r.Group(func(r chi.Router) {
		r.Use(midware.RequireAuth(a))

		r.Route("/api/billing", func(r chi.Router) {
			r.Post("/checkout-session", billinghttp.CreateCheckoutSession(a))
			r.Post("/portal-session", billinghttp.CreatePortalSession(a))
			r.Post("/checkout/sync", billinghttp.SyncCheckoutSession(a))
		})

		r.Group(func(r chi.Router) {
			r.Use(midware.RequireBillingAccess)

			r.Route("/api/settings", func(r chi.Router) {
				r.Get("/", settings.Get(a))
				r.Put("/", settings.Put(a))
				r.Route("/logo", func(r chi.Router) {
					r.Use(midware.LimitBodyMaxSize(5 << 20))
					r.Get("/", settings.GetLogo(a))
					r.Put("/", settings.PutLogo(a))
					r.Delete("/", settings.DeleteLogo(a))
				})
			})

			r.Route("/api/team", func(r chi.Router) {
				r.Use(midware.RequireOwner)
				r.Get("/", team.List(a))
				r.Post("/invites", team.CreateInvite(a))
				r.Delete("/invites/{inviteID}", team.DeleteInvite(a))
				r.Delete("/members/{memberID}", team.DeleteMember(a))
			})

			r.Get("/api/edits", editor.HandleINVBookData(a))

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
							r.Get("/{revisionNo}/docx", invoice.GenerateDOCXHandler(a))
							r.Post("/{revisionNo}/docx/quick", invoice.QuickDOCXHandler(a))
						})
					})
				})
			})
		})
	})
}
