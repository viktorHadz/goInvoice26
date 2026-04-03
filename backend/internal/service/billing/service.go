package billing

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/viktorHadz/goInvoice26/internal/billingcatalog"
	"github.com/viktorHadz/goInvoice26/internal/billingplan"
	"github.com/viktorHadz/goInvoice26/internal/billingstate"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/billingTx"
)

var (
	ErrNotConfigured        = errors.New("billing is not configured")
	ErrOwnerOnly            = errors.New("billing management is owner-only")
	ErrCustomerNotFound     = errors.New("stripe customer not found for account")
	ErrInvalidCheckoutSync  = errors.New("checkout session is not valid for this account")
	ErrCheckoutPending      = errors.New("checkout session is still confirming")
	ErrWebhookSignature     = errors.New("invalid stripe webhook signature")
	ErrSubscriptionNotFound = errors.New("stripe subscription not found for account")
	ErrInvalidPlan          = errors.New("invalid billing plan")
	ErrInvalidInterval      = errors.New("invalid billing interval")
	ErrPlanUnavailable      = errors.New("billing plan is not available")
	ErrPlanAlreadyActive    = errors.New("billing plan is already active")
)

type Config struct {
	AppBaseURL                 string
	StripeSecretKey            string
	StripeSingleMonthlyPriceID string
	StripeSingleYearlyPriceID  string
	StripeTeamMonthlyPriceID   string
	StripeTeamYearlyPriceID    string
	StripeTrialDays            int
	StripeWebhookSecret        string
	APIBaseURL                 string
	HTTPClient                 *http.Client
}

type Service struct {
	db                  *sql.DB
	appBaseURL          string
	stripeSecretKey     string
	billingCatalog      billingcatalog.Config
	stripeTrialDays     int
	stripeWebhookSecret string
	apiBaseURL          string
	httpClient          *http.Client
	publicCatalogMu     sync.RWMutex
	publicCatalog       models.PublicBillingCatalog
	publicCatalogAt     time.Time
}

type stripeCheckoutSession struct {
	ID                string            `json:"id"`
	URL               string            `json:"url"`
	Mode              string            `json:"mode"`
	Status            string            `json:"status"`
	PaymentStatus     string            `json:"payment_status"`
	Customer          string            `json:"customer"`
	CustomerEmail     string            `json:"customer_email"`
	Subscription      string            `json:"subscription"`
	ClientReferenceID string            `json:"client_reference_id"`
	Metadata          map[string]string `json:"metadata"`
}

type stripeSubscription struct {
	ID                string            `json:"id"`
	Customer          string            `json:"customer"`
	Status            string            `json:"status"`
	CancelAtPeriodEnd bool              `json:"cancel_at_period_end"`
	CurrentPeriodEnd  int64             `json:"current_period_end"`
	Metadata          map[string]string `json:"metadata"`
	Items             struct {
		Data []struct {
			ID    string `json:"id"`
			Price struct {
				ID string `json:"id"`
			} `json:"price"`
		} `json:"data"`
	} `json:"items"`
}

type stripeInvoice struct {
	Customer     string `json:"customer"`
	Subscription string `json:"subscription"`
}

type stripePrice struct {
	ID         string `json:"id"`
	Currency   string `json:"currency"`
	UnitAmount int64  `json:"unit_amount"`
	Recurring  struct {
		Interval string `json:"interval"`
	} `json:"recurring"`
}

type stripeEvent struct {
	Type string `json:"type"`
	Data struct {
		Object json.RawMessage `json:"object"`
	} `json:"data"`
}

const publicCatalogCacheTTL = 10 * time.Minute

func NewService(db *sql.DB, cfg Config) *Service {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	apiBaseURL := strings.TrimRight(strings.TrimSpace(cfg.APIBaseURL), "/")
	if apiBaseURL == "" {
		apiBaseURL = "https://api.stripe.com"
	}

	return &Service{
		db:              db,
		appBaseURL:      strings.TrimRight(strings.TrimSpace(cfg.AppBaseURL), "/"),
		stripeSecretKey: strings.TrimSpace(cfg.StripeSecretKey),
		billingCatalog: billingcatalog.Config{
			SingleMonthlyPriceID: strings.TrimSpace(cfg.StripeSingleMonthlyPriceID),
			SingleYearlyPriceID:  strings.TrimSpace(cfg.StripeSingleYearlyPriceID),
			TeamMonthlyPriceID:   strings.TrimSpace(cfg.StripeTeamMonthlyPriceID),
			TeamYearlyPriceID:    strings.TrimSpace(cfg.StripeTeamYearlyPriceID),
		},
		stripeTrialDays:     normalizeTrialDays(cfg.StripeTrialDays),
		stripeWebhookSecret: strings.TrimSpace(cfg.StripeWebhookSecret),
		apiBaseURL:          apiBaseURL,
		httpClient:          httpClient,
	}
}

func (s *Service) Configured() bool {
	return s.stripeSecretKey != "" && billingcatalog.AnyConfigured(s.billingCatalog)
}

