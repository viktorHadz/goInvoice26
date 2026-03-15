package pdf

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"github.com/viktorHadz/goInvoice26/internal/models"
)

type MarotoRenderer struct{}

type lineGroup struct {
	Title string
	Lines []models.InvoicePDFItem
	Order int
}

type totalRow struct {
	Label string
	Value string
	Bold  bool
	Shade bool
}

type infoBlock struct {
	Title string
	Body  string
}

var (
	gray50  = &props.Color{Red: 250, Green: 250, Blue: 250}
	gray100 = &props.Color{Red: 244, Green: 244, Blue: 245}
	gray200 = &props.Color{Red: 232, Green: 232, Blue: 234}
	gray300 = &props.Color{Red: 214, Green: 214, Blue: 217}
)

func (m *MarotoRenderer) RenderPDF(ctx context.Context, doc models.InvoicePDFData) ([]byte, error) {
	cfg := config.NewBuilder().
		WithOrientation(orientation.Vertical).
		WithPageSize(pagesize.A4).
		WithLeftMargin(15).
		WithRightMargin(15).
		WithTopMargin(15).
		WithBottomMargin(15).
		Build()

	mr := maroto.New(cfg)

	addHeader(mr, doc)
	addFromTo(mr, doc)
	addItemsSections(mr, doc)
	addTotalsSection(mr, doc)
	addFooterInfo(mr, doc)

	out, err := mr.Generate()
	if err != nil {
		return nil, fmt.Errorf("maroto generate: %w", err)
	}
	return out.GetBytes(), nil
}

func addHeader(mr core.Maroto, doc models.InvoicePDFData) {
	logoPath, hasLogo := resolveLocalLogoPath(doc.Issuer.LogoURL)

	if hasLogo {
		mr.AddRow(32,
			image.NewFromFileCol(6, logoPath, props.Rect{Center: false, Percent: 100}),

			text.NewCol(6, "Invoice: "+doc.InvoiceNumberLabel, props.Text{Size: 15, Align: align.Right}),
		)
	} else {
		mr.AddRow(20,
			text.NewCol(4, doc.InvoiceNumberLabel, props.Text{Size: 15, Style: fontstyle.Bold, Align: align.Left}),
		)
	}

	mr.AddRow(12,
		text.NewCol(6, "Issued at: "+doc.IssueAt, props.Text{Size: 10, Top: 8, Align: align.Left}),
		text.NewCol(6, formatDueLabel(doc.DueDate), props.Text{Size: 10, Top: 8, Align: align.Right}),
	)

	mr.AddRow(5)
	mr.AddRow(1, line.NewCol(12, props.Line{Color: gray300}))
	mr.AddRow(5)
}

func addFromTo(mr core.Maroto, doc models.InvoicePDFData) {
	from := collectNonEmptyLines(
		doc.Issuer.CompanyName,
		doc.Issuer.CompanyAddress,
		doc.Issuer.Email,
		doc.Issuer.Phone,
	)

	to := collectNonEmptyLines(
		doc.Client.Name,
		doc.Client.CompanyName,
		doc.Client.Address,
		doc.Client.Email,
	)
	addTwoColumnLinesCard(mr, "FROM", from, 9, "TO", to, 9)
	mr.AddRow(5)
	mr.AddRow(1, line.NewCol(12, props.Line{Color: gray300}))
	mr.AddRow(5)
}

// func addHeader(mr core.Maroto, doc models.InvoicePDFData) {
// 	titleBlock := joinNonEmpty(doc.Title, doc.InvoiceNumberLabel)
// 	metaBlock := joinNonEmpty("Issue date: "+doc.IssueAt, formatDueLabel(doc.DueDate))
// 	issuerLines := collectNonEmptyLines(
// 		doc.Issuer.CompanyName,
// 		doc.Issuer.CompanyAddress,
// 		doc.Issuer.Email,
// 		doc.Issuer.Phone,
// 	)
// 	firstIssuerLine := ""
// 	if len(issuerLines) > 0 {
// 		firstIssuerLine = issuerLines[0]
// 	}

