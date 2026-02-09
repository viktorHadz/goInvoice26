package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/traceid"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/config"
	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/httpx"
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
	r.Use(traceid.Middleware)
	// logger.Log(ctx, slog.LevelInfo, "Testing logger") |-> usecase example in main
	// slog.InfoContext(r.Context(), "All clients requested") |-> usecase globaly
	r.Use(httplog.RequestLogger(logger, opts))

	// Register routes
	httpx.RegisterAllRouters(r, &app.App{DB: dbConn})
	logger.Log(ctx, slog.LevelInfo, "Init:", "ENV:", cfg.Env, "DB:", cfg.DBPath, "API Listening on:", cfg.Port)
	http.ListenAndServe(cfg.Port, r)
}
