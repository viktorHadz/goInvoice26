package workspace_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/billingplan"
	"github.com/viktorHadz/goInvoice26/internal/db"
	billingsvc "github.com/viktorHadz/goInvoice26/internal/service/billing"
	"github.com/viktorHadz/goInvoice26/internal/service/logo"
	"github.com/viktorHadz/goInvoice26/internal/service/storage"
	"github.com/viktorHadz/goInvoice26/internal/service/workspace"
	"github.com/viktorHadz/goInvoice26/internal/transaction/authTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/billingTx"
)

type billingStub struct {
	cancelCount int
	cancelErr   error
}

func (b *billingStub) CancelSubscriptionImmediately(_ context.Context, _ int64) error {
	b.cancelCount++
	return b.cancelErr
}

func newWorkspaceService(t *testing.T, billing *billingStub) (*sql.DB, *storage.LocalStore, *workspace.Service, func()) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "workspace.sqlite")
	uploadRoot := filepath.Join(dir, "uploads")

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	store := storage.NewLocalStore(uploadRoot)
	var canceler interface {
		CancelSubscriptionImmediately(context.Context, int64) error
	}
	if billing != nil {
		canceler = billing
	}
	service := workspace.NewService(conn, canceler, store)

	return conn, store, service, func() {
		_ = conn.Close()
	}
}

func seedWorkspaceData(t *testing.T, conn *sql.DB, store *storage.LocalStore) int64 {
	t.Helper()

	ctx := context.Background()
	owner, err := authTx.CreateAccountOwner(ctx, conn, authTx.CreateGoogleUserParams{
		Name:      "Owner Example",
		Email:     "owner@example.com",
		GoogleSub: "owner-sub",
		Role:      authTx.UserRoleOwner,
	})
	if err != nil {
		t.Fatalf("CreateAccountOwner: %v", err)
	}

	if _, err := authTx.CreateMemberFromGoogle(ctx, conn, authTx.CreateGoogleUserParams{
		Name:      "Member Example",
		Email:     "member@example.com",
		GoogleSub: "member-sub",
		Role:      authTx.UserRoleMember,
		AccountID: owner.AccountID,
	}); err != nil {
		t.Fatalf("CreateMemberFromGoogle: %v", err)
	}
	if _, err := authTx.CreateInvite(
		ctx,
		conn,
		owner.AccountID,
		owner.ID,
		"invite@example.com",
		billingplan.TeamSeatLimit,
	); err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if _, err := conn.ExecContext(ctx, `INSERT INTO clients (account_id, name) VALUES (?, 'Client')`, owner.AccountID); err != nil {
		t.Fatalf("insert client: %v", err)
	}

	logoService := logo.NewService(conn, store)
	if _, err := logoService.Replace(ctx, owner.AccountID, bytes.NewReader([]byte("logo")), ".png", "image/png"); err != nil {
		t.Fatalf("Replace logo: %v", err)
	}

	return owner.AccountID
}

func TestDeleteAccount_RemovesWorkspaceRowsAndUploads(t *testing.T) {
	ctx := context.Background()
	billing := &billingStub{}
	conn, store, service, cleanup := newWorkspaceService(t, billing)
	defer cleanup()

	accountID := seedWorkspaceData(t, conn, store)
	if err := billingTx.UpdateAccountBilling(ctx, conn, billingTx.UpdateAccountBillingParams{
		AccountID:            accountID,
		StripeCustomerID:     "cus_123",
		StripeSubscriptionID: "sub_123",
		BillingStatus:        "active",
		BillingUpdatedAt:     now(),
	}); err != nil {
		t.Fatalf("seed billing state: %v", err)
	}

	if err := service.DeleteAccount(ctx, accountID); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}
	if billing.cancelCount != 1 {
		t.Fatalf("cancelCount = %d, want 1", billing.cancelCount)
	}

	assertCount := func(query string, args ...any) int {
		t.Helper()
		var count int
		if err := conn.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
			t.Fatalf("count query failed: %v", err)
		}
		return count
	}

	if got := assertCount(`SELECT COUNT(*) FROM accounts WHERE id = ?`, accountID); got != 0 {
		t.Fatalf("accounts count = %d, want 0", got)
	}
	if got := assertCount(`SELECT COUNT(*) FROM users WHERE account_id = ?`, accountID); got != 0 {
		t.Fatalf("users count = %d, want 0", got)
	}
	if got := assertCount(`SELECT COUNT(*) FROM allowed_users WHERE account_id = ?`, accountID); got != 0 {
		t.Fatalf("allowed_users count = %d, want 0", got)
	}
	if got := assertCount(`SELECT COUNT(*) FROM stored_files WHERE account_id = ?`, accountID); got != 0 {
		t.Fatalf("stored_files count = %d, want 0", got)
	}
	if got := assertCount(`SELECT COUNT(*) FROM clients WHERE account_id = ?`, accountID); got != 0 {
		t.Fatalf("clients count = %d, want 0", got)
	}
	if _, err := os.Stat(store.AccountDir(accountID)); !os.IsNotExist(err) {
		t.Fatalf("account uploads dir still exists, err=%v", err)
	}
}

func TestDeleteAccount_BlocksWhenSubscriptionExistsButCancellationUnavailable(t *testing.T) {
	ctx := context.Background()
	conn, store, service, cleanup := newWorkspaceService(t, nil)
	defer cleanup()

	accountID := seedWorkspaceData(t, conn, store)
	if err := billingTx.UpdateAccountBilling(ctx, conn, billingTx.UpdateAccountBillingParams{
		AccountID:            accountID,
		StripeCustomerID:     "cus_123",
		StripeSubscriptionID: "sub_123",
		BillingStatus:        "active",
		BillingUpdatedAt:     now(),
	}); err != nil {
		t.Fatalf("seed billing state: %v", err)
	}

	err := service.DeleteAccount(ctx, accountID)
	if !errors.Is(err, workspace.ErrDeleteBlockedByBilling) {
		t.Fatalf("DeleteAccount error = %v, want %v", err, workspace.ErrDeleteBlockedByBilling)
	}
	if _, statErr := os.Stat(store.AccountDir(accountID)); statErr != nil {
		t.Fatalf("account uploads dir missing after blocked delete: %v", statErr)
	}
}

func TestDeleteAccount_RestoresUploadsWhenDatabaseDeleteFails(t *testing.T) {
	ctx := context.Background()
	billing := &billingStub{cancelErr: billingsvc.ErrNotConfigured}
	conn, store, service, cleanup := newWorkspaceService(t, billing)
	defer cleanup()

	accountID := seedWorkspaceData(t, conn, store)
	if err := billingTx.UpdateAccountBilling(ctx, conn, billingTx.UpdateAccountBillingParams{
		AccountID:            accountID,
		StripeCustomerID:     "cus_123",
		StripeSubscriptionID: "sub_123",
		BillingStatus:        "active",
		BillingUpdatedAt:     now(),
	}); err != nil {
		t.Fatalf("seed billing state: %v", err)
	}

	err := service.DeleteAccount(ctx, accountID)
	if !errors.Is(err, workspace.ErrDeleteBlockedByBilling) && !errors.Is(err, billingsvc.ErrNotConfigured) {
		t.Fatalf("DeleteAccount error = %v, want billing-related error", err)
	}
	if _, statErr := os.Stat(store.AccountDir(accountID)); statErr != nil {
		t.Fatalf("account uploads dir missing after failed delete: %v", statErr)
	}
}

func now() time.Time {
	return time.Now().UTC()
}
