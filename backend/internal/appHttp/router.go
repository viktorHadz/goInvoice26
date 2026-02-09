package apphttp

import (
	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/appHttp/clients"
)

// Registers all service specific routers
func RegisterAllRouters(r chi.Router, a *app.App) {
	clients.Router(r, a)
	// product.Router(r)
	// editor.Router(r)
	// invoice.Router(r)
}
