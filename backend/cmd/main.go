package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/httprate"
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
	r.Use(httprate.Limit(
		10,             // requests
		10*time.Second, // per duration
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			logger.ErrorContext(r.Context(), "Rate limit exceeded",
				"path", r.URL.Path,
			)
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		}),
	))

	// Register routes
	httpx.RegisterAllRouters(r, &app.App{DB: dbConn})

	logger.Info("init",
		"db", cfg.DBPath,
		"port", cfg.Port,
	)

	if err := http.ListenAndServe(cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
