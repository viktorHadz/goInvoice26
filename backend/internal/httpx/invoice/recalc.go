package invoice

import (
	"math"

	"github.com/viktorHadz/goInvoice26/internal/models"
)

func RecalcInvoice(inv models.FEInvoiceIn) models.FEInvoiceIn {
	out := inv

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

	vatBps := clamp(out.Totals.VATRate, 0, 10000)

	discountRate := clamp(out.Totals.DiscountRate, 0, 10000)
	depositRate := clamp(out.Totals.DepositRate, 0, 10000)
	paidMinor := max(out.Totals.PaidMinor, 0)

	var discountMinor int64
	switch out.Totals.DiscountType {
	case "none":
		discountMinor = 0
		discountRate = 0
	case "percent":
		discountMinor = int64(math.Round(float64(subtotal*discountRate) / 10000.0))
	case "fixed":
		discountMinor = max(out.Totals.DiscountMinor, 0)
		discountRate = 0
	default:
		out.Totals.DiscountType = "none"
		discountMinor = 0
		discountRate = 0
	}
	discountMinor = clamp(discountMinor, 0, subtotal)

	subAfterDisc := max(subtotal-discountMinor, 0)
	vatMinor := max(int64(math.Round(float64(subAfterDisc*vatBps)/10000.0)), 0)
	totalMinor := max(subAfterDisc+vatMinor, 0)

	var depositMinor int64
	switch out.Totals.DepositType {
	case "none":
		depositMinor = 0
		depositRate = 0
	case "percent":
		depositMinor = int64(math.Round(float64(totalMinor*depositRate) / 10000.0))
	case "fixed":
		depositMinor = max(out.Totals.DepositMinor, 0)
		depositRate = 0
	default:
		out.Totals.DepositType = "none"
		depositMinor = 0
		depositRate = 0
	}
	depositMinor = clamp(depositMinor, 0, totalMinor)

	balanceDue := max(totalMinor-depositMinor-paidMinor, 0)

	// !CRITICAL: This logic must stay identical to the frontend invoice recalculation.
	// Any change here MUST be mirrored in the frontend to avoid total drift.
	// Totals (mirrors current frontend DTO meaning: discount/deposit minors are absolute amounts)
	out.Totals.SubtotalMinor = subtotal
	out.Totals.SubtotalAfterDisc = subAfterDisc
	out.Totals.VatAmountMinor = vatMinor
	out.Totals.TotalMinor = totalMinor
	out.Totals.BalanceDue = balanceDue

	out.Totals.VATRate = vatBps
	out.Totals.PaidMinor = paidMinor

	out.Totals.DiscountRate = discountRate
	out.Totals.DiscountMinor = discountMinor

	out.Totals.DepositRate = depositRate
	out.Totals.DepositMinor = depositMinor

	return out
}

func clamp(v, minV, maxV int64) int64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
