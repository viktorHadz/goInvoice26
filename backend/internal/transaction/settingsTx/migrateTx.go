package settingsTx

import (
	"context"
	"database/sql"
	"fmt"
)

func EnsureAccountSettingsColumns(ctx context.Context, tx *sql.Tx) error {
	if err := ensureTableColumn(ctx, tx, "account_settings", "show_item_type_headers", `
		ALTER TABLE account_settings
		ADD COLUMN show_item_type_headers INTEGER NOT NULL DEFAULT 1;
	`); err != nil {
		return fmt.Errorf("ensure account_settings.show_item_type_headers: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE account_settings
		SET show_item_type_headers = 1
		WHERE show_item_type_headers IS NULL;
	`); err != nil {
		return fmt.Errorf("backfill account_settings.show_item_type_headers: %w", err)
	}

	if err := ensureTableColumn(ctx, tx, "account_settings", "legacy_logo_url", `
		ALTER TABLE account_settings
		ADD COLUMN legacy_logo_url TEXT NOT NULL DEFAULT '';
	`); err != nil {
		return fmt.Errorf("ensure account_settings.legacy_logo_url: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE account_settings
		SET legacy_logo_url = ''
		WHERE legacy_logo_url IS NULL;
	`); err != nil {
		return fmt.Errorf("backfill account_settings.legacy_logo_url: %w", err)
	}

	return nil
}

func EnsureShowItemTypeHeadersColumn(ctx context.Context, tx *sql.Tx) error {
	hasTable, err := tableExists(ctx, tx, "user_settings")
	if err != nil {
		return err
	}
	if !hasTable {
		return nil
	}

	hasColumn, err := tableHasColumn(ctx, tx, "user_settings", "show_item_type_headers")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE user_settings
			ADD COLUMN show_item_type_headers INTEGER NOT NULL DEFAULT 1;
		`); err != nil {
			return fmt.Errorf("add user_settings.show_item_type_headers: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE user_settings
		SET show_item_type_headers = 1
		WHERE show_item_type_headers IS NULL;
	`); err != nil {
		return fmt.Errorf("backfill user_settings.show_item_type_headers: %w", err)
	}

	return nil
}

func EnsureUsersAccountIDColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "users", "account_id")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE users
			ADD COLUMN account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id);
		`); err != nil {
			return fmt.Errorf("add users.account_id: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE users
		SET account_id = 1
		WHERE account_id IS NULL;
	`); err != nil {
		return fmt.Errorf("backfill users.account_id: %w", err)
	}

	return nil
}

func MigrateLegacyUserSettings(ctx context.Context, tx *sql.Tx) error {
	hasLegacyTable, err := tableExists(ctx, tx, "user_settings")
	if err != nil {
		return err
	}

	if !hasLegacyTable {
		if _, err := tx.ExecContext(ctx, `
			INSERT OR IGNORE INTO account_settings (account_id)
			VALUES (1);
		`); err != nil {
			return fmt.Errorf("ensure default account_settings row: %w", err)
		}
		return nil
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT OR IGNORE INTO account_settings (
			account_id,
			company_name,
			email,
			phone,
			company_address,
			invoice_prefix,
			currency,
			date_format,
			payment_terms,
			payment_details,
			notes_footer,
			legacy_logo_url,
			show_item_type_headers,
			updated_at
		)
		SELECT
			1,
			company_name,
			email,
			phone,
			company_address,
			invoice_prefix,
			currency,
			date_format,
			payment_terms,
			payment_details,
			notes_footer,
			logo_url,
			show_item_type_headers,
			strftime('%Y-%m-%dT%H:%M:%fZ','now')
		FROM user_settings
		WHERE id = 1;
	`); err != nil {
		return fmt.Errorf("migrate legacy user_settings row: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT OR IGNORE INTO account_settings (account_id)
		VALUES (1);
	`); err != nil {
		return fmt.Errorf("ensure default account_settings row: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE account_settings
		SET legacy_logo_url = (
			SELECT COALESCE(logo_url, '')
			FROM user_settings
			WHERE id = 1
		)
		WHERE account_id = 1
			AND TRIM(COALESCE(legacy_logo_url, '')) = ''
			AND EXISTS (
				SELECT 1
				FROM user_settings
				WHERE id = 1
					AND TRIM(COALESCE(logo_url, '')) <> ''
			);
	`); err != nil {
		return fmt.Errorf("copy legacy logo url into account_settings: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `DROP TABLE IF EXISTS user_settings;`); err != nil {
		return fmt.Errorf("drop legacy user_settings table: %w", err)
	}

	return nil
}

func ensureTableColumn(ctx context.Context, tx *sql.Tx, tableName, columnName, alterStmt string) error {
	hasColumn, err := tableHasColumn(ctx, tx, tableName, columnName)
	if err != nil {
		return err
	}
	if hasColumn {
		return nil
	}

	if _, err := tx.ExecContext(ctx, alterStmt); err != nil {
		return err
	}
	return nil
}

func tableExists(ctx context.Context, tx *sql.Tx, tableName string) (bool, error) {
	var exists int
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM sqlite_master
		WHERE type = 'table' AND name = ?;
	`, tableName).Scan(&exists); err != nil {
		return false, fmt.Errorf("check table %s exists: %w", tableName, err)
	}
	return exists > 0, nil
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
