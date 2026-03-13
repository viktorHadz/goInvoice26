package pdf

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

type InvoicePDFRenderer interface {
	RenderPDF(ctx context.Context, doc models.InvoicePDFData) ([]byte, error)
}

// RenderPDF delegates to the chosen renderer implementation.
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

	lines := make([]models.InvoicePDFItem, 0, len(rawItems))
	for _, it := range rawItems {
		lines = append(lines, models.InvoicePDFItem{
			Name:      it.Name,
			LineType:  it.LineType,
			Quantity:  formatQuantity(it.Quantity),
			ItemPrice: formatPrice(it.UnitPriceMin),
			ItemTotal: formatPrice(it.LineTotalMin),
			SortOrder: it.SortOrder,
		})
	}

	return buildInvoicePDFData(overview, lines), nil
}

// BuildQuickInvoice builds a PDF from in-memory invoice data without persisting to DB.
func BuildQuickInvoice(invoice models.FEInvoiceIn, revisionNo int64) models.InvoicePDFData {
	lines := make([]models.InvoicePDFItem, 0, len(invoice.Lines))
	for _, line := range invoice.Lines {
		lines = append(lines, models.InvoicePDFItem{
			Name:      line.Name,
			LineType:  line.LineType,
			Quantity:  formatQuantity(line.Quantity),
			ItemPrice: formatPrice(line.UnitPriceMinor),
			ItemTotal: formatPrice(line.LineTotalMinor),
			SortOrder: line.SortOrder,
		})
	}

	var dueByDate sql.NullString
	if invoice.Overview.DueByDate != nil {
		dueByDate = sql.NullString{String: *invoice.Overview.DueByDate, Valid: true}
	}

	overview := &invoiceTx.InvoiceOverviewTotals{
		BaseNumber:        invoice.Overview.BaseNumber,
		RevisionNo:        revisionNo,
		IssueDate:         invoice.Overview.IssueDate,
		DueByDate:         dueByDate,
		ClientName:        invoice.Overview.ClientName,
		ClientCompanyName: invoice.Overview.ClientCompanyName,
		ClientAddress:     invoice.Overview.ClientAddress,
		ClientEmail:       invoice.Overview.ClientEmail,
		VATRate:           invoice.Totals.VATRate,
		VATAmountMin:      invoice.Totals.VatAmountMinor,
		DiscountType:      invoice.Totals.DiscountType,
		DiscountRate:      invoice.Totals.DiscountRate,
		DiscountMinor:     invoice.Totals.DiscountMinor,
		DepositType:       invoice.Totals.DepositType,
		DepositRate:       invoice.Totals.DepositRate,
		DepositMinor:      invoice.Totals.DepositMinor,
		SubtotalMinor:     invoice.Totals.SubtotalMinor,
		TotalMinor:        invoice.Totals.TotalMinor,
		PaidMinor:         invoice.Totals.PaidMinor,
	}

	return buildInvoicePDFData(overview, lines)
}

func buildInvoicePDFData(
	o *invoiceTx.InvoiceOverviewTotals,
	lines []models.InvoicePDFItem,
) models.InvoicePDFData {
	var dueDate *string
	if o.DueByDate.Valid {
		dueDate = &o.DueByDate.String
	}

	subtotalAfterDisc := o.SubtotalMinor - o.DiscountMinor
	balanceDue := o.TotalMinor - o.PaidMinor

	return models.InvoicePDFData{
		BaseNumber:     o.BaseNumber,
		RevisionNumber: fmt.Sprintf("%d", o.RevisionNo),
		IssueAt:        o.IssueDate,
		DueDate:        dueDate,
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
			SubtotalMinor:     o.SubtotalMinor,
			SubtotalAfterDisc: subtotalAfterDisc,
			TotalMinor:        o.TotalMinor,
			PaidMinor:         o.PaidMinor,
			BalanceDue:        balanceDue,
		},
	}
}

func formatPrice(minorUnits int64) string {
	pounds := minorUnits / 100
	pence := minorUnits % 100
	return fmt.Sprintf("£%d.%02d", pounds, pence)
}

func formatQuantity(qty int64) string {
	return fmt.Sprintf("%d", qty)
}
