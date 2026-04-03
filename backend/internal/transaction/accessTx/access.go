package accessTx

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/viktorHadz/goInvoice26/internal/billingplan"
	"github.com/viktorHadz/goInvoice26/internal/billingstate"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

const (
	AccessSourceDirect = "direct"
	AccessSourcePromo  = "promo"

	timestampLayout = "2006-01-02T15:04:05.000000000Z07:00"
)

var (
	ErrInvalidEmail             = errors.New("invalid email")
	ErrDirectAccessGrantExists  = errors.New("direct access grant already exists")
	ErrInvalidAccessPlan        = errors.New("invalid access plan")
	ErrInvalidPromoCode         = errors.New("invalid promo code")
	ErrPromoCodeExists          = errors.New("promo code already exists")
	ErrPromoCodeNotFound        = errors.New("promo code not found")
	ErrPromoCodeInactive        = errors.New("promo code inactive")
	ErrPromoCodeAlreadyRedeemed = errors.New("promo code already redeemed")
	ErrAccessAlreadyGranted     = errors.New("access already granted")
	promoCodePattern            = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]{2,63}$`)
)

func ListDirectAccessGrants(ctx context.Context, db *sql.DB) ([]models.DirectAccessGrant, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, email, COALESCE(plan, 'single'), COALESCE(note, ''), created_at
		FROM direct_access_grants
		ORDER BY created_at DESC, id DESC;
	`)
	if err != nil {
		return nil, fmt.Errorf("list direct access grants: %w", err)
	}
	defer rows.Close()

	var grants []models.DirectAccessGrant
	for rows.Next() {
		var grant models.DirectAccessGrant
		if err := rows.Scan(&grant.ID, &grant.Email, &grant.Plan, &grant.Note, &grant.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan direct access grant: %w", err)
		}
		grants = append(grants, grant)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate direct access grants: %w", err)
	}

	return grants, nil
}

func CreateDirectAccessGrant(
	ctx context.Context,
	db *sql.DB,
	email string,
	plan string,
	note string,
	createdByUserID int64,
) (models.DirectAccessGrant, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return models.DirectAccessGrant{}, err
	}
	normalizedPlan := billingplan.Normalize(plan)
	if strings.TrimSpace(plan) != "" && normalizedPlan == "" {
		return models.DirectAccessGrant{}, ErrInvalidAccessPlan
	}
	if normalizedPlan == "" {
		normalizedPlan = billingplan.PlanSingle
	}

	createdAt := formatTimestamp(time.Now())
	result, err := db.ExecContext(ctx, `
		INSERT INTO direct_access_grants (
			email,
			plan,
			note,
			created_by_user_id,
			created_at
		) VALUES (?, ?, ?, ?, ?);
	`, normalizedEmail, normalizedPlan, strings.TrimSpace(note), nullInt64(createdByUserID), createdAt)
	if err != nil {
		if isUniqueConstraintError(err) {
			return models.DirectAccessGrant{}, ErrDirectAccessGrantExists
		}
		return models.DirectAccessGrant{}, fmt.Errorf("insert direct access grant: %w", err)
	}

	grantID, err := result.LastInsertId()
	if err != nil {
		return models.DirectAccessGrant{}, fmt.Errorf("direct access grant id: %w", err)
	}

	return models.DirectAccessGrant{
		ID:        grantID,
		Email:     normalizedEmail,
		Plan:      normalizedPlan,
		Note:      strings.TrimSpace(note),
		CreatedAt: createdAt,
	}, nil
}

func DeleteDirectAccessGrant(ctx context.Context, db *sql.DB, grantID int64) (bool, error) {
	result, err := db.ExecContext(ctx, `
		DELETE FROM direct_access_grants
		WHERE id = ?;
	`, grantID)
	if err != nil {
		return false, fmt.Errorf("delete direct access grant: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete direct access grant rows affected: %w", err)
	}

	return affected > 0, nil
}

