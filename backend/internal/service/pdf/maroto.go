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
	"github.com/johnfercher/maroto/v2/pkg/components/page"
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
	renderClosingBlocks(mr, doc)

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

	issueDate := clean(doc.IssueAt)
	supplyDate := cleanPtr(doc.SupplyDate)
	dueDate := cleanPtr(doc.DueDate)
	logoPath, hasLogo := resolveLocalLogoPath(doc.Issuer.LogoPath)

	if hasLogo {
		mr.AddRow(invoiceTheme.row.headerLogo,
			image.NewFromFileCol(5, logoPath, props.Rect{Percent: 100, Top: 0.4}),
			text.NewCol(7, title, invoiceTheme.titleText(align.Right)),
		)
		mr.AddRow(invoiceTheme.row.headerText,
			col.New(5),
			text.NewCol(7, numberLabel, invoiceTheme.documentNoText(align.Right)),
		)
	} else {
		mr.AddRow(invoiceTheme.row.headerLogo,
			text.NewCol(12, title, invoiceTheme.titleText(align.Right)),
		)
		mr.AddRow(invoiceTheme.row.headerText,
			text.NewCol(12, numberLabel, invoiceTheme.documentNoText(align.Right)),
		)
	}

	renderHeaderMetaRow(mr, hasLogo, headerMetaValue("Issued", issueDate))
	if supplyDate != "" {
		renderHeaderMetaRow(mr, hasLogo, headerMetaValue("Supply", supplyDate))
	}
	renderHeaderMetaRow(mr, hasLogo, headerMetaValue("Due", dueDate))

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
	ensureFitsOrNewPage(mr, estimateDualPanelHeight(leftRows, rightRows, 42))
	renderDualPanelRows(mr, leftRows, rightRows, invoiceTheme.cell.party, invoiceTheme.cell.party, blankPartyRow())

	mr.AddRow(invoiceTheme.space.xl)
}

func renderItemTable(mr core.Maroto, doc models.InvoicePDFData) {
	renderSectionLabel(mr, "Line Items")

	mr.AddRows(
		row.New(invoiceTheme.row.tableHeader).
			WithStyle(invoiceTheme.cell.tableHeader).
			Add(
				text.NewCol(6, "Description", invoiceTheme.tableHeaderText(align.Left)),
				text.NewCol(1, "Qty", invoiceTheme.tableHeaderText(align.Center)),
				text.NewCol(1, "Time", invoiceTheme.tableHeaderText(align.Center)),
				text.NewCol(1, "Rate", invoiceTheme.tableHeaderText(align.Right)),
				text.NewCol(1, "Price", invoiceTheme.tableHeaderText(align.Right)),
				text.NewCol(2, "Amount", invoiceTheme.tableHeaderText(align.Right)),
			),
	)
	renderFullDivider(mr, invoiceTheme.line.divider)

	if len(doc.Lines) == 0 {
		mr.AddAutoRow(text.NewCol(12, "No line items.", invoiceTheme.text.emptyState))
		mr.AddRow(invoiceTheme.space.lg)
		return
	}

	groups := groupInvoicePDFItems(doc.Lines)
	renderedRows := 0

	for _, group := range groups {
		if len(group.Lines) == 0 {
			continue
		}

		if doc.ShowItemTypeHeaders {
			if renderedRows > 0 {
				mr.AddRow(invoiceTheme.space.sm)
			}
			mr.AddRow(invoiceTheme.row.groupLabel,
				text.NewCol(12, strings.ToUpper(group.Title), invoiceTheme.sectionLabelText(align.Left)),
			)
			renderFullDivider(mr, invoiceTheme.line.soft)
		}

		for i, ln := range group.Lines {
			mr.AddAutoRow(
				text.NewCol(6, clean(ln.Name), invoiceTheme.tableCellText(align.Left)),
				text.NewCol(1, clean(ln.Quantity), invoiceTheme.tableCellText(align.Center)),
				text.NewCol(1, clean(ln.TimeWorked), invoiceTheme.tableCellText(align.Center)),
				text.NewCol(1, clean(ln.HourlyRate), invoiceTheme.tableCellText(align.Right)),
				text.NewCol(1, clean(ln.ItemPrice), invoiceTheme.tableCellText(align.Right)),
				text.NewCol(2, clean(ln.ItemTotal), invoiceTheme.tableCellText(align.Right)),
			)
			renderedRows++

			if doc.ShowItemTypeHeaders && i < len(group.Lines)-1 {
				renderFullDivider(mr, invoiceTheme.line.soft)
			}
			if !doc.ShowItemTypeHeaders && renderedRows < len(doc.Lines) {
				renderFullDivider(mr, invoiceTheme.line.soft)
			}
		}
	}

	mr.AddRow(invoiceTheme.space.lg)
}

