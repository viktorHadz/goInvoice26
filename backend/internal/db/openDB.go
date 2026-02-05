package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDB(dbPath string) (*sql.DB, error) {
	dsn := dbPath + "?_loc=UTC&parseTime=true"

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil { // Ping() creates the DB if it doesnt exist
		db.Close()
		return nil, err
	}

	return db, nil
}
