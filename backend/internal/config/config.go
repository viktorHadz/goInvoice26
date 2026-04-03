package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                          string
	Port                         string
	DBPath                       string
	CORSOrigin                   string
	AppBaseURL                   string
	SessionCookieName            string
	GoogleClientID               string
	GoogleClientSecret           string
	GoogleRedirectURL            string
	StripeSecretKey              string
	StripePublishableKey         string
	StripeSingleMonthlyPriceID   string
	StripeSingleYearlyPriceID    string
	StripeTeamMonthlyPriceID     string
	StripeTeamYearlyPriceID      string
	StripeTrialDays              int
	StripeWebhookSecret          string
	PlatformAdminEmail           string
	AccessLedgerSecret           string
	PromoRedemptionRetentionDays int
}

func Load() (Config, error) {
	// Dev convenience | In prod, env vars come from systemd/docker
	_ = godotenv.Load()

	stripeTrialDays, err := getInt("STRIPE_TRIAL_DAYS", 7)
	if err != nil {
		return Config{}, err
	}
	legacyStripePriceID := get("STRIPE_PRICE_ID", "")
	legacySinglePriceID := get("STRIPE_SINGLE_PRICE_ID", legacyStripePriceID)
	legacyTeamPriceID := get("STRIPE_TEAM_PRICE_ID", "")

	cfg := Config{
		Env:                        get("ENV", "dev"),
		Port:                       get("PORT", "4206"),
		DBPath:                     must("DB_PATH"),
		CORSOrigin:                 get("CORS_ORIGIN", "http://localhost:5173"),
		AppBaseURL:                 get("APP_BASE_URL", get("CORS_ORIGIN", "http://localhost:5173")),
		SessionCookieName:          get("SESSION_COOKIE_NAME", "invoicer_session"),
		GoogleClientID:             must("GOOGLE_CLIENT_ID"),
		GoogleClientSecret:         must("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:          must("GOOGLE_REDIRECT_URL"),
		StripeSecretKey:            get("STRIPE_SECRET_KEY", ""),
		StripePublishableKey:       get("STRIPE_PUBLISHABLE_KEY", ""),
		StripeSingleMonthlyPriceID: get("STRIPE_SINGLE_MONTHLY_PRICE_ID", legacySinglePriceID),
		StripeSingleYearlyPriceID:  get("STRIPE_SINGLE_YEARLY_PRICE_ID", ""),
		StripeTeamMonthlyPriceID:   get("STRIPE_TEAM_MONTHLY_PRICE_ID", legacyTeamPriceID),
		StripeTeamYearlyPriceID:    get("STRIPE_TEAM_YEARLY_PRICE_ID", ""),
		StripeTrialDays:            stripeTrialDays,
		StripeWebhookSecret:        get("STRIPE_WEBHOOK_SECRET", ""),
		PlatformAdminEmail:         get("PLATFORM_ADMIN_EMAIL", ""),
		AccessLedgerSecret:         get("ACCESS_LEDGER_SECRET", ""),
	}

	promoRedemptionRetentionDays, err := getInt("PROMO_REDEMPTION_RETENTION_DAYS", 180)
	if err != nil {
		return Config{}, err
	}
	cfg.PromoRedemptionRetentionDays = promoRedemptionRetentionDays

	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Retrieves an .env key and returns empty string if the key is wrong
// Similar to get but doesnt provide fallback
func must(key string) string {
	v := os.Getenv(key)
	if v == "" {
		// return error via validate path. But need a placeholder value here
		return ""
	}
	return v
}

// Gets an environment variable. Must provide fallback param.
func get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid value for %s: %w", key, err)
	}

	return parsed, nil
}

// Helper that validates env retrieval
func validate(cfg Config) error {
	if cfg.DBPath == "" {
		return fmt.Errorf("missing required env var: DB_PATH")
	}
	if cfg.Port == "" {
		return fmt.Errorf("missing required env var: PORT")
	}
	if cfg.StripeTrialDays < 0 {
		return fmt.Errorf("STRIPE_TRIAL_DAYS must be greater than or equal to 0")
	}
	if cfg.PromoRedemptionRetentionDays < 0 {
		return fmt.Errorf("PROMO_REDEMPTION_RETENTION_DAYS must be greater than or equal to 0")
	}
	return nil
}

func (c Config) SecureCookies() bool {
	return c.Env == "prod" || c.Env == "production"
}

func (c Config) GoogleOAuthEnabled() bool {
	return c.GoogleClientID != "" && c.GoogleClientSecret != "" && c.GoogleRedirectURL != ""
}
