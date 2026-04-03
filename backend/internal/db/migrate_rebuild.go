package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func ensureStrictProductsTable(ctx context.Context, db *sql.DB) error {
	isStrict, err := tableDefinitionContains(ctx, db, "products",
		"client_id integer not null",
		"foreign key (account_id, client_id) references clients(account_id, id) on delete cascade",
		"unique (account_id, client_id, id)",
	)
	if err != nil {
		return err
	}
	if isStrict {
		return nil
	}

	return withSchemaRebuildTx(ctx, db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_account_id_id
			ON clients(account_id, id);
		`); err != nil {
			return fmt.Errorf("ensure clients composite key before products rebuild: %w", err)
		}

		var nullClientProducts int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM products
			WHERE client_id IS NULL;
		`).Scan(&nullClientProducts); err != nil {
			return fmt.Errorf("count products without client_id: %w", err)
		}
		if nullClientProducts > 0 {
			return fmt.Errorf("cannot migrate products: found %d rows without client ownership", nullClientProducts)
		}

		if _, err := tx.ExecContext(ctx, `ALTER TABLE products RENAME TO products_legacy;`); err != nil {
			return fmt.Errorf("rename legacy products table: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			CREATE TABLE products (
				id INTEGER PRIMARY KEY,
				account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id) ON DELETE CASCADE,
				product_type TEXT NOT NULL CHECK (product_type IN ('style','sample')),
				pricing_mode TEXT NOT NULL CHECK (pricing_mode IN ('flat','hourly')),
				name TEXT NOT NULL,
				flat_price_minor INTEGER CHECK (flat_price_minor IS NULL OR flat_price_minor >= 0),
				hourly_rate_minor INTEGER CHECK (hourly_rate_minor IS NULL OR hourly_rate_minor >= 0),
				default_minutes_worked INTEGER
					CHECK (default_minutes_worked IS NULL OR default_minutes_worked >= 0),
				client_id INTEGER NOT NULL,
				created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
				updated_at TEXT,
				FOREIGN KEY (account_id, client_id) REFERENCES clients(account_id, id) ON DELETE CASCADE,
				UNIQUE (account_id, client_id, id),
				CHECK (NOT (product_type = 'style' AND pricing_mode <> 'flat')),
				CHECK (
					(pricing_mode = 'flat' AND flat_price_minor IS NOT NULL) OR
					(pricing_mode = 'hourly' AND hourly_rate_minor IS NOT NULL)
				),
				CHECK (
					(product_type = 'sample' AND pricing_mode = 'hourly' AND default_minutes_worked IS NOT NULL)
					OR
					(NOT (product_type = 'sample' AND pricing_mode = 'hourly') AND default_minutes_worked IS NULL)
				)
			);
		`); err != nil {
			return fmt.Errorf("create strict products table: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO products (
				id,
				account_id,
				product_type,
				pricing_mode,
				name,
				flat_price_minor,
				hourly_rate_minor,
				default_minutes_worked,
				client_id,
				created_at,
				updated_at
			)
			SELECT
				id,
				account_id,
				product_type,
				pricing_mode,
				name,
				flat_price_minor,
				hourly_rate_minor,
				default_minutes_worked,
				client_id,
				created_at,
				updated_at
			FROM products_legacy;
		`); err != nil {
			return fmt.Errorf("copy legacy products into strict table: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `DROP TABLE products_legacy;`); err != nil {
			return fmt.Errorf("drop legacy products table: %w", err)
		}

		return nil
	})
}

func withSchemaRebuildTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("open dedicated schema rebuild connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `PRAGMA foreign_keys = OFF;`); err != nil {
		return fmt.Errorf("disable foreign keys for schema rebuild: %w", err)
	}
	defer func() {
		_, _ = conn.ExecContext(ctx, `PRAGMA foreign_keys = ON;`)
	}()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin schema rebuild tx: %w", err)
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}
	if err := foreignKeyCheckTx(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit schema rebuild tx: %w", err)
	}

	return nil
}

func foreignKeyCheckTx(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `PRAGMA foreign_key_check;`)
	if err != nil {
		return fmt.Errorf("run foreign_key_check: %w", err)
	}
	defer rows.Close()

	violations := make([]string, 0, 4)
	for rows.Next() {
		var (
			table  string
			rowID  sql.NullInt64
			parent sql.NullString
			fkID   sql.NullInt64
		)
		if err := rows.Scan(&table, &rowID, &parent, &fkID); err != nil {
			return fmt.Errorf("scan foreign_key_check row: %w", err)
		}

		violations = append(violations, fmt.Sprintf(
			"table=%s rowid=%v parent=%v fk=%v",
			table,
			rowID.Int64,
			parent.String,
			fkID.Int64,
		))
		if len(violations) >= 5 {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate foreign_key_check rows: %w", err)
	}
	if len(violations) > 0 {
		return fmt.Errorf("foreign key integrity check failed: %s", strings.Join(violations, "; "))
	}

	return nil
}

