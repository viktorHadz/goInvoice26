package db

import (
	"context"
	"database/sql"
	"fmt"
)

func ensureClientsAccountIDColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "clients", "account_id")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE clients
			ADD COLUMN account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id);
		`); err != nil {
			return fmt.Errorf("add clients.account_id: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE clients
		SET account_id = 1
		WHERE account_id IS NULL OR account_id <= 0;
	`); err != nil {
		return fmt.Errorf("backfill clients.account_id: %w", err)
	}

	return nil
}

func ensureProductsAccountIDColumn(ctx context.Context, tx *sql.Tx) error {
	hasColumn, err := tableHasColumn(ctx, tx, "products", "account_id")
	if err != nil {
		return err
	}

	if !hasColumn {
		if _, err := tx.ExecContext(ctx, `
			ALTER TABLE products
			ADD COLUMN account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id);
		`); err != nil {
			return fmt.Errorf("add products.account_id: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE products
		SET account_id = COALESCE(
			(
				SELECT c.account_id
				FROM clients c
				WHERE c.id = products.client_id
			),
			1
		)
		WHERE account_id IS NULL
		   OR account_id <= 0
		   OR account_id <> COALESCE(
				(
					SELECT c.account_id
					FROM clients c
					WHERE c.id = products.client_id
				),
				1
		   );
	`); err != nil {
		return fmt.Errorf("backfill products.account_id: %w", err)
	}

	return nil
}
