// internal/db/migrate.go
package db

import (
	"context"
	"database/sql"
	"fmt"
)

// ---------------------
// CLIENTS: cannot be deleted if they have any invoices
// PRODUCT TYPES:
// Style - flat price rate | Sample - Hourly Price * Hours Worked | Sample - Flat Price
// ---------------------

// Populates tables in an sqlite DB if they don't exist
func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("enable foreign_keys: %w", err)
	}

	stmts := []string{
		// -----------------------
		// Auth / access
		// -----------------------
		`CREATE TABLE IF NOT EXISTS allowed_users (
			id INTEGER PRIMARY KEY,
			email TEXT NOT NULL UNIQUE
		);`,

		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			name TEXT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		);`,

		// -----------------------
		// Clients
		// -----------------------
		`CREATE TABLE IF NOT EXISTS clients (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			company_name TEXT,
			address TEXT,
			email TEXT,
			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			updated_at TEXT
		);`,

		// -----------------------
		// Products
		// -----------------------
		`CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY,

			product_type TEXT NOT NULL CHECK (product_type IN ('style','sample')),
			pricing_mode TEXT NOT NULL CHECK (pricing_mode IN ('flat','hourly')),

			name TEXT NOT NULL,

			flat_price_minor INTEGER CHECK (flat_price_minor IS NULL OR flat_price_minor >= 0),
			hourly_rate_minor INTEGER CHECK (hourly_rate_minor IS NULL OR hourly_rate_minor >= 0),

			default_minutes_worked INTEGER
				CHECK (default_minutes_worked IS NULL OR default_minutes_worked >= 0),

			client_id INTEGER,

			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			updated_at TEXT,

			FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE,

			-- style must be flat
			CHECK (NOT (product_type = 'style' AND pricing_mode <> 'flat')),

			-- chosen pricing_mode must have its required price
			CHECK (
				(pricing_mode = 'flat'   AND flat_price_minor IS NOT NULL) OR
				(pricing_mode = 'hourly' AND hourly_rate_minor IS NOT NULL)
			),

			-- default minutes required only for hourly samples
			CHECK (
				(product_type = 'sample' AND pricing_mode = 'hourly' AND default_minutes_worked IS NOT NULL)
				OR
				(NOT (product_type = 'sample' AND pricing_mode = 'hourly') AND default_minutes_worked IS NULL)
			)
				
		);`,

		// -----------------------
		// Invoices
		// -----------------------
		`CREATE TABLE IF NOT EXISTS invoices (
			id INTEGER PRIMARY KEY,
			client_id INTEGER NOT NULL,

			base_number INTEGER NOT NULL UNIQUE CHECK (base_number > 0),

			status TEXT NOT NULL DEFAULT 'draft'
				CHECK (status IN ('draft','issued','paid','void')),

			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
			updated_at TEXT,

			FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE RESTRICT
		);`,

		`CREATE TABLE IF NOT EXISTS invoice_revisions (
			id INTEGER PRIMARY KEY,
			invoice_id INTEGER NOT NULL,
			revision_no INTEGER NOT NULL CHECK (revision_no >= 1),

			issue_date TEXT NOT NULL,
			due_by_date TEXT NOT NULL,
			
			-- Client snapshot (frozen at time of this revision)
			client_name TEXT NOT NULL,
			client_company_name TEXT NOT NULL DEFAULT '',
			client_address TEXT NOT NULL DEFAULT '',
			client_email TEXT NOT NULL DEFAULT '',
			
			note TEXT,

			vat_rate_bps INTEGER NOT NULL DEFAULT 2000 CHECK (vat_rate_bps >= 0),

			discount_type TEXT NOT NULL DEFAULT 'none'
				CHECK (discount_type IN ('none','percent','fixed')),

			discount_value INTEGER NOT NULL DEFAULT 0 CHECK (discount_value >= 0),

			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),

			FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
			UNIQUE (invoice_id, revision_no),

			CHECK (
				(discount_type = 'none'    AND discount_value = 0) OR
				(discount_type = 'percent' AND discount_value BETWEEN 0 AND 10000) OR
				(discount_type = 'fixed'   AND discount_value >= 0)
			)
		);`,

		// -----------------------
		// Invoice items
		// -----------------------
		`CREATE TABLE IF NOT EXISTS invoice_items (
			id INTEGER PRIMARY KEY,
			invoice_revision_id INTEGER NOT NULL,

			product_id INTEGER,

			name TEXT NOT NULL,

			line_type TEXT NOT NULL DEFAULT 'custom'
				CHECK (line_type IN ('style','sample','custom')),

			-- How this line is priced self-describing for the editor
			pricing_mode TEXT NOT NULL DEFAULT 'flat'
				CHECK (pricing_mode IN ('flat','hourly')),

			-- For flat lines: quantity * unit_price_minor
			quantity REAL NOT NULL DEFAULT 1 CHECK (quantity > 0),
			unit_price_minor INTEGER NOT NULL CHECK (unit_price_minor >= 0),

			-- For hourly lines: minutes_worked must be present
			minutes_worked INTEGER CHECK (minutes_worked IS NULL OR minutes_worked >= 0),

			sort_order INTEGER NOT NULL DEFAULT 1 CHECK (sort_order >= 1),

			FOREIGN KEY (invoice_revision_id) REFERENCES invoice_revisions(id) ON DELETE CASCADE,
			FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE SET NULL,

			UNIQUE (invoice_revision_id, sort_order),

			CHECK (
				(pricing_mode = 'flat'  AND minutes_worked IS NULL) OR
				(pricing_mode = 'hourly' AND minutes_worked IS NOT NULL)
			)
		);`,

		// -----------------------
		// Payments
		// -----------------------
		`CREATE TABLE IF NOT EXISTS payments (
			id INTEGER PRIMARY KEY,
			invoice_id INTEGER NOT NULL,

			kind TEXT NOT NULL DEFAULT 'payment'
				CHECK (kind IN ('deposit','payment')),

			amount_minor INTEGER NOT NULL CHECK (amount_minor > 0),

			label TEXT,

			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),

			FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE
		);`,

		// -----------------------
		// Indexes
		// -----------------------
		`CREATE INDEX IF NOT EXISTS idx_invoices_client_id ON invoices(client_id);`,
		`CREATE INDEX IF NOT EXISTS idx_revisions_invoice_id ON invoice_revisions(invoice_id);`,
		`CREATE INDEX IF NOT EXISTS idx_items_revision_id ON invoice_items(invoice_revision_id);`,
		`CREATE INDEX IF NOT EXISTS idx_revisions_invoice_id_revno ON invoice_revisions(invoice_id, revision_no);`,
		`CREATE INDEX IF NOT EXISTS idx_items_product_id ON invoice_items(product_id);`,
		`CREATE INDEX IF NOT EXISTS idx_products_client_id ON products(client_id);`,
		`CREATE INDEX IF NOT EXISTS idx_payments_invoice_id ON payments(invoice_id);`,
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("enable foreign_keys in tx: %w", err)
	}

	for i, stmt := range stmts {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("migration step %d failed: %w\nSQL: %s", i+1, err, stmt)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
