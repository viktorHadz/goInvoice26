package settingsTx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/models"
)

const StoredFileKindLogo = "logo"

var (
	ErrStartingInvoiceNumberLocked  = errors.New("starting invoice number cannot be changed while invoices exist")
	ErrStartingInvoiceNumberInvalid = errors.New("starting invoice number must be greater than 0")
)

type StoredFile struct {
	ID              int64
	AccountID       int64
	Kind            string
	StorageKey      string
	ContentType     string
	CreatedAt       string
	DeletePendingAt sql.NullString
}

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func Get(ctx context.Context, db *sql.DB, accountID int64) (models.Settings, error) {
	if err := ensureAccountSettingsRow(ctx, db, accountID); err != nil {
		return models.Settings{}, err
	}

	const q = `
		SELECT
			s.company_name,
			s.email,
			s.phone,
			s.company_address,
			s.invoice_prefix,
			s.currency,
			s.date_format,
			s.payment_terms,
			s.payment_details,
			s.notes_footer,
			COALESCE(s.logo_asset_id, 0),
			COALESCE(f.storage_key, ''),
			s.show_item_type_headers
		FROM account_settings s
		LEFT JOIN stored_files f
			ON f.id = s.logo_asset_id
		WHERE s.account_id = ?;
	`

	var s models.Settings
	err := db.QueryRowContext(ctx, q, accountID).Scan(
		&s.CompanyName,
		&s.Email,
		&s.Phone,
		&s.CompanyAddress,
		&s.InvoicePrefix,
		&s.Currency,
		&s.DateFormat,
		&s.PaymentTerms,
		&s.PaymentDetails,
		&s.NotesFooter,
		&s.LogoAssetID,
		&s.LogoStorageKey,
		&s.ShowItemTypeHeaders,
	)
	if err != nil {
		return models.Settings{}, fmt.Errorf("get settings: %w", err)
	}
	if s.LogoStorageKey == "" {
		s.LogoAssetID = 0
	}
	s.LogoURL = buildLogoURL(s.LogoAssetID)

	if err := db.QueryRowContext(ctx, `
		SELECT next_base_number
		FROM invoice_number_seq
		WHERE id = 1;
	`).Scan(&s.StartingInvoiceNumber); err != nil {
		return models.Settings{}, fmt.Errorf("get invoice number sequence: %w", err)
	}

	var invoiceCount int64
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM invoices;
	`).Scan(&invoiceCount); err != nil {
		return models.Settings{}, fmt.Errorf("count invoices: %w", err)
	}
	s.CanEditStartingInvoiceNumber = invoiceCount == 0

	return s, nil
}

func Upsert(ctx context.Context, db *sql.DB, accountID int64, s models.Settings) error {
	if s.StartingInvoiceNumber < 1 {
		return ErrStartingInvoiceNumberInvalid
	}

	const q = `
		INSERT INTO account_settings (
			account_id,
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
			show_item_type_headers,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ','now'))
		ON CONFLICT(account_id) DO UPDATE SET
			company_name = excluded.company_name,
			email = excluded.email,
			phone = excluded.phone,
			company_address = excluded.company_address,
			invoice_prefix = excluded.invoice_prefix,
			currency = excluded.currency,
			date_format = excluded.date_format,
			payment_terms = excluded.payment_terms,
			payment_details = excluded.payment_details,
			notes_footer = excluded.notes_footer,
			show_item_type_headers = excluded.show_item_type_headers,
			updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now');
	`

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin settings upsert tx: %w", err)
	}
	defer tx.Rollback()

	if err := ensureAccountSettingsRow(ctx, tx, accountID); err != nil {
		return err
	}

	var (
		invoiceCount    int64
		currentSequence int64
	)
	if err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM invoices;
	`).Scan(&invoiceCount); err != nil {
		return fmt.Errorf("count invoices: %w", err)
	}
	if err := tx.QueryRowContext(ctx, `
		SELECT next_base_number
		FROM invoice_number_seq
		WHERE id = 1;
	`).Scan(&currentSequence); err != nil {
		return fmt.Errorf("get current invoice sequence: %w", err)
	}

	if s.StartingInvoiceNumber != currentSequence {
		if invoiceCount > 0 {
			return ErrStartingInvoiceNumberLocked
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE invoice_number_seq
			SET next_base_number = ?
			WHERE id = 1;
		`, s.StartingInvoiceNumber); err != nil {
			return fmt.Errorf("update invoice number sequence: %w", err)
		}
	}

	if _, err := tx.ExecContext(
		ctx,
		q,
		accountID,
		s.CompanyName,
		s.Email,
		s.Phone,
		s.CompanyAddress,
		s.InvoicePrefix,
		s.Currency,
		s.DateFormat,
		s.PaymentTerms,
		s.PaymentDetails,
		s.NotesFooter,
		s.ShowItemTypeHeaders,
	); err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit settings upsert: %w", err)
	}

	return nil
}

func GetLogoFile(ctx context.Context, db *sql.DB, accountID int64) (StoredFile, bool, error) {
	if err := ensureAccountSettingsRow(ctx, db, accountID); err != nil {
		return StoredFile{}, false, err
	}

	file, ok, err := getLogoFile(ctx, db, accountID)
	if err != nil {
		return StoredFile{}, false, err
	}
	return file, ok, nil
}

func ReplaceLogo(ctx context.Context, db *sql.DB, accountID int64, storageKey, contentType string) (StoredFile, *StoredFile, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return StoredFile{}, nil, fmt.Errorf("begin replace logo tx: %w", err)
	}
	defer tx.Rollback()

	if err := ensureAccountSettingsRow(ctx, tx, accountID); err != nil {
		return StoredFile{}, nil, err
	}

	prev, ok, err := getLogoFile(ctx, tx, accountID)
	if err != nil {
		return StoredFile{}, nil, err
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO stored_files (
			account_id,
			kind,
			storage_key,
			content_type
		) VALUES (?, ?, ?, ?);
	`, accountID, StoredFileKindLogo, storageKey, contentType)
	if err != nil {
		return StoredFile{}, nil, fmt.Errorf("insert stored file: %w", err)
	}

	fileID, err := res.LastInsertId()
	if err != nil {
		return StoredFile{}, nil, fmt.Errorf("stored file id: %w", err)
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE account_settings
		SET
			logo_asset_id = ?,
			updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
		WHERE account_id = ?;
	`, fileID, accountID); err != nil {
		return StoredFile{}, nil, fmt.Errorf("assign logo asset: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return StoredFile{}, nil, fmt.Errorf("commit replace logo: %w", err)
	}

	out := StoredFile{
		ID:          fileID,
		AccountID:   accountID,
		Kind:        StoredFileKindLogo,
		StorageKey:  storageKey,
		ContentType: contentType,
	}
	if !ok {
		return out, nil, nil
	}

	return out, &prev, nil
}

func RemoveLogo(ctx context.Context, db *sql.DB, accountID int64) (*StoredFile, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin remove logo tx: %w", err)
	}
	defer tx.Rollback()

	if err := ensureAccountSettingsRow(ctx, tx, accountID); err != nil {
		return nil, err
	}

	prev, ok, err := getLogoFile(ctx, tx, accountID)
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE account_settings
		SET
			logo_asset_id = NULL,
			updated_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
		WHERE account_id = ?;
	`, accountID); err != nil {
		return nil, fmt.Errorf("clear logo asset: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit remove logo: %w", err)
	}

	if !ok {
		return nil, nil
	}

	return &prev, nil
}

func MarkStoredFileDeletePending(ctx context.Context, db *sql.DB, fileID int64) error {
	if _, err := db.ExecContext(ctx, `
		UPDATE stored_files
		SET delete_pending_at = strftime('%Y-%m-%dT%H:%M:%fZ','now')
		WHERE id = ?;
	`, fileID); err != nil {
		return fmt.Errorf("mark stored file delete pending: %w", err)
	}
	return nil
}

func DeleteStoredFile(ctx context.Context, db *sql.DB, fileID int64) error {
	if _, err := db.ExecContext(ctx, `
		DELETE FROM stored_files
		WHERE id = ?;
	`, fileID); err != nil {
		return fmt.Errorf("delete stored file row: %w", err)
	}
	return nil
}

func ListDeletePendingFiles(ctx context.Context, db *sql.DB) ([]StoredFile, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT
			id,
			account_id,
			kind,
			storage_key,
			content_type,
			created_at,
			delete_pending_at
		FROM stored_files
		WHERE delete_pending_at IS NOT NULL
		ORDER BY id ASC;
	`)
	if err != nil {
		return nil, fmt.Errorf("list delete pending files: %w", err)
	}
	defer rows.Close()

	var files []StoredFile
	for rows.Next() {
		var file StoredFile
		if err := rows.Scan(
			&file.ID,
			&file.AccountID,
			&file.Kind,
			&file.StorageKey,
			&file.ContentType,
			&file.CreatedAt,
			&file.DeletePendingAt,
		); err != nil {
			return nil, fmt.Errorf("scan delete pending file: %w", err)
		}
		files = append(files, file)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate delete pending files: %w", err)
	}

	return files, nil
}

func GetLegacyLogoURL(ctx context.Context, db *sql.DB, accountID int64) (string, error) {
	if err := ensureAccountSettingsRow(ctx, db, accountID); err != nil {
		return "", err
	}

	var logoURL sql.NullString
	if err := db.QueryRowContext(ctx, `
		SELECT legacy_logo_url
		FROM account_settings
		WHERE account_id = ?;
	`, accountID).Scan(&logoURL); err != nil {
		return "", fmt.Errorf("get legacy logo url: %w", err)
	}
	if !logoURL.Valid {
		return "", nil
	}
	return logoURL.String, nil
}

func ClearLegacyLogoURL(ctx context.Context, db *sql.DB, accountID int64) error {
	if err := ensureAccountSettingsRow(ctx, db, accountID); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `
		UPDATE account_settings
		SET legacy_logo_url = ''
		WHERE account_id = ?;
	`, accountID); err != nil {
		return fmt.Errorf("clear legacy logo url: %w", err)
	}
	return nil
}

func buildLogoURL(assetID int64) string {
	if assetID <= 0 {
		return ""
	}
	return fmt.Sprintf("/api/settings/logo?v=%d", assetID)
}

func ensureAccountSettingsRow(ctx context.Context, exec execer, accountID int64) error {
	if accountID <= 0 {
		accountID = 1
	}

	if _, err := exec.ExecContext(ctx, `
		INSERT OR IGNORE INTO accounts (id)
		VALUES (?);
	`, accountID); err != nil {
		return fmt.Errorf("ensure account row: %w", err)
	}
	if _, err := exec.ExecContext(ctx, `
		INSERT OR IGNORE INTO account_settings (account_id)
		VALUES (?);
	`, accountID); err != nil {
		return fmt.Errorf("ensure account settings row: %w", err)
	}

	return nil
}

type queryRowScanner interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func getLogoFile(ctx context.Context, q queryRowScanner, accountID int64) (StoredFile, bool, error) {
	const query = `
		SELECT
			f.id,
			f.account_id,
			f.kind,
			f.storage_key,
			f.content_type,
			f.created_at,
			f.delete_pending_at
		FROM account_settings s
		JOIN stored_files f
			ON f.id = s.logo_asset_id
		WHERE s.account_id = ?;
	`

	var file StoredFile
	err := q.QueryRowContext(ctx, query, accountID).Scan(
		&file.ID,
		&file.AccountID,
		&file.Kind,
		&file.StorageKey,
		&file.ContentType,
		&file.CreatedAt,
		&file.DeletePendingAt,
	)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return StoredFile{}, false, nil
	case err != nil:
		return StoredFile{}, false, fmt.Errorf("get logo file: %w", err)
	default:
		return file, true, nil
	}
}
