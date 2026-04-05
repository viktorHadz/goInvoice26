package app

import (
	"database/sql"

	"github.com/viktorHadz/goInvoice26/internal/service/auth"
	"github.com/viktorHadz/goInvoice26/internal/service/billing"
	"github.com/viktorHadz/goInvoice26/internal/service/logo"
	"github.com/viktorHadz/goInvoice26/internal/service/productimport"
	"github.com/viktorHadz/goInvoice26/internal/service/workspace"
)

type App struct {
	DB                           *sql.DB
	Auth                         *auth.Service
	Billing                      *billing.Service
	Logos                        *logo.Service
	ProductImports               *productimport.Coordinator
	Workspaces                   *workspace.Service
	AccessLedgerSecret           string
	PromoRedemptionRetentionDays int
}
