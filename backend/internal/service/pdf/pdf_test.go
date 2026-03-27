package pdf

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func TestBuildInvoicePDFData_InvoiceNumberLabelUsesDisplayMapping(t *testing.T) {
	overview := &invoiceTx.InvoiceOverviewTotals{
		BaseNumber:        7,
		RevisionNo:        1,
		IssueDate:         "2026-03-25",
		ClientName:        "Client",
		ClientCompanyName: "Client Co",
		ClientAddress:     "Address",
		ClientEmail:       "client@example.com",
	}

	settings := models.Settings{
		InvoicePrefix: "INV-",
		DateFormat:    "dd/mm/yyyy",
		Currency:      "GBP",
	}

	baseDoc := buildInvoicePDFData(overview, nil, settings)
	if baseDoc.InvoiceNumberLabel != "INV - 7" {
		t.Fatalf("base invoice label = %q, want %q", baseDoc.InvoiceNumberLabel, "INV - 7")
	}

	overview.RevisionNo = 2
	firstRevisionDoc := buildInvoicePDFData(overview, nil, settings)
	if firstRevisionDoc.InvoiceNumberLabel != "INV - 7.1" {
		t.Fatalf("first revision label = %q, want %q", firstRevisionDoc.InvoiceNumberLabel, "INV - 7.1")
	}
}

func TestBuildPartyBlock_NormalizesAddressCommas(t *testing.T) {
	block := buildPartyBlock(
		"ISSUED BY",
		"North Studio Ltd",
		"1 Design Yard,\nLondon,\nN1 1AA",
		"studio@example.com",
		"",
	)

	if len(block.details) < 2 {
		t.Fatalf("buildPartyBlock() details = %v, want address and email", block.details)
	}

	if got, want := block.details[0], "1 Design Yard, London, N1 1AA"; got != want {
		t.Fatalf("buildPartyBlock() address = %q, want %q", got, want)
	}
}

func TestBuildNoteRows_ExcludesSectionTitle(t *testing.T) {
	note := "Line one\nLine two"

	rows := buildNoteRows(&note)

	if len(rows) != 2 {
		t.Fatalf("buildNoteRows() len = %d, want 2", len(rows))
	}
	if got, want := rows[0].text, "Line one"; got != want {
		t.Fatalf("buildNoteRows()[0] = %q, want %q", got, want)
	}
	if got, want := rows[1].text, "Line two"; got != want {
		t.Fatalf("buildNoteRows()[1] = %q, want %q", got, want)
	}
}

func TestBuildTotalRows_UsesASCIIHyphenForNegativeSummaryValues(t *testing.T) {
	rows := buildTotalRows(models.InvoicePDFData{
		Currency: "GBP",
		Totals: models.TotalsCreateIn{
			SubtotalMinor:  10000,
			DiscountMinor:  500,
			VatAmountMinor: 1900,
			TotalMinor:     11400,
			DepositMinor:   1500,
			PaidMinor:      2000,
			BalanceDue:     7900,
		},
	})

	valuesByLabel := make(map[string]string, len(rows))
	for _, row := range rows {
		valuesByLabel[row.label] = row.value
	}

	for label, want := range map[string]string{
		"Discount": "-£5.00",
		"Deposit":  "-£15.00",
		"Paid":     "-£20.00",
	} {
		if got := valuesByLabel[label]; got != want {
			t.Fatalf("%s value = %q, want %q", label, got, want)
		}
	}
}

func TestFormatDurationMinutes(t *testing.T) {
	tests := []struct {
		name    string
		minutes int64
		want    string
	}{
		{name: "zero", minutes: 0, want: "0m"},
		{name: "below one hour", minutes: 45, want: "45m"},
		{name: "exact hour", minutes: 60, want: "1h"},
		{name: "mixed hours and minutes", minutes: 90, want: "1h 30m"},
		{name: "multi hour", minutes: 125, want: "2h 5m"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatDurationMinutes(tc.minutes); got != tc.want {
				t.Fatalf("formatDurationMinutes(%d) = %q, want %q", tc.minutes, got, tc.want)
			}
		})
	}
}

