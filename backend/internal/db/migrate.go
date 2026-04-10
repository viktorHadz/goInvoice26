package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/transaction/accessTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/authTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/billingTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/settingsTx"
)

// Migrate ensures the application schema exists and upgrades legacy databases
// to the current multi-tenant layout.
func Migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("enable foreign_keys: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `PRAGMA foreign_keys = ON;`); err != nil {
		return fmt.Errorf("enable foreign_keys in tx: %w", err)
	}

	if err := ensureBaseSchema(ctx, tx); err != nil {
		return err
	}
	if err := settingsTx.EnsureAccountSettingsColumns(ctx, tx); err != nil {
		return err
	}
	if err := settingsTx.EnsureShowItemTypeHeadersColumn(ctx, tx); err != nil {
		return err
	}
	if err := settingsTx.EnsureUsersAccountIDColumn(ctx, tx); err != nil {
		return err
	}
	if err := ensureClientsAccountIDColumn(ctx, tx); err != nil {
		return err
	}
	if err := ensureProductsAccountIDColumn(ctx, tx); err != nil {
		return err
	}
	if err := ensureInvoiceSupplyDateColumn(ctx, tx); err != nil {
		return err
	}
	if err := ensurePaymentReceiptNumberColumn(ctx, tx); err != nil {
		return err
	}
	if err := authTx.EnsureUsersGoogleSubColumn(ctx, tx); err != nil {
		return err
	}
	if err := billingTx.EnsureAccountsStripeCustomerIDColumn(ctx, tx); err != nil {
		return err
	}
	if err := billingTx.EnsureAccountsStripeSubscriptionIDColumn(ctx, tx); err != nil {
		return err
	}
	if err := billingTx.EnsureAccountsBillingPriceIDColumn(ctx, tx); err != nil {
		return err
	}
	if err := billingTx.EnsureAccountsBillingPlanColumn(ctx, tx); err != nil {
		return err
	}
	if err := billingTx.EnsureAccountsBillingIntervalColumn(ctx, tx); err != nil {
		return err
	}
	if err := billingTx.EnsureAccountsBillingEmailColumn(ctx, tx); err != nil {
		return err
	}
	if err := billingTx.EnsureAccountsBillingStatusColumn(ctx, tx); err != nil {
		return err
	}
	if err := billingTx.EnsureAccountsBillingCurrentPeriodEndColumn(ctx, tx); err != nil {
		return err
	}
	if err := billingTx.EnsureAccountsBillingCancelAtPeriodEndColumn(ctx, tx); err != nil {
		return err
	}
	if err := billingTx.EnsureAccountsBillingUpdatedAtColumn(ctx, tx); err != nil {
		return err
	}
	if err := authTx.EnsureUsersAvatarURLColumn(ctx, tx); err != nil {
		return err
	}
	if err := authTx.EnsureUsersRoleColumn(ctx, tx); err != nil {
		return err
	}
	if err := authTx.EnsureAllowedUsersAccountIDColumn(ctx, tx); err != nil {
		return err
	}
	if err := authTx.EnsureAllowedUsersCreatedAtColumn(ctx, tx); err != nil {
		return err
	}
	if err := authTx.EnsureAllowedUsersInvitedByUserIDColumn(ctx, tx); err != nil {
		return err
	}
	if err := accessTx.EnsureDirectAccessGrantsTable(ctx, tx); err != nil {
		return err
	}
	if err := accessTx.EnsurePromoCodesTable(ctx, tx); err != nil {
		return err
	}
	if err := accessTx.EnsurePromoCodeRedemptionsTable(ctx, tx); err != nil {
		return err
	}
	if err := accessTx.EnsurePromoCodeRedemptionClaimsTable(ctx, tx); err != nil {
		return err
	}
	if err := settingsTx.MigrateLegacyUserSettings(ctx, tx); err != nil {
		return err
	}
	if err := reconcileInvoiceStatusesToSavedPayments(ctx, tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	if err := migrateInvoiceNumberSequencesToAccounts(ctx, db); err != nil {
		return err
	}
	if err := ensureStrictProductsTable(ctx, db); err != nil {
		return err
	}
	if err := migrateInvoicesToAccounts(ctx, db); err != nil {
		return err
	}
	if err := ensurePostRebuildIndexes(ctx, db); err != nil {
		return err
	}
	if err := ensureTenantIntegrityTriggers(ctx, db); err != nil {
		return err
	}
	if err := validateTenantIntegrity(ctx, db); err != nil {
		return err
	}

	return nil
}
