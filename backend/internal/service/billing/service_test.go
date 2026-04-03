package billing_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/service/billing"
	"github.com/viktorHadz/goInvoice26/internal/transaction/billingTx"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func newBillingService(t *testing.T, transport roundTripFunc) (*sql.DB, *billing.Service, func()) {
	return newBillingServiceWithCatalog(
		t,
		transport,
		7,
		"price_single_monthly",
		"price_single_yearly",
		"price_team_monthly",
		"price_team_yearly",
	)
}

func newBillingServiceWithCatalog(
	t *testing.T,
	transport roundTripFunc,
	trialDays int,
	singleMonthlyPriceID string,
	singleYearlyPriceID string,
	teamMonthlyPriceID string,
	teamYearlyPriceID string,
) (*sql.DB, *billing.Service, func()) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "billing.sqlite")
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	service := billing.NewService(conn, billing.Config{
		AppBaseURL:                 "http://localhost:5173",
		StripeSecretKey:            "sk_test_123",
		StripeSingleMonthlyPriceID: singleMonthlyPriceID,
		StripeSingleYearlyPriceID:  singleYearlyPriceID,
		StripeTeamMonthlyPriceID:   teamMonthlyPriceID,
		StripeTeamYearlyPriceID:    teamYearlyPriceID,
		StripeTrialDays:            trialDays,
		StripeWebhookSecret:        "whsec_test",
		APIBaseURL:                 "https://stripe.test",
		HTTPClient: &http.Client{
			Transport: transport,
		},
	})

	return conn, service, func() {
		_ = conn.Close()
	}
}

func seedBillingRecord(t *testing.T, conn *sql.DB, accountID int64) {
	t.Helper()

	if err := billingTx.UpdateAccountBilling(context.Background(), conn, billingTx.UpdateAccountBillingParams{
		AccountID:               accountID,
		StripeCustomerID:        "cus_123",
		StripeSubscriptionID:    "sub_123",
		BillingPriceID:          "price_single_monthly",
		BillingEmail:            "owner@example.com",
		BillingStatus:           "active",
		BillingCurrentPeriodEnd: timePtr(time.Unix(1_750_000_000, 0).UTC()),
		BillingUpdatedAt:        time.Now(),
	}); err != nil {
		t.Fatalf("seed billing record: %v", err)
	}
}

func TestCancelSubscriptionAtPeriodEnd_UpdatesAccountState(t *testing.T) {
	ctx := context.Background()
	conn, service, cleanup := newBillingService(t, func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/subscriptions/sub_123" {
			t.Fatalf("unexpected stripe request %s %s", r.Method, r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("cancel_at_period_end"); got != "true" {
			t.Fatalf("cancel_at_period_end = %q, want true", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
			"id":"sub_123",
			"customer":"cus_123",
			"status":"active",
			"cancel_at_period_end":true,
			"current_period_end":1750000000,
			"items":{"data":[{"id":"si_123","price":{"id":"price_single_monthly"}}]}
		}`)),
		}, nil
	})
	defer cleanup()

	seedBillingRecord(t, conn, 1)

	if err := service.CancelSubscriptionAtPeriodEnd(ctx, 1); err != nil {
		t.Fatalf("CancelSubscriptionAtPeriodEnd: %v", err)
	}

	record, err := billingTx.GetAccountBilling(ctx, conn, 1)
	if err != nil {
		t.Fatalf("GetAccountBilling: %v", err)
	}
	if !record.BillingCancelAtPeriodEnd {
		t.Fatal("BillingCancelAtPeriodEnd = false, want true")
	}
	if record.BillingStatus != "active" {
		t.Fatalf("BillingStatus = %q, want active", record.BillingStatus)
	}
}

func TestCreateCheckoutSession_IncludesConfiguredTrial(t *testing.T) {
	ctx := context.Background()
	_, service, cleanup := newBillingServiceWithCatalog(t, func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/checkout/sessions" {
			t.Fatalf("unexpected stripe request %s %s", r.Method, r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("ParseForm: %v", err)
		}
		if got := r.Form.Get("subscription_data[trial_period_days]"); got != "7" {
			t.Fatalf("trial_period_days = %q, want 7", got)
		}
		if got := r.Form.Get("payment_method_collection"); got != "always" {
			t.Fatalf("payment_method_collection = %q, want always", got)
		}
		if got := r.Form.Get("line_items[0][price]"); got != "price_team_yearly" {
			t.Fatalf("line_items[0][price] = %q, want price_team_yearly", got)
		}
		if got := r.Form.Get("subscription_data[trial_settings][end_behavior][missing_payment_method]"); got != "cancel" {
			t.Fatalf("trial_settings end behavior = %q, want cancel", got)
		}
		if got := r.Form.Get("metadata[interval]"); got != "yearly" {
			t.Fatalf("metadata[interval] = %q, want yearly", got)
		}
		if got := r.Form.Get("success_url"); !strings.Contains(got, "redirect=%2Fapp%2Fclients") {
			t.Fatalf("success_url = %q, want redirect query", got)
		}
		if got := r.Form.Get("cancel_url"); !strings.Contains(got, "redirect=%2Fapp%2Fclients") {
			t.Fatalf("cancel_url = %q, want redirect query", got)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
			"id":"cs_test_123",
			"url":"https://checkout.stripe.test/session/cs_test_123"
		}`)),
		}, nil
	}, 7, "price_single_monthly", "price_single_yearly", "price_team_monthly", "price_team_yearly")
	defer cleanup()

	session, err := service.CreateCheckoutSession(
		ctx,
		1,
		"Workspace",
		"owner@example.com",
		"team",
		"yearly",
		"/app/clients",
	)
	if err != nil {
		t.Fatalf("CreateCheckoutSession: %v", err)
	}
	if session.URL == "" {
		t.Fatal("session URL = empty, want checkout URL")
	}
}

