package app

import (
	"database/sql"

	"github.com/viktorHadz/goInvoice26/internal/service/auth"
	"github.com/viktorHadz/goInvoice26/internal/service/logo"
)

type App struct {
	DB    *sql.DB
	Auth  *auth.Service
	Logos *logo.Service
}