func (s *Service) WebhooksConfigured() bool {
	return s.stripeSecretKey != "" && s.stripeWebhookSecret != ""
}

func (s *Service) PublicCatalog(ctx context.Context) (models.PublicBillingCatalog, error) {
	baseCatalog := models.PublicBillingCatalog{
		Configured:             s.Configured(),
		TrialDays:              s.stripeTrialDays,
		SingleMonthlyAvailable: billingcatalog.IntervalAvailable(billingplan.PlanSingle, billingcatalog.IntervalMonthly, s.billingCatalog),
		SingleYearlyAvailable:  billingcatalog.IntervalAvailable(billingplan.PlanSingle, billingcatalog.IntervalYearly, s.billingCatalog),
		TeamMonthlyAvailable:   billingcatalog.IntervalAvailable(billingplan.PlanTeam, billingcatalog.IntervalMonthly, s.billingCatalog),
		TeamYearlyAvailable:    billingcatalog.IntervalAvailable(billingplan.PlanTeam, billingcatalog.IntervalYearly, s.billingCatalog),
	}
	if !baseCatalog.Configured {
		return baseCatalog, nil
	}

	if cached, ok := s.cachedPublicCatalog(baseCatalog); ok {
		return cached, nil
	}

	catalog, err := s.loadPublicCatalog(ctx, baseCatalog)
	if err == nil {
		s.cachePublicCatalog(catalog)
		return catalog, nil
	}

	if cached, ok := s.cachedPublicCatalog(baseCatalog); ok {
		slog.WarnContext(ctx, "billing public catalog refresh failed, falling back to cached catalog", "err", err)
		return cached, nil
	}

	slog.WarnContext(ctx, "billing public catalog refresh failed, returning availability-only catalog", "err", err)
	return baseCatalog, nil
}

func (s *Service) BackfillPersistedSelections(ctx context.Context) error {
	if !s.Configured() {
		return nil
	}

	records, err := billingTx.ListAccountsMissingBillingSelection(ctx, s.db)
	if err != nil {
		return err
	}

	for _, record := range records {
		subscription, err := s.retrieveSubscription(ctx, record.StripeSubscriptionID)
		if err != nil {
			slog.WarnContext(
				ctx,
				"billing selection backfill failed to load subscription",
				"accountID", record.AccountID,
				"subscriptionID", shortStripeID(record.StripeSubscriptionID),
				"err", err,
			)
			continue
		}

		if err := s.applySubscription(ctx, record.AccountID, subscription, record.BillingEmail); err != nil {
			slog.WarnContext(
				ctx,
				"billing selection backfill failed to persist subscription",
				"accountID", record.AccountID,
				"subscriptionID", shortStripeID(record.StripeSubscriptionID),
				"err", err,
			)
		}
	}

	return nil
}

func (s *Service) CreateCheckoutSession(
	ctx context.Context,
	accountID int64,
	accountName string,
	customerEmail string,
	plan string,
	interval string,
	redirectPath string,
) (models.BillingSessionLink, error) {
	if !s.Configured() {
		slog.WarnContext(ctx, "billing checkout requested but service is not configured", "accountID", accountID)
		return models.BillingSessionLink{}, ErrNotConfigured
	}

	record, err := billingTx.GetAccountBilling(ctx, s.db, accountID)
	if err != nil {
		return models.BillingSessionLink{}, err
	}
	selection, err := s.resolveCheckoutSelection(plan, interval)
	if err != nil {
		return models.BillingSessionLink{}, err
	}
	priceID := billingcatalog.PriceIDFor(selection.Plan, selection.Interval, s.billingCatalog)
	if strings.TrimSpace(priceID) == "" {
		return models.BillingSessionLink{}, ErrPlanUnavailable
	}

	slog.InfoContext(
		ctx,
		"billing checkout session requested",
		"accountID", accountID,
		"hasCustomer", strings.TrimSpace(record.StripeCustomerID) != "",
		"hasEmail", strings.TrimSpace(customerEmail) != "",
		"priceID", shortStripeID(priceID),
		"plan", selection.Plan,
		"interval", selection.Interval,
		"trialDays", s.stripeTrialDays,
	)

	sanitizedRedirectPath := sanitizeRedirectPath(redirectPath)

	form := url.Values{}
	form.Set("mode", "subscription")
	form.Set(
		"success_url",
		s.appURL("/app/billing?checkout=success&session_id={CHECKOUT_SESSION_ID}&redirect="+url.QueryEscape(sanitizedRedirectPath)),
	)
	form.Set("cancel_url", s.appURL("/app/billing?checkout=canceled&redirect="+url.QueryEscape(sanitizedRedirectPath)))
	form.Set("payment_method_collection", "always")
	form.Set("line_items[0][price]", priceID)
	form.Set("line_items[0][quantity]", "1")
	if s.stripeTrialDays > 0 {
		form.Set("subscription_data[trial_period_days]", strconv.Itoa(s.stripeTrialDays))
		form.Set("subscription_data[trial_settings][end_behavior][missing_payment_method]", "cancel")
	}
	form.Set("client_reference_id", strconv.FormatInt(accountID, 10))
	form.Set("metadata[account_id]", strconv.FormatInt(accountID, 10))
	form.Set("metadata[account_name]", strings.TrimSpace(accountName))
	form.Set("metadata[plan]", selection.Plan)
	form.Set("metadata[interval]", selection.Interval)
	form.Set("subscription_data[metadata][account_id]", strconv.FormatInt(accountID, 10))
	form.Set("subscription_data[metadata][plan]", selection.Plan)
	form.Set("subscription_data[metadata][interval]", selection.Interval)
	if strings.TrimSpace(accountName) != "" {
		form.Set("subscription_data[metadata][account_name]", strings.TrimSpace(accountName))
	}
	if strings.TrimSpace(record.StripeCustomerID) != "" {
		form.Set("customer", strings.TrimSpace(record.StripeCustomerID))
	} else if strings.TrimSpace(customerEmail) != "" {
		form.Set("customer_email", strings.TrimSpace(customerEmail))
	}

	var session stripeCheckoutSession
	if err := s.doFormRequest(ctx, http.MethodPost, "/v1/checkout/sessions", form, &session); err != nil {
		return models.BillingSessionLink{}, err
	}
	if strings.TrimSpace(session.URL) == "" {
		return models.BillingSessionLink{}, errors.New("stripe checkout session missing url")
	}

	slog.InfoContext(
		ctx,
		"billing checkout session created",
		"accountID", accountID,
		"checkoutSessionID", shortStripeID(session.ID),
		"customerID", shortStripeID(session.Customer),
	)

	return models.BillingSessionLink{URL: session.URL}, nil
}

