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

// GenerateInvoicePDF delegates to the chosen renderer implementation.
func GenerateInvoicePDF(
	ctx context.Context,
	renderer InvoicePDFRenderer,
	doc models.InvoicePDFData,
) ([]byte, error) {
	return renderer.RenderPDF(ctx, doc)
}

func BuildInvoicePDFData(
	ctx context.Context,
	db *sql.DB,
	clientID int64,
	baseNo int64,
	revNo int64,
) (models.InvoicePDFData, error) {
	overview, err := invoiceTx.GetInvoiceOverviewTotals(ctx, db, clientID, baseNo, revNo)
	if err != nil {
		return models.InvoicePDFData{}, fmt.Errorf("get invoice overview: %w", err)
	}

	rawItems, err := invoiceTx.GetInvoiceItems(ctx, db, clientID, baseNo, revNo)
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
