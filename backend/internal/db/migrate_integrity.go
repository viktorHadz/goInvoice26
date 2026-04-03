package db

import (
	"context"
	"database/sql"
	"fmt"
)

func ensurePostRebuildIndexes(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE INDEX IF NOT EXISTS idx_clients_account_id ON clients(account_id);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_clients_account_id_id ON clients(account_id, id);`,
		`CREATE INDEX IF NOT EXISTS idx_invoices_account_id ON invoices(account_id);`,
		`CREATE INDEX IF NOT EXISTS idx_invoices_client_id ON invoices(client_id);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_id_account_client ON invoices(id, account_id, client_id);`,
		`CREATE INDEX IF NOT EXISTS idx_invoices_client_base ON invoices(account_id, client_id, base_number DESC);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_invoice_revisions_id_invoice ON invoice_revisions(id, invoice_id);`,
		`CREATE INDEX IF NOT EXISTS idx_products_account_client ON products(account_id, client_id);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_products_account_client_id ON products(account_id, client_id, id);`,
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("ensure post-rebuild index: %w", err)
		}
	}

	return nil
}

func ensureTenantIntegrityTriggers(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TRIGGER IF NOT EXISTS trg_accounts_id_immutable
		BEFORE UPDATE OF id ON accounts
		FOR EACH ROW
		BEGIN
			SELECT RAISE(ABORT, 'account identity is immutable');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_users_scope_immutable
		BEFORE UPDATE OF id, account_id ON users
		FOR EACH ROW
		BEGIN
			SELECT RAISE(ABORT, 'user ownership is immutable');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_allowed_users_scope_immutable
		BEFORE UPDATE OF id, account_id ON allowed_users
		FOR EACH ROW
		BEGIN
			SELECT RAISE(ABORT, 'invite ownership is immutable');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_auth_sessions_scope_immutable
		BEFORE UPDATE OF id, account_id, user_id ON auth_sessions
		FOR EACH ROW
		BEGIN
			SELECT RAISE(ABORT, 'session ownership is immutable');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_stored_files_scope_immutable
		BEFORE UPDATE OF id, account_id ON stored_files
		FOR EACH ROW
		BEGIN
			SELECT RAISE(ABORT, 'stored file ownership is immutable');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_account_settings_scope_immutable
		BEFORE UPDATE OF account_id ON account_settings
		FOR EACH ROW
		BEGIN
			SELECT RAISE(ABORT, 'settings ownership is immutable');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_invoice_number_seq_scope_immutable
		BEFORE UPDATE OF account_id ON invoice_number_seq
		FOR EACH ROW
		BEGIN
			SELECT RAISE(ABORT, 'invoice sequence ownership is immutable');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_clients_account_immutable
		BEFORE UPDATE OF id, account_id ON clients
		FOR EACH ROW
		BEGIN
			SELECT RAISE(ABORT, 'client ownership is immutable');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_products_scope_immutable
		BEFORE UPDATE OF id, account_id, client_id ON products
		FOR EACH ROW
		BEGIN
			SELECT RAISE(ABORT, 'product ownership is immutable');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_invoices_scope_immutable
		BEFORE UPDATE OF id, account_id, client_id ON invoices
		FOR EACH ROW
		BEGIN
			SELECT RAISE(ABORT, 'invoice ownership is immutable');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_invoice_revisions_scope_immutable
		BEFORE UPDATE OF id, invoice_id ON invoice_revisions
		FOR EACH ROW
		BEGIN
			SELECT RAISE(ABORT, 'invoice revision ownership is immutable');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_invoices_current_revision_matches_invoice_insert
		AFTER INSERT ON invoices
		FOR EACH ROW
		WHEN NEW.current_revision_id IS NOT NULL
			AND NOT EXISTS (
				SELECT 1
				FROM invoice_revisions r
				WHERE r.id = NEW.current_revision_id
				  AND r.invoice_id = NEW.id
			)
		BEGIN
			SELECT RAISE(ABORT, 'invoice current revision must belong to same invoice');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_invoices_current_revision_matches_invoice_update
		AFTER UPDATE OF current_revision_id ON invoices
		FOR EACH ROW
		WHEN NEW.current_revision_id IS NOT NULL
			AND NOT EXISTS (
				SELECT 1
				FROM invoice_revisions r
				WHERE r.id = NEW.current_revision_id
				  AND r.invoice_id = NEW.id
			)
		BEGIN
			SELECT RAISE(ABORT, 'invoice current revision must belong to same invoice');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_invoice_items_product_scope_insert
		BEFORE INSERT ON invoice_items
		FOR EACH ROW
		WHEN NEW.product_id IS NOT NULL
			AND NOT EXISTS (
				SELECT 1
				FROM invoice_revisions r
				JOIN invoices i
					ON i.id = r.invoice_id
				JOIN products p
					ON p.id = NEW.product_id
				WHERE r.id = NEW.invoice_revision_id
				  AND p.account_id = i.account_id
				  AND p.client_id = i.client_id
			)
		BEGIN
			SELECT RAISE(ABORT, 'invoice item product must belong to same account and client');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_invoice_items_product_scope_update
		BEFORE UPDATE OF invoice_revision_id, product_id ON invoice_items
		FOR EACH ROW
		WHEN NEW.product_id IS NOT NULL
			AND NOT EXISTS (
				SELECT 1
				FROM invoice_revisions r
				JOIN invoices i
					ON i.id = r.invoice_id
				JOIN products p
					ON p.id = NEW.product_id
				WHERE r.id = NEW.invoice_revision_id
				  AND p.account_id = i.account_id
				  AND p.client_id = i.client_id
			)
		BEGIN
			SELECT RAISE(ABORT, 'invoice item product must belong to same account and client');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_payments_applied_revision_matches_invoice_insert
		BEFORE INSERT ON payments
		FOR EACH ROW
		WHEN NEW.applied_in_revision_id IS NOT NULL
			AND NOT EXISTS (
				SELECT 1
				FROM invoice_revisions r
				WHERE r.id = NEW.applied_in_revision_id
				  AND r.invoice_id = NEW.invoice_id
			)
		BEGIN
			SELECT RAISE(ABORT, 'payment applied revision must belong to same invoice');
		END;`,
		`CREATE TRIGGER IF NOT EXISTS trg_payments_applied_revision_matches_invoice_update
		BEFORE UPDATE OF invoice_id, applied_in_revision_id ON payments
		FOR EACH ROW
		WHEN NEW.applied_in_revision_id IS NOT NULL
			AND NOT EXISTS (
				SELECT 1
				FROM invoice_revisions r
				WHERE r.id = NEW.applied_in_revision_id
				  AND r.invoice_id = NEW.invoice_id
			)
		BEGIN
			SELECT RAISE(ABORT, 'payment applied revision must belong to same invoice');
		END;`,
	}

	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("ensure tenant integrity trigger: %w", err)
		}
	}

	return nil
}