func (s *Service) ChangeSubscriptionPlan(ctx context.Context, accountID int64, plan string, interval string) error {
	if !s.Configured() {
		slog.WarnContext(ctx, "billing plan change requested but service is not configured", "accountID", accountID)
		return ErrNotConfigured
	}

	record, err := billingTx.GetAccountBilling(ctx, s.db, accountID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(record.StripeSubscriptionID) == "" {
		return ErrSubscriptionNotFound
	}
	currentSelection := billingcatalog.NormalizeSelection(record.BillingPlan, record.BillingInterval)
	if currentSelection.Plan == "" || currentSelection.Interval == "" {
		selectionFromPriceID := billingcatalog.DetermineFromPriceID(record.BillingPriceID, s.billingCatalog)
		if currentSelection.Plan == "" {
			currentSelection.Plan = selectionFromPriceID.Plan
		}
		if currentSelection.Interval == "" {
			currentSelection.Interval = selectionFromPriceID.Interval
		}
	}
	targetSelection, err := s.resolveChangeSelection(currentSelection, plan, interval)
	if err != nil {
		return err
	}
	if currentSelection == targetSelection {
		return ErrPlanAlreadyActive
	}
	priceID := billingcatalog.PriceIDFor(targetSelection.Plan, targetSelection.Interval, s.billingCatalog)

	subscription, err := s.retrieveSubscription(ctx, record.StripeSubscriptionID)
	if err != nil {
		return err
	}
	itemID := firstSubscriptionItemID(subscription)
	if itemID == "" {
		return errors.New("stripe subscription missing updatable item")
	}

	form := url.Values{}
	form.Set("items[0][id]", itemID)
	form.Set("items[0][price]", priceID)
	form.Set("proration_behavior", "create_prorations")
	form.Set("cancel_at_period_end", "false")
	form.Set("metadata[plan]", targetSelection.Plan)
	form.Set("metadata[interval]", targetSelection.Interval)

	updatedSubscription, err := s.updateSubscription(ctx, record.StripeSubscriptionID, form)
	if err != nil {
		return err
	}

	return s.applySubscription(ctx, accountID, updatedSubscription, record.BillingEmail)
}

func (s *Service) CreatePortalSession(ctx context.Context, accountID int64) (models.BillingSessionLink, error) {
	if !s.Configured() {
		slog.WarnContext(ctx, "billing portal requested but service is not configured", "accountID", accountID)
		return models.BillingSessionLink{}, ErrNotConfigured
	}

	record, err := billingTx.GetAccountBilling(ctx, s.db, accountID)
	if err != nil {
		return models.BillingSessionLink{}, err
	}
	if strings.TrimSpace(record.StripeCustomerID) == "" {
		slog.WarnContext(ctx, "billing portal requested without stripe customer", "accountID", accountID)
		return models.BillingSessionLink{}, ErrCustomerNotFound
	}

	slog.InfoContext(
		ctx,
		"billing portal session requested",
		"accountID", accountID,
		"customerID", shortStripeID(record.StripeCustomerID),
	)

	form := url.Values{}
	form.Set("customer", record.StripeCustomerID)
	form.Set("return_url", s.appURL("/app/billing"))

	var portal struct {
		URL string `json:"url"`
	}
	if err := s.doFormRequest(ctx, http.MethodPost, "/v1/billing_portal/sessions", form, &portal); err != nil {
		return models.BillingSessionLink{}, err
	}
	if strings.TrimSpace(portal.URL) == "" {
		return models.BillingSessionLink{}, errors.New("stripe portal session missing url")
	}

	slog.InfoContext(ctx, "billing portal session created", "accountID", accountID)

	return models.BillingSessionLink{URL: portal.URL}, nil
}

func (s *Service) SyncCheckoutSession(ctx context.Context, accountID int64, sessionID string) error {
	if !s.Configured() {
		slog.WarnContext(ctx, "billing checkout sync requested but service is not configured", "accountID", accountID)
		return ErrNotConfigured
	}

	slog.InfoContext(
		ctx,
		"billing checkout sync started",
		"accountID", accountID,
		"checkoutSessionID", shortStripeID(sessionID),
	)

	session, err := s.retrieveCheckoutSession(ctx, sessionID)
	if err != nil {
		return err
	}
	slog.DebugContext(
		ctx,
		"billing checkout session loaded",
		"accountID", accountID,
		"checkoutSessionID", shortStripeID(session.ID),
		"checkoutStatus", session.Status,
		"paymentStatus", session.PaymentStatus,
		"customerID", shortStripeID(session.Customer),
		"subscriptionID", shortStripeID(session.Subscription),
	)
	if !s.checkoutBelongsToAccount(accountID, session) {
		slog.WarnContext(
			ctx,
			"billing checkout sync rejected due to account mismatch",
			"accountID", accountID,
			"checkoutSessionID", shortStripeID(session.ID),
			"clientReferenceID", session.ClientReferenceID,
			"metadataAccountID", session.Metadata["account_id"],
		)
		return ErrInvalidCheckoutSync
	}

	if strings.TrimSpace(session.Customer) != "" || strings.TrimSpace(session.Subscription) != "" {
		if err := billingTx.LinkStripeIdentifiers(
			ctx,
			s.db,
			accountID,
			session.Customer,
			session.Subscription,
			session.CustomerEmail,
			time.Now(),
		); err != nil {
			return err
		}
	}

	if session.Status == "open" || strings.TrimSpace(session.Subscription) == "" {
		slog.InfoContext(
			ctx,
			"billing checkout sync is waiting for Stripe to finish provisioning subscription",
			"accountID", accountID,
			"checkoutSessionID", shortStripeID(session.ID),
			"checkoutStatus", session.Status,
			"paymentStatus", session.PaymentStatus,
			"subscriptionID", shortStripeID(session.Subscription),
		)
		return ErrCheckoutPending
	}

	subscription, err := s.retrieveSubscription(ctx, session.Subscription)
	if err != nil {
		return err
	}

	if err := s.applySubscription(ctx, accountID, subscription, session.CustomerEmail); err != nil {
		return err
	}

	slog.InfoContext(
		ctx,
		"billing checkout sync completed",
		"accountID", accountID,
		"checkoutSessionID", shortStripeID(session.ID),
		"subscriptionID", shortStripeID(subscription.ID),
		"subscriptionStatus", subscription.Status,
	)

	return nil
}

func (s *Service) CancelSubscriptionAtPeriodEnd(ctx context.Context, accountID int64) error {
	if !s.Configured() {
		slog.WarnContext(ctx, "billing cancellation requested but service is not configured", "accountID", accountID)
		return ErrNotConfigured
	}

	record, err := billingTx.GetAccountBilling(ctx, s.db, accountID)
	if err != nil {
		return err
	}

	subscriptionID := strings.TrimSpace(record.StripeSubscriptionID)
	if subscriptionID == "" {
		return ErrSubscriptionNotFound
	}
	if record.BillingCancelAtPeriodEnd {
		return nil
	}

	slog.InfoContext(ctx, "billing cancel-at-period-end requested", "accountID", accountID, "subscriptionID", shortStripeID(subscriptionID))

	form := url.Values{}
	form.Set("cancel_at_period_end", "true")

	subscription, err := s.updateSubscription(ctx, subscriptionID, form)
	if err != nil {
		return err
	}

	return s.applySubscription(ctx, accountID, subscription, record.BillingEmail)
}

func (s *Service) CancelSubscriptionImmediately(ctx context.Context, accountID int64) error {
	record, err := billingTx.GetAccountBilling(ctx, s.db, accountID)
	if err != nil {
		return err
	}

	subscriptionID := strings.TrimSpace(record.StripeSubscriptionID)
	if subscriptionID == "" {
		return nil
	}
	if !s.Configured() {
		slog.WarnContext(ctx, "immediate billing cancellation requested but service is not configured", "accountID", accountID)
		return ErrNotConfigured
	}

	slog.InfoContext(ctx, "billing immediate cancellation requested", "accountID", accountID, "subscriptionID", shortStripeID(subscriptionID))

	subscription, err := s.cancelSubscription(ctx, subscriptionID)
	if err != nil {
		return err
	}

	return s.applySubscription(ctx, accountID, subscription, record.BillingEmail)
}

func (s *Service) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	if !s.WebhooksConfigured() {
		slog.WarnContext(ctx, "stripe webhook received but billing webhooks are not configured")
		return ErrNotConfigured
	}
	if err := verifyStripeWebhookSignature(payload, signature, s.stripeWebhookSecret, 5*time.Minute); err != nil {
		slog.WarnContext(ctx, "stripe webhook signature verification failed", "err", err)
		return err
	}

	var event stripeEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return fmt.Errorf("decode stripe webhook: %w", err)
	}

	slog.InfoContext(ctx, "stripe webhook received", "eventType", event.Type)

	switch event.Type {
	case "checkout.session.completed":
		var session stripeCheckoutSession
		if err := json.Unmarshal(event.Data.Object, &session); err != nil {
			return fmt.Errorf("decode checkout.session.completed: %w", err)
		}
		if err := s.applyCheckoutCompletion(ctx, session); err != nil {
			return err
		}
		slog.InfoContext(
			ctx,
			"stripe webhook processed",
			"eventType", event.Type,
			"checkoutSessionID", shortStripeID(session.ID),
			"customerID", shortStripeID(session.Customer),
			"subscriptionID", shortStripeID(session.Subscription),
		)
		return nil
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		var subscription stripeSubscription
		if err := json.Unmarshal(event.Data.Object, &subscription); err != nil {
			return fmt.Errorf("decode %s: %w", event.Type, err)
		}
		accountID, ok, err := s.resolveAccountIDForSubscription(ctx, subscription)
		if err != nil {
			return err
		}
		if !ok {
			slog.WarnContext(
				ctx,
				"stripe subscription webhook could not be matched to an account",
				"eventType", event.Type,
				"subscriptionID", shortStripeID(subscription.ID),
				"customerID", shortStripeID(subscription.Customer),
			)
			return nil
		}
		if err := s.applySubscription(ctx, accountID, subscription, ""); err != nil {
			return err
		}
		slog.InfoContext(
			ctx,
			"stripe subscription webhook applied",
			"eventType", event.Type,
			"accountID", accountID,
			"subscriptionID", shortStripeID(subscription.ID),
			"subscriptionStatus", subscription.Status,
		)
		return nil
	case "invoice.paid", "invoice.payment_failed":
		var invoice stripeInvoice
		if err := json.Unmarshal(event.Data.Object, &invoice); err != nil {
			return fmt.Errorf("decode %s: %w", event.Type, err)
		}
		if strings.TrimSpace(invoice.Subscription) == "" {
			return nil
		}
		subscription, err := s.retrieveSubscription(ctx, invoice.Subscription)
		if err != nil {
			return err
		}
		accountID, ok, err := s.resolveAccountIDForSubscription(ctx, subscription)
		if err != nil {
			return err
		}
		if !ok {
			slog.WarnContext(
				ctx,
				"stripe invoice webhook could not be matched to an account",
				"eventType", event.Type,
				"subscriptionID", shortStripeID(invoice.Subscription),
				"customerID", shortStripeID(invoice.Customer),
			)
			return nil
		}
		if err := s.applySubscription(ctx, accountID, subscription, ""); err != nil {
			return err
		}
		slog.InfoContext(
			ctx,
			"stripe invoice webhook applied subscription state",
			"eventType", event.Type,
			"accountID", accountID,
			"subscriptionID", shortStripeID(subscription.ID),
			"subscriptionStatus", subscription.Status,
		)
		return nil
	default:
		slog.DebugContext(ctx, "stripe webhook ignored", "eventType", event.Type)
		return nil
	}
}

