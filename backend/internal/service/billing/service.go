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
	"time"

	"github.com/viktorHadz/goInvoice26/internal/billingstate"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/billingTx"
)

var (
	ErrNotConfigured       = errors.New("billing is not configured")
	ErrOwnerOnly           = errors.New("billing management is owner-only")
	ErrCustomerNotFound    = errors.New("stripe customer not found for account")
	ErrInvalidCheckoutSync = errors.New("checkout session is not valid for this account")
	ErrCheckoutPending     = errors.New("checkout session is still confirming")
	ErrWebhookSignature    = errors.New("invalid stripe webhook signature")
)

type Config struct {
	AppBaseURL          string
	StripeSecretKey     string
	StripePriceID       string
	StripeWebhookSecret string
	APIBaseURL          string
	HTTPClient          *http.Client
}

type Service struct {
	db                  *sql.DB
	appBaseURL          string
	stripeSecretKey     string
	stripePriceID       string
	stripeWebhookSecret string
	apiBaseURL          string
	httpClient          *http.Client
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

type stripeEvent struct {
	Type string `json:"type"`
	Data struct {
		Object json.RawMessage `json:"object"`
	} `json:"data"`
}

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
		db:                  db,
		appBaseURL:          strings.TrimRight(strings.TrimSpace(cfg.AppBaseURL), "/"),
		stripeSecretKey:     strings.TrimSpace(cfg.StripeSecretKey),
		stripePriceID:       strings.TrimSpace(cfg.StripePriceID),
		stripeWebhookSecret: strings.TrimSpace(cfg.StripeWebhookSecret),
		apiBaseURL:          apiBaseURL,
		httpClient:          httpClient,
	}
}

func (s *Service) Configured() bool {
	return s.stripeSecretKey != "" && s.stripePriceID != ""
}

func (s *Service) WebhooksConfigured() bool {
	return s.stripeSecretKey != "" && s.stripeWebhookSecret != ""
}

func (s *Service) CreateCheckoutSession(
	ctx context.Context,
	accountID int64,
	accountName string,
	customerEmail string,
) (models.BillingSessionLink, error) {
	if !s.Configured() {
		slog.WarnContext(ctx, "billing checkout requested but service is not configured", "accountID", accountID)
		return models.BillingSessionLink{}, ErrNotConfigured
	}

	record, err := billingTx.GetAccountBilling(ctx, s.db, accountID)
	if err != nil {
		return models.BillingSessionLink{}, err
	}

	slog.InfoContext(
		ctx,
		"billing checkout session requested",
		"accountID", accountID,
		"hasCustomer", strings.TrimSpace(record.StripeCustomerID) != "",
		"hasEmail", strings.TrimSpace(customerEmail) != "",
		"priceID", shortStripeID(s.stripePriceID),
	)

	form := url.Values{}
	form.Set("mode", "subscription")
	form.Set("success_url", s.appURL("/app/billing?checkout=success&session_id={CHECKOUT_SESSION_ID}"))
	form.Set("cancel_url", s.appURL("/app/billing?checkout=canceled"))
	form.Set("line_items[0][price]", s.stripePriceID)
	form.Set("line_items[0][quantity]", "1")
	form.Set("client_reference_id", strconv.FormatInt(accountID, 10))
	form.Set("metadata[account_id]", strconv.FormatInt(accountID, 10))
	form.Set("metadata[account_name]", strings.TrimSpace(accountName))
	form.Set("subscription_data[metadata][account_id]", strconv.FormatInt(accountID, 10))
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