func HasDirectAccessGrantForEmail(ctx context.Context, db *sql.DB, email string) (bool, error) {
	normalizedEmail, err := normalizeEmail(email)
	if err != nil {
		return false, err
	}

	return hasDirectAccessGrantForEmail(ctx, db, normalizedEmail)
}

func ListPromoCodes(ctx context.Context, db *sql.DB) ([]models.PromoCode, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			pc.id,
			pc.code,
			pc.duration_days,
			pc.active,
			pc.created_at,
			COUNT(pcrc.id) AS redemption_count
		FROM promo_codes pc
		LEFT JOIN promo_code_redemption_claims pcrc
			ON pcrc.promo_code_id = pc.id
		GROUP BY pc.id, pc.code, pc.duration_days, pc.active, pc.created_at
		ORDER BY pc.created_at DESC, pc.id DESC;
	`)
	if err != nil {
		return nil, fmt.Errorf("list promo codes: %w", err)
	}
	defer rows.Close()

	var promoCodes []models.PromoCode
	for rows.Next() {
		var promoCode models.PromoCode
		var active int
		if err := rows.Scan(
			&promoCode.ID,
			&promoCode.Code,
			&promoCode.DurationDays,
			&active,
			&promoCode.CreatedAt,
			&promoCode.RedemptionCount,
		); err != nil {
			return nil, fmt.Errorf("scan promo code: %w", err)
		}
		promoCode.Active = active != 0
		promoCodes = append(promoCodes, promoCode)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate promo codes: %w", err)
	}

	return promoCodes, nil
}

func CreatePromoCode(
	ctx context.Context,
	db *sql.DB,
	code string,
	durationDays int,
	createdByUserID int64,
) (models.PromoCode, error) {
	normalizedCode, err := normalizePromoCode(code)
	if err != nil {
		return models.PromoCode{}, err
	}
	if durationDays <= 0 {
		return models.PromoCode{}, ErrInvalidPromoCode
	}

	createdAt := formatTimestamp(time.Now())
	result, err := db.ExecContext(ctx, `
		INSERT INTO promo_codes (
			code,
			duration_days,
			active,
			created_by_user_id,
			created_at
		) VALUES (?, ?, 1, ?, ?);
	`, normalizedCode, durationDays, nullInt64(createdByUserID), createdAt)
	if err != nil {
		if isUniqueConstraintError(err) {
			return models.PromoCode{}, ErrPromoCodeExists
		}
		return models.PromoCode{}, fmt.Errorf("insert promo code: %w", err)
	}

	promoCodeID, err := result.LastInsertId()
	if err != nil {
		return models.PromoCode{}, fmt.Errorf("promo code id: %w", err)
	}

	return models.PromoCode{
		ID:              promoCodeID,
		Code:            normalizedCode,
		DurationDays:    durationDays,
		Active:          true,
		RedemptionCount: 0,
		CreatedAt:       createdAt,
	}, nil
}

func SetPromoCodeActive(ctx context.Context, db *sql.DB, promoCodeID int64, active bool) (bool, error) {
	result, err := db.ExecContext(ctx, `
		UPDATE promo_codes
		SET active = ?
		WHERE id = ?;
	`, boolInt(active), promoCodeID)
	if err != nil {
		return false, fmt.Errorf("update promo code active: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("update promo code rows affected: %w", err)
	}

	return affected > 0, nil
}

func RedeemPromoCode(
	ctx context.Context,
	db *sql.DB,
	accountID int64,
	redeemedByUserID int64,
	code string,
	now time.Time,
	ledgerSecret string,
	retentionDays int,
) (models.PromoRedemptionResult, error) {
	normalizedCode, err := normalizePromoCode(code)
	if err != nil {
		return models.PromoRedemptionResult{}, err
	}
	if retentionDays < 0 {
		retentionDays = 0
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return models.PromoRedemptionResult{}, fmt.Errorf("begin redeem promo code tx: %w", err)
	}
	defer tx.Rollback()

	if err := cleanupExpiredPromoRedemptionClaimsTx(ctx, tx, now); err != nil {
		return models.PromoRedemptionResult{}, err
	}

	if grantsAccess, err := accountCurrentlyHasAccessTx(ctx, tx, accountID, now); err != nil {
		return models.PromoRedemptionResult{}, err
	} else if grantsAccess {
		return models.PromoRedemptionResult{}, ErrAccessAlreadyGranted
	}

	ownerEmail, ok, err := ownerEmailForAccountTx(ctx, tx, accountID)
	if err != nil {
		return models.PromoRedemptionResult{}, err
	}
	if !ok {
		return models.PromoRedemptionResult{}, fmt.Errorf("load promo owner email: account owner not found")
	}
	ownerEmailHMAC := promoOwnerEmailHMAC(ownerEmail, ledgerSecret)

	var (
		promoCodeID  int64
		durationDays int
		active       int
	)
	err = tx.QueryRowContext(ctx, `
		SELECT id, duration_days, active
		FROM promo_codes
		WHERE code = ?
		LIMIT 1;
	`, normalizedCode).Scan(&promoCodeID, &durationDays, &active)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.PromoRedemptionResult{}, ErrPromoCodeNotFound
	case err != nil:
		return models.PromoRedemptionResult{}, fmt.Errorf("load promo code: %w", err)
	}
	if active == 0 {
		return models.PromoRedemptionResult{}, ErrPromoCodeInactive
	}

	var existingID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM promo_code_redemptions
		WHERE promo_code_id = ?
		  AND account_id = ?
		LIMIT 1;
	`, promoCodeID, accountID).Scan(&existingID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return models.PromoRedemptionResult{}, fmt.Errorf("check promo redemption: %w", err)
	default:
		return models.PromoRedemptionResult{}, ErrPromoCodeAlreadyRedeemed
	}

	var existingClaimID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM promo_code_redemption_claims
		WHERE promo_code_id = ?
		  AND owner_email_hmac = ?
		LIMIT 1;
	`, promoCodeID, ownerEmailHMAC).Scan(&existingClaimID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
	case err != nil:
		return models.PromoRedemptionResult{}, fmt.Errorf("check promo redemption claim: %w", err)
	default:
		return models.PromoRedemptionResult{}, ErrPromoCodeAlreadyRedeemed
	}

	redeemedAt := formatTimestamp(now)
	expiresAt := formatTimestamp(now.AddDate(0, 0, durationDays))
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO promo_code_redemptions (
			promo_code_id,
			account_id,
			redeemed_by_user_id,
			redeemed_at,
			expires_at
		) VALUES (?, ?, ?, ?, ?);
	`, promoCodeID, accountID, nullInt64(redeemedByUserID), redeemedAt, expiresAt); err != nil {
		if isUniqueConstraintError(err) {
			return models.PromoRedemptionResult{}, ErrPromoCodeAlreadyRedeemed
		}
		return models.PromoRedemptionResult{}, fmt.Errorf("insert promo redemption: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO promo_code_redemption_claims (
			promo_code_id,
			owner_email_hmac,
			redeemed_at,
			retention_until
		) VALUES (?, ?, ?, ?);
	`, promoCodeID, ownerEmailHMAC, redeemedAt, formatTimestamp(now.AddDate(0, 0, durationDays+retentionDays))); err != nil {
		if isUniqueConstraintError(err) {
			return models.PromoRedemptionResult{}, ErrPromoCodeAlreadyRedeemed
		}
		return models.PromoRedemptionResult{}, fmt.Errorf("insert promo redemption claim: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return models.PromoRedemptionResult{}, fmt.Errorf("commit promo redemption: %w", err)
	}

	return models.PromoRedemptionResult{
		Code:         normalizedCode,
		DurationDays: durationDays,
		ExpiresAt:    expiresAt,
	}, nil
}

func ResolveAccountAccess(ctx context.Context, db *sql.DB, accountID int64, now time.Time) (models.AccountAccessState, error) {
	ownerEmail, ok, err := ownerEmailForAccount(ctx, db, accountID)
	if err != nil {
		return models.AccountAccessState{}, err
	}
	if ok {
		grant, hasGrant, err := directAccessGrantForEmail(ctx, db, ownerEmail)
		if err != nil {
			return models.AccountAccessState{}, err
		}
		if hasGrant {
			return models.AccountAccessState{
				AccessGranted: true,
				Source:        AccessSourceDirect,
				Plan:          grant.Plan,
			}, nil
		}
	}

	var (
		code      string
		expiresAt string
	)
	err = db.QueryRowContext(ctx, `
		SELECT pc.code, pcr.expires_at
		FROM promo_code_redemptions pcr
		INNER JOIN promo_codes pc
			ON pc.id = pcr.promo_code_id
		WHERE pcr.account_id = ?
		ORDER BY pcr.expires_at DESC, pcr.id DESC
		LIMIT 1;
	`, accountID).Scan(&code, &expiresAt)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.AccountAccessState{}, nil
	case err != nil:
		return models.AccountAccessState{}, fmt.Errorf("load account promo access: %w", err)
	}

	expiry, err := parseTimestamp(expiresAt)
	if err != nil {
		return models.AccountAccessState{}, fmt.Errorf("parse promo expiry: %w", err)
	}

	return models.AccountAccessState{
		AccessGranted: expiry.After(now),
		Source:        AccessSourcePromo,
		ExpiresAt:     expiresAt,
		PromoCode:     code,
		PromoExpired:  !expiry.After(now),
	}, nil
}

func accountCurrentlyHasAccessTx(ctx context.Context, tx *sql.Tx, accountID int64, now time.Time) (bool, error) {
	var billingStatus string
	err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(billing_status, '')
		FROM accounts
		WHERE id = ?
		LIMIT 1;
	`, accountID).Scan(&billingStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("load account billing status: %w", err)
	}
	if billingstate.GrantsAccess(billingStatus) {
		return true, nil
	}

	ownerEmail, ok, err := ownerEmailForAccountTx(ctx, tx, accountID)
	if err != nil {
		return false, err
	}
	if ok {
		grant, hasGrant, err := directAccessGrantForEmailTx(ctx, tx, ownerEmail)
		if err != nil {
			return false, err
		}
		if hasGrant && grant.ID > 0 {
			return true, nil
		}
	}

	var redemptionID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM promo_code_redemptions
		WHERE account_id = ?
		  AND expires_at > ?
		ORDER BY expires_at DESC, id DESC
		LIMIT 1;
	`, accountID, formatTimestamp(now)).Scan(&redemptionID)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("load active promo redemption: %w", err)
	default:
		return true, nil
	}
}

func cleanupExpiredPromoRedemptionClaimsTx(ctx context.Context, tx *sql.Tx, now time.Time) error {
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM promo_code_redemption_claims
		WHERE retention_until <= ?;
	`, formatTimestamp(now)); err != nil {
		return fmt.Errorf("cleanup promo redemption claims: %w", err)
	}

	return nil
}

