package pdf

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/consts/pagesize"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"

	"github.com/viktorHadz/goInvoice26/internal/models"
)

type MarotoRenderer struct{}

func (m *MarotoRenderer) RenderPDF(ctx context.Context, doc models.InvoicePDFData) ([]byte, error) {
	cfg := config.NewBuilder().
		WithOrientation(orientation.Vertical).
		WithPageSize(pagesize.A4).
		WithLeftMargin(14).
		WithRightMargin(14).
		WithTopMargin(14).
		WithBottomMargin(12).
		WithPageNumber(props.PageNumber{
			Pattern: "Page {current} of {total}",
			Place:   props.LeftBottom,
			Size:    invoiceTheme.text.footer.Size,
			Color:   invoiceTheme.text.footer.Color,
		}).
		Build()

	mr := maroto.New(cfg)

	if err := registerFooter(mr, doc); err != nil {
		return nil, fmt.Errorf("register footer: %w", err)
	}

	renderHeader(mr, doc)
	renderMeta(mr, doc)
	renderItemTable(mr, doc)
	renderTotalsBlock(mr, doc)
	renderPaymentBlock(mr, doc)

	out, err := mr.Generate()
	if err != nil {
		return nil, fmt.Errorf("generate pdf: %w", err)
	}
	return out.GetBytes(), nil
}

func registerFooter(mr core.Maroto, doc models.InvoicePDFData) error {
	rows := []core.Row{
		row.New(invoiceTheme.space.xs),
		line.NewRow(invoiceTheme.row.footerRule, invoiceTheme.line.soft),
	}
	for _, ln := range linesOf(doc.NotesFooter) {
		rows = append(rows, text.NewRow(4.5, ln, invoiceTheme.text.footer))
	}
	return mr.RegisterFooter(rows...)
}

func renderHeader(mr core.Maroto, doc models.InvoicePDFData) {
	title := strings.ToUpper(clean(doc.Title))
	if title == "" {
		title = "INVOICE"
	}

	numberLabel := clean(doc.InvoiceNumberLabel)
	if numberLabel == "" {
		numberLabel = "—"
	}

	issueText := compactMeta("Issued", clean(doc.IssueAt))
	dueText := compactMeta("Due", cleanPtr(doc.DueDate))

	if path, ok := resolveLocalLogoPath(doc.Issuer.LogoURL); ok {
		mr.AddRow(invoiceTheme.row.headerLogo,
			image.NewFromFileCol(4, path, props.Rect{Percent: 78, Top: 1}),
			text.NewCol(8, title, invoiceTheme.titleText(align.Right)),
		)
		mr.AddRow(invoiceTheme.row.headerText,
			col.New(4),
			text.NewCol(8, numberLabel, invoiceTheme.documentNoText(align.Right)),
		)
		mr.AddRow(invoiceTheme.row.headerMeta,
			col.New(4),
			text.NewCol(4, issueText, invoiceTheme.metaText(align.Right)),
			text.NewCol(4, dueText, invoiceTheme.metaText(align.Right)),
		)
	} else {
		mr.AddRow(invoiceTheme.row.headerLogo,
			text.NewCol(12, title, invoiceTheme.titleText(align.Right)),
		)
		mr.AddRow(invoiceTheme.row.headerText,
			text.NewCol(12, numberLabel, invoiceTheme.documentNoText(align.Right)),
		)
		mr.AddRow(invoiceTheme.row.headerMeta,
			text.NewCol(6, issueText, invoiceTheme.metaText(align.Left)),
			text.NewCol(6, dueText, invoiceTheme.metaText(align.Right)),
		)
	}

	mr.AddRow(invoiceTheme.space.lg)
}

