package settingsTx

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/models"
)

var (
	ErrStartingInvoiceNumberLocked  = errors.New("starting invoice number cannot be changed while invoices exist")
	ErrStartingInvoiceNumberInvalid = errors.New("starting invoice number must be greater than 0")
)

func Get(ctx context.Context, db *sql.DB) (models.Settings, error) {
	const q = `
		SELECT
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
			logo_url,
			show_item_type_headers
		FROM user_settings
		WHERE id = 1;
	`

	var s models.Settings

	err := db.QueryRowContext(ctx, q).Scan(
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
		&s.LogoURL,
		&s.ShowItemTypeHeaders,
	)
	if err != nil {
		return models.Settings{}, fmt.Errorf("get settings: %w", err)
	}

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

func Upsert(ctx context.Context, db *sql.DB, s models.Settings) error {
	if s.StartingInvoiceNumber < 1 {
		return ErrStartingInvoiceNumberInvalid
	}

	const q = `
		INSERT INTO user_settings (
			id,
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
			logo_url,
			show_item_type_headers
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
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
			logo_url = excluded.logo_url,
			show_item_type_headers = excluded.show_item_type_headers;
	`

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin settings upsert tx: %w", err)
	}
	defer tx.Rollback()

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
		1,
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
		s.LogoURL,
		s.ShowItemTypeHeaders,
	); err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit settings upsert: %w", err)
	}

	return nil
}
