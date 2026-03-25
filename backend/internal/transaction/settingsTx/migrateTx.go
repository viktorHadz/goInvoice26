package settingsTx

import (
	"context"
	"database/sql"
	"fmt"
)

func EnsureShowItemTypeHeadersColumn(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `PRAGMA table_info(user_settings);`)
	if err != nil {
		return fmt.Errorf("table info user_settings: %w", err)
	}
	defer rows.Close()

	hasColumn := false
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
			return fmt.Errorf("scan table info user_settings: %w", err)
		}
		if name == "show_item_type_headers" {
			hasColumn = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate table info user_settings: %w", err)
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