func renderMeta(mr core.Maroto, doc models.InvoicePDFData) {
	clientName := clean(doc.Client.CompanyName)
	if clientName == "" {
		clientName = clean(doc.Client.Name)
	}

	left := buildPartyBlock("BILL TO", clientName, doc.Client.Address, clean(doc.Client.Email), "")
	right := buildPartyBlock("ISSUED BY", clean(doc.Issuer.CompanyName), doc.Issuer.CompanyAddress, clean(doc.Issuer.Email), clean(doc.Issuer.Phone))

	leftRows := left.rows()
	rightRows := right.rows()
	rowCount := maxInt(len(leftRows), len(rightRows))

	for i := 0; i < rowCount; i++ {
		l := blankPartyRow()
		if i < len(leftRows) {
			l = leftRows[i]
		}

		r := blankPartyRow()
		if i < len(rightRows) {
			r = rightRows[i]
		}

		mr.AddAutoRow(
			text.NewCol(6, l.text, l.style).WithStyle(invoiceTheme.cell.party),
			text.NewCol(6, r.text, r.style).WithStyle(invoiceTheme.cell.party),
		)
	}

	mr.AddRow(invoiceTheme.space.xl)
}

func renderItemTable(mr core.Maroto, doc models.InvoicePDFData) {
	renderSectionLabel(mr, "Line Items")

	mr.AddRows(
		row.New(invoiceTheme.row.tableHeader).
			WithStyle(invoiceTheme.cell.tableHeader).
			Add(
				text.NewCol(7, "Description", invoiceTheme.tableHeaderText(align.Left)),
				text.NewCol(1, "Qty", invoiceTheme.tableHeaderText(align.Center)),
				text.NewCol(2, "Unit Price", invoiceTheme.tableHeaderText(align.Right)),
				text.NewCol(2, "Amount", invoiceTheme.tableHeaderText(align.Right)),
			),
	)

	if len(doc.Lines) == 0 {
		mr.AddAutoRow(text.NewCol(12, "No line items.", invoiceTheme.text.emptyState))
		mr.AddRow(invoiceTheme.space.lg)
		return
	}

	groups := groupInvoicePDFItems(doc.Lines)
	rendered := 0

	for _, group := range groups {
		if len(group.Lines) == 0 {
			continue
		}

		if doc.ShowItemTypeHeaders {
			if rendered > 0 {
				mr.AddRow(invoiceTheme.space.sm)
			}
			mr.AddRow(invoiceTheme.row.groupLabel,
				text.NewCol(12, strings.ToUpper(group.Title), invoiceTheme.sectionLabelText(align.Left)),
			)
		}

		for _, ln := range group.Lines {
			if rendered > 0 {
				mr.AddRow(invoiceTheme.space.xxs, line.NewCol(12, invoiceTheme.line.soft))
			}

			mr.AddAutoRow(
				text.NewCol(7, clean(ln.Name), invoiceTheme.tableCellText(align.Left)),
				text.NewCol(1, clean(ln.Quantity), invoiceTheme.tableCellText(align.Center)),
				text.NewCol(2, clean(ln.ItemPrice), invoiceTheme.tableCellText(align.Right)),
				text.NewCol(2, clean(ln.ItemTotal), invoiceTheme.tableCellText(align.Right)),
			)
			rendered++
		}
	}

	mr.AddRow(invoiceTheme.space.lg)
}

func renderTotalsBlock(mr core.Maroto, doc models.InvoicePDFData) {
	renderSectionLabel(mr, "Summary")

	noteRows := buildNoteRows(doc.Note)
	totalRows := buildTotalRows(doc)
	rowCount := maxInt(len(noteRows), len(totalRows))

	for i := 0; i < rowCount; i++ {
		if i < len(totalRows) && totalRows[i].ruleAbove {
			mr.AddRow(invoiceTheme.space.xxs, col.New(6), line.NewCol(6, invoiceTheme.line.divider))
		}

		note := blankSectionLine(invoiceTheme.text.noteBody)
		if i < len(noteRows) {
			note = noteRows[i]
		}

		if i >= len(totalRows) {
			mr.AddAutoRow(
				text.NewCol(6, note.text, note.style),
				col.New(4),
				col.New(2),
			)
			continue
		}

		total := totalRows[i]
		mr.AddAutoRow(
			text.NewCol(6, note.text, note.style),
			text.NewCol(4, total.label, total.labelStyle).WithStyle(total.cellStyle),
			text.NewCol(2, total.value, total.valueStyle).WithStyle(total.cellStyle),
		)
	}

	mr.AddRow(invoiceTheme.space.xl)
}