func (s *Service) retrieveCheckoutSession(ctx context.Context, sessionID string) (stripeCheckoutSession, error) {
	var session stripeCheckoutSession
	if err := s.doRequest(ctx, http.MethodGet, "/v1/checkout/sessions/"+url.PathEscape(strings.TrimSpace(sessionID)), nil, &session); err != nil {
		return stripeCheckoutSession{}, err
	}

	return session, nil
}

func (s *Service) retrieveSubscription(ctx context.Context, subscriptionID string) (stripeSubscription, error) {
	var subscription stripeSubscription
	if err := s.doRequest(ctx, http.MethodGet, "/v1/subscriptions/"+url.PathEscape(strings.TrimSpace(subscriptionID)), nil, &subscription); err != nil {
		return stripeSubscription{}, err
	}

	return subscription, nil
}

func (s *Service) retrievePrice(ctx context.Context, priceID string) (stripePrice, error) {
	var price stripePrice
	if err := s.doRequest(ctx, http.MethodGet, "/v1/prices/"+url.PathEscape(strings.TrimSpace(priceID)), nil, &price); err != nil {
		return stripePrice{}, err
	}

	return price, nil
}

func (s *Service) updateSubscription(ctx context.Context, subscriptionID string, form url.Values) (stripeSubscription, error) {
	var subscription stripeSubscription
	if err := s.doFormRequest(
		ctx,
		http.MethodPost,
		"/v1/subscriptions/"+url.PathEscape(strings.TrimSpace(subscriptionID)),
		form,
		&subscription,
	); err != nil {
		return stripeSubscription{}, err
	}
	return subscription, nil
}