func TestBuildInvoicePDFPricing(t *testing.T) {
	minutes := int64(90)

	hourly := buildInvoicePDFPricing("hourly", 6000, &minutes, "GBP")
	if hourly.itemPrice != notApplicableValue {
		t.Fatalf("hourly itemPrice = %q, want %q", hourly.itemPrice, notApplicableValue)
	}
	if hourly.timeWorked != "1h 30m" {
		t.Fatalf("hourly timeWorked = %q, want %q", hourly.timeWorked, "1h 30m")
	}
	if hourly.hourlyRate != "£60.00/hr" {
		t.Fatalf("hourly hourlyRate = %q, want %q", hourly.hourlyRate, "£60.00/hr")
	}

	flat := buildInvoicePDFPricing("flat", 2500, &minutes, "GBP")
	if flat.itemPrice != "£25.00" {
		t.Fatalf("flat itemPrice = %q, want %q", flat.itemPrice, "£25.00")
	}
	if flat.timeWorked != notApplicableValue {
		t.Fatalf("flat timeWorked = %q, want %q", flat.timeWorked, notApplicableValue)
	}
	if flat.hourlyRate != notApplicableValue {
		t.Fatalf("flat hourlyRate = %q, want %q", flat.hourlyRate, notApplicableValue)
	}
}

func TestBuildQuickInvoice_HourlyLineIncludesPricingColumns(t *testing.T) {
	minutes := int64(90)
	note := "Invoice note"

	doc := BuildQuickInvoice(
		models.FEInvoiceIn{
			Overview: models.InvoiceCreateIn{
				ClientID:          1,
				BaseNumber:        42,
				IssueDate:         "2026-03-25",
				ClientName:        "Client",
				ClientCompanyName: "Client Co",
				ClientAddress:     "Address",
				ClientEmail:       "client@example.com",
				Note:              &note,
			},
			Lines: []models.LineCreateIn{
				{
					Name:           "Hourly sample",
					LineType:       "sample",
					PricingMode:    "hourly",
					Quantity:       1,
					MinutesWorked:  &minutes,
					UnitPriceMinor: 6000,
					LineTotalMinor: 9000,
					SortOrder:      1,
				},
				{
					Name:           "Flat line",
					LineType:       "custom",
					PricingMode:    "flat",
					Quantity:       1,
					UnitPriceMinor: 2500,
					LineTotalMinor: 2500,
					SortOrder:      2,
				},
			},
		},
		models.Settings{
			InvoicePrefix: "INV-",
			DateFormat:    "dd/mm/yyyy",
			Currency:      "GBP",
		},
		1,
	)

	if len(doc.Lines) != 2 {
		t.Fatalf("BuildQuickInvoice() lines len = %d, want 2", len(doc.Lines))
	}

	if got := doc.Lines[0].ItemPrice; got != notApplicableValue {
		t.Fatalf("hourly line ItemPrice = %q, want %q", got, notApplicableValue)
	}
	if got := doc.Lines[0].TimeWorked; got != "1h 30m" {
		t.Fatalf("hourly line TimeWorked = %q, want %q", got, "1h 30m")
	}
	if got := doc.Lines[0].HourlyRate; got != "£60.00/hr" {
		t.Fatalf("hourly line HourlyRate = %q, want %q", got, "£60.00/hr")
	}

	if got := doc.Lines[1].ItemPrice; got != "£25.00" {
		t.Fatalf("flat line ItemPrice = %q, want %q", got, "£25.00")
	}
	if got := doc.Lines[1].TimeWorked; got != notApplicableValue {
		t.Fatalf("flat line TimeWorked = %q, want %q", got, notApplicableValue)
	}
	if got := doc.Lines[1].HourlyRate; got != notApplicableValue {
		t.Fatalf("flat line HourlyRate = %q, want %q", got, notApplicableValue)
	}
}