// 	logoPath, hasLogo := resolveLocalLogoPath(doc.Issuer.LogoURL)
// 	if hasLogo {
// 		mr.AddRow(24,
// 			image.NewFromFileCol(3, logoPath, props.Rect{Center: false, Percent: 100}),
// 			text.NewCol(5, firstIssuerLine, props.Text{Size: 9, Top: 1, Align: align.Left}),
// 			text.NewCol(4, titleBlock, props.Text{Size: 17, Style: fontstyle.Bold, Align: align.Right}),
// 		)

// 		for _, ln := range issuerLines[1:] {
// 			mr.AddRow(5,
// 				text.NewCol(3, "", props.Text{}),
// 				text.NewCol(5, ln, props.Text{Size: 9, Align: align.Left}),
// 				text.NewCol(4, "", props.Text{}),
// 			)
// 		}
// 	} else {
// 		mr.AddRow(18,
// 			text.NewCol(8, firstIssuerLine, props.Text{Size: 9, Top: 1, Align: align.Left}),
// 			text.NewCol(4, titleBlock, props.Text{Size: 17, Style: fontstyle.Bold, Align: align.Right}),
// 		)
// 		for _, ln := range issuerLines[1:] {
// 			mr.AddRow(5,
// 				text.NewCol(8, ln, props.Text{Size: 9, Align: align.Left}),
// 				text.NewCol(4, "", props.Text{}),
// 			)
// 		}
// 	}

// 	mr.AddRow(maxFloat(10, estimateBlockHeight(metaBlock, 24, 4.2)),
// 		text.NewCol(8, "", props.Text{}),
// 		text.NewCol(4, metaBlock, props.Text{Size: 9, Align: align.Right}),
// 	)
// 	mr.AddRow(1, line.NewCol(12, props.Line{Color: gray300}))
// 	mr.AddRow(5)
// }

// func addPartySection(mr core.Maroto, doc models.InvoicePDFData) {
// 	leftTitle := "BILL TO"
// 	leftLines := collectNonEmptyLines(
// 		doc.Client.Name,
// 		doc.Client.CompanyName,
// 		doc.Client.Address,
// 		doc.Client.Email,
// 	)

// 	rightTitle := "NOTE"
// 	rightLines := collectNonEmptyLines(trimmedPtr(doc.Note))
// 	if len(rightLines) == 0 {
// 		rightTitle = "DETAILS"
// 		rightLines = collectNonEmptyLines(
// 			doc.InvoiceNumberLabel,
// 			"Issue date: "+doc.IssueAt,
// 			formatDueLabel(doc.DueDate),
// 		)
// 	}

// 	addTwoColumnLinesCard(
// 		mr,
// 		leftTitle, leftLines, 9.5,
// 		rightTitle, rightLines, 9,
// 	)
// 	mr.AddRow(7)
// }

func addItemsSections(mr core.Maroto, doc models.InvoicePDFData) {
	if len(doc.Lines) == 0 {
		addSectionTitle(mr, "Items")
		mr.AddRow(10, text.NewCol(12, "No line items", props.Text{Size: 9, Align: align.Center}))
		mr.AddRow(6)
		return
	}

	for _, grp := range groupInvoiceLines(doc.Lines) {
		if len(grp.Lines) == 0 {
			continue
		}

		addSectionTitle(mr, grp.Title)
		addItemsTableHeader(mr)

		for i, ln := range grp.Lines {
			addItemRow(mr, ln, i%2 == 1)
		}
		mr.AddRow(5)
	}
}

func addItemsTableHeader(mr core.Maroto) {
	header := row.New(8).Add(
		text.NewCol(7, "Description", props.Text{Style: fontstyle.Bold, Size: 9.5}),
		text.NewCol(1, "Qty", props.Text{Style: fontstyle.Bold, Size: 9.5, Align: align.Center}),
		text.NewCol(2, "Unit", props.Text{Style: fontstyle.Bold, Size: 9.5, Align: align.Right}),
		text.NewCol(2, "Amount", props.Text{Style: fontstyle.Bold, Size: 9.5, Align: align.Right}),
	)
	header.WithStyle(&props.Cell{BackgroundColor: gray200})
	mr.AddRows(header)
}

