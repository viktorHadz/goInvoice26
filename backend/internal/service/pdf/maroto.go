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

var (
	white   = &props.Color{Red: 255, Green: 255, Blue: 255}
	gray50  = &props.Color{Red: 248, Green: 248, Blue: 249}
	gray200 = &props.Color{Red: 232, Green: 232, Blue: 234}
	gray300 = &props.Color{Red: 214, Green: 214, Blue: 217}
	gray700 = &props.Color{Red: 70, Green: 70, Blue: 74}
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
	if rows := footerRows(doc); len(rows) > 0 {
		if err := mr.RegisterFooter(rows...); err != nil {
			return nil, fmt.Errorf("register footer: %w", err)
		}
	}

	addHeader(mr, doc)
	addParties(mr, doc)
	addItemsTable(mr, doc)
	addTotals(mr, doc)
	addNote(mr, doc)
	addPaymentInfoSections(mr, doc)

	out, err := mr.Generate()
	if err != nil {
		return nil, fmt.Errorf("maroto generate: %w", err)
	}
	return out.GetBytes(), nil
}

func addHeader(mr core.Maroto, doc models.InvoicePDFData) {
	logoPath, hasLogo := resolveLocalLogoPath(doc.Issuer.LogoURL)

	if hasLogo {
		mr.AddRow(24,
			image.NewFromFileCol(4, logoPath, props.Rect{Percent: 86}),
			text.NewCol(8, "INVOICE", props.Text{
				Size:  20,
				Style: fontstyle.Bold,
				Align: align.Right,
				Top:   3,
				Color: gray700,
			}),
		)
	} else {
		mr.AddRow(14,
			text.NewCol(12, "INVOICE", props.Text{
				Size:  20,
				Style: fontstyle.Bold,
				Align: align.Left,
				Color: gray700,
			}),
		)
	}

	mr.AddRow(8,
		text.NewCol(6, safe(doc.InvoiceNumberLabel), props.Text{
			Size:  12.5,
			Style: fontstyle.Bold,
			Align: align.Left,
			Top:   1,
		}),
		text.NewCol(3, "Issue date: "+safe(doc.IssueAt), props.Text{
			Size:  9,
			Align: align.Right,
			Top:   2,
		}),
		text.NewCol(3, dueLabel(doc.DueDate), props.Text{
			Size:  9,
			Align: align.Right,
			Top:   2,
		}),
	)

	mr.AddRow(3)
	mr.AddRow(1, line.NewCol(12, props.Line{Color: gray300}))
	mr.AddRow(5)
}

func addParties(mr core.Maroto, doc models.InvoicePDFData) {
	from := partyLines(
		doc.Issuer.CompanyName,
		"",
		doc.Issuer.CompanyAddress,
		doc.Issuer.Email,
		doc.Issuer.Phone,
	)

	to := partyLines(
		displayClientTitle(doc.Client.Name, doc.Client.CompanyName),
		"",
		doc.Client.Address,
		doc.Client.Email,
		"",
	)

	mr.AddRow(5,
		text.NewCol(6, "FROM", props.Text{
			Size:  8.8,
			Style: fontstyle.Bold,
			Align: align.Left,
		}),
		text.NewCol(6, "TO", props.Text{
			Size:  8.8,
			Style: fontstyle.Bold,
			Align: align.Left,
		}),
	)

	rows := len(from)
	if len(to) > rows {
		rows = len(to)
	}

	for i := 0; i < rows; i++ {
		left := ""
		right := ""

		if i < len(from) {
			left = from[i]
		}
		if i < len(to) {
			right = to[i]
		}

		h := maxPartyRowHeight(left, right)
		mr.AddRow(h,
			text.NewCol(6, left, props.Text{
				Size:  9,
				Align: align.Left,
				Top:   0.6,
			}),
			text.NewCol(6, right, props.Text{
				Size:  9,
				Align: align.Left,
				Top:   0.6,
			}),
		)
	}

	mr.AddRow(6)
}

func displayClientTitle(name, company string) string {
	company = strings.TrimSpace(company)
	if company != "" {
		return company
	}
	return strings.TrimSpace(name)
}

func maxPartyRowHeight(left, right string) float64 {
	h := partyTextHeight(left)
	if v := partyTextHeight(right); v > h {
		h = v
	}
	return h
}

func partyTextHeight(v string) float64 {
	v = strings.TrimSpace(v)
	if v == "" {
		return 4.8
	}

	lines := math.Ceil(float64(len([]rune(v))) / 34.0)
	if lines < 1 {
		lines = 1
	}

	if lines == 1 {
		return 4.8
	}

	return 4.8 + (lines-1)*1.3
}

