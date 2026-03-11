package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/httprate"
	"github.com/go-chi/traceid"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/config"
	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/httpx"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/logging"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	dbConn, err := db.OpenDB(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	if err := db.Migrate(ctx, dbConn); err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()

	logger, opts := logging.InitLogger(cfg)

	// Request-scoped tracing / logging
	r.Use(traceid.Middleware)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(30 * time.Second))

	// Request logger
	r.Use(httplog.RequestLogger(logger, opts))

	// Rate limit
	r.Use(httprate.Limit(
		10,
		10*time.Second,
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			logger.WarnContext(r.Context(), "rate limit exceeded", "path", r.URL.Path)
			res.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests")
		}),
	))

	httpx.RegisterAllRouters(r, &app.App{DB: dbConn})

	logger.Info("init",
		"env", cfg.Env,
		"db", cfg.DBPath,
		"port", cfg.Port,
	)

	if err := http.ListenAndServe(cfg.Port, r); err != nil {
		logger.Error("startup failed", "err", err)
		os.Exit(1)
	}
}
