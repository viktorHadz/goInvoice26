package app

import (
	"database/sql"
)

// Provides dependencies to the application
type App struct {
	DB *sql.DB
}