func renderPaymentBlock(mr core.Maroto, doc models.InvoicePDFData) {
	sections := make([]paymentSection, 0, 2)
	if clean(doc.PaymentTerms) != "" {
		sections = append(sections, paymentSection{
			Title: "PAYMENT TERMS",
			Lines: linesOf(doc.PaymentTerms),
		})
	}
	if clean(doc.PaymentDetails) != "" {
		sections = append(sections, paymentSection{
			Title: "PAYMENT DETAILS",
			Lines: linesOf(doc.PaymentDetails),
		})
	}
	if len(sections) == 0 {
		return
	}

	renderSectionLabel(mr, "Payment")

	if len(sections) == 2 && fitsPaymentColumns(mr, sections[0], sections[1]) {
		renderPaymentColumns(mr, sections[0], sections[1])
		return
	}

	for _, section := range sections {
		renderPaymentSection(mr, section)
	}
}

type partyBlock struct {
	label   string
	name    string
	details []string
}

type styledTextLine struct {
	text  string
	style props.Text
}

type totalLine struct {
	label      string
	value      string
	labelStyle props.Text
	valueStyle props.Text
	cellStyle  *props.Cell
	ruleAbove  bool
}

type paymentSection struct {
	Title string
	Lines []string
}

type itemGroup struct {
	Title string
	Lines []models.InvoicePDFItem
}

func buildPartyBlock(label, name, address, email, phone string) partyBlock {
	details := make([]string, 0, 3)

	addrOneLine := strings.Join(linesOf(address), ", ")
	if addrOneLine != "" {
		details = append(details, addrOneLine)
	}
	if email != "" {
		details = append(details, email)
	}
	if phone != "" {
		details = append(details, phone)
	}

	return partyBlock{
		label:   label,
		name:    name,
		details: details,
	}
}

func (p partyBlock) rows() []styledTextLine {
	rows := []styledTextLine{
		{text: p.label, style: invoiceTheme.text.partyLabel},
		{text: p.name, style: invoiceTheme.text.partyName},
	}

	for _, detail := range p.details {
		rows = append(rows, styledTextLine{text: detail, style: invoiceTheme.text.partyBody})
	}

	return rows
}

func buildNoteRows(note *string) []styledTextLine {
	rows := []styledTextLine{
		{text: "NOTES", style: invoiceTheme.text.noteLabel},
	}

	for _, ln := range linesOf(cleanPtr(note)) {
		rows = append(rows, styledTextLine{text: ln, style: invoiceTheme.text.noteBody})
	}

	return rows
}

func buildTotalRows(doc models.InvoicePDFData) []totalLine {
	rows := []totalLine{
		newTotalLine("Subtotal", formatMoney(doc.Totals.SubtotalMinor, doc.Currency)),
	}

	if doc.Totals.DiscountMinor > 0 {
		rows = append(rows, newTotalLine("Discount", "−"+formatMoney(doc.Totals.DiscountMinor, doc.Currency)))
	}

	rows = append(rows,
		newTotalLine("VAT", formatMoney(doc.Totals.VatAmountMinor, doc.Currency)),
		newTotalLine("Total", formatMoney(doc.Totals.TotalMinor, doc.Currency)),
	)

	if doc.Totals.DepositMinor > 0 {
		rows = append(rows, newTotalLine("Deposit", "−"+formatMoney(doc.Totals.DepositMinor, doc.Currency)))
	}
	if doc.Totals.PaidMinor > 0 {
		rows = append(rows, newTotalLine("Paid", "−"+formatMoney(doc.Totals.PaidMinor, doc.Currency)))
	}

	rows = append(rows, totalLine{
		label:      "Balance Due",
		value:      formatMoney(doc.Totals.BalanceDue, doc.Currency),
		labelStyle: invoiceTheme.balanceLabelText(),
		valueStyle: invoiceTheme.balanceValueText(),
		cellStyle:  invoiceTheme.cell.balance,
		ruleAbove:  true,
	})

	return rows
}

