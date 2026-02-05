package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/viktorHadz/goInvoice26/internal/app/clients"
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

	// Build deps
	q := clients.ClientQueries{DB: dbConn}
	svc := clients.ClientService{Q: q}
	clientAPI := clients.ClientAPI{Svc: svc}

	// Server
	r := gin.Default()
	clientAPI.Register(r)

	log.Printf("env=%s db=%s", cfg.Env, cfg.DBPath)
	log.Printf("API listening on :%s", cfg.Port)
	log.Fatal(r.Run(":" + cfg.Port))
}
