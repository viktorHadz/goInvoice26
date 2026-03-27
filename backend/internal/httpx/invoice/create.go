package invoice

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func CreateInvoice(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// return valid, true
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
			slog.WarnContext(r.Context(), "Received bad JSON for invoice",
				"json", &dtoInvoice,
			)
			return
		}

		slog.DebugContext(r.Context(), "invoice received from FE", "inv", &dtoInvoice)

		// Route param consistency - prevents mismatched path/body invoices
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

		// validate received invoice
		validInvoice, errs := ValidateInvoiceCreate(dtoInvoice)
		if len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}

		canonical := RecalcInvoice(validInvoice)
		slog.DebugContext(r.Context(), "canonical invoice ready", "inv", canonical)

		// Reject tampering - server totals must match FE totals
		if errs := verifyTotalsMatch(validInvoice.Totals, canonical.Totals); len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}
		if errs := ValidatePaidVsDepositTotal(canonical.Totals); len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}

		invID, revID, err := invoiceTx.Create(r.Context(), a, &canonical)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				res.Validation(w, res.Invalid("baseNumber", "invoice number already in use"))
				return
			}
			if errors.Is(err, invoiceTx.ErrPaymentTotalsMismatch) {
				res.Validation(w, res.Invalid("totals.paidMinor", "must match visible payments plus staged payments"))
				return
			}

			slog.ErrorContext(r.Context(),
				"create invoice failed",
				"client_id", clientID,
				"base_number", baseNumber,
				"err", err,
			)

			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		res.JSON(w, http.StatusCreated, map[string]any{
			"invoiceId":  invID,
			"revisionId": revID,
		})
	}
}

func CreateRevision(a *app.App) http.HandlerFunc {
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
			slog.WarnContext(r.Context(), "received bad JSON for invoice revision",
				"json", &dtoInvoice,
			)
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

		invoiceID, revisionID, revisionNo, err := invoiceTx.CreateRevision(r.Context(), a, &canonical)
		if err != nil {
			if errors.Is(err, invoiceTx.ErrInvoiceNotFound) {
				res.Error(w, http.StatusNotFound, "NOT_FOUND", "Invoice not found")
				return
			}
			if errors.Is(err, invoiceTx.ErrInvoiceDraftForRevision) {
				res.Error(w, http.StatusConflict, "INVOICE_DRAFT", "Issue the draft before saving a revision")
				return
			}
			if errors.Is(err, invoiceTx.ErrInvoiceVoidForRevision) {
				res.Error(w, http.StatusConflict, "INVOICE_VOID", "Invoice is void; revisions are not allowed")
				return
			}
			if errors.Is(err, invoiceTx.ErrInvoicePaidForRevision) {
				res.Error(w, http.StatusConflict, "INVOICE_PAID", "Reopen invoice to issued before saving a revision")
				return
			}
			if errors.Is(err, invoiceTx.ErrSourceRevisionInvalid) {
				res.Validation(w, res.Invalid("sourceRevisionNo", "must reference an existing revision before the new revision"))
				return
			}
			if errors.Is(err, invoiceTx.ErrPaymentTotalsMismatch) {
				res.Validation(w, res.Invalid("totals.paidMinor", "must match payments visible at source revision plus staged payments"))
				return
			}

			slog.ErrorContext(r.Context(),
				"create invoice revision failed",
				"client_id", clientID,
				"base_number", baseNumber,
				"err", err,
			)

			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		res.JSON(w, http.StatusCreated, map[string]any{
			"invoiceId":  invoiceID,
			"revisionId": revisionID,
			"revisionNo": revisionNo,
		})
	}
}

// verifyTotalsMatch returns field errors if server-recalculated totals differ from submitted totals.
func verifyTotalsMatch(submitted, recalc models.TotalsCreateIn) []res.FieldError {
	var errs []res.FieldError
	if submitted.TotalMinor != recalc.TotalMinor {
		errs = append(errs, res.Invalid("totals.totalMinor", "does not match server calculation"))
	}
	if submitted.BalanceDue != recalc.BalanceDue {
		errs = append(errs, res.Invalid("totals.balanceDueMinor", "does not match server calculation"))
	}
	if submitted.SubtotalMinor != recalc.SubtotalMinor {
		errs = append(errs, res.Invalid("totals.subtotalMinor", "does not match server calculation"))
	}
	if submitted.VatAmountMinor != recalc.VatAmountMinor {
		errs = append(errs, res.Invalid("totals.vatMinor", "does not match server calculation"))
	}
	return errs
}