func SweepExpiredPromoRedemptionClaims(ctx context.Context, db *sql.DB, now time.Time) error {
	if _, err := db.ExecContext(ctx, `
		DELETE FROM promo_code_redemption_claims
		WHERE retention_until <= ?;
	`, formatTimestamp(now)); err != nil {
		return fmt.Errorf("sweep promo redemption claims: %w", err)
	}

	return nil
}

func ownerEmailForAccount(ctx context.Context, db *sql.DB, accountID int64) (string, bool, error) {
	var email string
	err := db.QueryRowContext(ctx, `
		SELECT email
		FROM users
		WHERE account_id = ?
		  AND COALESCE(role, 'member') = 'owner'
		ORDER BY id
		LIMIT 1;
	`, accountID).Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("load owner email: %w", err)
	}

	return strings.ToLower(strings.TrimSpace(email)), true, nil
}

func ownerEmailForAccountTx(ctx context.Context, tx *sql.Tx, accountID int64) (string, bool, error) {
	var email string
	err := tx.QueryRowContext(ctx, `
		SELECT email
		FROM users
		WHERE account_id = ?
		  AND COALESCE(role, 'member') = 'owner'
		ORDER BY id
		LIMIT 1;
	`, accountID).Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("load owner email in tx: %w", err)
	}

	return strings.ToLower(strings.TrimSpace(email)), true, nil
}

