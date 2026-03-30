package authTx_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/transaction/authTx"
)

func newAuthDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "auth.sqlite")
	conn, err := sql.Open("sqlite3", dbPath)
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

func TestCreateInitialOwner_ClaimsDefaultAccountOnce(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAuthDB(t)
	defer cleanup()

	setupRequired, err := authTx.SetupRequired(ctx, conn)
	if err != nil {
		t.Fatalf("SetupRequired before owner: %v", err)
	}
	if !setupRequired {
		t.Fatal("SetupRequired before owner = false, want true")
	}

	user, err := authTx.CreateInitialOwner(ctx, conn, authTx.CreateGoogleUserParams{
		Name:      "Owner Example",
		Email:     "owner@example.com",
		GoogleSub: "google-sub-1",
		AvatarURL: "https://example.com/avatar.png",
		Role:      authTx.UserRoleOwner,
		AccountID: accountscope.DefaultAccountID,
	})
	if err != nil {
		t.Fatalf("CreateInitialOwner: %v", err)
	}
	if user.Role != authTx.UserRoleOwner {
		t.Fatalf("owner role = %q, want %q", user.Role, authTx.UserRoleOwner)
	}
	if user.AccountID != accountscope.DefaultAccountID {
		t.Fatalf("owner account_id = %d, want %d", user.AccountID, accountscope.DefaultAccountID)
	}
	if user.GoogleSub != "google-sub-1" {
		t.Fatalf("owner google_sub = %q, want google-sub-1", user.GoogleSub)
	}

	setupRequired, err = authTx.SetupRequired(ctx, conn)
	if err != nil {
		t.Fatalf("SetupRequired after owner: %v", err)
	}
	if setupRequired {
		t.Fatal("SetupRequired after owner = true, want false")
	}

	_, err = authTx.CreateInitialOwner(ctx, conn, authTx.CreateGoogleUserParams{
		Name:      "Second Owner",
		Email:     "second@example.com",
		GoogleSub: "google-sub-2",
		Role:      authTx.UserRoleOwner,
		AccountID: accountscope.DefaultAccountID,
	})
	if !errors.Is(err, authTx.ErrSetupAlreadyComplete) {
		t.Fatalf("second CreateInitialOwner error = %v, want %v", err, authTx.ErrSetupAlreadyComplete)
	}
}

func TestCreateSessionAndLookupByTokenHash(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAuthDB(t)
	defer cleanup()

	user, err := authTx.CreateInitialOwner(ctx, conn, authTx.CreateGoogleUserParams{
		Name:      "Owner Example",
		Email:     "owner@example.com",
		GoogleSub: "google-sub-1",
		Role:      authTx.UserRoleOwner,
		AccountID: accountscope.DefaultAccountID,
	})
	if err != nil {
		t.Fatalf("CreateInitialOwner: %v", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if err := authTx.CreateSession(ctx, conn, user.ID, user.AccountID, "session-hash", expiresAt); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	session, ok, err := authTx.GetSessionByTokenHash(ctx, conn, "session-hash", time.Now())
	if err != nil {
		t.Fatalf("GetSessionByTokenHash: %v", err)
	}
	if !ok {
		t.Fatal("GetSessionByTokenHash ok = false, want true")
	}
	if session.User.Email != "owner@example.com" {
		t.Fatalf("session user email = %q, want owner@example.com", session.User.Email)
	}
	if session.AccountID != accountscope.DefaultAccountID {
		t.Fatalf("session account_id = %d, want %d", session.AccountID, accountscope.DefaultAccountID)
	}
	if session.AccountName == "" {
		t.Fatal("session account name = empty, want owner workspace name")
	}
}

func TestInviteLifecycleAndMemberRemoval(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAuthDB(t)
	defer cleanup()

	owner, err := authTx.CreateInitialOwner(ctx, conn, authTx.CreateGoogleUserParams{
		Name:      "Owner Example",
		Email:     "owner@example.com",
		GoogleSub: "google-sub-owner",
		Role:      authTx.UserRoleOwner,
		AccountID: accountscope.DefaultAccountID,
	})
	if err != nil {
		t.Fatalf("CreateInitialOwner: %v", err)
	}

	invite, err := authTx.CreateInvite(ctx, conn, owner.AccountID, owner.ID, "teammate@example.com")
	if err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}
	if invite.Email != "teammate@example.com" {
		t.Fatalf("invite email = %q, want teammate@example.com", invite.Email)
	}

	invites, err := authTx.ListPendingInvites(ctx, conn, owner.AccountID)
	if err != nil {
		t.Fatalf("ListPendingInvites: %v", err)
	}
	if len(invites) != 1 || invites[0].Email != "teammate@example.com" {
		t.Fatalf("pending invites = %#v, want single teammate invite", invites)
	}

	member, err := authTx.CreateMemberFromGoogle(ctx, conn, authTx.CreateGoogleUserParams{
		Name:      "Teammate Example",
		Email:     "teammate@example.com",
		GoogleSub: "google-sub-member",
		Role:      authTx.UserRoleMember,
		AccountID: owner.AccountID,
	})
	if err != nil {
		t.Fatalf("CreateMemberFromGoogle: %v", err)
	}

	if err := authTx.DeleteInviteByEmail(ctx, conn, owner.AccountID, member.Email); err != nil {
		t.Fatalf("DeleteInviteByEmail: %v", err)
	}

	invites, err = authTx.ListPendingInvites(ctx, conn, owner.AccountID)
	if err != nil {
		t.Fatalf("ListPendingInvites after accept: %v", err)
	}
	if len(invites) != 0 {
		t.Fatalf("pending invites after accept = %#v, want empty", invites)
	}

	removed, err := authTx.RemoveMember(ctx, conn, owner.AccountID, owner.ID, member.ID)
	if err != nil {
		t.Fatalf("RemoveMember: %v", err)
	}
	if !removed {
		t.Fatal("RemoveMember removed = false, want true")
	}

	members, err := authTx.ListTeamMembers(ctx, conn, owner.AccountID)
	if err != nil {
		t.Fatalf("ListTeamMembers: %v", err)
	}
	if len(members) != 1 || members[0].Email != "owner@example.com" {
		t.Fatalf("team members after removal = %#v, want owner only", members)
	}
}

func TestRemoveMember_RejectsSelfAndOwnerRemoval(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAuthDB(t)
	defer cleanup()

	owner, err := authTx.CreateInitialOwner(ctx, conn, authTx.CreateGoogleUserParams{
		Name:      "Owner Example",
		Email:     "owner@example.com",
		GoogleSub: "google-sub-owner",
		Role:      authTx.UserRoleOwner,
		AccountID: accountscope.DefaultAccountID,
	})
	if err != nil {
		t.Fatalf("CreateInitialOwner: %v", err)
	}

	removed, err := authTx.RemoveMember(ctx, conn, owner.AccountID, owner.ID, owner.ID)
	if !errors.Is(err, authTx.ErrCannotRemoveSelf) {
		t.Fatalf("RemoveMember self error = %v, want %v", err, authTx.ErrCannotRemoveSelf)
	}
	if removed {
		t.Fatal("RemoveMember self removed = true, want false")
	}
}