func TestMarotoRenderer_RenderPDF(t *testing.T) {
	t.Parallel()

	renderer := &MarotoRenderer{}
	note := "Please reference the invoice number on payment."
	dueDate := "2026-04-15"

	tests := []struct {
		name string
		doc  models.InvoicePDFData
	}{
		{
			name: "full invoice",
			doc: models.InvoicePDFData{
				Title:               "Invoice",
				InvoiceNumberLabel:  "INV - 42",
				Currency:            "GBP",
				ShowItemTypeHeaders: true,
				IssueAt:             "26/03/2026",
				DueDate:             &dueDate,
				Note:                &note,
				Issuer: models.InvoicePDFIssuer{
					CompanyName:    "North Studio Ltd",
					Email:          "studio@example.com",
					Phone:          "+44 20 7123 4567",
					CompanyAddress: "1 Design Yard\nLondon\nN1 1AA",
				},
				Client: models.CreateClient{
					Name:        "Mila Hart",
					CompanyName: "Hart Retail",
					Address:     "14 Market Street\nLeeds\nLS1 4PL",
					Email:       "accounts@hart-retail.test",
				},
				Lines: []models.InvoicePDFItem{
					{Name: "Styling direction and concept notes", LineType: "style", Quantity: "2", ItemPrice: "£150.00", TimeWorked: "—", HourlyRate: "—", ItemTotal: "£300.00", SortOrder: 1},
					{Name: "Sample production oversight", LineType: "sample", Quantity: "1", ItemPrice: "—", TimeWorked: "1h 30m", HourlyRate: "£220.00/hr", ItemTotal: "£220.00", SortOrder: 2},
					{Name: "Final delivery pack", LineType: "other", Quantity: "1", ItemPrice: "£80.00", TimeWorked: "—", HourlyRate: "—", ItemTotal: "£80.00", SortOrder: 3},
				},
				Totals: models.TotalsCreateIn{
					SubtotalMinor:  60000,
					VatAmountMinor: 12000,
					TotalMinor:     72000,
					BalanceDue:     72000,
				},
				PaymentTerms:   "Payment due within 14 days.",
				PaymentDetails: "Account name: North Studio Ltd\nSort code: 00-11-22\nAccount number: 12345678",
				NotesFooter:    "Thank you for your business.",
			},
		},
		{
			name: "minimal invoice",
			doc: models.InvoicePDFData{
				Title:              "Invoice",
				InvoiceNumberLabel: "INV - 7",
				Currency:           "GBP",
				IssueAt:            "26/03/2026",
				Issuer: models.InvoicePDFIssuer{
					CompanyName: "North Studio Ltd",
				},
				Client: models.CreateClient{
					Name: "Client",
				},
				Lines: []models.InvoicePDFItem{
					{Name: "Consulting", Quantity: "1", ItemPrice: "£100.00", TimeWorked: "—", HourlyRate: "—", ItemTotal: "£100.00", SortOrder: 1},
				},
				Totals: models.TotalsCreateIn{
					SubtotalMinor: 10000,
					TotalMinor:    10000,
					BalanceDue:    10000,
				},
			},
		},
		{
			name: "empty line items",
			doc: models.InvoicePDFData{
				Title:              "Invoice",
				InvoiceNumberLabel: "INV - 8",
				Currency:           "GBP",
				IssueAt:            "26/03/2026",
				Issuer: models.InvoicePDFIssuer{
					CompanyName: "North Studio Ltd",
				},
				Client: models.CreateClient{
					Name: "Client",
				},
				Totals: models.TotalsCreateIn{
					SubtotalMinor: 0,
					TotalMinor:    0,
					BalanceDue:    0,
				},
			},
		},
		{
			name: "multi page invoice",
			doc: func() models.InvoicePDFData {
				doc := models.InvoicePDFData{
					Title:               "Invoice",
					InvoiceNumberLabel:  "INV - 99",
					Currency:            "GBP",
					ShowItemTypeHeaders: true,
					IssueAt:             "26/03/2026",
					Issuer: models.InvoicePDFIssuer{
						CompanyName:    "North Studio Ltd",
						CompanyAddress: "1 Design Yard\nLondon\nN1 1AA",
					},
					Client: models.CreateClient{
						Name:        "Client",
						CompanyName: "Longform Client Co",
						Address:     "14 Market Street\nLeeds\nLS1 4PL",
					},
					Totals: models.TotalsCreateIn{
						SubtotalMinor:  150000,
						VatAmountMinor: 30000,
						TotalMinor:     180000,
						BalanceDue:     180000,
					},
					PaymentTerms: "Payment due within 30 days.",
				}

				for i := 0; i < 70; i++ {
					lineType := "other"
					switch i % 3 {
					case 0:
						lineType = "style"
					case 1:
						lineType = "sample"
					}

					doc.Lines = append(doc.Lines, models.InvoicePDFItem{
						Name:       fmt.Sprintf("Long invoice line %02d with enough text to wrap across the description column cleanly", i+1),
						LineType:   lineType,
						Quantity:   "1",
						ItemPrice:  "£25.00",
						TimeWorked: "—",
						HourlyRate: "—",
						ItemTotal:  "£25.00",
						SortOrder:  int64(i + 1),
					})
				}

				return doc
			}(),
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pdfBytes, err := renderer.RenderPDF(context.Background(), tc.doc)
			if err != nil {
				t.Fatalf("RenderPDF() error = %v", err)
			}
			if len(pdfBytes) == 0 {
				t.Fatal("RenderPDF() returned empty PDF")
			}
			if len(pdfBytes) < 4 {
				t.Fatalf("RenderPDF() returned too few bytes for a PDF header: %d", len(pdfBytes))
			}
			if !strings.HasPrefix(string(pdfBytes), "%PDF") {
				t.Fatalf("RenderPDF() did not produce a PDF header: got %q", string(pdfBytes[:4]))
			}
		})
	}
}
