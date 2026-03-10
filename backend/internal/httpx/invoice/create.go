package invoice

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func createInvoice(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIDParam := chi.URLParam(r, "clientID")
		baseNumberParam := chi.URLParam(r, "baseNumber")

		clientID, err := strconv.ParseInt(clientIDParam, 10, 64)
		if err != nil || clientID < 1 {
			res.Error(w, res.Validation(res.Invalid("clientId", "invalid route param")))
			return
		}

		baseNumber, err := strconv.ParseInt(baseNumberParam, 10, 64)
		if err != nil || baseNumber < 1 {
			res.Error(w, res.Validation(res.Invalid("baseNumber", "invalid route param")))
			return
		}

		var dtoInvoice models.FEInvoiceIn
		if ok := res.DecodeJSON(w, r, &dtoInvoice); !ok {
			slog.WarnContext(r.Context(), "Received bad JSON for invoice",
				"json", &dtoInvoice,
			)
			return
		}

		slog.Debug("Invoice received from FE", "inv", &dtoInvoice)

		// Route param consistency (prevents mismatched path/body invoices)
		var routeErrs []res.FieldError
		if dtoInvoice.Overview.ClientID != clientID {
			routeErrs = append(routeErrs, res.Invalid("clientId", "does not match route param"))
		}
		if dtoInvoice.Overview.BaseNumber != baseNumber {
			routeErrs = append(routeErrs, res.Invalid("baseNumber", "does not match route param"))
		}
		if len(routeErrs) > 0 {
			res.Error(w, res.Validation(routeErrs...))
			return
		}

		// validate received invoice
		validInvoice, errs := ValidateInvoiceCreate(dtoInvoice)
		if len(errs) > 0 {
			res.Error(w, res.Validation(errs...))
			return
		}

		canonical := RecalcInvoice(validInvoice)
		slog.Debug("Canonical invoice ready", "inv", canonical)

		// Recalculated totals must match frontend-submitted totals (reject tampering).
		if errs := verifyTotalsMatch(validInvoice.Totals, canonical.Totals); len(errs) > 0 {
			res.Error(w, res.Validation(errs...))
			return
		}

		invID, revID, err := invoiceTx.Create(r.Context(), a, &canonical)
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				res.Error(w, res.Validation(res.Invalid("baseNumber", "invoice number already in use")))
				return
			}
			res.Error(w, res.Database(err))
			return
		}

		res.JSON(w, http.StatusCreated, map[string]int64{
			"invoiceId":  invID,
			"revisionId": revID,
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
