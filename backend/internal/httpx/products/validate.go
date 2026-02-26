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

	// ----- productType -----
	productType := ""
	if in.ProductType != nil {
		productType = *in.ProductType
	}
	productType, fe := validate.Text(productType, validate.TextRules{
		Field: "productType", Required: true, Min: 4, Max: 20, SingleLine: true, Trim: true,
	})
	if len(fe) > 0 {
		errs = append(errs, fe...)
	} else {
		if productType != "style" && productType != "sample" {
			errs = append(errs, res.Invalid("productType", "invalid value"))
		} else {
			out.ProductType = productType
		}
	}

	// ----- pricingMode -----
	pricingMode := ""
	if in.PricingMode != nil {
		pricingMode = *in.PricingMode
	}
	pricingMode, fe = validate.Text(pricingMode, validate.TextRules{
		Field: "pricingMode", Required: true, Min: 3, Max: 20, SingleLine: true, Trim: true,
	})
	if len(fe) > 0 {
		errs = append(errs, fe...)
	} else {
		if pricingMode != "flat" && pricingMode != "hourly" {
			errs = append(errs, res.Invalid("pricingMode", "invalid value"))
		} else {
			out.PricingMode = pricingMode
		}
	}

	productName := ""
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

	if out.ProductType == "style" && out.PricingMode == "hourly" {
		errs = append(errs, res.Invalid("pricingMode", "must be 'flat' for style"))
	}

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

	// pricingMode decides required number fields and dissallowed combos
	switch out.PricingMode {
	case "flat":
		out.FlatPriceMinor = parseMoneyRequired("flatPrice", in.FlatPrice)

		if in.HourlyRate != nil {
			errs = append(errs, res.Invalid("hourlyRate", "not allowed for flat pricing"))
		}
		if in.MinutesWorked != nil {
			errs = append(errs, res.Invalid("minutesWorked", "not allowed for flat pricing"))
		}

	case "hourly":
		out.HourlyRateMinor = parseMoneyRequired("hourlyRate", in.HourlyRate)
		out.MinutesWorked = parseIntRequired("minutesWorked", in.MinutesWorked)

		if in.FlatPrice != nil {
			errs = append(errs, res.Invalid("flatPrice", "not allowed for hourly pricing"))
		}
	}

	return out, errs
}