func (s *Service) cachedPublicCatalog(baseCatalog models.PublicBillingCatalog) (models.PublicBillingCatalog, bool) {
	s.publicCatalogMu.RLock()
	defer s.publicCatalogMu.RUnlock()

	if s.publicCatalogAt.IsZero() || time.Since(s.publicCatalogAt) > publicCatalogCacheTTL {
		return models.PublicBillingCatalog{}, false
	}

	catalog := s.publicCatalog
	catalog.Configured = baseCatalog.Configured
	catalog.TrialDays = baseCatalog.TrialDays
	catalog.SingleMonthlyAvailable = baseCatalog.SingleMonthlyAvailable
	catalog.SingleYearlyAvailable = baseCatalog.SingleYearlyAvailable
	catalog.TeamMonthlyAvailable = baseCatalog.TeamMonthlyAvailable
	catalog.TeamYearlyAvailable = baseCatalog.TeamYearlyAvailable
	return catalog, true
}

func (s *Service) cachePublicCatalog(catalog models.PublicBillingCatalog) {
	s.publicCatalogMu.Lock()
	defer s.publicCatalogMu.Unlock()

	s.publicCatalog = catalog
	s.publicCatalogAt = time.Now()
}

func (s *Service) loadPublicCatalog(
	ctx context.Context,
	baseCatalog models.PublicBillingCatalog,
) (models.PublicBillingCatalog, error) {
	catalog := baseCatalog

	var loadErr error
	loadPrice := func(priceID, interval string, assign func(string)) {
		if loadErr != nil || strings.TrimSpace(priceID) == "" {
			return
		}

		price, err := s.retrievePrice(ctx, priceID)
		if err != nil {
			loadErr = err
			return
		}
		assign(formatStripePriceLabel(price.Currency, price.UnitAmount, interval))
	}

	loadPrice(s.billingCatalog.SingleMonthlyPriceID, billingcatalog.IntervalMonthly, func(label string) {
		catalog.SingleMonthlyPriceLabel = label
	})
	loadPrice(s.billingCatalog.SingleYearlyPriceID, billingcatalog.IntervalYearly, func(label string) {
		catalog.SingleYearlyPriceLabel = label
	})
	loadPrice(s.billingCatalog.TeamMonthlyPriceID, billingcatalog.IntervalMonthly, func(label string) {
		catalog.TeamMonthlyPriceLabel = label
	})
	loadPrice(s.billingCatalog.TeamYearlyPriceID, billingcatalog.IntervalYearly, func(label string) {
		catalog.TeamYearlyPriceLabel = label
	})

	if loadErr != nil {
		return models.PublicBillingCatalog{}, loadErr
	}

	return catalog, nil
}

