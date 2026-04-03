package accessTx_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/transaction/accessTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/authTx"
)

func newAccessDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	conn, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "access.sqlite"))
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

func newOwnerAccount(t *testing.T, conn *sql.DB, email string) authTx.User {
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

	return owner
}

func TestResolveAccountAccess_DirectGrantWins(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAccessDB(t)
	defer cleanup()

	owner := newOwnerAccount(t, conn, "trusted@example.com")
	if _, err := accessTx.CreateDirectAccessGrant(ctx, conn, owner.Email, "team", "founder", owner.ID); err != nil {
		t.Fatalf("CreateDirectAccessGrant: %v", err)
	}

	access, err := accessTx.ResolveAccountAccess(ctx, conn, owner.AccountID, time.Now())
	if err != nil {
		t.Fatalf("ResolveAccountAccess: %v", err)
	}
	if !access.AccessGranted {
		t.Fatalf("AccessGranted = false, want true")
	}
	if access.Source != accessTx.AccessSourceDirect {
		t.Fatalf("Source = %q, want %q", access.Source, accessTx.AccessSourceDirect)
	}
	if access.Plan != "team" {
		t.Fatalf("Plan = %q, want team", access.Plan)
	}
}

func TestRedeemPromoCode_GrantsAccessAndExpires(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAccessDB(t)
	defer cleanup()

	owner := newOwnerAccount(t, conn, "promo@example.com")
	if _, err := accessTx.CreatePromoCode(ctx, conn, "EARLYBIRD14", 14, owner.ID); err != nil {
		t.Fatalf("CreatePromoCode: %v", err)
	}

	startedAt := time.Now().AddDate(0, 0, -20)
	if _, err := accessTx.RedeemPromoCode(ctx, conn, owner.AccountID, owner.ID, "EARLYBIRD14", startedAt, "test-secret", 180); err != nil {
		t.Fatalf("RedeemPromoCode: %v", err)
	}

	activeAccess, err := accessTx.ResolveAccountAccess(ctx, conn, owner.AccountID, startedAt.AddDate(0, 0, 1))
	if err != nil {
		t.Fatalf("ResolveAccountAccess active: %v", err)
	}
	if !activeAccess.AccessGranted || activeAccess.Source != accessTx.AccessSourcePromo {
		t.Fatalf("active access = %#v, want granted promo access", activeAccess)
	}
	if activeAccess.PromoCode != "EARLYBIRD14" {
		t.Fatalf("PromoCode = %q, want EARLYBIRD14", activeAccess.PromoCode)
	}

	expiredAccess, err := accessTx.ResolveAccountAccess(ctx, conn, owner.AccountID, time.Now())
	if err != nil {
		t.Fatalf("ResolveAccountAccess expired: %v", err)
	}
	if expiredAccess.AccessGranted {
		t.Fatalf("AccessGranted = true, want false after expiry")
	}
	if !expiredAccess.PromoExpired {
		t.Fatalf("PromoExpired = false, want true")
	}
	if expiredAccess.Source != accessTx.AccessSourcePromo {
		t.Fatalf("Source = %q, want %q", expiredAccess.Source, accessTx.AccessSourcePromo)
	}
}

func TestRedeemPromoCode_RejectsSecondUseForSameWorkspace(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAccessDB(t)
	defer cleanup()

	owner := newOwnerAccount(t, conn, "reuse@example.com")
	if _, err := accessTx.CreatePromoCode(ctx, conn, "ONCEONLY", 7, owner.ID); err != nil {
		t.Fatalf("CreatePromoCode: %v", err)
	}

	if _, err := accessTx.RedeemPromoCode(ctx, conn, owner.AccountID, owner.ID, "ONCEONLY", time.Now().AddDate(0, 0, -10), "test-secret", 180); err != nil {
		t.Fatalf("first RedeemPromoCode: %v", err)
	}

	_, err := accessTx.RedeemPromoCode(ctx, conn, owner.AccountID, owner.ID, "ONCEONLY", time.Now(), "test-secret", 180)
	if !errors.Is(err, accessTx.ErrPromoCodeAlreadyRedeemed) {
		t.Fatalf("RedeemPromoCode error = %v, want %v", err, accessTx.ErrPromoCodeAlreadyRedeemed)
	}
}

func TestRedeemPromoCode_RejectsReuseAfterAccountDeletion(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAccessDB(t)
	defer cleanup()

	owner := newOwnerAccount(t, conn, "same-person@example.com")
	if _, err := accessTx.CreatePromoCode(ctx, conn, "REUSEBLOCK", 7, owner.ID); err != nil {
		t.Fatalf("CreatePromoCode: %v", err)
	}

	if _, err := accessTx.RedeemPromoCode(ctx, conn, owner.AccountID, owner.ID, "REUSEBLOCK", time.Now().AddDate(0, 0, -10), "test-secret", 180); err != nil {
		t.Fatalf("RedeemPromoCode: %v", err)
	}
	if err := authTx.DeleteAccount(ctx, conn, owner.AccountID); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}

	newOwner := newOwnerAccount(t, conn, "same-person@example.com")
	_, err := accessTx.RedeemPromoCode(ctx, conn, newOwner.AccountID, newOwner.ID, "REUSEBLOCK", time.Now(), "test-secret", 180)
	if !errors.Is(err, accessTx.ErrPromoCodeAlreadyRedeemed) {
		t.Fatalf("RedeemPromoCode error = %v, want %v", err, accessTx.ErrPromoCodeAlreadyRedeemed)
	}
}

func TestRedeemPromoCode_AllowsReuseAfterSecretChangeAndAccountDeletion(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAccessDB(t)
	defer cleanup()

	owner := newOwnerAccount(t, conn, "rotate-secret@example.com")
	if _, err := accessTx.CreatePromoCode(ctx, conn, "SECRETROTATE", 7, owner.ID); err != nil {
		t.Fatalf("CreatePromoCode: %v", err)
	}

	if _, err := accessTx.RedeemPromoCode(
		ctx,
		conn,
		owner.AccountID,
		owner.ID,
		"SECRETROTATE",
		time.Now().AddDate(0, 0, -10),
		"old-secret",
		180,
	); err != nil {
		t.Fatalf("RedeemPromoCode first: %v", err)
	}

	if err := authTx.DeleteAccount(ctx, conn, owner.AccountID); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}

	newOwner := newOwnerAccount(t, conn, "rotate-secret@example.com")
	if _, err := accessTx.RedeemPromoCode(
		ctx,
		conn,
		newOwner.AccountID,
		newOwner.ID,
		"SECRETROTATE",
		time.Now(),
		"new-secret",
		180,
	); err != nil {
		t.Fatalf("RedeemPromoCode second after secret change: %v", err)
	}
}