func addItemsTable(mr core.Maroto, doc models.InvoicePDFData) {
	mr.AddRows(
		row.New(7).Add(
			text.NewCol(7, "Description", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
				Top:   2,
			}),
			text.NewCol(1, "Qty", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
				Top:   2,
				Align: align.Center,
			}),
			text.NewCol(2, "Unit", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
				Top:   2,
				Align: align.Right,
			}),
			text.NewCol(2, "Amount", props.Text{
				Size:  9,
				Style: fontstyle.Bold,
				Top:   2,
				Align: align.Right,
			}),
		).WithStyle(&props.Cell{BackgroundColor: gray200}),
	)

	if len(doc.Lines) == 0 {
		mr.AddRows(
			row.New(10).Add(
				text.NewCol(12, "No line items", props.Text{
					Size:  9,
					Align: align.Center,
					Top:   2.5,
				}),
			).WithStyle(&props.Cell{BackgroundColor: gray50}),
		)
		mr.AddRow(6)
		return
	}

	groups := groupInvoicePDFItems(doc.Lines)
	lineIndex := 0
	for _, group := range groups {
		if len(group.Lines) == 0 {
			continue
		}

		if doc.ShowItemTypeHeaders {
			mr.AddRows(
				row.New(6.5).Add(
					text.NewCol(12, group.Title, props.Text{
						Size:  8.8,
						Style: fontstyle.Bold,
						Align: align.Left,
						Top:   1.8,
						Color: gray700,
					}),
				).WithStyle(&props.Cell{BackgroundColor: gray200}),
			)
		}

		for _, ln := range group.Lines {
			h := blockHeight(safe(ln.Name), 42, 4.0) + 3
			if h < 9 {
				h = 9
			}

			r := row.New(h).Add(
				text.NewCol(7, safe(ln.Name), props.Text{
					Size:  9,
					Align: align.Left,
					Top:   2,
				}),
				text.NewCol(1, safe(ln.Quantity), props.Text{
					Size:  9,
					Align: align.Center,
					Top:   2,
				}),
				text.NewCol(2, safe(ln.ItemPrice), props.Text{
					Size:  9,
					Align: align.Right,
					Top:   2,
				}),
				text.NewCol(2, safe(ln.ItemTotal), props.Text{
					Size:  9,
					Style: fontstyle.Bold,
					Align: align.Right,
					Top:   2,
				}),
			)

			if lineIndex%2 == 1 {
				r.WithStyle(&props.Cell{BackgroundColor: gray50})
			} else {
				r.WithStyle(&props.Cell{BackgroundColor: white})
			}

			mr.AddRows(r)
			lineIndex++
		}
	}

	mr.AddRow(5)
}

func addTotals(mr core.Maroto, doc models.InvoicePDFData) {
	t := doc.Totals

	addTotalRow(mr, "Subtotal", formatMoney(t.SubtotalMinor, doc.Currency), false, false)

	if t.DiscountMinor > 0 {
		addTotalRow(mr, "Discount", formatMoney(-t.DiscountMinor, doc.Currency), false, false)
	}

	addTotalRow(mr, "VAT", formatMoney(t.VatAmountMinor, doc.Currency), false, false)
	addTotalRow(mr, "Total", formatMoney(t.TotalMinor, doc.Currency), true, false)

	if t.DepositMinor > 0 {
		addTotalRow(mr, "Deposit", formatMoney(-t.DepositMinor, doc.Currency), false, false)
	}
	if t.PaidMinor > 0 {
		addTotalRow(mr, "Paid", formatMoney(-t.PaidMinor, doc.Currency), false, false)
	}

	addTotalRow(mr, "Balance due", formatMoney(t.BalanceDue, doc.Currency), true, true)
	mr.AddRow(6)
}

func addTotalRow(mr core.Maroto, label, value string, bold, shaded bool) {
	size := float64(9)
	style := fontstyle.Normal
	bg := white

	if bold {
		size = 9.6
		style = fontstyle.Bold
	}
	if shaded {
		bg = gray200
	}

	r := row.New(7.5).Add(
		text.NewCol(7, "", props.Text{}),
		text.NewCol(3, label, props.Text{
			Size:  size,
			Style: style,
			Align: align.Right,
			Top:   2,
		}),
		text.NewCol(2, value, props.Text{
			Size:  size,
			Style: style,
			Align: align.Right,
			Top:   2,
		}),
	)
	r.WithStyle(&props.Cell{BackgroundColor: bg})
	mr.AddRows(r)
}

func addNote(mr core.Maroto, doc models.InvoicePDFData) {
	if doc.Note == nil || strings.TrimSpace(*doc.Note) == "" {
		return
	}

	body := normalizeMultiline(*doc.Note)

	mr.AddRow(6,
		text.NewCol(12, "NOTE", props.Text{
			Size:  8.8,
			Style: fontstyle.Bold,
			Align: align.Left,
		}),
	)

	addMultilineBodyRows(mr, body)

	mr.AddRow(6)
}

func addPaymentInfoSections(mr core.Maroto, doc models.InvoicePDFData) {
	addInfoSection(mr, "Payment terms", doc.PaymentTerms)
	addInfoSection(mr, "Payment details", doc.PaymentDetails)
}

