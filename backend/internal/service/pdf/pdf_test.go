package pdf

import (
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