func migrateInvoiceNumberSequencesToAccounts(ctx context.Context, db *sql.DB) error {
	hasAccountID, err := dbTableHasColumn(ctx, db, "invoice_number_seq", "account_id")
	if err != nil {
		return err
	}
	if hasAccountID {
		if _, err := db.ExecContext(ctx, `
			INSERT OR IGNORE INTO invoice_number_seq (account_id, next_base_number)
			VALUES (1, 1);
		`); err != nil {
			return fmt.Errorf("ensure default invoice sequence row: %w", err)
		}
		return nil
	}

	if _, err := db.ExecContext(ctx, `
		ALTER TABLE invoice_number_seq RENAME TO invoice_number_seq_legacy;
	`); err != nil {
		return fmt.Errorf("rename invoice_number_seq legacy table: %w", err)
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE invoice_number_seq (
			account_id INTEGER PRIMARY KEY REFERENCES accounts(id) ON DELETE CASCADE,
			next_base_number INTEGER NOT NULL CHECK (next_base_number > 0)
		);
	`); err != nil {
		return fmt.Errorf("create account-scoped invoice_number_seq: %w", err)
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO invoice_number_seq (account_id, next_base_number)
		VALUES (
			1,
			MAX(
				COALESCE((SELECT next_base_number FROM invoice_number_seq_legacy LIMIT 1), 1),
				COALESCE((SELECT MAX(base_number) FROM invoices), 0) + 1
			)
		);
	`); err != nil {
		return fmt.Errorf("copy legacy invoice sequence: %w", err)
	}

	if _, err := db.ExecContext(ctx, `DROP TABLE invoice_number_seq_legacy;`); err != nil {
		return fmt.Errorf("drop legacy invoice_number_seq: %w", err)
	}

	return nil
}

func migrateInvoicesToAccounts(ctx context.Context, db *sql.DB) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("open dedicated invoice migration connection: %w", err)
	}
	defer conn.Close()

	hasAccountID, err := connTableHasColumn(ctx, conn, "invoices", "account_id")
	if err != nil {
		return err
	}
	isStrict, err := tableDefinitionContains(ctx, db, "invoices",
		"foreign key (account_id, client_id) references clients(account_id, id) on delete restrict",
		"unique (id, account_id, client_id)",
	)
	if err != nil {
		return err
	}
	if hasAccountID && isStrict {
		return nil
	}

	return withSchemaRebuildTx(ctx, db, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `
			CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_account_id_id
			ON clients(account_id, id);
		`); err != nil {
			return fmt.Errorf("ensure clients composite key before invoices rebuild: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			CREATE UNIQUE INDEX IF NOT EXISTS idx_invoice_revisions_id_invoice
			ON invoice_revisions(id, invoice_id);
		`); err != nil {
			return fmt.Errorf("ensure invoice revisions composite key before invoices rebuild: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `ALTER TABLE invoices RENAME TO invoices_legacy;`); err != nil {
			return fmt.Errorf("rename legacy invoices table: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			CREATE TABLE invoices (
				id INTEGER PRIMARY KEY,
				account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id) ON DELETE CASCADE,
				client_id INTEGER NOT NULL,
				current_revision_id INTEGER
					REFERENCES invoice_revisions(id)
					ON DELETE SET NULL
					DEFERRABLE INITIALLY DEFERRED,
				base_number INTEGER NOT NULL CHECK (base_number > 0),
				status TEXT NOT NULL DEFAULT 'draft'
					CHECK (status IN ('draft','issued','paid','void')),
				created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
				FOREIGN KEY (account_id, client_id) REFERENCES clients(account_id, id) ON DELETE RESTRICT,
				UNIQUE (account_id, base_number),
				UNIQUE (id, account_id, client_id)
			);
		`); err != nil {
			return fmt.Errorf("create strict invoices table: %w", err)
		}

		selectAccountID := "account_id"
		if !hasAccountID {
			selectAccountID = "1"
		}

		if _, err := tx.ExecContext(ctx, fmt.Sprintf(`
			INSERT INTO invoices (
				id,
				account_id,
				client_id,
				current_revision_id,
				base_number,
				status,
				created_at
			)
			SELECT
				id,
				%s,
				client_id,
				current_revision_id,
				base_number,
				status,
				created_at
			FROM invoices_legacy;
		`, selectAccountID)); err != nil {
			return fmt.Errorf("copy legacy invoices into strict table: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `DROP TABLE invoices_legacy;`); err != nil {
			return fmt.Errorf("drop legacy invoices table: %w", err)
		}

		return nil
	})
}
