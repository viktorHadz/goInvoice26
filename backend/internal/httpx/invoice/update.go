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

func UpdateInvoice(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}
		baseNumber, ok := params.ValidateParam(w, r, "baseNumber")
		if !ok {
			return
		}

		var dtoInvoice models.FEInvoiceIn
		if ok := res.DecodeJSON(w, r, &dtoInvoice); !ok {
			return
		}

		var routeErrs []res.FieldError
		if dtoInvoice.Overview.ClientID != clientID {
			routeErrs = append(routeErrs, res.Invalid("clientId", "does not match route parameter"))
		}
		if dtoInvoice.Overview.BaseNumber != baseNumber {
			routeErrs = append(routeErrs, res.Invalid("baseNumber", "does not match route parameter"))
		}
		if len(routeErrs) > 0 {
			res.Validation(w, routeErrs...)
			return
		}

		validInvoice, errs := ValidateInvoiceCreate(dtoInvoice)
		if len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}

		canonical := RecalcInvoice(validInvoice)

		if errs := verifyTotalsMatch(validInvoice.Totals, canonical.Totals); len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}
		if errs := ValidatePaidVsDepositTotal(canonical.Totals); len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}

		invoiceID, revisionID, err := invoiceTx.UpdateDraft(r.Context(), a, &canonical)
		if err != nil {
			switch {
			case errors.Is(err, invoiceTx.ErrInvoiceNotFound):
				res.Error(w, http.StatusNotFound, "NOT_FOUND", "Invoice not found")
				return
			case errors.Is(err, invoiceTx.ErrInvoiceIssuedForDraftUpdate):
				res.Error(w, http.StatusConflict, "INVOICE_ISSUED", "Issued invoices must be saved as revisions")
				return
			case errors.Is(err, invoiceTx.ErrInvoicePaidForDraftUpdate):
				res.Error(w, http.StatusConflict, "INVOICE_PAID", "Paid invoices cannot be edited")
				return
			case errors.Is(err, invoiceTx.ErrInvoiceVoidForDraftUpdate):
				res.Error(w, http.StatusConflict, "INVOICE_VOID", "Void invoices cannot be edited")
				return
			case errors.Is(err, invoiceTx.ErrDraftInvoiceHasRevisions):
				res.Error(w, http.StatusConflict, "DRAFT_HAS_REVISIONS", "Draft invoice has revisions and can no longer be updated in place")
				return
			case errors.Is(err, invoiceTx.ErrPaymentTotalsMismatch):
				res.Validation(w, res.Invalid("totals.paidMinor", "must match visible payments plus staged payments"))
				return
			}

			slog.ErrorContext(r.Context(),
				"update draft invoice failed",
				"client_id", clientID,
				"base_number", baseNumber,
				"err", err,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		res.JSON(w, http.StatusOK, map[string]any{
			"invoiceId":  invoiceID,
			"revisionId": revisionID,
		})
	}
}
