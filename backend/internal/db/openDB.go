package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func OpenDB(dbPath string) (*sql.DB, error) {
	if err := ensureParentDir(dbPath); err != nil {
		return nil, err
	}

	dsn := dbPath + "?_loc=UTC&parseTime=true&_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL"

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil { // Ping() creates the DB if it doesnt exist
		db.Close()
		return nil, err
	}

	if _, err := db.Exec(`
		PRAGMA foreign_keys = ON;
		PRAGMA busy_timeout = 5000;
		PRAGMA journal_mode = WAL;
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("configure sqlite pragmas: %w", err)
	}

	return db, nil
}

func ensureParentDir(dbPath string) error {
	// Skip special SQLite DSNs that are not ordinary filesystem paths.
	if dbPath == "" || dbPath == ":memory:" || strings.HasPrefix(dbPath, "file:") {
		return nil
	}

	dir := filepath.Dir(dbPath)
	if dir == "." || dir == "" {
		return nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create database directory %s: %w", dir, err)
	}

	return nil
}