func newTotalLine(label, value string) totalLine {
	return totalLine{
		label:      label,
		value:      value,
		labelStyle: invoiceTheme.totalLabelText(),
		valueStyle: invoiceTheme.totalValueText(),
		cellStyle:  invoiceTheme.cell.total,
	}
}

func renderSectionLabel(mr core.Maroto, title string) {
	if !mr.FitlnCurrentPage(invoiceTheme.row.sectionLabel + invoiceTheme.row.tableHeader) {
		mr.AddRow(999)
	}

	mr.AddRow(invoiceTheme.row.sectionLabel,
		text.NewCol(12, strings.ToUpper(title), invoiceTheme.sectionLabelText(align.Left)),
	)
	mr.AddRow(invoiceTheme.space.xs)
}

func fitsPaymentColumns(mr core.Maroto, left, right paymentSection) bool {
	bodyRows := maxInt(len(left.Lines), len(right.Lines))
	estimatedHeight := invoiceTheme.row.tableHeader + float64(bodyRows)*6 + invoiceTheme.space.lg
	return mr.FitlnCurrentPage(estimatedHeight)
}

func renderPaymentColumns(mr core.Maroto, left, right paymentSection) {
	mr.AddRows(
		row.New(invoiceTheme.row.tableHeader).
			Add(
				text.NewCol(6, left.Title, invoiceTheme.text.paymentLabel).WithStyle(invoiceTheme.cell.payment),
				text.NewCol(6, right.Title, invoiceTheme.text.paymentLabel).WithStyle(invoiceTheme.cell.payment),
			),
	)

	rows := maxInt(len(left.Lines), len(right.Lines))
	for i := 0; i < rows; i++ {
		leftText := ""
		if i < len(left.Lines) {
			leftText = left.Lines[i]
		}

		rightText := ""
		if i < len(right.Lines) {
			rightText = right.Lines[i]
		}

		mr.AddAutoRow(
			text.NewCol(6, leftText, invoiceTheme.text.paymentBody).WithStyle(invoiceTheme.cell.payment),
			text.NewCol(6, rightText, invoiceTheme.text.paymentBody).WithStyle(invoiceTheme.cell.payment),
		)
	}

	mr.AddRow(invoiceTheme.space.lg)
}

func renderPaymentSection(mr core.Maroto, section paymentSection) {
	estimatedHeight := invoiceTheme.row.tableHeader + float64(len(section.Lines))*6 + invoiceTheme.space.md
	if !mr.FitlnCurrentPage(estimatedHeight) {
		mr.AddRow(999)
	}

	mr.AddRows(
		row.New(invoiceTheme.row.tableHeader).
			Add(text.NewCol(12, section.Title, invoiceTheme.text.paymentLabel).WithStyle(invoiceTheme.cell.payment)),
	)

	for _, ln := range section.Lines {
		mr.AddAutoRow(text.NewCol(12, ln, invoiceTheme.text.paymentBody).WithStyle(invoiceTheme.cell.payment))
	}

	mr.AddRow(invoiceTheme.space.md)
}

func blankPartyRow() styledTextLine {
	return styledTextLine{style: invoiceTheme.text.partyBody}
}

func blankSectionLine(style props.Text) styledTextLine {
	return styledTextLine{style: style}
}

func compactMeta(label, value string) string {
	if value == "" {
		return ""
	}
	return label + " " + value
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

func linesOf(v string) []string {
	v = strings.ReplaceAll(v, "\r\n", "\n")
	v = strings.ReplaceAll(v, "\r", "\n")
	v = strings.ReplaceAll(v, "\\n", "\n")

	parts := strings.Split(strings.TrimSpace(v), "\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func clean(v string) string { return strings.TrimSpace(v) }

func cleanPtr(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
