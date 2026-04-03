package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/httprate"
	"github.com/go-chi/traceid"
	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/config"
	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/httpx"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/logging"
	authsvc "github.com/viktorHadz/goInvoice26/internal/service/auth"
	billingsvc "github.com/viktorHadz/goInvoice26/internal/service/billing"
	"github.com/viktorHadz/goInvoice26/internal/service/logo"
	"github.com/viktorHadz/goInvoice26/internal/service/storage"
	"github.com/viktorHadz/goInvoice26/internal/service/workspace"
	"github.com/viktorHadz/goInvoice26/internal/transaction/accessTx"
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
	if err := accessTx.SweepExpiredPromoRedemptionClaims(ctx, dbConn, time.Now()); err != nil {
		log.Fatal(err)
	}

	logoStore := storage.NewLocalStore(storage.DefaultRootDir)
	logoService := logo.NewService(dbConn, logoStore)
	billingService := billingsvc.NewService(dbConn, billingsvc.Config{
		AppBaseURL:                 cfg.AppBaseURL,
		StripeSecretKey:            cfg.StripeSecretKey,
		StripeSingleMonthlyPriceID: cfg.StripeSingleMonthlyPriceID,
		StripeSingleYearlyPriceID:  cfg.StripeSingleYearlyPriceID,
		StripeTeamMonthlyPriceID:   cfg.StripeTeamMonthlyPriceID,
		StripeTeamYearlyPriceID:    cfg.StripeTeamYearlyPriceID,
		StripeTrialDays:            cfg.StripeTrialDays,
		StripeWebhookSecret:        cfg.StripeWebhookSecret,
	})
	authService := authsvc.NewService(dbConn, authsvc.Config{
		AppBaseURL:                  cfg.AppBaseURL,
		GoogleClientID:              cfg.GoogleClientID,
		GoogleClientSecret:          cfg.GoogleClientSecret,
		GoogleRedirectURL:           cfg.GoogleRedirectURL,
		SessionCookieName:           cfg.SessionCookieName,
		SecureCookies:               cfg.SecureCookies(),
		BillingConfigured:           billingService.Configured(),
		BillingTrialDays:            cfg.StripeTrialDays,
		BillingSingleMonthlyPriceID: cfg.StripeSingleMonthlyPriceID,
		BillingSingleYearlyPriceID:  cfg.StripeSingleYearlyPriceID,
		BillingTeamMonthlyPriceID:   cfg.StripeTeamMonthlyPriceID,
		BillingTeamYearlyPriceID:    cfg.StripeTeamYearlyPriceID,
		PlatformAdminEmail:          cfg.PlatformAdminEmail,
	})
	if err := logoService.CleanupTemp(); err != nil {
		log.Fatal(err)
	}
	if err := logoService.MigrateLegacyLogo(ctx, accountscope.DefaultAccountID); err != nil {
		log.Fatal(err)
	}
	if err := logoService.SweepPendingDeletes(ctx); err != nil {
		log.Fatal(err)
	}
	workspaceService := workspace.NewService(dbConn, billingService, logoStore)

	r := chi.NewRouter()

	logger, opts := logging.InitLogger(cfg)

	if err := billingService.BackfillPersistedSelections(ctx); err != nil {
		logger.Warn("billing selection backfill failed", "err", err)
	}

	// Request-scoped tracing / logging
	r.Use(traceid.Middleware)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(30 * time.Second))

	// Request logger
	r.Use(httplog.RequestLogger(logger, opts))

	// Rate limit
	r.Use(httprate.Limit(
		50,
		10*time.Second,
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			logger.WarnContext(r.Context(), "rate limit exceeded", "path", r.URL.Path)
			res.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", "Too many requests")
		}),
	))

	httpx.RegisterAllRouters(r, &app.App{
		DB:                           dbConn,
		Auth:                         authService,
		Billing:                      billingService,
		Logos:                        logoService,
		Workspaces:                   workspaceService,
		AccessLedgerSecret:           cfg.AccessLedgerSecret,
		PromoRedemptionRetentionDays: cfg.PromoRedemptionRetentionDays,
	})

	logger.Info("init",
		"env", cfg.Env,
		"db", cfg.DBPath,
		"port", cfg.Port,
		"billingConfigured", billingService.Configured(),
		"billingWebhooksConfigured", billingService.WebhooksConfigured(),
		"hasStripeSecretKey", strings.TrimSpace(cfg.StripeSecretKey) != "",
		"hasStripeSingleMonthlyPriceID", strings.TrimSpace(cfg.StripeSingleMonthlyPriceID) != "",
		"hasStripeSingleYearlyPriceID", strings.TrimSpace(cfg.StripeSingleYearlyPriceID) != "",
		"hasStripeTeamMonthlyPriceID", strings.TrimSpace(cfg.StripeTeamMonthlyPriceID) != "",
		"hasStripeTeamYearlyPriceID", strings.TrimSpace(cfg.StripeTeamYearlyPriceID) != "",
		"stripeTrialDays", cfg.StripeTrialDays,
		"hasStripeWebhookSecret", strings.TrimSpace(cfg.StripeWebhookSecret) != "",
		"hasAccessLedgerSecret", strings.TrimSpace(cfg.AccessLedgerSecret) != "",
		"promoRedemptionRetentionDays", cfg.PromoRedemptionRetentionDays,
	)

	if err := http.ListenAndServe(cfg.Port, r); err != nil {
		logger.Error("startup failed", "err", err)
		os.Exit(1)
	}
}