func (s *Service) cancelSubscription(ctx context.Context, subscriptionID string) (stripeSubscription, error) {
	var subscription stripeSubscription
	if err := s.doRequest(
		ctx,
		http.MethodDelete,
		"/v1/subscriptions/"+url.PathEscape(strings.TrimSpace(subscriptionID)),
		nil,
		&subscription,
	); err != nil {
		return stripeSubscription{}, err
	}
	return subscription, nil
}

func (s *Service) applyCheckoutCompletion(ctx context.Context, session stripeCheckoutSession) error {
	accountID, ok := billingTx.AccountIDFromString(session.ClientReferenceID)
	if !ok {
		accountID, ok = billingTx.AccountIDFromString(session.Metadata["account_id"])
	}
	if !ok {
		return nil
	}

	return billingTx.LinkStripeIdentifiers(
		ctx,
		s.db,
		accountID,
		session.Customer,
		session.Subscription,
		session.CustomerEmail,
		time.Now(),
	)
}

func (s *Service) applySubscription(
	ctx context.Context,
	accountID int64,
	subscription stripeSubscription,
	billingEmail string,
) error {
	selection := determineSubscriptionSelection(subscription, s.billingCatalog)

	var currentPeriodEnd *time.Time
	if subscription.CurrentPeriodEnd > 0 {
		ts := time.Unix(subscription.CurrentPeriodEnd, 0).UTC()
		currentPeriodEnd = &ts
	}

	return billingTx.UpdateAccountBilling(ctx, s.db, billingTx.UpdateAccountBillingParams{
		AccountID:                accountID,
		StripeCustomerID:         strings.TrimSpace(subscription.Customer),
		StripeSubscriptionID:     strings.TrimSpace(subscription.ID),
		BillingPriceID:           firstSubscriptionPriceID(subscription),
		BillingPlan:              selection.Plan,
		BillingInterval:          selection.Interval,
		BillingEmail:             strings.TrimSpace(billingEmail),
		BillingStatus:            billingstate.Normalize(subscription.Status),
		BillingCurrentPeriodEnd:  currentPeriodEnd,
		BillingCancelAtPeriodEnd: subscription.CancelAtPeriodEnd,
		BillingUpdatedAt:         time.Now(),
	})
}