func addItemRow(mr core.Maroto, ln models.InvoicePDFItem, shade bool) {
	rowHeight := maxFloat(8, estimateBlockHeight(ln.Name, 42, 3.5))
	r := row.New(rowHeight).Add(
		text.NewCol(7, ln.Name, props.Text{Size: 9.5, Top: 1, Align: align.Left}),
		text.NewCol(1, ln.Quantity, props.Text{Size: 9, Top: 1, Align: align.Center}),
		text.NewCol(2, ln.ItemPrice, props.Text{Size: 9, Top: 1, Align: align.Right}),
		text.NewCol(2, ln.ItemTotal, props.Text{Size: 9, Top: 1, Style: fontstyle.Bold, Align: align.Right}),
	)
	if shade {
		r.WithStyle(&props.Cell{BackgroundColor: gray50})
	}
	mr.AddRows(r)
}

func addTotalsSection(mr core.Maroto, doc models.InvoicePDFData) {
	rows := makeTotalRows(doc)
	mr.AddRow(2)
	mr.AddRow(1, line.NewCol(12, props.Line{Color: gray300}))
	mr.AddRow(5)

	addSectionTitle(mr, "Totals")
	for _, tr := range rows {
		addTotalRow(mr, tr)
	}
	mr.AddRow(8)
}

func makeTotalRows(doc models.InvoicePDFData) []totalRow {
	t := doc.Totals
	rows := []totalRow{
		{Label: "Subtotal", Value: formatMoney(t.SubtotalMinor, doc.Currency)},
	}
	if t.DiscountMinor > 0 {
		rows = append(rows, totalRow{Label: "Discount", Value: formatMoney(-t.DiscountMinor, doc.Currency)})
	}
	rows = append(rows,
		totalRow{Label: "VAT", Value: formatMoney(t.VatAmountMinor, doc.Currency)},
		totalRow{Label: "Total", Value: formatMoney(t.TotalMinor, doc.Currency)},
	)
	if t.DepositMinor > 0 {
		rows = append(rows, totalRow{Label: "Deposit", Value: formatMoney(-t.DepositMinor, doc.Currency)})
	}
	if t.PaidMinor > 0 {
		rows = append(rows, totalRow{Label: "Paid", Value: formatMoney(-t.PaidMinor, doc.Currency)})
	}
	rows = append(rows, totalRow{
		Label: "Balance due",
		Value: formatMoney(t.BalanceDue, doc.Currency),
		Bold:  true,
		Shade: true,
	})
	return rows
}

func addTotalRow(mr core.Maroto, tr totalRow) {
	size, style := float64(9), fontstyle.Normal
	if tr.Bold {
		size, style = 10, fontstyle.Bold
	}

	r := row.New(8).Add(
		text.NewCol(7, "", props.Text{}),
		text.NewCol(3, tr.Label, props.Text{Size: size, Style: style, Align: align.Right}),
		text.NewCol(2, tr.Value, props.Text{Size: size, Style: style, Align: align.Right}),
	)
	if tr.Shade {
		r.WithStyle(&props.Cell{BackgroundColor: gray100})
	}
	mr.AddRows(r)
}

func addFooterInfo(mr core.Maroto, doc models.InvoicePDFData) {
	blocks := []infoBlock{
		{Title: "Payment terms", Body: doc.PaymentTerms},
		{Title: "Payment details", Body: doc.PaymentDetails},
		{Title: "Notes", Body: doc.NotesFooter},
	}

	for _, b := range blocks {
		body := strings.TrimSpace(b.Body)
		if body == "" {
			continue
		}
		addSectionTitle(mr, b.Title)
		mr.AddRow(
			estimateBlockHeight(body, 88, 4),
			text.NewCol(12, body, props.Text{Size: 9, Top: 1, Align: align.Left}),
		)
		mr.AddRow(4)
	}
}

