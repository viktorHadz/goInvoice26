package app

import (
	"database/sql"
)

// Provides dependencies to the app
type App struct {
	DB *sql.DB
}

type ErrorResponse struct {
	Error string `json:"error"`
}