func (s *Service) resolveAccountIDForSubscription(
	ctx context.Context,
	subscription stripeSubscription,
) (int64, bool, error) {
	if accountID, ok := billingTx.AccountIDFromString(subscription.Metadata["account_id"]); ok {
		return accountID, true, nil
	}

	record, ok, err := billingTx.FindAccountBillingBySubscriptionID(ctx, s.db, subscription.ID)
	if err != nil {
		return 0, false, err
	}
	if ok {
		return record.AccountID, true, nil
	}

	record, ok, err = billingTx.FindAccountBillingByCustomerID(ctx, s.db, subscription.Customer)
	if err != nil {
		return 0, false, err
	}
	if ok {
		return record.AccountID, true, nil
	}

	return 0, false, nil
}

func (s *Service) checkoutBelongsToAccount(accountID int64, session stripeCheckoutSession) bool {
	if sessionID, ok := billingTx.AccountIDFromString(session.ClientReferenceID); ok && sessionID == accountID {
		return true
	}
	if sessionID, ok := billingTx.AccountIDFromString(session.Metadata["account_id"]); ok && sessionID == accountID {
		return true
	}

	return false
}

func (s *Service) doFormRequest(ctx context.Context, method, path string, form url.Values, dst any) error {
	return s.doRequest(
		ctx,
		method,
		path,
		strings.NewReader(form.Encode()),
		dst,
		withHeader("Content-Type", "application/x-www-form-urlencoded"),
	)
}

func normalizeTrialDays(days int) int {
	if days < 0 {
		return 0
	}
	return days
}

func (s *Service) resolveCheckoutSelection(plan string, interval string) (billingcatalog.Selection, error) {
	selection := billingcatalog.DefaultCheckoutSelection(plan, interval, s.billingCatalog)
	if selection.Plan == "" {
		return billingcatalog.Selection{}, ErrInvalidPlan
	}
	if selection.Interval == "" {
		if billingcatalog.NormalizeInterval(interval) == "" && strings.TrimSpace(interval) != "" {
			return billingcatalog.Selection{}, ErrInvalidInterval
		}
		return billingcatalog.Selection{}, ErrPlanUnavailable
	}
	if !billingcatalog.IntervalAvailable(selection.Plan, selection.Interval, s.billingCatalog) {
		return billingcatalog.Selection{}, ErrPlanUnavailable
	}
	return selection, nil
}

func (s *Service) resolveChangeSelection(
	current billingcatalog.Selection,
	plan string,
	interval string,
) (billingcatalog.Selection, error) {
	normalizedPlan := billingplan.Normalize(plan)
	if normalizedPlan == "" {
		return billingcatalog.Selection{}, ErrInvalidPlan
	}

	normalizedInterval := billingcatalog.NormalizeInterval(interval)
	switch {
	case strings.TrimSpace(interval) == "":
		normalizedInterval = current.Interval
	case normalizedInterval == "":
		return billingcatalog.Selection{}, ErrInvalidInterval
	}
	if normalizedInterval == "" {
		normalizedInterval = billingcatalog.DefaultIntervalForPlan(normalizedPlan, s.billingCatalog)
	}

	selection := billingcatalog.Selection{
		Plan:     normalizedPlan,
		Interval: normalizedInterval,
	}
	if selection.Interval == "" || !billingcatalog.IntervalAvailable(selection.Plan, selection.Interval, s.billingCatalog) {
		return billingcatalog.Selection{}, ErrPlanUnavailable
	}
	return selection, nil
}

type requestOpt func(*http.Request)

func withHeader(key, value string) requestOpt {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}

func (s *Service) doRequest(
	ctx context.Context,
	method, path string,
	body io.Reader,
	dst any,
	opts ...requestOpt,
) error {
	if strings.TrimSpace(s.stripeSecretKey) == "" {
		return ErrNotConfigured
	}

	req, err := http.NewRequestWithContext(ctx, method, s.apiBaseURL+path, body)
	if err != nil {
		return fmt.Errorf("create stripe request: %w", err)
	}
	req.SetBasicAuth(s.stripeSecretKey, "")
	for _, opt := range opts {
		opt(req)
	}

	slog.DebugContext(ctx, "stripe request started", "method", method, "path", path)

	res, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("stripe request failed: %w", err)
	}
	defer res.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("read stripe response: %w", err)
	}

	slog.DebugContext(ctx, "stripe request completed", "method", method, "path", path, "status", res.StatusCode)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		slog.ErrorContext(
			ctx,
			"stripe request returned non-success status",
			"method", method,
			"path", path,
			"status", res.StatusCode,
			"body", truncateForLog(string(payload), 800),
		)
		return fmt.Errorf("stripe %s %s failed: status %d: %s", method, path, res.StatusCode, strings.TrimSpace(string(payload)))
	}
	if dst == nil || len(bytes.TrimSpace(payload)) == 0 {
		return nil
	}
	if err := json.Unmarshal(payload, dst); err != nil {
		return fmt.Errorf("decode stripe response: %w", err)
	}

	return nil
}

