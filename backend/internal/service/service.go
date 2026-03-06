package service

// import (
// 	"strings"

// 	"github.com/viktorHadz/goInvoice26/internal/models"
// )

// type ComputedLine struct {
// 	ProductID         *int64
// 	Name              string
// 	LineType          string
// 	PricingMode       string
// 	Quantity          int64
// 	UnitPriceMinor    int64
// 	MinutesWorked     *int64
// 	LineSubtotalMinor int64
// 	SortOrder         int64
// }

// type ComputedInvoice struct {
// 	IssueDate string
// 	DueByDate *string

// 	ClientName        *string
// 	ClientCompanyName *string
// 	ClientAddress     *string
// 	ClientEmail       *string
// 	Note              *string

// 	VATRate       int64
// 	DiscountType  string
// 	DiscountMinor int64

// 	Lines  []ComputedLine
// 	Totals models.InvoiceTotals
// }

// func Compute(in models.InvoiceCreateIn) ComputedInvoice {
// 	vatRate := int64(2000)
// 	if in.VATRate != nil {
// 		vatRate = *in.VATRate
// 	}

// 	dt := "none"
// 	if in.DiscountType != nil && strings.TrimSpace(*in.DiscountType) != "" {
// 		dt = *in.DiscountType
// 	}

// 	discMinor := int64(0)
// 	if in.DiscountMinor != nil {
// 		discMinor = *in.DiscountMinor
// 	}

// 	lines := make([]ComputedLine, 0, len(in.Lines))
// 	subtotal := int64(0)

// 	for i := range in.Lines {
// 		ln := in.Lines[i]

// 		var (
// 			name      = ""
// 			lineType  = ""
// 			pricing   = ""
// 			qty       = int64(0)
// 			unit      = int64(0)
// 			mins      *int64
// 			lineSub   int64
// 			sortOrder = int64(i + 1)
// 		)

// 		if ln.Name != nil {
// 			name = *ln.Name
// 		}
// 		if ln.LineType != nil {
// 			lineType = *ln.LineType
// 		}
// 		if ln.PricingMode != nil {
// 			pricing = *ln.PricingMode
// 		}
// 		if ln.Quantity != nil {
// 			qty = *ln.Quantity
// 		}
// 		if ln.UnitPriceMinor != nil {
// 			unit = *ln.UnitPriceMinor
// 		}
// 		mins = ln.MinutesWorked

// 		switch pricing {
// 		case "flat":
// 			lineSub = qty * unit
// 		case "hourly":
// 			// rounded: (minutes * hourlyRate) / 60
// 			m := int64(0)
// 			if mins != nil {
// 				m = *mins
// 			}

// 			lineMins := (m*unit + 30) / 60
// 			lineSub = qty * lineMins
// 		default:
// 			lineSub = 0
// 		}

// 		subtotal += lineSub

// 		lines = append(lines, ComputedLine{
// 			ProductID:         ln.ProductID,
// 			Name:              name,
// 			LineType:          lineType,
// 			PricingMode:       pricing,
// 			Quantity:          qty,
// 			UnitPriceMinor:    unit,
// 			MinutesWorked:     mins,
// 			LineSubtotalMinor: lineSub,
// 			SortOrder:         sortOrder,
// 		})
// 	}

// 	discount := int64(0)
// 	switch dt {
// 	case "none":
// 		discount = 0
// 	case "percent":
// 		discount = (subtotal * discMinor) / 10000
// 	case "fixed":
// 		discount = min(discMinor, subtotal)
// 	default:
// 		discount = 0
// 	}

// 	after := subtotal - discount
// 	vat := (after * vatRate) / 10000
// 	total := after + vat

// 	return ComputedInvoice{
// 		IssueDate: in.IssueDate,
// 		DueByDate: in.DueByDate,

// 		ClientName:        in.ClientName,
// 		ClientCompanyName: in.ClientCompanyName,
// 		ClientAddress:     in.ClientAddress,
// 		ClientEmail:       in.ClientEmail,
// 		Note:              in.Note,

// 		VATRate:       vatRate,
// 		DiscountType:  dt,
// 		DiscountMinor: discount,

// 		Lines: lines,
// 		Totals: models.InvoiceTotals{
// 			SubtotalMinor:      subtotal,
// 			DiscountMinor:      discount,
// 			AfterDiscountMinor: after,
// 			VATMinor:           vat,
// 			TotalMinor:         total,
// 		},
// 	}
// }
