package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/viktorHadz/goInvoice26/internal/app"
	apphttp "github.com/viktorHadz/goInvoice26/internal/appHttp"
	"github.com/viktorHadz/goInvoice26/internal/config"
	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/logging"
)

func main() {
	// Config init
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("CONFIIG INIT")

	// Ctx init
	ctx := context.Background()

	// DB init
	dbConn, err := db.OpenDB(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	if err := db.Migrate(ctx, dbConn); err != nil {
		log.Fatal(err)
	}

	// Create a server and mux
	r := chi.NewRouter()

	// Logger init
	logger, opts := logging.InitLogger(cfg)

	// Register middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(httplog.RequestLogger(logger, opts))
	logger.Log(ctx, slog.LevelInfo, "Testing logger")

	// Register routes
	apphttp.RegisterAllRouters(r, &app.App{DB: dbConn})

	log.Printf("env=%s db=%s", cfg.Env, cfg.DBPath)
	log.Printf("API listening on %s", cfg.Port)
	http.ListenAndServe(cfg.Port, r)
}
