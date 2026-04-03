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
	"github.com/viktorHadz/goInvoice26/internal/billingplan"
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

func TestCreateAccountOwner_CreatesIndependentAccounts(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAuthDB(t)
	defer cleanup()

	first, err := authTx.CreateAccountOwner(ctx, conn, authTx.CreateGoogleUserParams{
		Name:      "First Workspace",
		Email:     "first@example.com",
		GoogleSub: "google-sub-first",
		Role:      authTx.UserRoleOwner,
	})
	if err != nil {
		t.Fatalf("CreateAccountOwner first: %v", err)
	}

	second, err := authTx.CreateAccountOwner(ctx, conn, authTx.CreateGoogleUserParams{
		Name:      "Second Workspace",
		Email:     "second@example.com",
		GoogleSub: "google-sub-second",
		Role:      authTx.UserRoleOwner,
	})
	if err != nil {
		t.Fatalf("CreateAccountOwner second: %v", err)
	}

	if first.AccountID <= 0 || second.AccountID <= 0 {
		t.Fatalf("account ids should be populated, got first=%d second=%d", first.AccountID, second.AccountID)
	}
	if first.AccountID == second.AccountID {
		t.Fatalf("expected distinct account ids, both were %d", first.AccountID)
	}

	var seqCount int
	if err := conn.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM invoice_number_seq
		WHERE account_id IN (?, ?)
	`, first.AccountID, second.AccountID).Scan(&seqCount); err != nil {
		t.Fatalf("count invoice sequences: %v", err)
	}
	if seqCount != 2 {
		t.Fatalf("invoice sequence rows = %d, want 2", seqCount)
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

	invite, err := authTx.CreateInvite(
		ctx,
		conn,
		owner.AccountID,
		owner.ID,
		"teammate@example.com",
		billingplan.TeamSeatLimit,
	)
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

func TestDeleteAccount_RemovesUsersInvitesAndScopedData(t *testing.T) {
	ctx := context.Background()
	conn, cleanup := newAuthDB(t)
	defer cleanup()

	owner, err := authTx.CreateAccountOwner(ctx, conn, authTx.CreateGoogleUserParams{
		Name:      "Owner Example",
		Email:     "owner@example.com",
		GoogleSub: "google-sub-owner",
		Role:      authTx.UserRoleOwner,
	})
	if err != nil {
		t.Fatalf("CreateAccountOwner: %v", err)
	}

	member, err := authTx.CreateMemberFromGoogle(ctx, conn, authTx.CreateGoogleUserParams{
		Name:      "Teammate Example",
		Email:     "member@example.com",
		GoogleSub: "google-sub-member",
		Role:      authTx.UserRoleMember,
		AccountID: owner.AccountID,
	})
	if err != nil {
		t.Fatalf("CreateMemberFromGoogle: %v", err)
	}

	if _, err := authTx.CreateInvite(
		ctx,
		conn,
		owner.AccountID,
		owner.ID,
		"invitee@example.com",
		billingplan.TeamSeatLimit,
	); err != nil {
		t.Fatalf("CreateInvite: %v", err)
	}

	if err := authTx.CreateSession(ctx, conn, owner.ID, owner.AccountID, "owner-session", time.Now().Add(24*time.Hour)); err != nil {
		t.Fatalf("CreateSession owner: %v", err)
	}
	if err := authTx.CreateSession(ctx, conn, member.ID, member.AccountID, "member-session", time.Now().Add(24*time.Hour)); err != nil {
		t.Fatalf("CreateSession member: %v", err)
	}

	if _, err := conn.ExecContext(ctx, `
		INSERT INTO clients (account_id, name)
		VALUES (?, 'Client');
	`, owner.AccountID); err != nil {
		t.Fatalf("insert client: %v", err)
	}

	if err := authTx.DeleteAccount(ctx, conn, owner.AccountID); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}

	assertCount := func(query string, args ...any) int {
		t.Helper()
		var count int
		if err := conn.QueryRowContext(ctx, query, args...).Scan(&count); err != nil {
			t.Fatalf("count query failed: %v", err)
		}
		return count
	}

	if got := assertCount(`SELECT COUNT(*) FROM accounts WHERE id = ?`, owner.AccountID); got != 0 {
		t.Fatalf("accounts count = %d, want 0", got)
	}
	if got := assertCount(`SELECT COUNT(*) FROM users WHERE account_id = ?`, owner.AccountID); got != 0 {
		t.Fatalf("users count = %d, want 0", got)
	}
	if got := assertCount(`SELECT COUNT(*) FROM allowed_users WHERE account_id = ?`, owner.AccountID); got != 0 {
		t.Fatalf("allowed_users count = %d, want 0", got)
	}
	if got := assertCount(`SELECT COUNT(*) FROM auth_sessions WHERE account_id = ?`, owner.AccountID); got != 0 {
		t.Fatalf("auth_sessions count = %d, want 0", got)
	}
	if got := assertCount(`SELECT COUNT(*) FROM clients WHERE account_id = ?`, owner.AccountID); got != 0 {
		t.Fatalf("clients count = %d, want 0", got)
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

func TestCreateInvite_RejectsWhenTeamSeatLimitReached(t *testing.T) {
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

	for i := 0; i < billingplan.TeamSeatLimit-1; i++ {
		email := "member" + string(rune('a'+i)) + "@example.com"
		if _, err := authTx.CreateMemberFromGoogle(ctx, conn, authTx.CreateGoogleUserParams{
			Name:      "Member Example",
			Email:     email,
			GoogleSub: "google-sub-" + string(rune('a'+i)),
			Role:      authTx.UserRoleMember,
			AccountID: owner.AccountID,
		}); err != nil {
			t.Fatalf("CreateMemberFromGoogle %d: %v", i, err)
		}
	}

	_, err = authTx.CreateInvite(
		ctx,
		conn,
		owner.AccountID,
		owner.ID,
		"overflow@example.com",
		billingplan.TeamSeatLimit,
	)
	if !errors.Is(err, authTx.ErrTeamSeatLimitReached) {
		t.Fatalf("CreateInvite error = %v, want %v", err, authTx.ErrTeamSeatLimitReached)
	}
}
