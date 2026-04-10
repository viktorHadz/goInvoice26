package pdf

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/service/invoiceformat"
	"github.com/viktorHadz/goInvoice26/internal/service/storage"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/settingsTx"
)

type InvoicePDFRenderer interface {
	RenderPDF(ctx context.Context, doc models.InvoicePDFData) ([]byte, error)
}

// RenderPDF assigns to the chosen renderer.
func RenderPDF(
	ctx context.Context,
	renderer InvoicePDFRenderer,
	doc models.InvoicePDFData,
) ([]byte, error) {
	return renderer.RenderPDF(ctx, doc)
}

func BuildInvoiceFromDB(
	ctx context.Context,
	db *sql.DB,
	clientID int64,
	baseNo int64,
	revNo int64,
) (models.InvoicePDFData, error) {
	overview, err := invoiceTx.QueryInvoiceSummary(ctx, db, clientID, baseNo, revNo)
	if err != nil {
		return models.InvoicePDFData{}, fmt.Errorf("get invoice overview: %w", err)
	}

	rawItems, err := invoiceTx.QueryInvoiceLines(ctx, db, clientID, baseNo, revNo)
	if err != nil {
		return models.InvoicePDFData{}, fmt.Errorf("get invoice items: %w", err)
	}

	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return models.InvoicePDFData{}, fmt.Errorf("get account scope: %w", err)
	}

	settings, err := settingsTx.Get(ctx, db, accountID)
	if err != nil {
		return models.InvoicePDFData{}, fmt.Errorf("get settings: %w", err)
	}

	lines := make([]models.InvoicePDFItem, 0, len(rawItems))
	for _, it := range rawItems {
		pricingMode := ""
		if it.PricingMode != nil {
			pricingMode = *it.PricingMode
		}
		pricing := buildInvoicePDFPricing(pricingMode, it.UnitPriceMin, it.MinutesWorked, settings.Currency)

		lines = append(lines, models.InvoicePDFItem{
			Name:       it.Name,
			LineType:   it.LineType,
			Quantity:   formatQuantity(it.Quantity),
			ItemPrice:  pricing.itemPrice,
			TimeWorked: pricing.timeWorked,
			HourlyRate: pricing.hourlyRate,
			ItemTotal:  formatMoney(it.LineTotalMin, settings.Currency),
			SortOrder:  it.SortOrder,
		})
	}

	return buildInvoicePDFData(overview, lines, settings), nil
}

