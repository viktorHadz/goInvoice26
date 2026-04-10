package invoice

import (
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/validate"
)

// Validates the 3 parts of FEInvoiceIn, and returns a validated FEInvoiceIn
// - On error returns FEInvoiceIn{}, []FieldError
func ValidateInvoiceCreate(inv models.FEInvoiceIn) (models.FEInvoiceIn, []res.FieldError) {
	slog.Debug(
		"VALIDATE_INVOICE_CREATE",
		"invoice", inv,
	)

	var errors []res.FieldError
	validated := models.FEInvoiceIn{}

	overview, errs := validateOverview(inv.Overview)
	if len(errs) > 0 {
		errors = append(errors, errs...)
	}
	validated.Overview = overview

	lines, errs := validateLines(inv.Lines)
	if len(errs) > 0 {
		errors = append(errors, errs...)
	}
	validated.Lines = lines

	totals, errs := validateTotals(inv.Totals)
	if len(errs) > 0 {
		errors = append(errors, errs...)
	}
	validated.Totals = totals

	if len(errors) > 0 {
		return models.FEInvoiceIn{}, errors
	}

	return validated, nil
}

func validateOverview(o models.InvoiceCreateIn) (models.InvoiceCreateIn, []res.FieldError) {
	var out models.InvoiceCreateIn
	var errs []res.FieldError

	// numeric required
	if o.ClientID < 1 {
		errs = append(errs, res.Invalid("clientId", "must be greater than 0"))
	} else {
		out.ClientID = o.ClientID
	}

	if o.BaseNumber < 1 {
		errs = append(errs, res.Invalid("baseNumber", "must be greater than 0"))
	} else {
		out.BaseNumber = o.BaseNumber
	}
	if o.SourceRevisionNo != nil {
		if *o.SourceRevisionNo < 1 {
			errs = append(errs, res.Invalid("sourceRevisionNo", "must be greater than 0"))
		} else {
			source := *o.SourceRevisionNo
			out.SourceRevisionNo = &source
		}
	}

	// dates
	issueDate, dateErrs := validateISODateRequired("issueDate", o.IssueDate)
	errs = append(errs, dateErrs...)
	out.IssueDate = issueDate
	if o.SupplyDate != nil {
		supply, dateErrs := validateISODateOptional("supplyDate", o.SupplyDate)
		errs = append(errs, dateErrs...)
		if supply != nil && *supply != issueDate {
			out.SupplyDate = supply
		}
	}

	if o.DueByDate != nil {
		due, dateErrs := validateISODateOptional("dueByDate", o.DueByDate)
		errs = append(errs, dateErrs...)
		out.DueByDate = due
	}

	// text fields
	clientName, textErrs := validate.Text(o.ClientName, validate.TextRules{
		Field:      "clientName",
		Required:   true,
		Min:        0,
		Max:        100,
		SingleLine: true,
		Trim:       true,
	})
	errs = append(errs, textErrs...)
	out.ClientName = clientName

	clientCompanyName, textErrs := validate.Text(o.ClientCompanyName, validate.TextRules{
		Field:      "clientCompanyName",
		Required:   false,
		Min:        0,
		Max:        100,
		SingleLine: true,
		Trim:       true,
	})
	errs = append(errs, textErrs...)
	out.ClientCompanyName = clientCompanyName

	clientAddress, textErrs := validate.Text(o.ClientAddress, validate.TextRules{
		Field:      "clientAddress",
		Required:   false,
		Min:        0,
		Max:        200,
		SingleLine: true,
		Trim:       true,
	})
	errs = append(errs, textErrs...)
	out.ClientAddress = clientAddress

	clientEmail, emailErrs := validate.Email("clientEmail", o.ClientEmail, 100)
	errs = append(errs, emailErrs...)
	out.ClientEmail = clientEmail

	if o.Note != nil {
		note, textErrs := validate.Text(*o.Note, validate.TextRules{
			Field:      "note",
			Required:   false,
			Min:        0,
			Max:        1000,
			SingleLine: true,
			Trim:       true,
		})
		errs = append(errs, textErrs...)
		out.Note = &note
	}

	return out, errs
}