func addInfoSection(mr core.Maroto, title, body string) {
	body = normalizeMultiline(body)
	if body == "" {
		return
	}

	mr.AddRow(6,
		text.NewCol(12, strings.ToUpper(title), props.Text{
			Size:  8.8,
			Style: fontstyle.Bold,
			Align: align.Left,
		}),
	)

	addMultilineBodyRows(mr, body)

	mr.AddRow(4)
}

func partyLines(name, company, address, email, phone string) []string {
	var out []string

	add := func(v string) {
		v = strings.TrimSpace(v)
		if v != "" {
			out = append(out, v)
		}
	}

	add(name)
	add(company)
	add(flattenAddress(address))
	add(phone)
	add(email)

	return out
}
func flattenAddress(v string) string {
	v = normalizeMultiline(v)
	if v == "" {
		return ""
	}

	parts := make([]string, 0, 4)
	for _, ln := range multilineLines(v) {
		ln = strings.TrimSpace(ln)
		ln = strings.Trim(ln, ", ")
		if ln != "" {
			parts = append(parts, ln)
		}
	}

	return strings.Join(parts, ", ")
}

type itemGroup struct {
	Title string
	Lines []models.InvoicePDFItem
}

func groupInvoicePDFItems(lines []models.InvoicePDFItem) []itemGroup {
	sorted := append([]models.InvoicePDFItem(nil), lines...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].SortOrder < sorted[j].SortOrder
	})

	groups := []itemGroup{
		{Title: "Styles"},
		{Title: "Samples"},
		{Title: "Other Items"},
	}

	for _, ln := range sorted {
		switch strings.TrimSpace(strings.ToLower(ln.LineType)) {
		case "style":
			groups[0].Lines = append(groups[0].Lines, ln)
		case "sample":
			groups[1].Lines = append(groups[1].Lines, ln)
		default:
			groups[2].Lines = append(groups[2].Lines, ln)
		}
	}

	return groups
}

func footerRows(doc models.InvoicePDFData) []core.Row {
	note := normalizeMultiline(doc.NotesFooter)
	if note == "" {
		return nil
	}

	note = clampFooter(note, 6) // keep footer sane

	h := blockHeight(note, 94, 3.2) + 1.6
	if h < 5.2 {
		h = 5.2
	}

	return []core.Row{
		row.New(0.7).Add(
			line.NewCol(12, props.Line{Color: gray300}),
		),
		row.New(h).Add(
			text.NewCol(12, note, props.Text{
				Size:  7.2,
				Align: align.Left,
				Top:   0.9,
			}),
		),
	}
}
func clampFooter(v string, maxLines int) string {
	lines := multilineLines(normalizeMultiline(v))
	if len(lines) <= maxLines {
		return strings.Join(lines, "\n")
	}

	lines = lines[:maxLines]
	last := strings.TrimSpace(lines[len(lines)-1])
	if last != "" {
		lines[len(lines)-1] = last + "…"
	} else {
		lines[len(lines)-1] = "…"
	}
	return strings.Join(lines, "\n")
}

func dueLabel(v *string) string {
	if v == nil || strings.TrimSpace(*v) == "" {
		return ""
	}
	return "Due date: " + strings.TrimSpace(*v)
}

func safe(v string) string {
	return strings.TrimSpace(v)
}

func resolveLocalLogoPath(v string) (string, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", false
	}
	v = strings.TrimPrefix(v, "/")
	if _, err := os.Stat(v); err == nil {
		return v, true
	}
	return "", false
}

func blockHeight(content string, charsPerLine int, lineHeight float64) float64 {
	content = strings.TrimSpace(content)
	if content == "" {
		return 8
	}

	var lines float64
	for _, part := range strings.Split(content, "\n") {
		part = strings.TrimSpace(part)
		if part == "" {
			lines++
			continue
		}
		n := math.Ceil(float64(len([]rune(part))) / float64(charsPerLine))
		if n < 1 {
			n = 1
		}
		lines += n
	}

	return lines*lineHeight + 2
}

func normalizeMultiline(v string) string {
	v = strings.ReplaceAll(v, "\r\n", "\n")
	v = strings.ReplaceAll(v, "\r", "\n")
	v = strings.ReplaceAll(v, "\\n", "\n")
	return strings.TrimSpace(v)
}

func multilineLines(v string) []string {
	parts := strings.Split(v, "\n")
	out := make([]string, 0, len(parts))
	for _, ln := range parts {
		ln = strings.TrimSpace(ln)
		if ln != "" {
			out = append(out, ln)
		}
	}
	return out
}

func addMultilineBodyRows(mr core.Maroto, body string) {
	for _, ln := range multilineLines(body) {
		h := blockHeight(ln, 94, 4.1) + 1.8
		if h < 6.2 {
			h = 6.2
		}
		mr.AddRows(
			row.New(h).Add(
				text.NewCol(12, ln, props.Text{
					Size:  9,
					Align: align.Left,
					Top:   1.2,
				}),
			).WithStyle(&props.Cell{BackgroundColor: gray50}),
		)
	}
}