func addSectionTitle(mr core.Maroto, title string) {
	r := row.New(8).Add(
		text.NewCol(12, strings.ToUpper(strings.TrimSpace(title)), props.Text{
			Style: fontstyle.Bold,
			Size:  9,
		}),
	)
	r.WithStyle(&props.Cell{BackgroundColor: gray100})
	mr.AddRows(r)
}

func addTwoColumnLinesCard(
	mr core.Maroto,
	leftTitle string, leftLines []string, leftSize float64,
	rightTitle string, rightLines []string, rightSize float64,
) {
	headerRow := row.New(6).Add(
		text.NewCol(4, leftTitle, props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Left}),
		text.NewCol(5, "", props.Text{}), // spacer
		text.NewCol(4, rightTitle, props.Text{Style: fontstyle.Bold, Size: 9, Align: align.Left}),
	)
	headerRow.WithStyle(&props.Cell{BackgroundColor: gray100})
	mr.AddRows(headerRow)

	rows := max(len(leftLines), len(rightLines))
	if rows == 0 {
		rows = 1
	}

	for i := 0; i < rows; i++ {
		left := ""
		right := ""

		if i < len(leftLines) {
			left = leftLines[i]
		}
		if i < len(rightLines) {
			right = rightLines[i]
		}

		bodyRow := row.New(6).Add(
			text.NewCol(4, left, props.Text{Size: leftSize, Align: align.Left}),
			text.NewCol(5, "", props.Text{}), // spacer
			text.NewCol(4, right, props.Text{Size: rightSize, Align: align.Left}),
		)

		bodyRow.WithStyle(&props.Cell{BackgroundColor: gray50})
		mr.AddRows(bodyRow)
	}
}

func groupInvoiceLines(lines []models.InvoicePDFItem) []lineGroup {
	grouped := map[string][]models.InvoicePDFItem{
		"Styles":      {},
		"Samples":     {},
		"Other Items": {},
	}

	for _, ln := range lines {
		switch normalizeLineType(ln.LineType) {
		case "style":
			grouped["Styles"] = append(grouped["Styles"], ln)
		case "sample":
			grouped["Samples"] = append(grouped["Samples"], ln)
		default:
			grouped["Other Items"] = append(grouped["Other Items"], ln)
		}
	}

	out := []lineGroup{
		{Title: "Styles", Lines: grouped["Styles"], Order: 1},
		{Title: "Samples", Lines: grouped["Samples"], Order: 2},
		{Title: "Other Items", Lines: grouped["Other Items"], Order: 3},
	}

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Order < out[j].Order
	})
	return out
}

func normalizeLineType(v string) string {
	s := strings.ToLower(strings.TrimSpace(v))
	switch s {
	case "style", "styles":
		return "style"
	case "sample", "samples":
		return "sample"
	default:
		return "other"
	}
}

func trimmedPtr(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func joinNonEmpty(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, "\n")
}

func collectNonEmptyLines(parts ...string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		for _, line := range strings.Split(part, "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				out = append(out, line)
			}
		}
	}
	return out
}

func formatDueLabel(v *string) string {
	if v == nil || strings.TrimSpace(*v) == "" {
		return ""
	}
	return "Due date: " + *v
}

func resolveLocalLogoPath(v string) (string, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", false
	}

	clean := strings.TrimPrefix(v, "/")
	if _, err := os.Stat(clean); err == nil {
		return clean, true
	}
	return "", false
}

func estimateBlockHeight(content string, charsPerLine int, lineHeight float64) float64 {
	content = strings.TrimSpace(content)
	if content == "" {
		return 8
	}

	lines := strings.Split(content, "\n")
	var visualLines float64
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			visualLines++
			continue
		}

		count := math.Ceil(float64(len([]rune(l))) / float64(charsPerLine))
		if count < 1 {
			count = 1
		}
		visualLines += count
	}
	return maxFloat(8, visualLines*lineHeight+2)
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
