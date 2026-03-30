package authTx

import (
	"context"
	"database/sql"
	"fmt"
)

func EnsureUsersGoogleSubColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "users", "google_sub")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE users
			ADD COLUMN google_sub TEXT;
		`); err != nil {
			return fmt.Errorf("add users.google_sub: %w", err)
		}
	}

	return nil
}

func EnsureUsersAvatarURLColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "users", "avatar_url")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE users
			ADD COLUMN avatar_url TEXT NOT NULL DEFAULT '';
		`); err != nil {
			return fmt.Errorf("add users.avatar_url: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE users
		SET avatar_url = ''
		WHERE avatar_url IS NULL;
	`); err != nil {
		return fmt.Errorf("backfill users.avatar_url: %w", err)
	}

	return nil
}

func EnsureUsersRoleColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "users", "role")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE users
			ADD COLUMN role TEXT NOT NULL DEFAULT 'member';
		`); err != nil {
			return fmt.Errorf("add users.role: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE users
		SET role = 'member'
		WHERE role IS NULL OR TRIM(role) = '';
	`); err != nil {
		return fmt.Errorf("backfill users.role: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE users
		SET role = 'owner'
		WHERE id = (
			SELECT id
			FROM users
			ORDER BY id
			LIMIT 1
		)
		AND NOT EXISTS (
			SELECT 1
			FROM users
			WHERE role = 'owner'
		);
	`); err != nil {
		return fmt.Errorf("ensure owner role: %w", err)
	}

	return nil
}

func EnsureAllowedUsersAccountIDColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "allowed_users", "account_id")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE allowed_users
			ADD COLUMN account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id);
		`); err != nil {
			return fmt.Errorf("add allowed_users.account_id: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE allowed_users
		SET account_id = 1
		WHERE account_id IS NULL OR account_id <= 0;
	`); err != nil {
		return fmt.Errorf("backfill allowed_users.account_id: %w", err)
	}

	return nil
}

func EnsureAllowedUsersCreatedAtColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "allowed_users", "created_at")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE allowed_users
			ADD COLUMN created_at TEXT NOT NULL DEFAULT '';
		`); err != nil {
			return fmt.Errorf("add allowed_users.created_at: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE allowed_users
		SET created_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
		WHERE created_at IS NULL OR TRIM(created_at) = '';
	`); err != nil {
		return fmt.Errorf("backfill allowed_users.created_at: %w", err)
	}

	return nil
}

func EnsureAllowedUsersInvitedByUserIDColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "allowed_users", "invited_by_user_id")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE allowed_users
			ADD COLUMN invited_by_user_id INTEGER;
		`); err != nil {
			return fmt.Errorf("add allowed_users.invited_by_user_id: %w", err)
		}
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