func directAccessGrantForEmail(ctx context.Context, db *sql.DB, normalizedEmail string) (models.DirectAccessGrant, bool, error) {
	var grant models.DirectAccessGrant
	err := db.QueryRowContext(ctx, `
		SELECT id, email, COALESCE(plan, 'single'), COALESCE(note, ''), created_at
		FROM direct_access_grants
		WHERE LOWER(email) = LOWER(?)
		LIMIT 1;
	`, normalizedEmail).Scan(&grant.ID, &grant.Email, &grant.Plan, &grant.Note, &grant.CreatedAt)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.DirectAccessGrant{}, false, nil
	case err != nil:
		return models.DirectAccessGrant{}, false, fmt.Errorf("load direct access grant by email: %w", err)
	default:
		grant.Plan = billingplan.Normalize(grant.Plan)
		if grant.Plan == "" {
			grant.Plan = billingplan.PlanSingle
		}
		return grant, true, nil
	}
}

func hasDirectAccessGrantForEmail(ctx context.Context, db *sql.DB, normalizedEmail string) (bool, error) {
	_, ok, err := directAccessGrantForEmail(ctx, db, normalizedEmail)
	return ok, err
}

func directAccessGrantForEmailTx(ctx context.Context, tx *sql.Tx, normalizedEmail string) (models.DirectAccessGrant, bool, error) {
	var grant models.DirectAccessGrant
	err := tx.QueryRowContext(ctx, `
		SELECT id, email, COALESCE(plan, 'single'), COALESCE(note, ''), created_at
		FROM direct_access_grants
		WHERE LOWER(email) = LOWER(?)
		LIMIT 1;
	`, normalizedEmail).Scan(&grant.ID, &grant.Email, &grant.Plan, &grant.Note, &grant.CreatedAt)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.DirectAccessGrant{}, false, nil
	case err != nil:
		return models.DirectAccessGrant{}, false, fmt.Errorf("load direct access grant by email in tx: %w", err)
	default:
		grant.Plan = billingplan.Normalize(grant.Plan)
		if grant.Plan == "" {
			grant.Plan = billingplan.PlanSingle
		}
		return grant, true, nil
	}
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

func promoOwnerEmailHMAC(email string, secret string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if strings.TrimSpace(secret) == "" {
		sum := sha256.Sum256([]byte(normalized))
		return hex.EncodeToString(sum[:])
	}

	mac := hmac.New(sha256.New, []byte(strings.TrimSpace(secret)))
	_, _ = mac.Write([]byte(normalized))
	return hex.EncodeToString(mac.Sum(nil))
}

func normalizePromoCode(code string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(code))
	if !promoCodePattern.MatchString(normalized) {
		return "", ErrInvalidPromoCode
	}

	return normalized, nil
}

func formatTimestamp(ts time.Time) string {
	return ts.UTC().Format(timestampLayout)
}

func parseTimestamp(value string) (time.Time, error) {
	parsed, err := time.Parse(timestampLayout, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, err
	}

	return parsed, nil
}

func nullInt64(v int64) sql.NullInt64 {
	if v <= 0 {
		return sql.NullInt64{}
	}

	return sql.NullInt64{Int64: v, Valid: true}
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func isUniqueConstraintError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