func (s *Service) appURL(path string) string {
	base := strings.TrimRight(strings.TrimSpace(s.appBaseURL), "/")
	if base == "" {
		return path
	}
	if strings.HasPrefix(path, "/") {
		return base + path
	}

	return base + "/" + path
}

func firstSubscriptionPriceID(subscription stripeSubscription) string {
	for _, item := range subscription.Items.Data {
		if strings.TrimSpace(item.Price.ID) != "" {
			return strings.TrimSpace(item.Price.ID)
		}
	}

	return ""
}

func firstSubscriptionItemID(subscription stripeSubscription) string {
	for _, item := range subscription.Items.Data {
		if strings.TrimSpace(item.ID) != "" {
			return strings.TrimSpace(item.ID)
		}
	}

	return ""
}

func determineSubscriptionSelection(
	subscription stripeSubscription,
	cfg billingcatalog.Config,
) billingcatalog.Selection {
	selection := billingcatalog.NormalizeSelection(subscription.Metadata["plan"], subscription.Metadata["interval"])
	selectionFromPriceID := billingcatalog.DetermineFromPriceID(firstSubscriptionPriceID(subscription), cfg)

	if selection.Plan == "" {
		selection.Plan = selectionFromPriceID.Plan
	}
	if selection.Interval == "" {
		selection.Interval = selectionFromPriceID.Interval
	}

	return selection
}

func sanitizeRedirectPath(path string) string {
	normalized := strings.TrimSpace(path)
	if normalized == "" || !strings.HasPrefix(normalized, "/") || strings.HasPrefix(normalized, "//") {
		return "/app"
	}

	return normalized
}

func formatStripePriceLabel(currency string, unitAmount int64, interval string) string {
	if unitAmount <= 0 {
		return ""
	}

	divisor := int64(100)
	if zeroDecimalCurrency(currency) {
		divisor = 1
	}

	whole := unitAmount / divisor
	remainder := unitAmount % divisor

	amountLabel := ""
	switch {
	case divisor == 1:
		amountLabel = fmt.Sprintf("%s%d", currencySymbol(currency), whole)
	case remainder == 0:
		amountLabel = fmt.Sprintf("%s%d", currencySymbol(currency), whole)
	default:
		amountLabel = fmt.Sprintf("%s%d.%02d", currencySymbol(currency), whole, remainder)
	}

	suffix := interval
	switch suffix {
	case billingcatalog.IntervalMonthly:
		suffix = "month"
	case billingcatalog.IntervalYearly:
		suffix = "year"
	}
	if suffix == "" {
		suffix = "period"
	}

	return amountLabel + " / " + suffix
}

func currencySymbol(currency string) string {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "GBP":
		return "£"
	case "USD":
		return "$"
	case "EUR":
		return "€"
	case "AUD":
		return "A$"
	case "CAD":
		return "C$"
	case "JPY":
		return "¥"
	case "NZD":
		return "NZ$"
	default:
		return strings.ToUpper(strings.TrimSpace(currency)) + " "
	}
}

func zeroDecimalCurrency(currency string) bool {
	switch strings.ToUpper(strings.TrimSpace(currency)) {
	case "BIF", "CLP", "DJF", "GNF", "JPY", "KMF", "KRW", "MGA", "PYG", "RWF", "UGX", "VND", "VUV", "XAF", "XOF", "XPF":
		return true
	default:
		return false
	}
}

func verifyStripeWebhookSignature(payload []byte, signatureHeader, secret string, tolerance time.Duration) error {
	secret = strings.TrimSpace(secret)
	signatureHeader = strings.TrimSpace(signatureHeader)
	if secret == "" || signatureHeader == "" {
		return ErrWebhookSignature
	}

	timestamp, signatures, err := parseStripeSignatureHeader(signatureHeader)
	if err != nil {
		return err
	}
	if tolerance > 0 && time.Since(timestamp) > tolerance {
		return ErrWebhookSignature
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strconv.FormatInt(timestamp.Unix(), 10)))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	for _, candidate := range signatures {
		if hmac.Equal([]byte(candidate), []byte(expected)) {
			return nil
		}
	}

	return ErrWebhookSignature
}

func parseStripeSignatureHeader(header string) (time.Time, []string, error) {
	var (
		timestamp  time.Time
		signatures []string
	)

	for part := range strings.SplitSeq(header, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "t":
			secs, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
			if err != nil {
				return time.Time{}, nil, ErrWebhookSignature
			}
			timestamp = time.Unix(secs, 0).UTC()
		case "v1":
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				signatures = append(signatures, trimmed)
			}
		}
	}

	if timestamp.IsZero() || len(signatures) == 0 {
		return time.Time{}, nil, ErrWebhookSignature
	}

	return timestamp, signatures, nil
}

func shortStripeID(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	if len(id) <= 12 {
		return id
	}

	return id[:8] + "..." + id[len(id)-4:]
}

func truncateForLog(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}

	return value[:max-3] + "..."
}
