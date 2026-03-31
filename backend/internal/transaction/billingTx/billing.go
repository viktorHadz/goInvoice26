package billingTx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/viktorHadz/goInvoice26/internal/billingstate"
)

const timestampLayout = "2006-01-02T15:04:05.000000000Z07:00"

type AccountBilling struct {
	AccountID                int64
	StripeCustomerID         string
	StripeSubscriptionID     string
	BillingPriceID           string
	BillingEmail             string
	BillingStatus            string
	BillingCurrentPeriodEnd  string
	BillingCancelAtPeriodEnd bool
	BillingUpdatedAt         string
}

type UpdateAccountBillingParams struct {
	AccountID                int64
	StripeCustomerID         string
	StripeSubscriptionID     string
	BillingPriceID           string
	BillingEmail             string
	BillingStatus            string
	BillingCurrentPeriodEnd  *time.Time
	BillingCancelAtPeriodEnd bool
	BillingUpdatedAt         time.Time
}

func GetAccountBilling(ctx context.Context, db *sql.DB, accountID int64) (AccountBilling, error) {
	var (
		record AccountBilling
		cancel int64
	)

	if err := db.QueryRowContext(ctx, `
		SELECT
			id,
			COALESCE(stripe_customer_id, ''),
			COALESCE(stripe_subscription_id, ''),
			COALESCE(billing_price_id, ''),
			COALESCE(billing_email, ''),
			COALESCE(billing_status, ''),
			COALESCE(billing_current_period_end, ''),
			COALESCE(billing_cancel_at_period_end, 0),
			COALESCE(billing_updated_at, '')
		FROM accounts
		WHERE id = ?
		LIMIT 1;
	`, accountID).Scan(
		&record.AccountID,
		&record.StripeCustomerID,
		&record.StripeSubscriptionID,
		&record.BillingPriceID,
		&record.BillingEmail,
		&record.BillingStatus,
		&record.BillingCurrentPeriodEnd,
		&cancel,
		&record.BillingUpdatedAt,
	); err != nil {
		return AccountBilling{}, fmt.Errorf("get account billing: %w", err)
	}

	record.BillingStatus = billingstate.Normalize(record.BillingStatus)
	record.BillingCancelAtPeriodEnd = cancel > 0

	return record, nil
}

func FindAccountBillingByCustomerID(ctx context.Context, db *sql.DB, customerID string) (AccountBilling, bool, error) {
	return findAccountBillingByField(ctx, db, "stripe_customer_id", customerID)
}

func FindAccountBillingBySubscriptionID(ctx context.Context, db *sql.DB, subscriptionID string) (AccountBilling, bool, error) {
	return findAccountBillingByField(ctx, db, "stripe_subscription_id", subscriptionID)
}

func LinkStripeIdentifiers(
	ctx context.Context,
	db *sql.DB,
	accountID int64,
	customerID string,
	subscriptionID string,
	billingEmail string,
	updatedAt time.Time,
) error {
	if accountID <= 0 {
		return errors.New("account id is required")
	}

	if _, err := db.ExecContext(ctx, `
		UPDATE accounts
		SET stripe_customer_id = CASE WHEN ? <> '' THEN ? ELSE stripe_customer_id END,
			stripe_subscription_id = CASE WHEN ? <> '' THEN ? ELSE stripe_subscription_id END,
			billing_email = CASE
				WHEN ? <> '' THEN ?
				WHEN COALESCE(billing_email, '') = '' THEN ?
				ELSE billing_email
			END,
			billing_updated_at = ?
		WHERE id = ?;
	`,
		strings.TrimSpace(customerID),
		strings.TrimSpace(customerID),
		strings.TrimSpace(subscriptionID),
		strings.TrimSpace(subscriptionID),
		strings.TrimSpace(billingEmail),
		strings.TrimSpace(billingEmail),
		strings.TrimSpace(billingEmail),
		formatTimestamp(updatedAt),
		accountID,
	); err != nil {
		return fmt.Errorf("link stripe identifiers: %w", err)
	}

	return nil
}