func validateLines(lines []models.LineCreateIn) ([]models.LineCreateIn, []res.FieldError) {
	var out []models.LineCreateIn
	var errs []res.FieldError

	if len(lines) == 0 {
		return nil, []res.FieldError{
			res.Invalid("lines", "must contain at least one item"),
		}
	}

	for i, ln := range lines {
		var clean models.LineCreateIn
		prefix := func(field string) string { return fmt.Sprintf("lines[%d].%s", i, field) }

		// productId
		if ln.ProductID != nil {
			if *ln.ProductID < 1 {
				errs = append(errs, res.Invalid(prefix("productId"), "must be greater than 0"))
			} else {
				id := *ln.ProductID
				clean.ProductID = &id
			}
		}

		// name
		name, textErrs := validate.Text(ln.Name, validate.TextRules{
			Field:      prefix("name"),
			Required:   true,
			Min:        1,
			Max:        200,
			SingleLine: true,
			Trim:       true,
		})
		errs = append(errs, textErrs...)
		clean.Name = name

		// enums
		lineType := strings.TrimSpace(ln.LineType)
		switch lineType {
		case "custom", "style", "sample":
			clean.LineType = lineType
		default:
			errs = append(errs, res.Invalid(prefix("lineType"), "must be one of: custom, style, sample"))
		}

		pricingMode := strings.TrimSpace(ln.PricingMode)
		switch pricingMode {
		case "flat", "hourly":
			clean.PricingMode = pricingMode
		default:
			errs = append(errs, res.Invalid(prefix("pricingMode"), "must be one of: flat, hourly"))
		}

		// quantity
		if ln.Quantity < 1 {
			errs = append(errs, res.Invalid(prefix("quantity"), "must be greater than 0"))
		} else {
			clean.Quantity = ln.Quantity
		}

		// minutesWorked
		if ln.MinutesWorked != nil {
			if *ln.MinutesWorked < 0 {
				errs = append(errs, res.Invalid(prefix("minutesWorked"), "must be 0 or greater"))
			} else {
				m := *ln.MinutesWorked
				clean.MinutesWorked = &m
			}
		}

		// unitPriceMinor
		if ln.UnitPriceMinor < 0 {
			errs = append(errs, res.Invalid(prefix("unitPriceMinor"), "must be 0 or greater"))
		} else {
			clean.UnitPriceMinor = ln.UnitPriceMinor
		}

		// sortOrder
		if ln.SortOrder < 1 {
			errs = append(errs, res.Invalid(prefix("sortOrder"), "must be 1 or greater"))
		} else {
			clean.SortOrder = ln.SortOrder
		}

		// cross-field rules
		switch pricingMode {
		case "hourly":
			if ln.MinutesWorked == nil {
				errs = append(errs, res.Required(prefix("minutesWorked")))
			}
		case "flat":
			// DB CHECK requires minutes_worked IS NULL for flat lines
			if ln.MinutesWorked != nil {
				errs = append(errs, res.Invalid(prefix("minutesWorked"), "must be null for flat pricing"))
			}
		}

		var expectedLineTotal int64
		if pricingMode == "hourly" && ln.MinutesWorked != nil {
			// Match frontend: round(qty * unit * minutes / 60)
			expectedLineTotal = int64(math.Round(
				(float64(ln.Quantity) * float64(ln.UnitPriceMinor) * float64(*ln.MinutesWorked)) / 60.0,
			))
		} else {
			expectedLineTotal = ln.Quantity * ln.UnitPriceMinor
		}
		if ln.LineTotalMinor < 0 {
			errs = append(errs, res.Invalid(prefix("lineTotalMinor"), "must be 0 or greater"))
		} else if ln.LineTotalMinor != expectedLineTotal {
			if pricingMode == "hourly" {
				errs = append(errs, res.Invalid(prefix("lineTotalMinor"), "does not match rounded(quantity * unitPriceMinor * minutesWorked / 60)"))
			} else {
				errs = append(errs, res.Invalid(prefix("lineTotalMinor"), "does not match quantity * unitPriceMinor"))
			}
		} else {
			clean.LineTotalMinor = ln.LineTotalMinor
		}

		out = append(out, clean)
	}

	return out, errs
}

