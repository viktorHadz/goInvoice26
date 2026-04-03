package auth

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/transaction/accessTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/authTx"
)

func newAuthServiceTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	conn, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "auth-service.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	return conn, func() {
		_ = conn.Close()
	}
}

func createOwnerAndSession(t *testing.T, conn *sql.DB, email string) (authTx.User, string) {
	t.Helper()

	owner, err := authTx.CreateAccountOwner(context.Background(), conn, authTx.CreateGoogleUserParams{
		Name:      "Owner Example",
		Email:     email,
		GoogleSub: "sub-" + email,
		Role:      authTx.UserRoleOwner,
	})
	if err != nil {
		t.Fatalf("CreateAccountOwner: %v", err)
	}

	token := "session-" + email
	if err := authTx.CreateSession(context.Background(), conn, owner.ID, owner.AccountID, hashToken(token), time.Now().Add(24*time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	return owner, token
}

func TestResolveSession_UsesDirectAccessGrant(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAuthServiceTestDB(t)
	defer cleanup()

	owner, token := createOwnerAndSession(t, conn, "vikecah@gmail.com")
	if _, err := accessTx.CreateDirectAccessGrant(ctx, conn, owner.Email, "team", "platform invite", owner.ID); err != nil {
		t.Fatalf("CreateDirectAccessGrant: %v", err)
	}

	service := NewService(conn, Config{
		BillingConfigured:  true,
		BillingTrialDays:   7,
		PlatformAdminEmail: "vikecah@gmail.com",
	})

	principal, ok, err := service.ResolveSession(ctx, token)
	if err != nil {
		t.Fatalf("ResolveSession: %v", err)
	}
	if !ok {
		t.Fatal("ResolveSession ok = false, want true")
	}
	if !principal.Billing.AccessGranted {
		t.Fatalf("AccessGranted = false, want true")
	}
	if principal.Billing.AccessSource != accessTx.AccessSourceDirect {
		t.Fatalf("AccessSource = %q, want %q", principal.Billing.AccessSource, accessTx.AccessSourceDirect)
	}
	if principal.Billing.Plan != "team" {
		t.Fatalf("Plan = %q, want team for direct access", principal.Billing.Plan)
	}

	status, _, err := service.Status(ctx, token)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.CanManagePlatformAccess {
		t.Fatalf("CanManagePlatformAccess = false, want true")
	}
}

func TestResolveSession_ReportsExpiredPromo(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAuthServiceTestDB(t)
	defer cleanup()

	owner, token := createOwnerAndSession(t, conn, "promo-owner@example.com")
	if _, err := accessTx.CreatePromoCode(ctx, conn, "SPRING14", 14, owner.ID); err != nil {
		t.Fatalf("CreatePromoCode: %v", err)
	}
	if _, err := accessTx.RedeemPromoCode(ctx, conn, owner.AccountID, owner.ID, "SPRING14", time.Now().AddDate(0, 0, -20), "test-secret", 180); err != nil {
		t.Fatalf("RedeemPromoCode: %v", err)
	}

	service := NewService(conn, Config{
		BillingConfigured:  true,
		BillingTrialDays:   7,
		PlatformAdminEmail: "vikecah@gmail.com",
	})

	principal, ok, err := service.ResolveSession(ctx, token)
	if err != nil {
		t.Fatalf("ResolveSession: %v", err)
	}
	if !ok {
		t.Fatal("ResolveSession ok = false, want true")
	}
	if principal.Billing.AccessGranted {
		t.Fatalf("AccessGranted = true, want false")
	}
	if !principal.Billing.PromoExpired {
		t.Fatalf("PromoExpired = false, want true")
	}
	if principal.Billing.AccessSource != accessTx.AccessSourcePromo {
		t.Fatalf("AccessSource = %q, want %q", principal.Billing.AccessSource, accessTx.AccessSourcePromo)
	}
	if principal.Billing.AccessExpiresAt == "" {
		t.Fatal("AccessExpiresAt = empty, want expiry date")
	}
	if principal.Billing.Plan != "" {
		t.Fatalf("Plan = %q, want empty after promo expiry", principal.Billing.Plan)
	}
}

func TestResolveSession_PlatformAdminBypassesBillingWithoutGrant(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAuthServiceTestDB(t)
	defer cleanup()

	_, token := createOwnerAndSession(t, conn, "vikecah@gmail.com")

	service := NewService(conn, Config{
		BillingConfigured:  true,
		BillingTrialDays:   7,
		PlatformAdminEmail: "vikecah@gmail.com",
	})

	principal, ok, err := service.ResolveSession(ctx, token)
	if err != nil {
		t.Fatalf("ResolveSession: %v", err)
	}
	if !ok {
		t.Fatal("ResolveSession ok = false, want true")
	}
	if !principal.Billing.AccessGranted {
		t.Fatal("AccessGranted = false, want true")
	}
	if principal.Billing.AccessSource != accessTx.AccessSourceDirect {
		t.Fatalf("AccessSource = %q, want %q", principal.Billing.AccessSource, accessTx.AccessSourceDirect)
	}
	if principal.Billing.Plan != "team" {
		t.Fatalf("Plan = %q, want team", principal.Billing.Plan)
	}
}