func TestCancelSubscriptionImmediately_DeletesSubscriptionAndMarksCanceled(t *testing.T) {
	ctx := context.Background()
	conn, service, cleanup := newBillingService(t, func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodDelete || r.URL.Path != "/v1/subscriptions/sub_123" {
			t.Fatalf("unexpected stripe request %s %s", r.Method, r.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
			"id":"sub_123",
			"customer":"cus_123",
			"status":"canceled",
			"cancel_at_period_end":false,
			"current_period_end":1750000000,
			"items":{"data":[{"id":"si_123","price":{"id":"price_single_monthly"}}]}
		}`)),
		}, nil
	})
	defer cleanup()

	seedBillingRecord(t, conn, 1)

	if err := service.CancelSubscriptionImmediately(ctx, 1); err != nil {
		t.Fatalf("CancelSubscriptionImmediately: %v", err)
	}

	record, err := billingTx.GetAccountBilling(ctx, conn, 1)
	if err != nil {
		t.Fatalf("GetAccountBilling: %v", err)
	}
	if record.BillingStatus != "canceled" {
		t.Fatalf("BillingStatus = %q, want canceled", record.BillingStatus)
	}
	if record.BillingCancelAtPeriodEnd {
		t.Fatal("BillingCancelAtPeriodEnd = true, want false")
	}
}

func TestCancelSubscriptionAtPeriodEnd_ReturnsNotFoundWithoutSubscription(t *testing.T) {
	ctx := context.Background()
	_, service, cleanup := newBillingService(t, func(r *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected stripe request %s %s", r.Method, r.URL.Path)
		return nil, nil
	})
	defer cleanup()

	err := service.CancelSubscriptionAtPeriodEnd(ctx, 1)
	if err == nil || !strings.Contains(err.Error(), billing.ErrSubscriptionNotFound.Error()) {
		t.Fatalf("CancelSubscriptionAtPeriodEnd error = %v, want %v", err, billing.ErrSubscriptionNotFound)
	}
}

func TestChangeSubscriptionPlan_UpdatesSubscriptionItemPrice(t *testing.T) {
	ctx := context.Background()
	conn, service, cleanup := newBillingService(t, func(r *http.Request) (*http.Response, error) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1/subscriptions/sub_123":
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body: io.NopCloser(strings.NewReader(`{
				"id":"sub_123",
				"customer":"cus_123",
				"status":"active",
				"cancel_at_period_end":false,
				"current_period_end":1750000000,
				"items":{"data":[{"id":"si_123","price":{"id":"price_single_monthly"}}]}
			}`)),
			}, nil
		case r.Method == http.MethodPost && r.URL.Path == "/v1/subscriptions/sub_123":
			if err := r.ParseForm(); err != nil {
				t.Fatalf("ParseForm: %v", err)
			}
			if got := r.Form.Get("items[0][id]"); got != "si_123" {
				t.Fatalf("items[0][id] = %q, want si_123", got)
			}
			if got := r.Form.Get("items[0][price]"); got != "price_team_yearly" {
				t.Fatalf("items[0][price] = %q, want price_team_yearly", got)
			}
			if got := r.Form.Get("proration_behavior"); got != "create_prorations" {
				t.Fatalf("proration_behavior = %q, want create_prorations", got)
			}
			if got := r.Form.Get("cancel_at_period_end"); got != "false" {
				t.Fatalf("cancel_at_period_end = %q, want false", got)
			}
			if got := r.Form.Get("metadata[interval]"); got != "yearly" {
				t.Fatalf("metadata[interval] = %q, want yearly", got)
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body: io.NopCloser(strings.NewReader(`{
				"id":"sub_123",
				"customer":"cus_123",
				"status":"active",
				"cancel_at_period_end":false,
				"current_period_end":1750000000,
				"items":{"data":[{"id":"si_123","price":{"id":"price_team_yearly"}}]}
			}`)),
			}, nil
		default:
			t.Fatalf("unexpected stripe request %s %s", r.Method, r.URL.Path)
			return nil, nil
		}
	})
	defer cleanup()

	seedBillingRecord(t, conn, 1)

	if err := service.ChangeSubscriptionPlan(ctx, 1, "team", "yearly"); err != nil {
		t.Fatalf("ChangeSubscriptionPlan: %v", err)
	}

	record, err := billingTx.GetAccountBilling(ctx, conn, 1)
	if err != nil {
		t.Fatalf("GetAccountBilling: %v", err)
	}
	if record.BillingPriceID != "price_team_yearly" {
		t.Fatalf("BillingPriceID = %q, want price_team_yearly", record.BillingPriceID)
	}
}

func TestHandleWebhook_TransitionsTrialToActive(t *testing.T) {
	ctx := context.Background()
	conn, service, cleanup := newBillingService(t, func(r *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected stripe request %s %s", r.Method, r.URL.Path)
		return nil, nil
	})
	defer cleanup()

	if err := billingTx.UpdateAccountBilling(ctx, conn, billingTx.UpdateAccountBillingParams{
		AccountID:               1,
		StripeCustomerID:        "cus_123",
		StripeSubscriptionID:    "sub_123",
		BillingPriceID:          "price_single_monthly",
		BillingPlan:             "single",
		BillingInterval:         "monthly",
		BillingEmail:            "owner@example.com",
		BillingStatus:           "trialing",
		BillingCurrentPeriodEnd: timePtr(time.Unix(1_750_000_000, 0).UTC()),
		BillingUpdatedAt:        time.Now(),
	}); err != nil {
		t.Fatalf("seed billing record: %v", err)
	}

	payload := []byte(`{
		"type":"customer.subscription.updated",
		"data":{"object":{
			"id":"sub_123",
			"customer":"cus_123",
			"status":"active",
			"cancel_at_period_end":false,
			"current_period_end":1751000000,
			"items":{"data":[{"id":"si_123","price":{"id":"price_single_monthly"}}]}
		}}
	}`)
	if err := service.HandleWebhook(ctx, payload, signedStripeWebhookHeader(payload, "whsec_test")); err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}

	record, err := billingTx.GetAccountBilling(ctx, conn, 1)
	if err != nil {
		t.Fatalf("GetAccountBilling: %v", err)
	}
	if record.BillingStatus != "active" {
		t.Fatalf("BillingStatus = %q, want active", record.BillingStatus)
	}
}

func TestHandleWebhook_TransitionsTrialToCanceled(t *testing.T) {
	ctx := context.Background()
	conn, service, cleanup := newBillingService(t, func(r *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected stripe request %s %s", r.Method, r.URL.Path)
		return nil, nil
	})
	defer cleanup()

	if err := billingTx.UpdateAccountBilling(ctx, conn, billingTx.UpdateAccountBillingParams{
		AccountID:               1,
		StripeCustomerID:        "cus_123",
		StripeSubscriptionID:    "sub_123",
		BillingPriceID:          "price_single_monthly",
		BillingPlan:             "single",
		BillingInterval:         "monthly",
		BillingEmail:            "owner@example.com",
		BillingStatus:           "trialing",
		BillingCurrentPeriodEnd: timePtr(time.Unix(1_750_000_000, 0).UTC()),
		BillingUpdatedAt:        time.Now(),
	}); err != nil {
		t.Fatalf("seed billing record: %v", err)
	}

	payload := []byte(`{
		"type":"customer.subscription.deleted",
		"data":{"object":{
			"id":"sub_123",
			"customer":"cus_123",
			"status":"canceled",
			"cancel_at_period_end":false,
			"current_period_end":1750000000,
			"items":{"data":[{"id":"si_123","price":{"id":"price_single_monthly"}}]}
		}}
	}`)
	if err := service.HandleWebhook(ctx, payload, signedStripeWebhookHeader(payload, "whsec_test")); err != nil {
		t.Fatalf("HandleWebhook: %v", err)
	}

	record, err := billingTx.GetAccountBilling(ctx, conn, 1)
	if err != nil {
		t.Fatalf("GetAccountBilling: %v", err)
	}
	if record.BillingStatus != "canceled" {
		t.Fatalf("BillingStatus = %q, want canceled", record.BillingStatus)
	}
}

func timePtr(ts time.Time) *time.Time {
	return &ts
}

func signedStripeWebhookHeader(payload []byte, secret string) string {
	ts := time.Now().Unix()
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(strconv.FormatInt(ts, 10)))
	_, _ = mac.Write([]byte("."))
	_, _ = mac.Write(payload)
	return fmt.Sprintf("t=%d,v1=%s", ts, hex.EncodeToString(mac.Sum(nil)))
}
