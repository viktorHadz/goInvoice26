package products

import (
	"encoding/json"

	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/validate"
)

func ValidateCreate(in models.ProductCreateIn, routeClientID int64) (models.ProductCreate, []res.FieldError) {
	out := models.ProductCreate{ClientID: routeClientID}
	var errs []res.FieldError

	// strings
	// productType
	var productType string
	if in.ProductType != nil {
		productType = *in.ProductType
	}
	productType, fe := validate.Text(productType, validate.TextRules{
		Field: "productType", Required: true, Min: 4, Max: 20, SingleLine: true, Trim: true,
	})
	if len(fe) > 0 {
		errs = append(errs, fe...)
	} else {
		out.ProductType = productType
	}

	// pricingMode
	var pricingMode string
	if in.PricingMode != nil {
		pricingMode = *in.PricingMode
	}
	pricingMode, fe = validate.Text(pricingMode, validate.TextRules{
		Field: "pricingMode", Required: true, Min: 3, Max: 20, SingleLine: true, Trim: true,
	})
	if len(fe) > 0 {
		errs = append(errs, fe...)
	} else {
		out.PricingMode = pricingMode
	}

	// productName
	var productName string
	if in.ProductName != nil {
		productName = *in.ProductName
	}
	productName, fe = validate.Text(productName, validate.TextRules{
		Field: "productName", Required: true, Min: 2, Max: 80, SingleLine: true, Trim: true,
	})
	if len(fe) > 0 {
		errs = append(errs, fe...)
	} else {
		out.ProductName = productName
	}

	// productType / pricingMode
	// style => flat only
	// sampleFlat => flat only
	// sampleHourly => hourly only
	switch out.ProductType {
	case "style", "sampleFlat":
		if out.PricingMode != "" && out.PricingMode != "flat" {
			errs = append(errs, res.Invalid("pricingMode", "must be 'flat' for this productType"))
		}
	case "sampleHourly":
		if out.PricingMode != "" && out.PricingMode != "hourly" {
			errs = append(errs, res.Invalid("pricingMode", "must be 'hourly' for this productType"))
		}
	default:
		if out.ProductType != "" {
			errs = append(errs, res.Invalid("productType", "invalid value"))
		}
	}

	// numeric helpers
	min0 := int64(0)

	parseMoneyRequired := func(field string, n *json.Number) *int64 {
		if n == nil {
			errs = append(errs, res.Required(field))
			return nil
		}
		v, fe := validate.MoneyMinor(n.String(), validate.MoneyRules{
			Field: field, Required: true, MinMinor: &min0, Trim: true,
		})
		if len(fe) > 0 {
			errs = append(errs, fe...)
			return nil
		}
		return &v
	}

	parseIntRequired := func(field string, n *json.Number) *int64 {
		if n == nil {
			errs = append(errs, res.Required(field))
			return nil
		}
		v, fe := validate.Int64(n.String(), validate.IntRules{
			Field: field, Required: true, Min: &min0, Trim: true,
		})
		if len(fe) > 0 {
			errs = append(errs, fe...)
			return nil
		}
		return &v
	}

	// pricingMode decides required numeric fields/disallowed extras
	switch out.PricingMode {
	case "flat":
		out.FlatPriceMinor = parseMoneyRequired("flatPrice", in.FlatPrice)

		// strict: reject extras
		if in.HourlyRate != nil {
			errs = append(errs, res.Invalid("hourlyRate", "not allowed for flat pricing"))
		}
		if in.MinutesWorked != nil {
			errs = append(errs, res.Invalid("minutesWorked", "not allowed for flat pricing"))
		}

	case "hourly":
		out.HourlyRateMinor = parseMoneyRequired("hourlyRate", in.HourlyRate)
		out.MinutesWorked = parseIntRequired("minutesWorked", in.MinutesWorked)

		// strictly reject extras
		if in.FlatPrice != nil {
			errs = append(errs, res.Invalid("flatPrice", "not allowed for hourly pricing"))
		}

	default:
		// pricingMode invalid already flagged above nothing else to do
	}

	return out, errs
}