func BuildPaymentReceiptFromDB(
	ctx context.Context,
	db *sql.DB,
	clientID int64,
	baseNo int64,
	receiptNo int64,
) (models.InvoicePDFData, error) {
	receipt, err := invoiceTx.QueryPaymentReceiptByNumber(ctx, db, clientID, baseNo, receiptNo)
	if err != nil {
		return models.InvoicePDFData{}, fmt.Errorf("get payment receipt: %w", err)
	}

	overview, err := invoiceTx.QueryInvoiceSummary(ctx, db, clientID, baseNo, receipt.AppliedRevisionNo)
	if err != nil {
		return models.InvoicePDFData{}, fmt.Errorf("get payment receipt invoice summary: %w", err)
	}

	accountID, err := accountscope.Require(ctx)
	if err != nil {
		return models.InvoicePDFData{}, fmt.Errorf("get account scope: %w", err)
	}

	settings, err := settingsTx.Get(ctx, db, accountID)
	if err != nil {
		return models.InvoicePDFData{}, fmt.Errorf("get settings: %w", err)
	}

	var paidUpToReceipt int64
	if err := db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(amount_minor), 0)
		FROM payments
		WHERE invoice_id = ?
		  AND payment_type = 'payment'
		  AND receipt_no <= ?;
	`, receipt.InvoiceID, receipt.ReceiptNo).Scan(&paidUpToReceipt); err != nil {
		return models.InvoicePDFData{}, fmt.Errorf("sum payment receipts to receipt number: %w", err)
	}

	return buildPaymentReceiptPDFData(overview, receipt, paidUpToReceipt, settings), nil
}

// BuildQuickInvoice builds a PDF from in-memory invoice data without saving to DB.
func BuildQuickInvoice(
	invoice models.FEInvoiceIn,
	settings models.Settings,
	revisionNo int64,
) models.InvoicePDFData {
	lines := make([]models.InvoicePDFItem, 0, len(invoice.Lines))
	for _, line := range invoice.Lines {
		pricing := buildInvoicePDFPricing(
			line.PricingMode,
			line.UnitPriceMinor,
			line.MinutesWorked,
			settings.Currency,
		)

		lines = append(lines, models.InvoicePDFItem{
			Name:       line.Name,
			LineType:   line.LineType,
			Quantity:   formatQuantity(line.Quantity),
			ItemPrice:  pricing.itemPrice,
			TimeWorked: pricing.timeWorked,
			HourlyRate: pricing.hourlyRate,
			ItemTotal:  formatMoney(line.LineTotalMinor, settings.Currency),
			SortOrder:  line.SortOrder,
		})
	}

	var dueByDate sql.NullString
	if invoice.Overview.DueByDate != nil {
		dueByDate = sql.NullString{
			String: *invoice.Overview.DueByDate,
			Valid:  true,
		}
	}

	var note sql.NullString
	if invoice.Overview.Note != nil && *invoice.Overview.Note != "" {
		note = sql.NullString{
			String: *invoice.Overview.Note,
			Valid:  true,
		}
	}

	overview := &invoiceTx.InvoiceOverviewTotals{
		BaseNumber:        invoice.Overview.BaseNumber,
		RevisionNo:        revisionNo,
		IssueDate:         invoice.Overview.IssueDate,
		SupplyDate:        nullStringFromPointer(invoice.Overview.SupplyDate),
		DueByDate:         dueByDate,
		ClientName:        invoice.Overview.ClientName,
		ClientCompanyName: invoice.Overview.ClientCompanyName,
		ClientAddress:     invoice.Overview.ClientAddress,
		ClientEmail:       invoice.Overview.ClientEmail,
		Note:              note,

		VATRate:       invoice.Totals.VATRate,
		VATAmountMin:  invoice.Totals.VatAmountMinor,
		DiscountType:  invoice.Totals.DiscountType,
		DiscountRate:  invoice.Totals.DiscountRate,
		DiscountMinor: invoice.Totals.DiscountMinor,
		DepositType:   invoice.Totals.DepositType,
		DepositRate:   invoice.Totals.DepositRate,
		DepositMinor:  invoice.Totals.DepositMinor,
		SubtotalMinor: invoice.Totals.SubtotalMinor,
		TotalMinor:    invoice.Totals.TotalMinor,
		PaidMinor:     invoice.Totals.PaidMinor,
	}

	return buildInvoicePDFData(overview, lines, settings)
}

func buildInvoicePDFData(
	o *invoiceTx.InvoiceOverviewTotals,
	lines []models.InvoicePDFItem,
	s models.Settings,
) models.InvoicePDFData {
	var dueDate *string
	if o.DueByDate.Valid {
		v := formatDate(o.DueByDate.String, s.DateFormat)
		dueDate = &v
	}
	var supplyDate *string
	if o.SupplyDate.Valid {
		v := formatDate(o.SupplyDate.String, s.DateFormat)
		supplyDate = &v
	}

	var note *string
	if o.Note.Valid {
		v := o.Note.String
		note = &v
	}

	subtotalAfterDisc := o.SubtotalMinor - o.DiscountMinor
	balanceDue := o.TotalMinor - o.PaidMinor
	if balanceDue < 0 {
		balanceDue = 0
	}

	logoPath := ""
	if s.LogoStorageKey != "" {
		logoPath = storage.NewLocalStore(storage.DefaultRootDir).Path(s.LogoStorageKey)
	}

	return models.InvoicePDFData{
		DocumentKind:        "invoice",
		Title:               "Invoice",
		InvoiceNumberLabel:  invoiceformat.FormatInvoiceNumber(s.InvoicePrefix, o.BaseNumber, o.RevisionNo),
		Currency:            fallbackCurrency(s.Currency),
		ShowItemTypeHeaders: s.ShowItemTypeHeaders,

		IssueAt:    formatDate(o.IssueDate, s.DateFormat),
		SupplyDate: supplyDate,
		DueDate:    dueDate,
		Note:       note,

		Issuer: models.InvoicePDFIssuer{
			CompanyName:    s.CompanyName,
			Email:          s.Email,
			Phone:          s.Phone,
			CompanyAddress: s.CompanyAddress,
			LogoPath:       logoPath,
		},
		Client: models.CreateClient{
			Name:        o.ClientName,
			CompanyName: o.ClientCompanyName,
			Address:     o.ClientAddress,
			Email:       o.ClientEmail,
		},
		Lines: lines,
		Totals: models.TotalsCreateIn{
			VATRate:           o.VATRate,
			VatAmountMinor:    o.VATAmountMin,
			DiscountType:      o.DiscountType,
			DiscountRate:      o.DiscountRate,
			DiscountMinor:     o.DiscountMinor,
			DepositType:       o.DepositType,
			DepositRate:       o.DepositRate,
			DepositMinor:      o.DepositMinor,
			PaidMinor:         o.PaidMinor,
			SubtotalAfterDisc: subtotalAfterDisc,
			SubtotalMinor:     o.SubtotalMinor,
			TotalMinor:        o.TotalMinor,
			BalanceDue:        balanceDue,
		},
		PaymentTerms:   s.PaymentTerms,
		PaymentDetails: s.PaymentDetails,
		NotesFooter:    s.NotesFooter,
	}
}

func buildPaymentReceiptPDFData(
	o *invoiceTx.InvoiceOverviewTotals,
	receipt *invoiceTx.PaymentReceiptRow,
	paidUpToReceipt int64,
	s models.Settings,
) models.InvoicePDFData {
	referenceNumberLabel := invoiceformat.FormatInvoiceNumber(s.InvoicePrefix, o.BaseNumber, receipt.AppliedRevisionNo)
	receiptNumberLabel := invoiceformat.FormatPaymentReceiptNumber(s.InvoicePrefix, o.BaseNumber, receipt.ReceiptNo)

	balanceDue := o.TotalMinor - paidUpToReceipt
	if balanceDue < 0 {
		balanceDue = 0
	}

	logoPath := ""
	if s.LogoStorageKey != "" {
		logoPath = storage.NewLocalStore(storage.DefaultRootDir).Path(s.LogoStorageKey)
	}

	lines := []models.InvoicePDFItem{
		{
			Name:      fmt.Sprintf("Payment received for %s", referenceNumberLabel),
			LineType:  "custom",
			Quantity:  "1",
			ItemPrice: formatMoney(receipt.AmountMinor, s.Currency),
			ItemTotal: formatMoney(receipt.AmountMinor, s.Currency),
			SortOrder: 1,
		},
	}

	var note *string
	if receipt.Label.Valid && receipt.Label.String != "" {
		value := receipt.Label.String
		note = &value
	}

	paymentDetails := fmt.Sprintf("Reference invoice: %s", referenceNumberLabel)
	if note != nil {
		paymentDetails += "\nReceipt note: " + *note
	}

	return models.InvoicePDFData{
		DocumentKind:         "payment_receipt",
		Title:                "Payment Receipt",
		InvoiceNumberLabel:   receiptNumberLabel,
		ReferenceNumberLabel: referenceNumberLabel,
		ReceiptAmountMinor:   receipt.AmountMinor,
		Currency:             fallbackCurrency(s.Currency),
		ShowItemTypeHeaders:  false,

		IssueAt: formatDate(receipt.PaymentDate, s.DateFormat),
		Note:    note,

		Issuer: models.InvoicePDFIssuer{
			CompanyName:    s.CompanyName,
			Email:          s.Email,
			Phone:          s.Phone,
			CompanyAddress: s.CompanyAddress,
			LogoPath:       logoPath,
		},
		Client: models.CreateClient{
			Name:        o.ClientName,
			CompanyName: o.ClientCompanyName,
			Address:     o.ClientAddress,
			Email:       o.ClientEmail,
		},
		Lines: lines,
		Totals: models.TotalsCreateIn{
			DepositType:   o.DepositType,
			DepositRate:   o.DepositRate,
			DepositMinor:  o.DepositMinor,
			PaidMinor:     paidUpToReceipt,
			SubtotalMinor: receipt.AmountMinor,
			TotalMinor:    o.TotalMinor,
			BalanceDue:    balanceDue,
		},
		PaymentDetails: paymentDetails,
		NotesFooter:    s.NotesFooter,
	}
}

func nullStringFromPointer(value *string) sql.NullString {
	if value == nil || *value == "" {
		return sql.NullString{}
	}

	return sql.NullString{
		String: *value,
		Valid:  true,
	}
}

func fallbackCurrency(v string) string {
	switch v {
	case "EUR", "USD", "GBP":
		return v
	default:
		return "GBP"
	}
}

func formatMoney(minorUnits int64, currency string) string {
	sign := ""
	if minorUnits < 0 {
		sign = "-"
		minorUnits = -minorUnits
	}

	symbol := currencySymbol(currency)
	major := minorUnits / 100
	minor := minorUnits % 100

	return fmt.Sprintf("%s%s%d.%02d", sign, symbol, major, minor)
}

func currencySymbol(code string) string {
	switch code {
	case "EUR":
		return "€"
	case "USD":
		return "$"
	default:
		return "£"
	}
}

func formatQuantity(qty int64) string {
	return fmt.Sprintf("%d", qty)
}

func formatDate(input string, dateFormat string) string {
	t, err := time.Parse("2006-01-02", input)
	if err != nil {
		return input
	}

	switch dateFormat {
	case "mm/dd/yyyy":
		return t.Format("01/02/2006")
	case "yyyy-mm-dd":
		return t.Format("2006-01-02")
	default:
		return t.Format("02/01/2006")
	}
}
