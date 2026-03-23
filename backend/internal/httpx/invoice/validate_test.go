package invoice

import (
	"testing"

	"github.com/viktorHadz/goInvoice26/internal/models"
)

func validInvoiceInput() models.FEInvoiceIn {
	return models.FEInvoiceIn{
		Overview: models.InvoiceCreateIn{
			ClientID:          1,
			BaseNumber:        1,
			IssueDate:         "2026-03-23",
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
		Payments: []models.PaymentCreateIn{
			{
				AmountMinor: 2000,
				PaymentDate: "2026-03-23",
			},
		},
	}
}

func TestValidateInvoiceCreate_PaymentDateValidation(t *testing.T) {
	in := validInvoiceInput()
	in.Payments[0].PaymentDate = "03/23/2026"

	_, errs := ValidateInvoiceCreate(in)
	if len(errs) == 0 {
		t.Fatalf("expected validation error for invalid payment date")
	}
}

func TestValidateInvoiceCreate_PaymentAmountValidation(t *testing.T) {
	in := validInvoiceInput()
	in.Payments[0].AmountMinor = 0

	_, errs := ValidateInvoiceCreate(in)
	if len(errs) == 0 {
		t.Fatalf("expected validation error for non-positive payment amount")
	}
}

func TestValidatePaidVsDepositTotal(t *testing.T) {
	errs := ValidatePaidVsDepositTotal(models.TotalsCreateIn{
		TotalMinor:   1000,
		DepositMinor: 200,
		PaidMinor:    801,
	})
	if len(errs) == 0 {
		t.Fatalf("expected paid vs deposit validation error")
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
