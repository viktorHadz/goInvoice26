package invoice

import (
	"testing"

	"github.com/viktorHadz/goInvoice26/internal/models"
)

func validInvoiceInput() models.FEInvoiceIn {
	supplyDate := "2026-03-24"
	return models.FEInvoiceIn{
		Overview: models.InvoiceCreateIn{
			ClientID:          1,
			BaseNumber:        1,
			IssueDate:         "2026-03-23",
			SupplyDate:        &supplyDate,
			DueByDate:         nil,
			ClientName:        "Client",
			ClientCompanyName: "Company",
			ClientAddress:     "Address",
			ClientEmail:       "client@example.com",
			Note:              nil,
		},
		Lines: []models.LineCreateIn{
			{
				ProductID:      nil,
				Name:           "Line",
				LineType:       "custom",
				PricingMode:    "flat",
				Quantity:       1,
				MinutesWorked:  nil,
				UnitPriceMinor: 10000,
				LineTotalMinor: 10000,
				SortOrder:      1,
			},
		},
		Totals: models.TotalsCreateIn{
			VATRate:          2000,
			VatAmountMinor:   2000,
			DepositType:      "none",
			DepositRate:      0,
			DepositMinor:     0,
			DiscountType:     "none",
			DiscountRate:     0,
			DiscountMinor:    0,
			PaidMinor:        2000,
			SubtotalAfterDisc: 10000,
			SubtotalMinor:    10000,
			TotalMinor:       12000,
			BalanceDue:       10000,
		},
	}
}

func TestValidateInvoiceCreate_NormalizesSupplyDateMatchingIssueDate(t *testing.T) {
	in := validInvoiceInput()
	in.Overview.SupplyDate = &in.Overview.IssueDate

	got, errs := ValidateInvoiceCreate(in)
	if len(errs) > 0 {
		t.Fatalf("ValidateInvoiceCreate() errors = %v, want none", errs)
	}
	if got.Overview.SupplyDate != nil {
		t.Fatalf("SupplyDate = %v, want nil when it matches issue date", *got.Overview.SupplyDate)
	}
}

func TestValidatePaymentReceiptCreate_PaymentDateValidation(t *testing.T) {
	_, errs := ValidatePaymentReceiptCreate(models.PaymentReceiptCreateIn{
		AmountMinor: 2000,
		PaymentDate: "03/23/2026",
	})
	if len(errs) == 0 {
		t.Fatalf("expected validation error for invalid payment date")
	}
}

func TestValidatePaymentReceiptCreate_PaymentAmountValidation(t *testing.T) {
	_, errs := ValidatePaymentReceiptCreate(models.PaymentReceiptCreateIn{
		AmountMinor: 0,
		PaymentDate: "2026-03-23",
	})
	if len(errs) == 0 {
		t.Fatalf("expected validation error for non-positive payment amount")
	}
}

func TestValidatePaidVsDepositTotal(t *testing.T) {
	errs := ValidatePaidVsDepositTotal(models.TotalsCreateIn{
		TotalMinor:   1000,
		DepositMinor: 200,
		PaidMinor:    1001,
	})
	if len(errs) == 0 {
		t.Fatalf("expected paid vs total validation error")
	}

	okErrs := ValidatePaidVsDepositTotal(models.TotalsCreateIn{
		TotalMinor:   1000,
		DepositMinor: 200,
		PaidMinor:    900,
	})
	if len(okErrs) > 0 {
		t.Fatalf("ValidatePaidVsDepositTotal() errors = %v, want none", okErrs)
	}
}

func TestValidateInvoiceCreate_SourceRevisionNoValidation(t *testing.T) {
	in := validInvoiceInput()
	invalid := int64(0)
	in.Overview.SourceRevisionNo = &invalid

	_, errs := ValidateInvoiceCreate(in)
	if len(errs) == 0 {
		t.Fatalf("expected validation error for invalid sourceRevisionNo")
	}
}
