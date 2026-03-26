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
					{Name: "Styling direction and concept notes", LineType: "style", Quantity: "2", ItemPrice: "£150.00", ItemTotal: "£300.00", SortOrder: 1},
					{Name: "Sample production oversight", LineType: "sample", Quantity: "1", ItemPrice: "£220.00", ItemTotal: "£220.00", SortOrder: 2},
					{Name: "Final delivery pack", LineType: "other", Quantity: "1", ItemPrice: "£80.00", ItemTotal: "£80.00", SortOrder: 3},
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
					{Name: "Consulting", Quantity: "1", ItemPrice: "£100.00", ItemTotal: "£100.00", SortOrder: 1},
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
						Name:      fmt.Sprintf("Long invoice line %02d with enough text to wrap across the description column cleanly", i+1),
						LineType:  lineType,
						Quantity:  "1",
						ItemPrice: "£25.00",
						ItemTotal: "£25.00",
						SortOrder: int64(i + 1),
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
