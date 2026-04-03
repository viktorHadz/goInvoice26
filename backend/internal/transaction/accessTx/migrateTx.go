package accessTx

import (
	"context"
	"database/sql"
	"fmt"
)

func EnsureDirectAccessGrantsTable(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS direct_access_grants (
			id INTEGER PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			plan TEXT NOT NULL DEFAULT 'single',
			note TEXT NOT NULL DEFAULT '',
			created_by_user_id INTEGER,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		);
	`); err != nil {
		return fmt.Errorf("ensure direct_access_grants table: %w", err)
	}
	hasPlanColumn, err := tableHasColumn(ctx, tx, "direct_access_grants", "plan")
	if err != nil {
		return err
	}
	if !hasPlanColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE direct_access_grants
			ADD COLUMN plan TEXT NOT NULL DEFAULT 'single';
		`); err != nil {
			return fmt.Errorf("add direct_access_grants.plan: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE direct_access_grants
		SET plan = 'single'
		WHERE plan IS NULL OR TRIM(plan) = '';
	`); err != nil {
		return fmt.Errorf("backfill direct_access_grants.plan: %w", err)
	}

	return nil
}

func EnsurePromoCodesTable(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS promo_codes (
			id INTEGER PRIMARY KEY,
			code TEXT NOT NULL UNIQUE,
			duration_days INTEGER NOT NULL CHECK (duration_days > 0),
			active INTEGER NOT NULL DEFAULT 1,
			created_by_user_id INTEGER,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		);
	`); err != nil {
		return fmt.Errorf("ensure promo_codes table: %w", err)
	}

	return nil
}

func EnsurePromoCodeRedemptionsTable(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS promo_code_redemptions (
			id INTEGER PRIMARY KEY,
			promo_code_id INTEGER NOT NULL REFERENCES promo_codes(id) ON DELETE CASCADE,
			account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
			redeemed_by_user_id INTEGER,
			redeemed_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			expires_at TEXT NOT NULL,
			UNIQUE (promo_code_id, account_id)
		);
	`); err != nil {
		return fmt.Errorf("ensure promo_code_redemptions table: %w", err)
	}

	return nil
}

func EnsurePromoCodeRedemptionClaimsTable(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS promo_code_redemption_claims (
			id INTEGER PRIMARY KEY,
			promo_code_id INTEGER NOT NULL REFERENCES promo_codes(id) ON DELETE CASCADE,
			owner_email_hmac TEXT NOT NULL,
			redeemed_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			retention_until TEXT NOT NULL,
			UNIQUE (promo_code_id, owner_email_hmac)
		);
	`); err != nil {
		return fmt.Errorf("ensure promo_code_redemption_claims table: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_promo_code_redemption_claims_retention_until
		ON promo_code_redemption_claims(retention_until);
	`); err != nil {
		return fmt.Errorf("ensure promo_code_redemption_claims retention index: %w", err)
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