func validateTotals(t models.TotalsCreateIn) (models.TotalsCreateIn, []res.FieldError) {
	var out models.TotalsCreateIn
	var errs []res.FieldError

	// VAT rate - basis points
	if t.VATRate < 0 || t.VATRate > 10000 {
		errs = append(errs, res.Invalid("totals.vatRate", "must be between 0 and 10000"))
	} else {
		out.VATRate = t.VATRate
	}

	if t.VatAmountMinor < 0 {
		errs = append(errs, res.Invalid("totals.vatMinor", "must be 0 or greater"))
	} else {
		out.VatAmountMinor = t.VatAmountMinor
	}

	switch strings.TrimSpace(t.DepositType) {
	case "none", "percent", "fixed":
		out.DepositType = strings.TrimSpace(t.DepositType)
	default:
		errs = append(errs, res.Invalid("totals.depositType", "must be one of: none, percent, fixed"))
	}

	switch strings.TrimSpace(t.DiscountType) {
	case "none", "percent", "fixed":
		out.DiscountType = strings.TrimSpace(t.DiscountType)
	default:
		errs = append(errs, res.Invalid("totals.discountType", "must be one of: none, percent, fixed"))
	}

	if t.DepositRate < 0 || t.DepositRate > 10000 {
		errs = append(errs, res.Invalid("totals.depositRate", "must be between 0 and 10000"))
	} else {
		out.DepositRate = t.DepositRate
	}
	if t.DepositMinor < 0 {
		errs = append(errs, res.Invalid("totals.depositMinor", "must be 0 or greater"))
	} else {
		out.DepositMinor = t.DepositMinor
	}

	if t.DiscountRate < 0 || t.DiscountRate > 10000 {
		errs = append(errs, res.Invalid("totals.discountRate", "must be between 0 and 10000"))
	} else {
		out.DiscountRate = t.DiscountRate
	}
	if t.DiscountMinor < 0 {
		errs = append(errs, res.Invalid("totals.discountMinor", "must be 0 or greater"))
	} else {
		out.DiscountMinor = t.DiscountMinor
	}

	if out.DepositType != "percent" && out.DepositRate != 0 {
		errs = append(errs, res.Invalid("totals.depositRate", "must be 0 unless depositType is percent"))
	}

	if out.DiscountType != "percent" && out.DiscountRate != 0 {
		errs = append(errs, res.Invalid("totals.discountRate", "must be 0 unless discountType is percent"))
	}

	if t.PaidMinor < 0 {
		errs = append(errs, res.Invalid("totals.paidMinor", "must be 0 or greater"))
	} else {
		out.PaidMinor = t.PaidMinor
	}

	if t.SubtotalAfterDisc < 0 {
		errs = append(errs, res.Invalid("totals.subtotalAfterDiscountMinor", "must be 0 or greater"))
	} else {
		out.SubtotalAfterDisc = t.SubtotalAfterDisc
	}

	if t.SubtotalMinor < 0 {
		errs = append(errs, res.Invalid("totals.subtotalMinor", "must be 0 or greater"))
	} else {
		out.SubtotalMinor = t.SubtotalMinor
	}

	if t.TotalMinor < 0 {
		errs = append(errs, res.Invalid("totals.totalMinor", "must be 0 or greater"))
	} else {
		out.TotalMinor = t.TotalMinor
	}

	if t.BalanceDue < 0 {
		errs = append(errs, res.Invalid("totals.balanceDueMinor", "must be 0 or greater"))
	} else {
		out.BalanceDue = t.BalanceDue
	}

	return out, errs
}

// ValidatePaidVsDepositTotal ensures paidMinor does not exceed totalMinor (post-recalc canonical totals).
func ValidatePaidVsDepositTotal(t models.TotalsCreateIn) []res.FieldError {
	maxPaid := max(t.TotalMinor, 0)
	if t.PaidMinor > maxPaid {
		return []res.FieldError{res.Invalid("totals.paidMinor", "cannot exceed invoice total")}
	}
	return nil
}

func ValidatePaymentReceiptCreate(in models.PaymentReceiptCreateIn) (models.PaymentReceiptCreateIn, []res.FieldError) {
	var out models.PaymentReceiptCreateIn
	var errs []res.FieldError

	if in.AmountMinor <= 0 {
		errs = append(errs, res.Invalid("amountMinor", "must be greater than 0"))
	} else {
		out.AmountMinor = in.AmountMinor
	}

	paymentDate, dateErrs := validateISODateRequired("paymentDate", in.PaymentDate)
	errs = append(errs, dateErrs...)
	out.PaymentDate = paymentDate

	label, labelErrs := validateOptionalPaymentReceiptLabel(in.Label, "label")
	errs = append(errs, labelErrs...)
	out.Label = label

	return out, errs
}

func ValidatePaymentReceiptUpdate(in models.PaymentReceiptUpdateIn) (models.PaymentReceiptUpdateIn, []res.FieldError) {
	var out models.PaymentReceiptUpdateIn
	var errs []res.FieldError

	paymentDate, dateErrs := validateISODateRequired("paymentDate", in.PaymentDate)
	errs = append(errs, dateErrs...)
	out.PaymentDate = paymentDate

	label, labelErrs := validateOptionalPaymentReceiptLabel(in.Label, "label")
	errs = append(errs, labelErrs...)
	out.Label = label

	return out, errs
}

func validateOptionalPaymentReceiptLabel(value *string, field string) (*string, []res.FieldError) {
	if value == nil {
		return nil, nil
	}

	label, textErrs := validate.Text(*value, validate.TextRules{
		Field:      field,
		Required:   false,
		Min:        0,
		Max:        120,
		SingleLine: true,
		Trim:       true,
	})
	if len(textErrs) > 0 {
		return nil, textErrs
	}

	return &label, nil
}

func validateISODateRequired(field, value string) (string, []res.FieldError) {
	value = strings.TrimSpace(value)
	if value == "" {
		return value, []res.FieldError{res.Required(field)}
	}
	if _, err := time.Parse("2006-01-02", value); err != nil {
		return value, []res.FieldError{res.Invalid(field, "must be a valid ISO date (YYYY-MM-DD)")}
	}
	return value, nil
}

func validateISODateOptional(field string, value *string) (*string, []res.FieldError) {
	if value == nil {
		return nil, nil
	}

	v := strings.TrimSpace(*value)
	if v == "" {
		return nil, nil
	}

	if _, err := time.Parse("2006-01-02", v); err != nil {
		return value, []res.FieldError{res.Invalid(field, "must be a valid ISO date (YYYY-MM-DD)")}
	}

	return &v, nil
}