func renderClosingBlocks(mr core.Maroto, doc models.InvoicePDFData) {
	totalRows := buildTotalRows(doc)
	noteRows := buildNoteRows(doc.Note)
	paymentSections := buildPaymentSections(doc)

	renderTotalsBlock(mr, totalRows, noteRows)

	if len(paymentSections) > 0 {
		renderPaymentBlock(mr, paymentSections)
	}
}

func renderTotalsBlock(mr core.Maroto, totalRows []totalLine, noteRows []styledTextLine) {
	ensureFitsOrNewPage(mr, estimateSummaryHeight(totalRows, noteRows))
	renderOffsetSectionLabel(mr, "Summary", 7, 5)
	renderTotalsPanel(mr, totalRows)

	if len(noteRows) > 0 {
		mr.AddRow(invoiceTheme.space.xl)
		renderNotePanel(mr, noteRows)
	}

	mr.AddRow(invoiceTheme.space.xl)
}

func renderPaymentBlock(mr core.Maroto, sections []paymentSection) {
	ensureFitsOrNewPage(mr, estimatePaymentHeight(sections))
	renderSectionLabel(mr, "Payment")

	if len(sections) == 2 {
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

	addrOneLine := joinAddressParts(address)
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
	if cleanPtr(note) == "" {
		return nil
	}

	rows := make([]styledTextLine, 0, len(linesOf(*note)))
	for _, ln := range linesOf(*note) {
		rows = append(rows, styledTextLine{text: ln, style: invoiceTheme.text.noteBody})
	}

	return rows
}

func buildTotalRows(doc models.InvoicePDFData) []totalLine {
	if doc.DocumentKind == "payment_receipt" {
		rows := []totalLine{
			newTotalLine("Payment Amount", formatMoney(doc.ReceiptAmountMinor, doc.Currency)),
			newTotalLine("Invoice Total", formatMoney(doc.Totals.TotalMinor, doc.Currency)),
		}
		if doc.Totals.DepositMinor > 0 {
			rows = append(rows, newTotalLine("Requested Deposit", formatMoney(doc.Totals.DepositMinor, doc.Currency)))
		}
		rows = append(rows, newTotalLine("Total Paid", formatMoney(doc.Totals.PaidMinor, doc.Currency)))
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

	rows := []totalLine{
		newTotalLine("Subtotal", formatMoney(doc.Totals.SubtotalMinor, doc.Currency)),
	}

	if doc.Totals.DiscountMinor > 0 {
		rows = append(rows, newTotalLine("Discount", formatMoney(-doc.Totals.DiscountMinor, doc.Currency)))
	}

	rows = append(rows,
		newTotalLine("VAT", formatMoney(doc.Totals.VatAmountMinor, doc.Currency)),
		newTotalLine("Total", formatMoney(doc.Totals.TotalMinor, doc.Currency)),
	)

	if doc.Totals.DepositMinor > 0 {
		rows = append(rows, newTotalLine("Requested Deposit", formatMoney(doc.Totals.DepositMinor, doc.Currency)))
	}
	if doc.Totals.PaidMinor > 0 {
		rows = append(rows, newTotalLine("Paid", formatMoney(-doc.Totals.PaidMinor, doc.Currency)))
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
	mr.AddRow(invoiceTheme.row.sectionLabel,
		text.NewCol(12, strings.ToUpper(title), invoiceTheme.sectionLabelText(align.Left)),
	)
	mr.AddRow(invoiceTheme.space.xs)
}

func renderOffsetSectionLabel(mr core.Maroto, title string, offset, span int) {
	mr.AddRow(invoiceTheme.row.sectionLabel,
		col.New(offset),
		text.NewCol(span, strings.ToUpper(title), invoiceTheme.sectionLabelText(align.Left)),
	)
	mr.AddRow(invoiceTheme.space.xs)
}

func renderPaymentColumns(mr core.Maroto, left, right paymentSection) {
	leftRows := buildPaymentRows(left)
	rightRows := buildPaymentRows(right)
	renderDualPanelRows(mr, leftRows, rightRows, invoiceTheme.cell.payment, invoiceTheme.cell.payment, blankStyledLine(invoiceTheme.text.paymentBody))
	mr.AddRow(invoiceTheme.space.lg)
}

func renderPaymentSection(mr core.Maroto, section paymentSection) {
	renderSinglePanelRows(mr, buildPaymentRows(section), invoiceTheme.cell.payment)
	mr.AddRow(invoiceTheme.space.md)
}

func renderNotePanel(mr core.Maroto, rows []styledTextLine) {
	mr.AddRow(invoiceTheme.row.sectionLabel,
		text.NewCol(12, "NOTES", invoiceTheme.text.noteLabel),
	)
	renderSinglePanelRows(mr, rows, invoiceTheme.cell.note)
}

func blankPartyRow() styledTextLine {
	return styledTextLine{style: invoiceTheme.text.partyBody}
}

func blankStyledLine(style props.Text) styledTextLine {
	return styledTextLine{style: style}
}

func buildPaymentRows(section paymentSection) []styledTextLine {
	rows := []styledTextLine{
		{text: section.Title, style: invoiceTheme.text.paymentLabel},
	}
	for _, ln := range section.Lines {
		rows = append(rows, styledTextLine{text: ln, style: invoiceTheme.text.paymentBody})
	}
	return rows
}

func buildPaymentSections(doc models.InvoicePDFData) []paymentSection {
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

	return sections
}

func headerMetaValue(label, value string) string {
	if value == "" {
		return ""
	}
	return label + " " + value
}

func renderHeaderMetaRow(mr core.Maroto, hasLogo bool, value string) {
	if value == "" {
		return
	}

	if hasLogo {
		mr.AddRow(invoiceTheme.row.headerMeta,
			col.New(5),
			text.NewCol(7, value, invoiceTheme.metaText(align.Right)),
		)
		return
	}

	mr.AddRow(invoiceTheme.row.headerMeta,
		text.NewCol(12, value, invoiceTheme.metaText(align.Right)),
	)
}

func renderTotalsPanel(mr core.Maroto, rows []totalLine) {
	for i, total := range rows {
		if i == 0 {
			mr.AddRow(invoiceTheme.space.xs,
				col.New(7),
				col.New(5).WithStyle(total.cellStyle),
			)
		}
		if total.ruleAbove {
			mr.AddRow(invoiceTheme.space.xxs,
				col.New(7),
				line.NewCol(5, invoiceTheme.line.divider),
			)
		}

		mr.AddAutoRow(
			col.New(7),
			text.NewCol(3, total.label, total.labelStyle).WithStyle(total.cellStyle),
			text.NewCol(2, total.value, total.valueStyle).WithStyle(total.cellStyle),
		)

		if i == len(rows)-1 {
			mr.AddRow(invoiceTheme.space.xs,
				col.New(7),
				col.New(5).WithStyle(total.cellStyle),
			)
		}
	}
}

func renderSinglePanelRows(mr core.Maroto, rows []styledTextLine, style *props.Cell) {
	if len(rows) == 0 {
		return
	}

	mr.AddRow(invoiceTheme.space.xs, col.New(12).WithStyle(style))
	for _, item := range rows {
		mr.AddAutoRow(text.NewCol(12, item.text, item.style).WithStyle(style))
	}
	mr.AddRow(invoiceTheme.space.xs, col.New(12).WithStyle(style))
}

func renderDualPanelRows(
	mr core.Maroto,
	leftRows []styledTextLine,
	rightRows []styledTextLine,
	leftStyle *props.Cell,
	rightStyle *props.Cell,
	blank styledTextLine,
) {
	rowCount := maxInt(len(leftRows), len(rightRows))

	mr.AddRow(invoiceTheme.space.sm,
		col.New(5).WithStyle(leftStyle),
		col.New(2),
		col.New(5).WithStyle(rightStyle),
	)

	for i := 0; i < rowCount; i++ {
		left := blank
		if i < len(leftRows) {
			left = leftRows[i]
		}

		right := blank
		if i < len(rightRows) {
			right = rightRows[i]
		}

		mr.AddAutoRow(
			text.NewCol(5, left.text, left.style).WithStyle(leftStyle),
			col.New(2),
			text.NewCol(5, right.text, right.style).WithStyle(rightStyle),
		)
	}

	mr.AddRow(invoiceTheme.space.sm,
		col.New(5).WithStyle(leftStyle),
		col.New(2),
		col.New(5).WithStyle(rightStyle),
	)
}

func renderFullDivider(mr core.Maroto, rule props.Line) {
	mr.AddRow(invoiceTheme.space.xxs, line.NewCol(12, rule))
}

func ensureFitsOrNewPage(mr core.Maroto, estimatedHeight float64) {
	if !mr.FitlnCurrentPage(estimatedHeight) {
		mr.AddPages(page.New())
	}
}

func estimateDualPanelHeight(leftRows, rightRows []styledTextLine, charsPerLine int) float64 {
	lines := maxInt(estimateStyledRows(leftRows, charsPerLine), estimateStyledRows(rightRows, charsPerLine))
	return invoiceTheme.space.sm*2 + float64(lines)*5.6
}

func estimateSummaryHeight(totalRows []totalLine, noteRows []styledTextLine) float64 {
	height := invoiceTheme.row.sectionLabel + invoiceTheme.space.xs + invoiceTheme.space.xs*2 + float64(len(totalRows))*6.2 + invoiceTheme.space.xl
	if len(noteRows) > 0 {
		height += invoiceTheme.space.md + invoiceTheme.row.sectionLabel + invoiceTheme.space.xs*2 + float64(estimateStyledRows(noteRows, 92))*5.6
	}
	return height
}

func estimatePaymentHeight(sections []paymentSection) float64 {
	base := invoiceTheme.row.sectionLabel + invoiceTheme.space.xs
	if len(sections) == 1 {
		return base + invoiceTheme.space.xs*2 + float64(estimateStyledRows(buildPaymentRows(sections[0]), 92))*5.6 + invoiceTheme.space.md
	}

	leftRows := buildPaymentRows(sections[0])
	rightRows := buildPaymentRows(sections[1])
	return base + estimateDualPanelHeight(leftRows, rightRows, 42) + invoiceTheme.space.lg
}

func estimateStyledRows(rows []styledTextLine, charsPerLine int) int {
	total := 0
	for _, row := range rows {
		total += estimateTextLines(row.text, charsPerLine)
	}
	return total
}

func estimateTextLines(text string, charsPerLine int) int {
	if strings.TrimSpace(text) == "" {
		return 1
	}

	lines := 0
	for _, part := range linesOf(text) {
		runes := len([]rune(part))
		count := (runes + charsPerLine - 1) / charsPerLine
		if count < 1 {
			count = 1
		}
		lines += count
	}
	if lines == 0 {
		return 1
	}
	return lines
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
	if _, err := os.Stat(v); err == nil {
		return v, true
	}

	trimmed := strings.TrimPrefix(v, "/")
	if trimmed != v {
		if _, err := os.Stat(trimmed); err == nil {
			return trimmed, true
		}
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

func joinAddressParts(v string) string {
	parts := linesOf(v)
	for i, part := range parts {
		parts[i] = strings.TrimRight(strings.TrimSpace(part), ",")
	}
	return strings.Join(parts, ", ")
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