func validateTenantIntegrity(ctx context.Context, db *sql.DB) error {
	type validationQuery struct {
		name  string
		query string
	}

	checks := []validationQuery{
		{
			name: "invoice current revision ownership",
			query: `
				SELECT i.id
				FROM invoices i
				LEFT JOIN invoice_revisions r
					ON r.id = i.current_revision_id
				WHERE i.current_revision_id IS NOT NULL
				  AND (r.id IS NULL OR r.invoice_id <> i.id)
				LIMIT 1;
			`,
		},
		{
			name: "payment applied revision ownership",
			query: `
				SELECT p.id
				FROM payments p
				LEFT JOIN invoice_revisions r
					ON r.id = p.applied_in_revision_id
				WHERE p.applied_in_revision_id IS NOT NULL
				  AND (r.id IS NULL OR r.invoice_id <> p.invoice_id)
				LIMIT 1;
			`,
		},
		{
			name: "invoice item product ownership",
			query: `
				SELECT it.id
				FROM invoice_items it
				JOIN invoice_revisions r
					ON r.id = it.invoice_revision_id
				JOIN invoices i
					ON i.id = r.invoice_id
				LEFT JOIN products p
					ON p.id = it.product_id
				WHERE it.product_id IS NOT NULL
				  AND (p.id IS NULL OR p.account_id <> i.account_id OR p.client_id <> i.client_id)
				LIMIT 1;
			`,
		},
	}

	for _, check := range checks {
		var offendingID sql.NullInt64
		err := db.QueryRowContext(ctx, check.query).Scan(&offendingID)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return fmt.Errorf("validate %s: %w", check.name, err)
		}
		if offendingID.Valid {
			return fmt.Errorf("%s violated for row id %d", check.name, offendingID.Int64)
		}
	}

	return nil
}
