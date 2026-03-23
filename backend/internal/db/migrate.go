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

		`CREATE TABLE IF NOT EXISTS user_settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			company_name TEXT NOT NULL DEFAULT '',
			email TEXT NOT NULL DEFAULT '',
			phone TEXT NOT NULL DEFAULT '',
			company_address TEXT NOT NULL DEFAULT '',
			invoice_prefix TEXT NOT NULL DEFAULT 'INV-',
			currency TEXT NOT NULL DEFAULT 'GBP',
			date_format TEXT NOT NULL DEFAULT 'dd/mm/yyyy',
			payment_terms TEXT NOT NULL DEFAULT 'Please make payment within 14 days.',
			payment_details TEXT NOT NULL DEFAULT '',
			notes_footer TEXT NOT NULL DEFAULT '',
			logo_url TEXT NOT NULL DEFAULT ''
		);`,
		`INSERT OR IGNORE INTO user_settings (
			id,
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
			logo_url
		) VALUES (
			1,
			'',
			'',
			'',
			'',
			'INV-',
			'GBP',
			'dd/mm/yyyy',
			'Please make payment within 14 days.',
			'',
			'',
			''
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

			-- stores latest invoice revision (if 1.1, 1.2, 1.3 it returns 1.3's id)  
			current_revision_id INTEGER
				REFERENCES invoice_revisions(id)
				ON DELETE SET NULL
				DEFERRABLE INITIALLY DEFERRED,

			base_number INTEGER NOT NULL UNIQUE CHECK (base_number > 0),

			status TEXT NOT NULL DEFAULT 'draft'
				CHECK (status IN ('draft','issued','paid','void')),

			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),

			FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE RESTRICT
		);`,

		// -----------------------
		// Invoice number allocator
		// -----------------------
		`CREATE TABLE IF NOT EXISTS invoice_number_seq (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			next_base_number INTEGER NOT NULL CHECK (next_base_number > 0)
		);`,

		// ensures the single row exists - initial value cam be adjusted in allocator for migrations
		`INSERT OR IGNORE INTO invoice_number_seq (id, next_base_number) VALUES (1, 1);`,

		// -----------------------
		// Invoice Revisions
		// -----------------------
		`CREATE TABLE IF NOT EXISTS invoice_revisions (
			id INTEGER PRIMARY KEY,
			invoice_id INTEGER NOT NULL,
			revision_no INTEGER NOT NULL CHECK (revision_no >= 1),

			issue_date TEXT NOT NULL,
			due_by_date TEXT,

			updated_at TEXT,

			-- Client Snapshot - copied and stored for each invoice 
			client_name TEXT NOT NULL,
			client_company_name TEXT NOT NULL DEFAULT '',
			client_address TEXT NOT NULL DEFAULT '',
			client_email TEXT NOT NULL DEFAULT '',
			
			note TEXT,

			vat_rate INTEGER NOT NULL DEFAULT 2000 CHECK (vat_rate BETWEEN 0 AND 10000),

			-- disc/depo type - 'none', 'percent', 'fixed'
			-- rate fields are basis points: 1000 = 10%, 10000 = 100%
			-- minor fields are currency minor units: 1000 = £10.00
			discount_type TEXT NOT NULL DEFAULT 'none'
				CHECK (discount_type IN ('none','percent','fixed')),
			discount_rate INTEGER NOT NULL DEFAULT 0 CHECK (discount_rate BETWEEN 0 AND 10000),
			discount_minor INTEGER NOT NULL DEFAULT 0 CHECK (discount_minor >= 0),

			deposit_type TEXT NOT NULL DEFAULT 'none'
				CHECK (deposit_type IN ('none','percent','fixed')),
			deposit_rate INTEGER NOT NULL DEFAULT 0 CHECK (deposit_rate BETWEEN 0 AND 10000),
			deposit_minor INTEGER NOT NULL DEFAULT 0 CHECK (deposit_minor >= 0),

			subtotal_minor INTEGER NOT NULL CHECK (subtotal_minor >= 0),
			vat_amount_minor INTEGER NOT NULL CHECK (vat_amount_minor >= 0),
			total_minor INTEGER NOT NULL CHECK (total_minor >= 0),

			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),

			FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
			UNIQUE (invoice_id, revision_no),

			CHECK (
				(discount_type = 'none' AND discount_rate = 0 AND discount_minor = 0) OR
				(discount_type = 'percent' AND discount_rate BETWEEN 0 AND 10000) OR
				(discount_type = 'fixed' AND discount_rate = 0)
			),
			CHECK (
				(deposit_type = 'none' AND deposit_rate = 0 AND deposit_minor = 0) OR
				(deposit_type = 'percent' AND deposit_rate BETWEEN 0 AND 10000) OR
				(deposit_type = 'fixed' AND deposit_rate = 0)
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

			pricing_mode TEXT NOT NULL DEFAULT 'flat'
				CHECK (pricing_mode IN ('flat','hourly')),

			-- For flat lines: quantity * unit_price_minor
			quantity INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
			unit_price_minor INTEGER NOT NULL CHECK (unit_price_minor >= 0),
			line_total_minor INTEGER NOT NULL DEFAULT 0 CHECK (line_total_minor >= 0),

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

			payment_type TEXT NOT NULL DEFAULT 'payment'
				CHECK (payment_type IN ('deposit','payment')),
			amount_minor INTEGER NOT NULL CHECK (amount_minor > 0),
			payment_date TEXT NOT NULL,
			applied_in_revision_id INTEGER
				REFERENCES invoice_revisions(id)
				ON DELETE SET NULL
				DEFERRABLE INITIALLY DEFERRED,

			label TEXT,

			created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ','now')),

			FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE
		);`,

		// -----------------------
		// Views
		// -----------------------
		`CREATE INDEX IF NOT EXISTS idx_invoices_current_revision_id ON invoices(current_revision_id);`,
		// Load current items for invoice/invoice_revision without finding revision_id
		`CREATE VIEW IF NOT EXISTS invoice_current_items AS
		SELECT
			i.id AS invoice_id,
			i.current_revision_id AS revision_id,

			r.revision_no,

			it.id AS item_id,
			it.sort_order,
			it.product_id,
			it.name,
			it.line_type,
			it.pricing_mode,
			it.quantity,
			it.unit_price_minor,
			it.minutes_worked,
			it.line_total_minor

		FROM invoices i
		JOIN invoice_revisions r
			ON r.id = i.current_revision_id
		JOIN invoice_items it
			ON it.invoice_revision_id = r.id
		ORDER BY i.id, it.sort_order;`,

		// Initial invoice and revision number data
		`CREATE VIEW IF NOT EXISTS invoice_book_rows AS
			SELECT
				i.id,
				i.client_id,
				i.base_number,
				i.status,
				i.current_revision_id,
				r.revision_no,
				r.issue_date,
				r.due_by_date,
				r.updated_at
			FROM invoices i
			JOIN invoice_revisions r
			ON r.id = i.current_revision_id;`,

		// items for a specific invoice revision
		`CREATE VIEW IF NOT EXISTS invoice_revision_items AS
		SELECT
			r.invoice_id,
			r.id AS revision_id,
			r.revision_no,
			it.id AS item_id,
			it.sort_order,
			it.product_id,
			it.name,
			it.line_type,
			it.pricing_mode,
			it.quantity,
			it.unit_price_minor,
			it.minutes_worked,
			it.line_total_minor
		FROM invoice_revisions r
		JOIN invoice_items it
			ON it.invoice_revision_id = r.id;
		`,

		// -----------------------
		// Indexes
		// -----------------------
		`CREATE INDEX IF NOT EXISTS idx_invoices_client_id ON invoices(client_id);`,
		// Invoice Book
		`CREATE INDEX IF NOT EXISTS idx_invoices_client_base ON invoices(client_id, base_number DESC);`,
		`CREATE INDEX IF NOT EXISTS idx_revisions_invoice_id ON invoice_revisions(invoice_id);`,
		`CREATE INDEX IF NOT EXISTS idx_revisions_invoice_id_revno ON invoice_revisions(invoice_id, revision_no);`,
		`CREATE INDEX IF NOT EXISTS idx_items_revision_id ON invoice_items(invoice_revision_id);`,
		`CREATE INDEX IF NOT EXISTS idx_items_product_id ON invoice_items(product_id);`,
		`CREATE INDEX IF NOT EXISTS idx_products_client_id ON products(client_id);`,
		`CREATE INDEX IF NOT EXISTS idx_payments_invoice_id ON payments(invoice_id);`,
		`CREATE INDEX IF NOT EXISTS idx_payments_invoice_revision ON payments(invoice_id, applied_in_revision_id);`,
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
			return fmt.Errorf("migration step %d failed: %w\nSQL: %s,\nctx: %v,", i+1, err, stmt, ctx)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

/*

USAGE EXAMPLES FOR VIEWS

--- load all items for specific invoice ---
	SELECT * FROM invoice_latest_items WHERE invoice_id = ?;

--- load items for a specific revision ---
	SELECT * FROM invoice_revision_items WHERE invoice_id = ? AND revision_no = ? ORDER BY sort_order;

-- load current revision's invoice items
	SELECT * FROM invoice_current_items WHERE invoice_id = ?;
*/
