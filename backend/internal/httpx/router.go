package httpx

import (
	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/clients"
	"github.com/viktorHadz/goInvoice26/internal/httpx/products"
)

// Registers all service specific routers
func RegisterAllRouters(r chi.Router, a *app.App) {
	clients.Router(r, a)
	products.Router(r, a)
	// editor.Router(r)
	// invoice.Router(r)
}
