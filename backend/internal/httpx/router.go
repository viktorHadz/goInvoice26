package httpx

import (
	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/clients"
)

// Registers all service specific routers
func RegisterAllRouters(r chi.Router, a *app.App) {
	clients.Router(r, a)
	// products mounted inside clients for sane name convention
	// e.g. /api/clients/{clientID}/products....
	// editor.Router(r)
	// invoice.Router(r)
}
