package app

import "database/sql"

type UserSettings struct {
	CompanyName       string
	Email             *string
	Phone             *string
	CompanyAddress    *string
	InvoicePrefix     string
	Currency          string
	DateFormat        string
	CustomItemsPrefix string
	PaymentTerms      *string
	PaymentDetails    *string
	NotesFooter       *string
	LogoURL           *string
}

type App struct {
	DB     *sql.DB
	UsrCfg *UserSettings
}
