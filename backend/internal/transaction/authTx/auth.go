package authTx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/viktorHadz/goInvoice26/internal/models"
)

const (
	UserRoleOwner  = "owner"
	UserRoleMember = "member"

	timestampLayout = "2006-01-02T15:04:05.000000000Z07:00"
)

var (
	ErrSetupAlreadyComplete = errors.New("owner setup already complete")
	ErrSetupRequired        = errors.New("owner setup required")
	ErrInvalidEmail         = errors.New("invalid email")
	ErrInviteAlreadyExists  = errors.New("invite already exists")
	ErrMemberAlreadyExists  = errors.New("member already exists")
	ErrCannotRemoveSelf     = errors.New("cannot remove self")
	ErrCannotRemoveOwner    = errors.New("cannot remove owner")
)

type User struct {
	ID        int64
	Name      string
	Email     string
	GoogleSub string
	AvatarURL string
	Role      string
	AccountID int64
}

type Session struct {
	ID                       int64
	UserID                   int64
	AccountID                int64
	AccountName              string
	BillingStatus            string
	BillingCurrentPeriodEnd  string
	BillingCancelAtPeriodEnd bool
	StripeCustomerID         string
	ExpiresAt                time.Time
	User                     User
}

type CreateGoogleUserParams struct {
	Name      string
	Email     string
	GoogleSub string
	AvatarURL string
	Role      string
	AccountID int64
}

