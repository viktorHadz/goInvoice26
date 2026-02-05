package main

import (
	"context"
	"log"

	"github.com/viktorHadz/goInvoice26/internal/config"
	"github.com/viktorHadz/goInvoice26/internal/db"
)

func main() {
	// Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// DB
	ctx := context.Background()

	dbConn, err := db.OpenDB(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	if err := db.Migrate(ctx, dbConn); err != nil {
		log.Fatal(err)
	}

	log.Printf("env=%s db=%s", cfg.Env, cfg.DBPath)
	log.Printf("API listening on :%s", cfg.Port)

	// Server
}
