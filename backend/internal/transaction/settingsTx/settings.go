package settingsTx

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/models"
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
			custom_items_prefix,
			payment_terms,
			payment_details,
			notes_footer,
			logo_url
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
		&s.CustomItemsPrefix,
		&s.PaymentTerms,
		&s.PaymentDetails,
		&s.NotesFooter,
		&s.LogoURL,
	)
	if err != nil {
		return models.Settings{}, fmt.Errorf("get settings: %w", err)
	}

	return s, nil
}

func Upsert(ctx context.Context, db *sql.DB, s models.Settings) error {
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
			custom_items_prefix,
			payment_terms,
			payment_details,
			notes_footer,
			logo_url
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			company_name = excluded.company_name,
			email = excluded.email,
			phone = excluded.phone,
			company_address = excluded.company_address,
			invoice_prefix = excluded.invoice_prefix,
			currency = excluded.currency,
			date_format = excluded.date_format,
			custom_items_prefix = excluded.custom_items_prefix,
			payment_terms = excluded.payment_terms,
			payment_details = excluded.payment_details,
			notes_footer = excluded.notes_footer,
			logo_url = excluded.logo_url;
	`

	_, err := db.ExecContext(
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
		s.CustomItemsPrefix,
		s.PaymentTerms,
		s.PaymentDetails,
		s.NotesFooter,
		s.LogoURL,
	)
	if err != nil {
		return fmt.Errorf("upsert settings: %w", err)
	}

	return nil
}