func SetupRequired(ctx context.Context, db *sql.DB) (bool, error) {
	var count int64
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users;`).Scan(&count); err != nil {
		return false, fmt.Errorf("count users: %w", err)
	}

	return count == 0, nil
}

func GetUserByGoogleSub(ctx context.Context, db *sql.DB, googleSub string) (User, bool, error) {
	var user User
	err := db.QueryRowContext(ctx, `
		SELECT id, COALESCE(name, ''), email, COALESCE(google_sub, ''), COALESCE(avatar_url, ''), COALESCE(role, ?), account_id
		FROM users
		WHERE google_sub = ?
		LIMIT 1;
	`, UserRoleMember, googleSub).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.GoogleSub,
		&user.AvatarURL,
		&user.Role,
		&user.AccountID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, fmt.Errorf("get user by google_sub: %w", err)
	}

	return user, true, nil
}

func GetUserByEmail(ctx context.Context, db *sql.DB, email string) (User, bool, error) {
	var user User
	err := db.QueryRowContext(ctx, `
		SELECT id, COALESCE(name, ''), email, COALESCE(google_sub, ''), COALESCE(avatar_url, ''), COALESCE(role, ?), account_id
		FROM users
		WHERE LOWER(email) = LOWER(?)
		LIMIT 1;
	`, UserRoleMember, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.GoogleSub,
		&user.AvatarURL,
		&user.Role,
		&user.AccountID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, false, nil
	}
	if err != nil {
		return User{}, false, fmt.Errorf("get user by email: %w", err)
	}

	return user, true, nil
}

func CreateInitialOwner(ctx context.Context, db *sql.DB, params CreateGoogleUserParams) (User, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, fmt.Errorf("begin create initial owner tx: %w", err)
	}
	defer tx.Rollback()

	var count int64
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users;`).Scan(&count); err != nil {
		return User{}, fmt.Errorf("count users in setup tx: %w", err)
	}
	if count > 0 {
		return User{}, ErrSetupAlreadyComplete
	}

	if params.AccountID <= 0 {
		params.AccountID = 1
	}
	if params.Role == "" {
		params.Role = UserRoleOwner
	}

	accountName := strings.TrimSpace(params.Name)
	if accountName == "" {
		accountName = strings.TrimSpace(strings.Split(params.Email, "@")[0])
	}
	if accountName == "" {
		accountName = "Workspace"
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT OR IGNORE INTO accounts (id, name)
		VALUES (?, ?);
	`, params.AccountID, accountName); err != nil {
		return User{}, fmt.Errorf("ensure default account: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE accounts
		SET name = ?
		WHERE id = ?
		  AND (TRIM(name) = '' OR name = 'Default account');
	`, accountName, params.AccountID); err != nil {
		return User{}, fmt.Errorf("update account name: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO users (
			name,
			email,
			password_hash,
			account_id,
			google_sub,
			avatar_url,
			role
		) VALUES (?, ?, '', ?, ?, ?, ?);
	`,
		strings.TrimSpace(params.Name),
		strings.TrimSpace(params.Email),
		params.AccountID,
		strings.TrimSpace(params.GoogleSub),
		strings.TrimSpace(params.AvatarURL),
		params.Role,
	)
	if err != nil {
		return User{}, fmt.Errorf("insert initial owner: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("get inserted owner id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return User{}, fmt.Errorf("commit initial owner tx: %w", err)
	}

	return GetUserByID(ctx, db, userID)
}

func CreateMemberFromGoogle(ctx context.Context, db *sql.DB, params CreateGoogleUserParams) (User, error) {
	if params.AccountID <= 0 {
		params.AccountID = 1
	}
	if params.Role == "" {
		params.Role = UserRoleMember
	}

	result, err := db.ExecContext(ctx, `
		INSERT INTO users (
			name,
			email,
			password_hash,
			account_id,
			google_sub,
			avatar_url,
			role
		) VALUES (?, ?, '', ?, ?, ?, ?);
	`,
		strings.TrimSpace(params.Name),
		strings.TrimSpace(params.Email),
		params.AccountID,
		strings.TrimSpace(params.GoogleSub),
		strings.TrimSpace(params.AvatarURL),
		params.Role,
	)
	if err != nil {
		return User{}, fmt.Errorf("insert member user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return User{}, fmt.Errorf("get inserted member id: %w", err)
	}

	return GetUserByID(ctx, db, userID)
}

func GetUserByID(ctx context.Context, db *sql.DB, userID int64) (User, error) {
	var user User
	if err := db.QueryRowContext(ctx, `
		SELECT id, COALESCE(name, ''), email, COALESCE(google_sub, ''), COALESCE(avatar_url, ''), COALESCE(role, ?), account_id
		FROM users
		WHERE id = ?
		LIMIT 1;
	`, UserRoleMember, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.GoogleSub,
		&user.AvatarURL,
		&user.Role,
		&user.AccountID,
	); err != nil {
		return User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func UpdateGoogleIdentity(ctx context.Context, db *sql.DB, userID int64, googleSub, name, avatarURL string) error {
	if _, err := db.ExecContext(ctx, `
		UPDATE users
		SET google_sub = ?,
			name = ?,
			avatar_url = ?
		WHERE id = ?;
	`,
		strings.TrimSpace(googleSub),
		strings.TrimSpace(name),
		strings.TrimSpace(avatarURL),
		userID,
	); err != nil {
		return fmt.Errorf("update google identity: %w", err)
	}

	return nil
}

func UpdateUserProfile(ctx context.Context, db *sql.DB, userID int64, name, avatarURL string) error {
	if _, err := db.ExecContext(ctx, `
		UPDATE users
		SET name = ?,
			avatar_url = ?
		WHERE id = ?;
	`,
		strings.TrimSpace(name),
		strings.TrimSpace(avatarURL),
		userID,
	); err != nil {
		return fmt.Errorf("update user profile: %w", err)
	}

	return nil
}

func AllowedAccountIDForEmail(ctx context.Context, db *sql.DB, email string) (int64, bool, error) {
	var accountID int64
	err := db.QueryRowContext(ctx, `
		SELECT account_id
		FROM allowed_users
		WHERE LOWER(email) = LOWER(?)
		LIMIT 1;
	`, email).Scan(&accountID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("lookup allowed user: %w", err)
	}

	return accountID, true, nil
}

func CreateSession(ctx context.Context, db *sql.DB, userID, accountID int64, tokenHash string, expiresAt time.Time) error {
	if err := CleanupExpiredSessions(ctx, db, time.Now()); err != nil {
		return err
	}

	ts := formatTimestamp(expiresAt)
	now := formatTimestamp(time.Now())

	if _, err := db.ExecContext(ctx, `
		INSERT INTO auth_sessions (
			user_id,
			account_id,
			token_hash,
			expires_at,
			last_seen_at
		) VALUES (?, ?, ?, ?, ?);
	`, userID, accountID, tokenHash, ts, now); err != nil {
		return fmt.Errorf("create auth session: %w", err)
	}

	return nil
}

func GetSessionByTokenHash(ctx context.Context, db *sql.DB, tokenHash string, now time.Time) (Session, bool, error) {
	var (
		session           Session
		expiresAtText     string
		cancelAtPeriodEnd int64
	)

	err := db.QueryRowContext(ctx, `
		SELECT
			s.id,
			s.user_id,
			s.account_id,
			s.expires_at,
			COALESCE(a.name, ''),
			COALESCE(a.billing_status, 'inactive'),
			COALESCE(a.billing_current_period_end, ''),
			COALESCE(a.billing_cancel_at_period_end, 0),
			COALESCE(a.stripe_customer_id, ''),
			u.id,
			COALESCE(u.name, ''),
			u.email,
			COALESCE(u.google_sub, ''),
			COALESCE(u.avatar_url, ''),
			COALESCE(u.role, ?)
		FROM auth_sessions s
		INNER JOIN users u ON u.id = s.user_id
		INNER JOIN accounts a ON a.id = s.account_id
		WHERE s.token_hash = ?
		  AND s.expires_at > ?
		LIMIT 1;
	`, UserRoleMember, tokenHash, formatTimestamp(now)).Scan(
		&session.ID,
		&session.UserID,
		&session.AccountID,
		&expiresAtText,
		&session.AccountName,
		&session.BillingStatus,
		&session.BillingCurrentPeriodEnd,
		&cancelAtPeriodEnd,
		&session.StripeCustomerID,
		&session.User.ID,
		&session.User.Name,
		&session.User.Email,
		&session.User.GoogleSub,
		&session.User.AvatarURL,
		&session.User.Role,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, fmt.Errorf("get session by token hash: %w", err)
	}

	expiresAt, err := time.Parse(timestampLayout, expiresAtText)
	if err != nil {
		return Session{}, false, fmt.Errorf("parse session expiry: %w", err)
	}
	session.ExpiresAt = expiresAt
	session.BillingCancelAtPeriodEnd = cancelAtPeriodEnd > 0

	return session, true, nil
}

func TouchSession(ctx context.Context, db *sql.DB, sessionID int64, seenAt time.Time) error {
	if _, err := db.ExecContext(ctx, `
		UPDATE auth_sessions
		SET last_seen_at = ?
		WHERE id = ?;
	`, formatTimestamp(seenAt), sessionID); err != nil {
		return fmt.Errorf("touch auth session: %w", err)
	}

	return nil
}

func DeleteSessionByTokenHash(ctx context.Context, db *sql.DB, tokenHash string) error {
	if tokenHash == "" {
		return nil
	}

	if _, err := db.ExecContext(ctx, `
		DELETE FROM auth_sessions
		WHERE token_hash = ?;
	`, tokenHash); err != nil {
		return fmt.Errorf("delete auth session: %w", err)
	}

	return nil
}

func CleanupExpiredSessions(ctx context.Context, db *sql.DB, now time.Time) error {
	if _, err := db.ExecContext(ctx, `
		DELETE FROM auth_sessions
		WHERE expires_at <= ?;
	`, formatTimestamp(now)); err != nil {
		return fmt.Errorf("cleanup expired sessions: %w", err)
	}

	return nil
}

func formatTimestamp(ts time.Time) string {
	return ts.UTC().Format(timestampLayout)
}

func ListTeamMembers(ctx context.Context, db *sql.DB, accountID int64) ([]models.TeamMember, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			id,
			COALESCE(name, ''),
			email,
			COALESCE(avatar_url, ''),
			COALESCE(role, ?),
			created_at
		FROM users
		WHERE account_id = ?
		ORDER BY
			CASE COALESCE(role, ?) WHEN ? THEN 0 ELSE 1 END,
			LOWER(COALESCE(name, email)),
			id;
	`, UserRoleMember, accountID, UserRoleMember, UserRoleOwner)
	if err != nil {
		return nil, fmt.Errorf("list team members: %w", err)
	}
	defer rows.Close()

	var members []models.TeamMember
	for rows.Next() {
		var member models.TeamMember
		if err := rows.Scan(
			&member.ID,
			&member.Name,
			&member.Email,
			&member.AvatarURL,
			&member.Role,
			&member.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan team member: %w", err)
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate team members: %w", err)
	}

	return members, nil
}

