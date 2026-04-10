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
		SupplyDate:        nullStringPtr(in.SupplyDate),
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

func toEditorHistory(in []invoiceTx.InvoiceHistoryRow) []models.InvoiceEditorHistoryItem {
	out := make([]models.InvoiceEditorHistoryItem, 0, len(in))
	for _, row := range in {
		entry := models.InvoiceEditorHistoryItem{
			ID:        row.ID,
			Type:      row.Type,
			CreatedAt: row.CreatedAt,
			Label:     nullStringPtr(row.Label),
		}

		if row.RevisionNo.Valid {
			revisionNo := row.RevisionNo.Int64
			entry.RevisionNo = &revisionNo
		}
		if row.ReceiptNo.Valid {
			receiptNo := row.ReceiptNo.Int64
			entry.ReceiptNo = &receiptNo
		}
		if row.IssueDate.Valid {
			issueDate := row.IssueDate.String
			entry.IssueDate = &issueDate
		}
		if row.DueByDate.Valid {
			dueByDate := row.DueByDate.String
			entry.DueByDate = &dueByDate
		}
		if row.PaymentDate.Valid {
			paymentDate := row.PaymentDate.String
			entry.PaymentDate = &paymentDate
		}
		if row.AmountMinor.Valid {
			amountMinor := row.AmountMinor.Int64
			entry.AmountMinor = &amountMinor
		}

		out = append(out, entry)
	}
	return out
}

func toSelectedReceipt(in *invoiceTx.PaymentReceiptRow) *models.InvoiceEditorReceipt {
	if in == nil {
		return nil
	}

	return &models.InvoiceEditorReceipt{
		ID:                in.ID,
		ReceiptNo:         in.ReceiptNo,
		PaymentDate:       in.PaymentDate,
		AmountMinor:       in.AmountMinor,
		Label:             nullStringPtr(in.Label),
		AppliedRevisionNo: in.AppliedRevisionNo,
		CreatedAt:         in.CreatedAt,
	}
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

		history, err := invoiceTx.QueryInvoiceHistory(r.Context(), a.DB, clientID, baseNo)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "DB_ERROR - error while getting invoice history",
				"err", err,
				"clientID", clientID,
				"baseNumber", baseNo,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		out := models.InvoiceEditorResponse{
			Status:  summary.Status,
			Totals:   toEditorTotals(*summary),
			Lines:    toEditorLines(lines),
			History:  toEditorHistory(history),
			Payments: toEditorPayments(payments),
		}

		res.JSON(w, http.StatusOK, out)
	}
}

func GetPaymentReceipt(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}
		baseNo, ok := params.ValidateParam(w, r, "baseNo")
		if !ok {
			return
		}
		receiptNo, ok := params.ValidateParam(w, r, "receiptNo")
		if !ok {
			return
		}

		receipt, err := invoiceTx.QueryPaymentReceiptByNumber(r.Context(), a.DB, clientID, baseNo, receiptNo)
		if err != nil {
			if errors.Is(err, invoiceTx.ErrPaymentReceiptNotFound) {
				res.NotFound(w, "Payment receipt not found")
				return
			}
			slog.ErrorContext(
				r.Context(), "DB_ERROR - error while getting payment receipt",
				"err", err,
				"clientID", clientID,
				"baseNumber", baseNo,
				"receiptNo", receiptNo,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		summary, err := invoiceTx.QueryInvoiceSummary(r.Context(), a.DB, clientID, baseNo, receipt.AppliedRevisionNo)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "DB_ERROR - error while getting payment receipt invoice summary",
				"err", err,
				"clientID", clientID,
				"baseNumber", baseNo,
				"receiptNo", receiptNo,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		lines, err := invoiceTx.QueryInvoiceLines(r.Context(), a.DB, clientID, baseNo, receipt.AppliedRevisionNo)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "DB_ERROR - error while getting payment receipt invoice lines",
				"err", err,
				"clientID", clientID,
				"baseNumber", baseNo,
				"receiptNo", receiptNo,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		history, err := invoiceTx.QueryInvoiceHistory(r.Context(), a.DB, clientID, baseNo)
		if err != nil {
			slog.ErrorContext(
				r.Context(), "DB_ERROR - error while getting payment receipt invoice history",
				"err", err,
				"clientID", clientID,
				"baseNumber", baseNo,
				"receiptNo", receiptNo,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		out := models.InvoiceEditorResponse{
			Status:          summary.Status,
			Totals:          toEditorTotals(*summary),
			Lines:           toEditorLines(lines),
			History:         toEditorHistory(history),
			SelectedReceipt: toSelectedReceipt(receipt),
		}

		res.JSON(w, http.StatusOK, out)
	}
}
