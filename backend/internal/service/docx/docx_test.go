package docx

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/viktorHadz/goInvoice26/internal/models"
)

func TestRenderDOCX_CreatesArchiveWithFooterAndEmbeddedLogo(t *testing.T) {
	logoPath := writeTestLogo(t)
	note := "Please reference the invoice number.\nThank you."
	dueDate := "05/04/2026"

	data, err := RenderDOCX(models.InvoicePDFData{
		Title:               "Invoice",
		InvoiceNumberLabel:  "INV-42-Rev-1",
		Currency:            "GBP",
		ShowItemTypeHeaders: true,
		IssueAt:             "28/03/2026",
		DueDate:             &dueDate,
		Note:                &note,
		Issuer: models.InvoicePDFIssuer{
			CompanyName:    "North Studio Ltd",
			Email:          "studio@example.com",
			Phone:          "+44 20 7123 4567",
			CompanyAddress: "1 Design Yard\nLondon\nN1 1AA",
			LogoPath:       logoPath,
		},
		Client: models.CreateClient{
			Name:        "Mila Hart",
			CompanyName: "Hart Retail",
			Address:     "14 Market Street\nLeeds\nLS1 4PL",
			Email:       "accounts@hart-retail.test",
		},
		Lines: []models.InvoicePDFItem{
			{Name: "Styling direction", LineType: "style", Quantity: "2", ItemPrice: "£150.00", ItemTotal: "£300.00", SortOrder: 1},
			{Name: "Sample production oversight", LineType: "sample", Quantity: "1", TimeWorked: "1h 30m", HourlyRate: "£220.00/hr", ItemTotal: "£220.00", SortOrder: 2},
		},
		Totals: models.TotalsCreateIn{
			SubtotalMinor:  100000,
			DiscountMinor:  5000,
			VatAmountMinor: 19000,
			TotalMinor:     114000,
			DepositMinor:   15000,
			PaidMinor:      20000,
			BalanceDue:     79000,
		},
		PaymentTerms:   "Payment due in 14 days",
		PaymentDetails: "Sort code: 00-00-00\nAccount: 12345678",
		NotesFooter:    "VAT registration available on request",
	})
	if err != nil {
		t.Fatalf("RenderDOCX() error = %v", err)
	}

	files := unzipFileMap(t, data)

	for _, name := range []string{
		"[Content_Types].xml",
		"_rels/.rels",
		"docProps/app.xml",
		"docProps/core.xml",
		"word/document.xml",
		"word/settings.xml",
		"word/styles.xml",
		"word/_rels/document.xml.rels",
		"word/footer1.xml",
		"word/media/logo.png",
	} {
		if _, ok := files[name]; !ok {
			t.Fatalf("archive missing %s", name)
		}
	}

	documentXML := files["word/document.xml"]
	footerXML := files["word/footer1.xml"]
	relsXML := files["word/_rels/document.xml.rels"]
	contentTypes := files["[Content_Types].xml"]

	for _, want := range []string{
		"INV-42-Rev-1",
		"North Studio Ltd",
		"Hart Retail",
		"Sample production oversight",
		"Balance Due",
		"£790.00",
		"Payment due in 14 days",
		`r:id="rId2"`,
		`r:embed="rId3"`,
	} {
		if !strings.Contains(documentXML, want) {
			t.Fatalf("document XML missing %q", want)
		}
	}

	if strings.Contains(documentXML, "Footer Notes") {
		t.Fatalf("document XML should not render footer notes heading in the body")
	}
	if strings.Contains(documentXML, "VAT registration available on request") {
		t.Fatalf("document XML should not contain footer note body text")
	}

	if !strings.Contains(footerXML, "VAT registration available on request") {
		t.Fatalf("footer XML missing footer note text")
	}
	if strings.Contains(footerXML, "Footer Notes") {
		t.Fatalf("footer XML should not contain footer heading text")
	}
	if !strings.Contains(footerXML, " PAGE ") {
		t.Fatalf("footer XML missing page number field")
	}

	for _, want := range []string{
		`Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer"`,
		`Target="footer1.xml"`,
		`Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image"`,
		`Target="media/logo.png"`,
		`Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings"`,
		`Target="settings.xml"`,
	} {
		if !strings.Contains(relsXML, want) {
			t.Fatalf("document relationships missing %q", want)
		}
	}

	for _, want := range []string{
		`Extension="png"`,
		`PartName="/word/footer1.xml"`,
		`PartName="/word/settings.xml"`,
	} {
		if !strings.Contains(contentTypes, want) {
			t.Fatalf("content types missing %q", want)
		}
	}
}

func TestRenderDOCX_UsesNoLineItemsFallback(t *testing.T) {
	data, err := RenderDOCX(models.InvoicePDFData{
		Title:              "Invoice",
		InvoiceNumberLabel: "INV-7",
		Currency:           "USD",
		IssueAt:            "28/03/2026",
		Totals: models.TotalsCreateIn{
			SubtotalMinor: 1000,
			TotalMinor:    1000,
			BalanceDue:    1000,
		},
	})
	if err != nil {
		t.Fatalf("RenderDOCX() error = %v", err)
	}

	files := unzipFileMap(t, data)
	documentXML, ok := files["word/document.xml"]
	if !ok {
		t.Fatal("document.xml not found")
	}

	if !strings.Contains(documentXML, "No line items.") {
		t.Fatalf("document XML missing no line items fallback")
	}
	if !strings.Contains(documentXML, `w:top w:val="single"`) {
		t.Fatalf("document XML missing full table borders")
	}
	footerXML, ok := files["word/footer1.xml"]
	if !ok {
		t.Fatalf("footer should be created for page numbers")
	}
	if !strings.Contains(footerXML, " PAGE ") {
		t.Fatalf("footer XML missing page number field")
	}
}

func unzipFileMap(t *testing.T, data []byte) map[string]string {
	t.Helper()

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("zip.NewReader() error = %v", err)
	}

	files := make(map[string]string, len(zr.File))
	for _, file := range zr.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("Open(%s) error = %v", file.Name, err)
		}
		body, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("ReadAll(%s) error = %v", file.Name, err)
		}
		files[file.Name] = string(body)
	}

	return files
}

func writeTestLogo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "logo.png")

	img := image.NewNRGBA(image.Rect(0, 0, 12, 6))
	for y := 0; y < 6; y++ {
		for x := 0; x < 12; x++ {
			img.Set(x, y, color.NRGBA{R: 20, G: 180, B: 160, A: 255})
		}
	}

	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("os.Create() error = %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if err := png.Encode(f, img); err != nil {
		t.Fatalf("png.Encode() error = %v", err)
	}

	return path
}