func UpdateAccountBilling(ctx context.Context, db *sql.DB, params UpdateAccountBillingParams) error {
	if params.AccountID <= 0 {
		return errors.New("account id is required")
	}

	status := billingstate.Normalize(params.BillingStatus)
	periodEnd := ""
	if params.BillingCurrentPeriodEnd != nil {
		periodEnd = formatTimestamp(*params.BillingCurrentPeriodEnd)
	}

	if _, err := db.ExecContext(ctx, `
		UPDATE accounts
		SET stripe_customer_id = CASE
				WHEN ? <> '' THEN ?
				ELSE stripe_customer_id
			END,
			stripe_subscription_id = CASE
				WHEN ? <> '' THEN ?
				WHEN ? = ? THEN ''
				ELSE stripe_subscription_id
			END,
			billing_price_id = CASE
				WHEN ? <> '' THEN ?
				ELSE billing_price_id
			END,
			billing_email = CASE
				WHEN ? <> '' THEN ?
				ELSE billing_email
			END,
			billing_status = ?,
			billing_current_period_end = ?,
			billing_cancel_at_period_end = ?,
			billing_updated_at = ?
		WHERE id = ?;
	`,
		strings.TrimSpace(params.StripeCustomerID),
		strings.TrimSpace(params.StripeCustomerID),
		strings.TrimSpace(params.StripeSubscriptionID),
		strings.TrimSpace(params.StripeSubscriptionID),
		status,
		billingstate.StatusInactive,
		strings.TrimSpace(params.BillingPriceID),
		strings.TrimSpace(params.BillingPriceID),
		strings.TrimSpace(params.BillingEmail),
		strings.TrimSpace(params.BillingEmail),
		status,
		periodEnd,
		boolToInt(params.BillingCancelAtPeriodEnd),
		formatTimestamp(params.BillingUpdatedAt),
		params.AccountID,
	); err != nil {
		return fmt.Errorf("update account billing: %w", err)
	}

	return nil
}

func AccountIDFromString(raw string) (int64, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, false
	}

	accountID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || accountID <= 0 {
		return 0, false
	}

	return accountID, true
}

func findAccountBillingByField(ctx context.Context, db *sql.DB, field, value string) (AccountBilling, bool, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return AccountBilling{}, false, nil
	}

	var (
		record AccountBilling
		cancel int64
	)

	query := fmt.Sprintf(`
		SELECT
			id,
			COALESCE(stripe_customer_id, ''),
			COALESCE(stripe_subscription_id, ''),
			COALESCE(billing_price_id, ''),
			COALESCE(billing_email, ''),
			COALESCE(billing_status, ''),
			COALESCE(billing_current_period_end, ''),
			COALESCE(billing_cancel_at_period_end, 0),
			COALESCE(billing_updated_at, '')
		FROM accounts
		WHERE %s = ?
		LIMIT 1;
	`, field)
	err := db.QueryRowContext(ctx, query, value).Scan(
		&record.AccountID,
		&record.StripeCustomerID,
		&record.StripeSubscriptionID,
		&record.BillingPriceID,
		&record.BillingEmail,
		&record.BillingStatus,
		&record.BillingCurrentPeriodEnd,
		&cancel,
		&record.BillingUpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return AccountBilling{}, false, nil
	}
	if err != nil {
		return AccountBilling{}, false, fmt.Errorf("find account billing by %s: %w", field, err)
	}

	record.BillingStatus = billingstate.Normalize(record.BillingStatus)
	record.BillingCancelAtPeriodEnd = cancel > 0

	return record, true, nil
}

func formatTimestamp(ts time.Time) string {
	return ts.UTC().Format(timestampLayout)
}

func boolToInt(v bool) int64 {
	if v {
		return 1
	}
	return 0
}
