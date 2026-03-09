package invoice

import (
	"log/slog"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

type verifyResponse struct {
	Invoice models.FEInvoiceIn `json:"invoice"`
}

func verifyInvoice(a *app.App) http.HandlerFunc {
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

		var invoice models.FEInvoiceIn
		if ok := res.DecodeJSON(w, r, &invoice); !ok {
			return
		}

		// Route/body consistency
		var routeErrs []res.FieldError
		if invoice.Overview.ClientID != clientID {
			routeErrs = append(routeErrs, res.Invalid("clientId", "does not match route param"))
		}
		if invoice.Overview.BaseNumber != baseNumber {
			routeErrs = append(routeErrs, res.Invalid("baseNumber", "does not match route param"))
		}
		if len(routeErrs) > 0 {
			res.Error(w, res.Validation(routeErrs...))
			return
		}

		validInvoice, errs := ValidateInvoiceCreate(invoice)
		if len(errs) > 0 {
			res.Error(w, res.Validation(errs...))
			return
		}

		canonical := recalcInvoice(validInvoice)
		slog.Debug("invoice verified", "clientID", clientID, "baseNumber", baseNumber)

		res.JSON(w, http.StatusOK, verifyResponse{Invoice: canonical})
	}
}

func recalcInvoice(inv models.FEInvoiceIn) models.FEInvoiceIn {
	out := inv

	// Line totals
	var subtotal int64
	for i := range out.Lines {
		ln := out.Lines[i]
		var lt int64
		if ln.PricingMode == "hourly" && ln.MinutesWorked != nil {
			lt = int64(math.Round(
				(float64(ln.Quantity) * float64(ln.UnitPriceMinor) * float64(*ln.MinutesWorked)) / 60.0,
			))
		} else {
			lt = ln.Quantity * ln.UnitPriceMinor
		}
		if lt < 0 {
			lt = 0
		}
		out.Lines[i].LineTotalMinor = lt
		subtotal += lt
	}

	// Totals (mirrors current frontend DTO meaning: discount/deposit minors are absolute amounts)
	discountMinor := out.Totals.DiscountMinor
	if discountMinor < 0 {
		discountMinor = 0
	}
	if discountMinor > subtotal {
		discountMinor = subtotal
	}

	subAfterDisc := subtotal - discountMinor
	if subAfterDisc < 0 {
		subAfterDisc = 0
	}

	vatBps := out.Totals.VATRate
	if vatBps < 0 {
		vatBps = 0
	}
	if vatBps > 10000 {
		vatBps = 10000
	}

	vatMinor := int64(math.Round((float64(subAfterDisc) * float64(vatBps)) / 10000.0))
	if vatMinor < 0 {
		vatMinor = 0
	}

	totalMinor := subAfterDisc + vatMinor
	if totalMinor < 0 {
		totalMinor = 0
	}

	depositMinor := out.Totals.DepositMinor
	if depositMinor < 0 {
		depositMinor = 0
	}
	if depositMinor > totalMinor {
		depositMinor = totalMinor
	}

	paidMinor := out.Totals.PaidMinor
	if paidMinor < 0 {
		paidMinor = 0
	}

	balanceDue := totalMinor - depositMinor - paidMinor
	if balanceDue < 0 {
		balanceDue = 0
	}

	out.Totals.SubtotalMinor = subtotal
	out.Totals.SubtotalAfterDisc = subAfterDisc
	out.Totals.VatAmountMinor = vatMinor
	out.Totals.TotalMinor = totalMinor
	out.Totals.BalanceDue = balanceDue

	out.Totals.DepositMinor = depositMinor
	out.Totals.DiscountMinor = discountMinor
	out.Totals.PaidMinor = paidMinor
	out.Totals.VATRate = vatBps

	return out
}

