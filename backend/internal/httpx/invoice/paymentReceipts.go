package invoice

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func CreatePaymentReceipt(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}
		baseNumber, ok := params.ValidateParam(w, r, "baseNumber")
		if !ok {
			return
		}

		var dto models.PaymentReceiptCreateIn
		if ok := res.DecodeJSON(w, r, &dto); !ok {
			return
		}

		valid, errs := ValidatePaymentReceiptCreate(dto)
		if len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}

		invoiceID, receiptID, receiptNo, err := invoiceTx.CreatePaymentReceipt(r.Context(), a, clientID, baseNumber, &valid)
		if err != nil {
			switch {
			case errors.Is(err, invoiceTx.ErrInvoiceNotFound):
				res.Error(w, http.StatusNotFound, "NOT_FOUND", "Invoice not found")
				return
			case errors.Is(err, invoiceTx.ErrInvoiceDraftForReceipt):
				res.Error(w, http.StatusConflict, "INVOICE_DRAFT", "Issue the invoice before recording a payment receipt")
				return
			case errors.Is(err, invoiceTx.ErrInvoicePaidForReceipt):
				res.Error(w, http.StatusConflict, "INVOICE_PAID", "Invoice is already fully paid")
				return
			case errors.Is(err, invoiceTx.ErrInvoiceVoidForReceipt):
				res.Error(w, http.StatusConflict, "INVOICE_VOID", "Invoice is void; payment receipts are not allowed")
				return
			}

			slog.ErrorContext(r.Context(),
				"create payment receipt failed",
				"client_id", clientID,
				"base_number", baseNumber,
				"err", err,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		res.JSON(w, http.StatusCreated, map[string]any{
			"invoiceId": invoiceID,
			"receiptId": receiptID,
			"receiptNo": receiptNo,
		})
	}
}

func UpdatePaymentReceipt(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}
		baseNumber, ok := params.ValidateParam(w, r, "baseNumber")
		if !ok {
			return
		}
		receiptNo, ok := params.ValidateParam(w, r, "receiptNo")
		if !ok {
			return
		}

		var dto models.PaymentReceiptUpdateIn
		if ok := res.DecodeJSON(w, r, &dto); !ok {
			return
		}

		valid, errs := ValidatePaymentReceiptUpdate(dto)
		if len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}

		receiptID, err := invoiceTx.UpdatePaymentReceiptMetadata(r.Context(), a, clientID, baseNumber, receiptNo, &valid)
		if err != nil {
			switch {
			case errors.Is(err, invoiceTx.ErrInvoiceNotFound), errors.Is(err, invoiceTx.ErrPaymentReceiptNotFound):
				res.Error(w, http.StatusNotFound, "NOT_FOUND", "Payment receipt not found")
				return
			case errors.Is(err, invoiceTx.ErrInvoiceDraftForReceipt):
				res.Error(w, http.StatusConflict, "INVOICE_DRAFT", "Draft invoices do not support payment receipts")
				return
			case errors.Is(err, invoiceTx.ErrInvoiceVoidForReceipt):
				res.Error(w, http.StatusConflict, "INVOICE_VOID", "Invoice is void; payment receipts are not editable")
				return
			}

			slog.ErrorContext(r.Context(),
				"update payment receipt failed",
				"client_id", clientID,
				"base_number", baseNumber,
				"receipt_no", receiptNo,
				"err", err,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		res.JSON(w, http.StatusOK, map[string]any{
			"receiptId": receiptID,
			"receiptNo": receiptNo,
		})
	}
}
