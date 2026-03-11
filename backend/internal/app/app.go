package app

import (
	"database/sql"
)

type UserConfig struct {
	Name        string
	Address     string
	Email       string
	Phone       string
	CompanyName string
	Logo        string
}

// Provides dependencies to the application
type App struct {
	DB         *sql.DB // DB mounted in main
	UserConfig UserConfig
}
