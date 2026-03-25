package editor

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func nullStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func toEditorTotals(in invoiceTx.InvoiceOverviewTotals) models.InvoiceEditorTotals {
	return models.InvoiceEditorTotals{
		BaseNumber:        in.BaseNumber,
		RevisionNo:        in.RevisionNo,
		IssueDate:         in.IssueDate,
		DueByDate:         nullStringPtr(in.DueByDate),
		ClientName:        in.ClientName,
		ClientCompanyName: in.ClientCompanyName,
		ClientAddress:     in.ClientAddress,
		ClientEmail:       in.ClientEmail,
		Note:              nullStringPtr(in.Note),

		VATRate:       in.VATRate,
		VATAmountMin:  in.VATAmountMin,
		DiscountType:  in.DiscountType,
		DiscountRate:  in.DiscountRate,
		DiscountMinor: in.DiscountMinor,
		DepositType:   in.DepositType,
		DepositRate:   in.DepositRate,
		DepositMinor:  in.DepositMinor,
		SubtotalMinor: in.SubtotalMinor,
		TotalMinor:    in.TotalMinor,
		PaidMinor:     in.PaidMinor,
	}
}

func toEditorLines(in []invoiceTx.ItemLine) []models.InvoiceEditorLine {
	out := make([]models.InvoiceEditorLine, 0, len(in))
	for _, line := range in {
		out = append(out, models.InvoiceEditorLine{
			ProductID:     line.ProductID,
			PricingMode:   line.PricingMode,
			MinutesWorked: line.MinutesWorked,
			Name:          line.Name,
			LineType:      line.LineType,
			Quantity:      line.Quantity,
			UnitPriceMin:  line.UnitPriceMin,
			LineTotalMin:  line.LineTotalMin,
			SortOrder:     line.SortOrder,
		})
	}
	return out
}

func toEditorPayments(in []invoiceTx.PaymentRow) []models.InvoiceEditorPayment {
	out := make([]models.InvoiceEditorPayment, 0, len(in))
	for _, p := range in {
		out = append(out, models.InvoiceEditorPayment{
			ID:          p.ID,
			AmountMinor: p.AmountMinor,
			PaymentDate: p.PaymentDate,
			PaymentType: p.PaymentType,
			Label:       nullStringPtr(p.Label),
		})
	}
	return out
}

func GetInvoice(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}
		baseNo, ok := params.ValidateParam(w, r, "baseNo")
		if !ok {
			return
		}
		revNo, ok := params.ValidateParam(w, r, "revNo")
		if !ok {
			return
		}

		summary, err := invoiceTx.QueryInvoiceSummary(r.Context(), a.DB, clientID, baseNo, revNo)
		if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			res.NotFound(w, "Invoice revision not found")
			return
		}
			slog.ErrorContext(
				r.Context(), "DB_ERROR - error while getting invoice summary",
				"err", err,
				"clientID", clientID,
				"baseNumber", baseNo,
				"revisionNumber", revNo,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		lines, err := invoiceTx.QueryInvoiceLines(r.Context(), a.DB, clientID, baseNo, revNo)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "DB_ERROR - error while getting invoice lines",
				"err", err,
				"clientID", clientID,
				"baseNumber", baseNo,
				"revisionNumber", revNo,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		payments, err := invoiceTx.QueryInvoicePaymentsForRevision(r.Context(), a.DB, clientID, baseNo, revNo)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "DB_ERROR - error while getting invoice payments",
				"err", err,
				"clientID", clientID,
				"baseNumber", baseNo,
				"revisionNumber", revNo,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		out := models.InvoiceEditorResponse{
			Status: summary.Status,
			Totals:   toEditorTotals(*summary),
			Lines:    toEditorLines(lines),
			Payments: toEditorPayments(payments),
		}

		res.JSON(w, http.StatusOK, out)
	}
}