func ListPendingInvites(ctx context.Context, db *sql.DB, accountID int64) ([]models.TeamInvite, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			au.id,
			au.email,
			au.created_at
		FROM allowed_users au
		LEFT JOIN users u
			ON u.account_id = au.account_id
		   AND LOWER(u.email) = LOWER(au.email)
		WHERE au.account_id = ?
		  AND u.id IS NULL
		ORDER BY au.created_at DESC, au.id DESC;
	`, accountID)
	if err != nil {
		return nil, fmt.Errorf("list pending invites: %w", err)
	}
	defer rows.Close()

	var invites []models.TeamInvite
	for rows.Next() {
		var invite models.TeamInvite
		if err := rows.Scan(&invite.ID, &invite.Email, &invite.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan pending invite: %w", err)
		}
		invites = append(invites, invite)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending invites: %w", err)
	}

	return invites, nil
}

func CreateInvite(ctx context.Context, db *sql.DB, accountID, invitedByUserID int64, email string) (models.TeamInvite, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return models.TeamInvite{}, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return models.TeamInvite{}, fmt.Errorf("begin create invite tx: %w", err)
	}
	defer tx.Rollback()

	var existingMemberID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM users
		WHERE account_id = ?
		  AND LOWER(email) = LOWER(?)
		LIMIT 1;
	`, accountID, normalizedEmail).Scan(&existingMemberID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return models.TeamInvite{}, fmt.Errorf("check existing member before invite: %w", err)
	default:
		return models.TeamInvite{}, ErrMemberAlreadyExists
	}

	var existingInviteID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM allowed_users
		WHERE account_id = ?
		  AND LOWER(email) = LOWER(?)
		LIMIT 1;
	`, accountID, normalizedEmail).Scan(&existingInviteID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return models.TeamInvite{}, fmt.Errorf("check existing invite: %w", err)
	default:
		return models.TeamInvite{}, ErrInviteAlreadyExists
	}

	createdAt := formatTimestamp(time.Now())
	result, err := tx.ExecContext(ctx, `
		INSERT INTO allowed_users (
			account_id,
			email,
			invited_by_user_id,
			created_at
		) VALUES (?, ?, ?, ?);
	`, accountID, normalizedEmail, nullInt64(invitedByUserID), createdAt)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return models.TeamInvite{}, ErrInviteAlreadyExists
		}
		return models.TeamInvite{}, fmt.Errorf("insert invite: %w", err)
	}

	inviteID, err := result.LastInsertId()
	if err != nil {
		return models.TeamInvite{}, fmt.Errorf("get invite id: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return models.TeamInvite{}, fmt.Errorf("commit create invite tx: %w", err)
	}

	return models.TeamInvite{
		ID:        inviteID,
		Email:     normalizedEmail,
		CreatedAt: createdAt,
	}, nil
}

func DeleteInvite(ctx context.Context, db *sql.DB, accountID, inviteID int64) (bool, error) {
	result, err := db.ExecContext(ctx, `
		DELETE FROM allowed_users
		WHERE id = ?
		  AND account_id = ?;
	`, inviteID, accountID)
	if err != nil {
		return false, fmt.Errorf("delete invite: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete invite rows affected: %w", err)
	}

	return affected > 0, nil
}

func DeleteInviteByEmail(ctx context.Context, db *sql.DB, accountID int64, email string) error {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `
		DELETE FROM allowed_users
		WHERE account_id = ?
		  AND LOWER(email) = LOWER(?);
	`, accountID, normalizedEmail); err != nil {
		return fmt.Errorf("delete invite by email: %w", err)
	}

	return nil
}

func RemoveMember(ctx context.Context, db *sql.DB, accountID, actingUserID, memberUserID int64) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin remove member tx: %w", err)
	}
	defer tx.Rollback()

	var (
		memberEmail string
		memberRole  string
	)
	err = tx.QueryRowContext(ctx, `
		SELECT email, COALESCE(role, ?)
		FROM users
		WHERE id = ?
		  AND account_id = ?
		LIMIT 1;
	`, UserRoleMember, memberUserID, accountID).Scan(&memberEmail, &memberRole)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("load member before remove: %w", err)
	}

	if memberUserID == actingUserID {
		return false, ErrCannotRemoveSelf
	}
	if memberRole == UserRoleOwner {
		return false, ErrCannotRemoveOwner
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM auth_sessions
		WHERE user_id = ?;
	`, memberUserID); err != nil {
		return false, fmt.Errorf("delete member sessions: %w", err)
	}

	result, err := tx.ExecContext(ctx, `
		DELETE FROM users
		WHERE id = ?
		  AND account_id = ?;
	`, memberUserID, accountID)
	if err != nil {
		return false, fmt.Errorf("delete member: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("member delete rows affected: %w", err)
	}
	if affected == 0 {
		return false, nil
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM allowed_users
		WHERE account_id = ?
		  AND LOWER(email) = LOWER(?);
	`, accountID, memberEmail); err != nil {
		return false, fmt.Errorf("cleanup member invite: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit remove member tx: %w", err)
	}

	return true, nil
}

func normalizeEmail(email string) (string, error) {
	address, err := mail.ParseAddress(strings.TrimSpace(email))
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidEmail, err)
	}

	normalized := strings.ToLower(strings.TrimSpace(address.Address))
	if normalized == "" {
		return "", ErrInvalidEmail
	}

	return normalized, nil
}

func nullInt64(v int64) sql.NullInt64 {
	if v <= 0 {
		return sql.NullInt64{}
	}

	return sql.NullInt64{Int64: v, Valid: true}
}
