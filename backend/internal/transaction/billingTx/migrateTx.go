package billingTx

import (
	"context"
	"database/sql"
	"fmt"
)

func EnsureAccountsStripeCustomerIDColumn(ctx context.Context, tx *sql.Tx) error {
	return ensureAccountsTextColumn(ctx, tx, "stripe_customer_id")
}

func EnsureAccountsStripeSubscriptionIDColumn(ctx context.Context, tx *sql.Tx) error {
	return ensureAccountsTextColumn(ctx, tx, "stripe_subscription_id")
}

func EnsureAccountsBillingPriceIDColumn(ctx context.Context, tx *sql.Tx) error {
	return ensureAccountsTextColumn(ctx, tx, "billing_price_id")
}

func EnsureAccountsBillingEmailColumn(ctx context.Context, tx *sql.Tx) error {
	return ensureAccountsTextColumn(ctx, tx, "billing_email")
}

func EnsureAccountsBillingStatusColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "accounts", "billing_status")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE accounts
			ADD COLUMN billing_status TEXT NOT NULL DEFAULT 'inactive';
		`); err != nil {
			return fmt.Errorf("add accounts.billing_status: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE accounts
		SET billing_status = 'inactive'
		WHERE billing_status IS NULL OR TRIM(billing_status) = '';
	`); err != nil {
		return fmt.Errorf("backfill accounts.billing_status: %w", err)
	}

	return nil
}

func EnsureAccountsBillingCurrentPeriodEndColumn(ctx context.Context, tx *sql.Tx) error {
	return ensureAccountsTextColumn(ctx, tx, "billing_current_period_end")
}

func EnsureAccountsBillingCancelAtPeriodEndColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "accounts", "billing_cancel_at_period_end")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE accounts
			ADD COLUMN billing_cancel_at_period_end INTEGER NOT NULL DEFAULT 0;
		`); err != nil {
			return fmt.Errorf("add accounts.billing_cancel_at_period_end: %w", err)
		}
	}

	return nil
}

func EnsureAccountsBillingUpdatedAtColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "accounts", "billing_updated_at")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE accounts
			ADD COLUMN billing_updated_at TEXT NOT NULL DEFAULT '';
		`); err != nil {
			return fmt.Errorf("add accounts.billing_updated_at: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE accounts
		SET billing_updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
		WHERE billing_updated_at IS NULL OR TRIM(billing_updated_at) = '';
	`); err != nil {
		return fmt.Errorf("backfill accounts.billing_updated_at: %w", err)
	}

	return nil
}

func ensureAccountsTextColumn(ctx context.Context, tx *sql.Tx, columnName string) error {
	hasColumn, err := tableHasColumn(ctx, tx, "accounts", columnName)
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
			ALTER TABLE accounts
			ADD COLUMN %s TEXT NOT NULL DEFAULT '';
		`, columnName)); err != nil {
			return fmt.Errorf("add accounts.%s: %w", columnName, err)
		}
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE accounts
		SET %s = ''
		WHERE %s IS NULL;
	`, columnName, columnName)); err != nil {
		return fmt.Errorf("backfill accounts.%s: %w", columnName, err)
	}

	return nil
}

func tableHasColumn(ctx context.Context, tx *sql.Tx, tableName, columnName string) (bool, error) {
	rows, err := tx.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s);`, tableName))
	if err != nil {
		return false, fmt.Errorf("table info %s: %w", tableName, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid       int
			name      string
			typ       string
			notNull   int
			dfltValue sql.NullString
			pk        int
		)
		if err := rows.Scan(&cid, &name, &typ, &notNull, &dfltValue, &pk); err != nil {
			return false, fmt.Errorf("scan table info %s: %w", tableName, err)
		}
		if name == columnName {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate table info %s: %w", tableName, err)
	}

	return false, nil
}
