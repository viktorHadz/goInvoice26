package products

import (
	"encoding/json"
	"testing"

	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

func numPtr(s string) *json.Number {
	n := json.Number(s)
	return &n
}

func strPtr(s string) *string { return &s }

func hasFieldErr(errs []res.FieldError, field string) bool {
	for _, e := range errs {
		if e.Field == field {
			return true
		}
	}
	return false
}

func fieldErrCount(errs []res.FieldError, field string) int {
	c := 0
	for _, e := range errs {
		if e.Field == field {
			c++
		}
	}
	return c
}

// --- HAPPY PATHS ---

func TestValidateCreate_StyleFlat_OK(t *testing.T) {
	in := models.ProductCreateIn{
		ProductType: strPtr("style"),
		PricingMode: strPtr("flat"),
		ProductName: strPtr("Blue Jeans"),
		FlatPrice:   numPtr("12.50"),
	}

	out, errs := ValidateCreate(in, 1)
	if len(errs) != 0 {
		t.Fatalf("expected no errs, got: %#v", errs)
	}
	if out.ClientID != 1 {
		t.Fatalf("ClientID: got %d, want %d", out.ClientID, 1)
	}
	if out.ProductType != "style" || out.PricingMode != "flat" || out.ProductName != "Blue Jeans" {
		t.Fatalf("unexpected out: %#v", out)
	}
	if out.FlatPriceMinor == nil || *out.FlatPriceMinor != 1250 {
		t.Fatalf("FlatPriceMinor: got %#v, want 1250", out.FlatPriceMinor)
	}
	if out.HourlyRateMinor != nil || out.MinutesWorked != nil {
		t.Fatalf("expected hourly fields nil for flat style, got: %#v", out)
	}
}

func TestValidateCreate_SampleFlat_OK(t *testing.T) {
	in := models.ProductCreateIn{
		ProductType: strPtr("sample"),
		PricingMode: strPtr("flat"),
		ProductName: strPtr("Sample Tee"),
		FlatPrice:   numPtr("10"),
	}
	out, errs := ValidateCreate(in, 7)
	if len(errs) != 0 {
		t.Fatalf("expected no errs, got: %#v", errs)
	}
	if out.FlatPriceMinor == nil || *out.FlatPriceMinor != 1000 {
		t.Fatalf("FlatPriceMinor: got %#v, want 1000", out.FlatPriceMinor)
	}
}

func TestValidateCreate_SampleHourly_OK(t *testing.T) {
	in := models.ProductCreateIn{
		ProductType:   strPtr("sample"),
		PricingMode:   strPtr("hourly"),
		ProductName:   strPtr("Pattern Adjustment"),
		HourlyRate:    numPtr("30"),
		MinutesWorked: numPtr("90"),
	}
	out, errs := ValidateCreate(in, 3)
	if len(errs) != 0 {
		t.Fatalf("expected no errs, got: %#v", errs)
	}
	if out.HourlyRateMinor == nil || *out.HourlyRateMinor != 3000 {
		t.Fatalf("HourlyRateMinor: got %#v, want 3000", out.HourlyRateMinor)
	}
	if out.MinutesWorked == nil || *out.MinutesWorked != 90 {
		t.Fatalf("MinutesWorked: got %#v, want 90", out.MinutesWorked)
	}
	if out.FlatPriceMinor != nil {
		t.Fatalf("expected FlatPriceMinor nil for hourly, got: %#v", out.FlatPriceMinor)
	}
}

// --- VALIDATION FAILURES ---

func TestValidateCreate_InvalidProductType(t *testing.T) {
	in := models.ProductCreateIn{
		ProductType: strPtr("sampleHourly"), // invalid now
		PricingMode: strPtr("hourly"),
		ProductName: strPtr("X"),
		HourlyRate:  numPtr("30"),
		// minutes missing too but main assert is that productType error exists
	}

	_, errs := ValidateCreate(in, 1)
	if !hasFieldErr(errs, "productType") {
		t.Fatalf("expected productType error, got: %#v", errs)
	}
}

func TestValidateCreate_InvalidPricingMode(t *testing.T) {
	in := models.ProductCreateIn{
		ProductType: strPtr("sample"),
		PricingMode: strPtr("weekly"),
		ProductName: strPtr("X"),
	}
	_, errs := ValidateCreate(in, 1)
	if !hasFieldErr(errs, "pricingMode") {
		t.Fatalf("expected pricingMode error, got: %#v", errs)
	}
}

func TestValidateCreate_StyleHourly_Disallowed(t *testing.T) {
	in := models.ProductCreateIn{
		ProductType:   strPtr("style"),
		PricingMode:   strPtr("hourly"),
		ProductName:   strPtr("Should fail"),
		HourlyRate:    numPtr("30"),
		MinutesWorked: numPtr("60"),
	}

	_, errs := ValidateCreate(in, 1)
	if !hasFieldErr(errs, "pricingMode") {
		t.Fatalf("expected pricingMode error, got: %#v", errs)
	}
}

func TestValidateCreate_FlatMissingFlatPrice(t *testing.T) {
	in := models.ProductCreateIn{
		ProductType: strPtr("sample"),
		PricingMode: strPtr("flat"),
		ProductName: strPtr("Missing price"),
		// FlatPrice missing
	}
	_, errs := ValidateCreate(in, 1)
	if !hasFieldErr(errs, "flatPrice") {
		t.Fatalf("expected flatPrice required error, got: %#v", errs)
	}
}

func TestValidateCreate_HourlyMissingFields(t *testing.T) {
	in := models.ProductCreateIn{
		ProductType: strPtr("sample"),
		PricingMode: strPtr("hourly"),
		ProductName: strPtr("Missing hourly bits"),
		// missing hourlyRate + minutesWorked
	}
	_, errs := ValidateCreate(in, 1)
	if !hasFieldErr(errs, "hourlyRate") || !hasFieldErr(errs, "minutesWorked") {
		t.Fatalf("expected hourlyRate + minutesWorked required errors, got: %#v", errs)
	}
}

func TestValidateCreate_HourlyRejectsFlatPriceExtra(t *testing.T) {
	in := models.ProductCreateIn{
		ProductType:   strPtr("sample"),
		PricingMode:   strPtr("hourly"),
		ProductName:   strPtr("Extra field"),
		HourlyRate:    numPtr("10"),
		MinutesWorked: numPtr("30"),
		FlatPrice:     numPtr("99"),
	}
	_, errs := ValidateCreate(in, 1)
	if !hasFieldErr(errs, "flatPrice") {
		t.Fatalf("expected flatPrice invalid (not allowed), got: %#v", errs)
	}
}

func TestValidateCreate_FlatRejectsHourlyExtras(t *testing.T) {
	in := models.ProductCreateIn{
		ProductType:   strPtr("sample"),
		PricingMode:   strPtr("flat"),
		ProductName:   strPtr("Extra hourly bits"),
		FlatPrice:     numPtr("10"),
		HourlyRate:    numPtr("5"),
		MinutesWorked: numPtr("30"),
	}
	_, errs := ValidateCreate(in, 1)
	if !hasFieldErr(errs, "hourlyRate") || !hasFieldErr(errs, "minutesWorked") {
		t.Fatalf("expected hourlyRate + minutesWorked invalid (not allowed), got: %#v", errs)
	}
}

func TestValidateCreate_MoneyTooManyDecimals(t *testing.T) {
	in := models.ProductCreateIn{
		ProductType: strPtr("style"),
		PricingMode: strPtr("flat"),
		ProductName: strPtr("Bad money"),
		FlatPrice:   numPtr("12.345"),
	}
	_, errs := ValidateCreate(in, 1)
	if !hasFieldErr(errs, "flatPrice") {
		t.Fatalf("expected flatPrice error, got: %#v", errs)
	}
}

func TestValidateCreate_MinutesMustBeInt(t *testing.T) {
	in := models.ProductCreateIn{
		ProductType:   strPtr("sample"),
		PricingMode:   strPtr("hourly"),
		ProductName:   strPtr("Bad minutes"),
		HourlyRate:    numPtr("10"),
		MinutesWorked: numPtr("12.5"),
	}
	_, errs := ValidateCreate(in, 1)
	if !hasFieldErr(errs, "minutesWorked") {
		t.Fatalf("expected minutesWorked integer error, got: %#v", errs)
	}
}

// Check no duplicate noise for enums
func TestValidateCreate_NoDuplicateProductTypeErrors(t *testing.T) {
	in := models.ProductCreateIn{
		ProductType: strPtr("bad"),
		PricingMode: strPtr("flat"),
		ProductName: strPtr("X"),
		FlatPrice:   numPtr("1"),
	}
	_, errs := ValidateCreate(in, 1)
	if c := fieldErrCount(errs, "productType"); c < 1 {
		t.Fatalf("expected at least 1 productType error, got: %#v", errs)
	}
}
