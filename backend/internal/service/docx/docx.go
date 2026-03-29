package docx

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/viktorHadz/goInvoice26/internal/models"
)

const (
	wordNamespace               = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"
	relationshipNamespace       = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	drawingNamespace            = "http://schemas.openxmlformats.org/drawingml/2006/main"
	pictureNamespace            = "http://schemas.openxmlformats.org/drawingml/2006/picture"
	wordDrawingNamespace        = "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing"
	documentStylesRelID         = "rId1"
	documentFooterRelID         = "rId2"
	documentLogoRelID           = "rId3"
	documentSettingsRelID       = "rId4"
	defaultLogoName             = "logo"
	maxLogoWidthEMU       int64 = 1_900_000
	maxLogoHeightEMU      int64 = 750_000
	defaultLogoWidthEMU   int64 = 1_600_000
	defaultLogoHeightEMU  int64 = 600_000
	emuPerPixel           int64 = 9_525
)

type archiveFile struct {
	name string
	data []byte
}

type paragraphOptions struct {
	align         string
	bold          bool
	size          int
	spacingBefore int
	spacingAfter  int
	topBorder     bool
}

type itemGroup struct {
	Title string
	Lines []models.InvoicePDFItem
}

type summaryRow struct {
	label     string
	value     string
	highlight bool
}

type embeddedImage struct {
	archivePath string
	fileName    string
	extension   string
	contentType string
	data        []byte
	widthEMU    int64
	heightEMU   int64
}

func RenderDOCX(doc models.InvoicePDFData) ([]byte, error) {
	var out bytes.Buffer

	logo := loadEmbeddedLogo(doc.Issuer.LogoURL)
	footerLines := linesOf(doc.NotesFooter)
	hasFooter := true

	zw := zip.NewWriter(&out)
	files := []archiveFile{
		xmlFile("[Content_Types].xml", contentTypesXML(logo, hasFooter)),
		xmlFile("_rels/.rels", rootRelationshipsXML()),
		xmlFile("docProps/app.xml", appPropsXML()),
		xmlFile("docProps/core.xml", corePropsXML(doc)),
		xmlFile("word/document.xml", documentXML(doc, logo, hasFooter)),
		xmlFile("word/settings.xml", settingsXML()),
		xmlFile("word/styles.xml", stylesXML()),
		xmlFile("word/_rels/document.xml.rels", documentRelationshipsXML(logo, hasFooter)),
	}

	if hasFooter {
		files = append(files, xmlFile("word/footer1.xml", footerXML(footerLines)))
	}
	if logo != nil {
		files = append(files, archiveFile{
			name: logo.archivePath,
			data: append([]byte(nil), logo.data...),
		})
	}

	for _, file := range files {
		w, err := zw.Create(file.name)
		if err != nil {
			return nil, fmt.Errorf("create %s: %w", file.name, err)
		}
		if _, err := w.Write(file.data); err != nil {
			return nil, fmt.Errorf("write %s: %w", file.name, err)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close docx archive: %w", err)
	}

	return out.Bytes(), nil
}

func xmlFile(name, content string) archiveFile {
	return archiveFile{name: name, data: []byte(content)}
}

func contentTypesXML(logo *embeddedImage, hasFooter bool) string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">`)
	b.WriteString(`<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>`)
	b.WriteString(`<Default Extension="xml" ContentType="application/xml"/>`)
	if logo != nil {
		b.WriteString(fmt.Sprintf(
			`<Default Extension="%s" ContentType="%s"/>`,
			escapeXML(logo.extension),
			escapeXML(logo.contentType),
		))
	}
	b.WriteString(`<Override PartName="/docProps/app.xml" ContentType="application/vnd.openxmlformats-officedocument.extended-properties+xml"/>`)
	b.WriteString(`<Override PartName="/docProps/core.xml" ContentType="application/vnd.openxmlformats-package.core-properties+xml"/>`)
	b.WriteString(`<Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>`)
	b.WriteString(`<Override PartName="/word/settings.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"/>`)
	b.WriteString(`<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>`)
	if hasFooter {
		b.WriteString(`<Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>`)
	}
	b.WriteString(`</Types>`)

	return b.String()
}

func rootRelationshipsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/package/2006/relationships/metadata/core-properties" Target="docProps/core.xml"/>
  <Relationship Id="rId3" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/extended-properties" Target="docProps/app.xml"/>
</Relationships>`
}

func appPropsXML() string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Properties xmlns="http://schemas.openxmlformats.org/officeDocument/2006/extended-properties"
 xmlns:vt="http://schemas.openxmlformats.org/officeDocument/2006/docPropsVTypes">
  <Application>GoInvoicer</Application>
</Properties>`
}

func corePropsXML(doc models.InvoicePDFData) string {
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	title := clean(doc.Title)
	if title == "" {
		title = "Invoice"
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<cp:coreProperties xmlns:cp="http://schemas.openxmlformats.org/package/2006/metadata/core-properties"
 xmlns:dc="http://purl.org/dc/elements/1.1/"
 xmlns:dcterms="http://purl.org/dc/terms/"
 xmlns:dcmitype="http://purl.org/dc/dcmitype/"
 xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <dc:title>%s</dc:title>
  <dc:creator>GoInvoicer</dc:creator>
  <cp:lastModifiedBy>GoInvoicer</cp:lastModifiedBy>
  <dcterms:created xsi:type="dcterms:W3CDTF">%s</dcterms:created>
  <dcterms:modified xsi:type="dcterms:W3CDTF">%s</dcterms:modified>
</cp:coreProperties>`, escapeXML(title), now, now)
}

func documentRelationshipsXML(logo *embeddedImage, hasFooter bool) string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">`)
	b.WriteString(fmt.Sprintf(
		`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`,
		documentStylesRelID,
	))
	if hasFooter {
		b.WriteString(fmt.Sprintf(
			`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer" Target="footer1.xml"/>`,
			documentFooterRelID,
		))
	}
	if logo != nil {
		target := strings.TrimPrefix(logo.archivePath, "word/")
		b.WriteString(fmt.Sprintf(
			`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="%s"/>`,
			documentLogoRelID,
			escapeXML(target),
		))
	}
	b.WriteString(fmt.Sprintf(
		`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings" Target="settings.xml"/>`,
		documentSettingsRelID,
	))
	b.WriteString(`</Relationships>`)

	return b.String()
}

func stylesXML() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:styles xmlns:w="%s">
  <w:docDefaults>
    <w:rPrDefault>
      <w:rPr>
        <w:rFonts w:ascii="Calibri" w:hAnsi="Calibri" w:cs="Calibri"/>
        <w:sz w:val="22"/>
        <w:szCs w:val="22"/>
      </w:rPr>
    </w:rPrDefault>
    <w:pPrDefault>
      <w:pPr>
        <w:spacing w:after="120"/>
      </w:pPr>
    </w:pPrDefault>
  </w:docDefaults>
</w:styles>`, wordNamespace)
}

func settingsXML() string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:settings xmlns:w="%s"><w:updateFields w:val="true"/></w:settings>`, wordNamespace)
}

func documentXML(doc models.InvoicePDFData, logo *embeddedImage, hasFooter bool) string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(fmt.Sprintf(
		`<w:document xmlns:w="%s" xmlns:r="%s" xmlns:wp="%s" xmlns:a="%s" xmlns:pic="%s"><w:body>`,
		wordNamespace,
		relationshipNamespace,
		wordDrawingNamespace,
		drawingNamespace,
		pictureNamespace,
	))

	if logo != nil {
		b.WriteString(headerTableXML(doc, logo))
	} else {
		for _, p := range headerTextParagraphs(doc) {
			b.WriteString(p)
		}
	}

	b.WriteString(paragraph(" ", paragraphOptions{spacingAfter: 120}))
	b.WriteString(partyTableXML(doc))
	b.WriteString(paragraph(" ", paragraphOptions{spacingAfter: 160}))
	b.WriteString(sectionHeading("Line Items"))
	b.WriteString(lineItemsTableXML(doc))
	b.WriteString(paragraph(" ", paragraphOptions{spacingAfter: 160}))
	b.WriteString(sectionHeading("Summary"))
	b.WriteString(summaryTableXML(doc))

	if cleanPtr(doc.Note) != "" {
		b.WriteString(paragraph(" ", paragraphOptions{spacingAfter: 120}))
		b.WriteString(sectionHeading("Notes"))
		for _, line := range linesOf(*doc.Note) {
			b.WriteString(paragraph(line, paragraphOptions{spacingAfter: 60}))
		}
	}

	if clean(doc.PaymentTerms) != "" {
		b.WriteString(paragraph(" ", paragraphOptions{spacingAfter: 120}))
		b.WriteString(sectionHeading("Payment Terms"))
		for _, line := range linesOf(doc.PaymentTerms) {
			b.WriteString(paragraph(line, paragraphOptions{spacingAfter: 60}))
		}
	}

	if clean(doc.PaymentDetails) != "" {
		b.WriteString(paragraph(" ", paragraphOptions{spacingAfter: 120}))
		b.WriteString(sectionHeading("Payment Details"))
		for _, line := range linesOf(doc.PaymentDetails) {
			b.WriteString(paragraph(line, paragraphOptions{spacingAfter: 60}))
		}
	}

	b.WriteString(`<w:sectPr>`)
	if hasFooter {
		b.WriteString(fmt.Sprintf(`<w:footerReference w:type="default" r:id="%s"/>`, documentFooterRelID))
	}
	b.WriteString(`<w:pgSz w:w="11906" w:h="16838"/>`)
	b.WriteString(`<w:pgMar w:top="1440" w:right="1080" w:bottom="1080" w:left="1080" w:header="708" w:footer="708" w:gutter="0"/>`)
	b.WriteString(`</w:sectPr>`)
	b.WriteString(`</w:body></w:document>`)

	return b.String()
}

func footerXML(lines []string) string {
	var b strings.Builder

	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(fmt.Sprintf(`<w:ftr xmlns:w="%s">`, wordNamespace))

	for i, line := range lines {
		b.WriteString(paragraph(line, paragraphOptions{
			align:         "center",
			size:          18,
			spacingBefore: 40,
			spacingAfter:  20,
			topBorder:     i == 0,
		}))
	}
	b.WriteString(pageNumberParagraphXML())

	b.WriteString(`</w:ftr>`)
	return b.String()
}

func headerTableXML(doc models.InvoicePDFData, logo *embeddedImage) string {
	var b strings.Builder

	b.WriteString(`<w:tbl>`)
	b.WriteString(`<w:tblPr><w:tblW w:w="0" w:type="auto"/><w:tblLayout w:type="fixed"/>` + whiteTableXML() + `</w:tblPr>`)
	b.WriteString(`<w:tblGrid><w:gridCol w:w="2600"/><w:gridCol w:w="6760"/></w:tblGrid>`)
	b.WriteString(`<w:tr>`)
	b.WriteString(tableCellXML(
		[]string{imageParagraphXML(documentLogoRelID, logo)},
		2600,
		false,
		true,
		1,
	))
	b.WriteString(tableCellXML(headerTextParagraphs(doc), 6760, false, true, 1))
	b.WriteString(`</w:tr></w:tbl>`)

	return b.String()
}

func headerTextParagraphs(doc models.InvoicePDFData) []string {
	title := clean(doc.Title)
	if title == "" {
		title = "Invoice"
	}

	numberLabel := clean(doc.InvoiceNumberLabel)
	if numberLabel == "" {
		numberLabel = "Invoice"
	}

	paragraphs := []string{
		paragraph(strings.ToUpper(title), paragraphOptions{
			align:        "right",
			bold:         true,
			size:         32,
			spacingAfter: 40,
		}),
		paragraph(numberLabel, paragraphOptions{
			align:        "right",
			bold:         true,
			size:         24,
			spacingAfter: 120,
		}),
	}

	for _, row := range []string{
		metaLine("Issued", doc.IssueAt),
		metaLine("Due", cleanPtr(doc.DueDate)),
	} {
		if row == "" {
			continue
		}
		paragraphs = append(paragraphs, paragraph(row, paragraphOptions{
			align:        "right",
			size:         20,
			spacingAfter: 80,
		}))
	}

	return paragraphs
}

func imageParagraphXML(relID string, logo *embeddedImage) string {
	if logo == nil {
		return paragraph(" ", paragraphOptions{})
	}

	var b strings.Builder

	b.WriteString(`<w:p><w:pPr><w:jc w:val="left"/></w:pPr><w:r><w:drawing>`)
	b.WriteString(`<wp:inline distT="0" distB="0" distL="0" distR="0">`)
	b.WriteString(fmt.Sprintf(`<wp:extent cx="%d" cy="%d"/>`, logo.widthEMU, logo.heightEMU))
	b.WriteString(`<wp:effectExtent l="0" t="0" r="0" b="0"/>`)
	b.WriteString(fmt.Sprintf(`<wp:docPr id="1" name="%s"/>`, escapeXML(logo.fileName)))
	b.WriteString(`<wp:cNvGraphicFramePr><a:graphicFrameLocks noChangeAspect="1"/></wp:cNvGraphicFramePr>`)
	b.WriteString(`<a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture">`)
	b.WriteString(`<pic:pic>`)
	b.WriteString(fmt.Sprintf(`<pic:nvPicPr><pic:cNvPr id="0" name="%s"/><pic:cNvPicPr/></pic:nvPicPr>`, escapeXML(logo.fileName)))
	b.WriteString(`<pic:blipFill>`)
	b.WriteString(fmt.Sprintf(`<a:blip r:embed="%s"/>`, relID))
	b.WriteString(`<a:stretch><a:fillRect/></a:stretch>`)
	b.WriteString(`</pic:blipFill>`)
	b.WriteString(`<pic:spPr>`)
	b.WriteString(fmt.Sprintf(`<a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm>`, logo.widthEMU, logo.heightEMU))
	b.WriteString(`<a:prstGeom prst="rect"><a:avLst/></a:prstGeom>`)
	b.WriteString(`</pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline>`)
	b.WriteString(`</w:drawing></w:r></w:p>`)

	return b.String()
}

func paragraph(text string, opts paragraphOptions) string {
	var b strings.Builder

	b.WriteString(`<w:p>`)

	if opts.align != "" || opts.spacingBefore > 0 || opts.spacingAfter > 0 || opts.topBorder {
		b.WriteString(`<w:pPr>`)
		if opts.align != "" {
			b.WriteString(fmt.Sprintf(`<w:jc w:val="%s"/>`, opts.align))
		}
		if opts.spacingBefore > 0 || opts.spacingAfter > 0 {
			b.WriteString(fmt.Sprintf(`<w:spacing w:before="%d" w:after="%d"/>`, opts.spacingBefore, opts.spacingAfter))
		}
		if opts.topBorder {
			b.WriteString(`<w:pBdr><w:top w:val="single" w:sz="6" w:space="10" w:color="D4D4D8"/></w:pBdr>`)
		}
		b.WriteString(`</w:pPr>`)
	}

	b.WriteString(`<w:r>`)
	if opts.bold || opts.size > 0 {
		b.WriteString(`<w:rPr>`)
		if opts.bold {
			b.WriteString(`<w:b/>`)
		}
		if opts.size > 0 {
			b.WriteString(fmt.Sprintf(`<w:sz w:val="%d"/><w:szCs w:val="%d"/>`, opts.size, opts.size))
		}
		b.WriteString(`</w:rPr>`)
	}
	b.WriteString(`<w:t xml:space="preserve">`)
	b.WriteString(escapeXML(text))
	b.WriteString(`</w:t></w:r></w:p>`)

	return b.String()
}

func sectionHeading(text string) string {
	return paragraph(strings.ToUpper(clean(text)), paragraphOptions{
		bold:         true,
		size:         22,
		spacingAfter: 80,
	})
}

func partyTableXML(doc models.InvoicePDFData) string {
	leftName := clean(doc.Client.CompanyName)
	if leftName == "" {
		leftName = clean(doc.Client.Name)
	}

	leftParagraphs := []string{
		paragraph("BILL TO", paragraphOptions{bold: true, size: 20, spacingAfter: 20}),
		paragraph(defaultText(leftName), paragraphOptions{bold: true, size: 22, spacingAfter: 40}),
	}
	for _, detail := range nonEmptyLines(joinAddressParts(doc.Client.Address), clean(doc.Client.Email)) {
		leftParagraphs = append(leftParagraphs, paragraph(detail, paragraphOptions{spacingAfter: 20}))
	}

	rightParagraphs := []string{
		paragraph("ISSUED BY", paragraphOptions{bold: true, size: 20, spacingAfter: 20}),
		paragraph(defaultText(clean(doc.Issuer.CompanyName)), paragraphOptions{bold: true, size: 22, spacingAfter: 40}),
	}
	for _, detail := range nonEmptyLines(
		joinAddressParts(doc.Issuer.CompanyAddress),
		clean(doc.Issuer.Email),
		clean(doc.Issuer.Phone),
	) {
		rightParagraphs = append(rightParagraphs, paragraph(detail, paragraphOptions{spacingAfter: 20}))
	}

	var b strings.Builder
	b.WriteString(`<w:tbl>`)
	b.WriteString(`<w:tblPr><w:tblW w:w="0" w:type="auto"/><w:tblLayout w:type="fixed"/>` + tableBordersXML() + tableCellMarginsXML() + `</w:tblPr>`)
	b.WriteString(`<w:tblGrid><w:gridCol w:w="4680"/><w:gridCol w:w="4680"/></w:tblGrid>`)
	b.WriteString(`<w:tr>`)
	b.WriteString(tableCellXML(leftParagraphs, 4680, false, false, 1))
	b.WriteString(tableCellXML(rightParagraphs, 4680, false, false, 1))
	b.WriteString(`</w:tr></w:tbl>`)

	return b.String()
}

func lineItemsTableXML(doc models.InvoicePDFData) string {
	widths := []int{3600, 700, 900, 1500, 1300, 1500}

	var b strings.Builder
	b.WriteString(`<w:tbl>`)
	b.WriteString(`<w:tblPr>`)
	b.WriteString(`<w:tblW w:w="0" w:type="auto"/>`)
	b.WriteString(`<w:tblLayout w:type="fixed"/>`)
	b.WriteString(tableBordersXML())
	b.WriteString(tableCellMarginsXML())
	b.WriteString(`</w:tblPr>`)
	b.WriteString(`<w:tblGrid>`)
	for _, width := range widths {
		b.WriteString(fmt.Sprintf(`<w:gridCol w:w="%d"/>`, width))
	}
	b.WriteString(`</w:tblGrid>`)

	b.WriteString(`<w:tr>`)
	for idx, heading := range []string{"Description", "Qty", "Time", "Rate", "Price", "Amount"} {
		align := "left"
		if idx > 0 {
			align = "center"
		}
		if idx >= 3 {
			align = "right"
		}
		b.WriteString(tableCellXML(
			[]string{paragraph(heading, paragraphOptions{bold: true, size: 20, align: align})},
			widths[idx],
			true,
			false,
			1,
		))
	}
	b.WriteString(`</w:tr>`)

	groups := groupInvoiceLines(doc.Lines)
	totalWidth := 0
	for _, width := range widths {
		totalWidth += width
	}

	hasLines := false
	for _, group := range groups {
		if len(group.Lines) == 0 {
			continue
		}
		hasLines = true

		if doc.ShowItemTypeHeaders {
			b.WriteString(`<w:tr>`)
			b.WriteString(tableCellXML(
				[]string{paragraph(strings.ToUpper(group.Title), paragraphOptions{bold: true, size: 20})},
				totalWidth,
				false,
				true,
				len(widths),
			))
			b.WriteString(`</w:tr>`)
		}

		for _, line := range group.Lines {
			b.WriteString(`<w:tr>`)
			b.WriteString(tableCellXML([]string{paragraph(defaultText(clean(line.Name)), paragraphOptions{})}, widths[0], false, false, 1))
			b.WriteString(tableCellXML([]string{paragraph(defaultText(clean(line.Quantity)), paragraphOptions{align: "center"})}, widths[1], false, false, 1))
			b.WriteString(tableCellXML([]string{paragraph(defaultText(clean(line.TimeWorked)), paragraphOptions{align: "center"})}, widths[2], false, false, 1))
			b.WriteString(tableCellXML([]string{paragraph(defaultText(clean(line.HourlyRate)), paragraphOptions{align: "right"})}, widths[3], false, false, 1))
			b.WriteString(tableCellXML([]string{paragraph(defaultText(clean(line.ItemPrice)), paragraphOptions{align: "right"})}, widths[4], false, false, 1))
			b.WriteString(tableCellXML([]string{paragraph(defaultText(clean(line.ItemTotal)), paragraphOptions{align: "right"})}, widths[5], false, false, 1))
			b.WriteString(`</w:tr>`)
		}
	}

	if !hasLines {
		b.WriteString(`<w:tr>`)
		b.WriteString(tableCellXML(
			[]string{paragraph("No line items.", paragraphOptions{})},
			totalWidth,
			false,
			false,
			len(widths),
		))
		b.WriteString(`</w:tr>`)
	}

	b.WriteString(`</w:tbl>`)
	return b.String()
}

func summaryTableXML(doc models.InvoicePDFData) string {
	rows := buildSummaryRows(doc)
	widths := []int{4200, 1800}

	var b strings.Builder
	b.WriteString(`<w:tbl>`)
	b.WriteString(`<w:tblPr>`)
	b.WriteString(`<w:tblW w:w="0" w:type="auto"/>`)
	b.WriteString(`<w:tblLayout w:type="fixed"/>`)
	b.WriteString(tableBordersXML())
	b.WriteString(tableCellMarginsXML())
	b.WriteString(`</w:tblPr>`)
	b.WriteString(fmt.Sprintf(`<w:tblGrid><w:gridCol w:w="%d"/><w:gridCol w:w="%d"/></w:tblGrid>`, widths[0], widths[1]))

	for _, row := range rows {
		b.WriteString(`<w:tr>`)
		b.WriteString(tableCellXML(
			[]string{paragraph(row.label, paragraphOptions{bold: row.highlight})},
			widths[0],
			row.highlight,
			false,
			1,
		))
		b.WriteString(tableCellXML(
			[]string{paragraph(row.value, paragraphOptions{align: "right", bold: row.highlight})},
			widths[1],
			row.highlight,
			false,
			1,
		))
		b.WriteString(`</w:tr>`)
	}

	b.WriteString(`</w:tbl>`)
	return b.String()
}

func tableCellXML(
	contents []string,
	width int,
	shaded bool,
	borderless bool,
	gridSpan int,
) string {
	var b strings.Builder

	b.WriteString(`<w:tc><w:tcPr>`)
	b.WriteString(fmt.Sprintf(`<w:tcW w:w="%d" w:type="dxa"/>`, width))
	b.WriteString(`<w:vAlign w:val="top"/>`)
	if shaded {
		b.WriteString(`<w:shd w:val="clear" w:color="auto" w:fill="F3F4F6"/>`)
	}
	if borderless {
		b.WriteString(`<w:tcBorders><w:top w:val="nil"/><w:left w:val="nil"/><w:bottom w:val="nil"/><w:right w:val="nil"/></w:tcBorders>`)
	}
	if gridSpan > 1 {
		b.WriteString(fmt.Sprintf(`<w:gridSpan w:val="%d"/>`, gridSpan))
	}
	b.WriteString(`</w:tcPr>`)

	if len(contents) == 0 {
		contents = []string{paragraph(" ", paragraphOptions{})}
	}

	for _, content := range contents {
		b.WriteString(content)
	}

	b.WriteString(`</w:tc>`)
	return b.String()
}

func tableBordersXML() string {
	return `<w:tblBorders>
<w:top w:val="single" w:sz="4" w:space="0" w:color="D4D4D8"/><w:left w:val="single" w:sz="4" w:space="0" w:color="D4D4D8"/><w:bottom w:val="single" w:sz="4" w:space="0" w:color="D4D4D8"/><w:right w:val="single" w:sz="4" w:space="0" w:color="D4D4D8"/>
<w:insideH w:val="single" w:sz="4" w:space="0" w:color="E4E4E7"/>
<w:insideV w:val="single" w:sz="4" w:space="0" w:color="E4E4E7"/>
</w:tblBorders>`
}

func whiteTableXML() string {
	return `<w:tblBorders><w:top w:val="single" w:sz="4" w:space="0" w:color="FFFFFF"/><w:left w:val="single" w:sz="4" w:space="0" w:color="FFFFFF"/><w:bottom w:val="single" w:sz="4" w:space="0" w:color="FFFFFF"/><w:right w:val="single" w:sz="4" w:space="0" w:color="FFFFFF"/><w:insideH w:val="single" w:sz="4" w:space="0" w:color="FFFFFF"/><w:insideV w:val="single" w:sz="4" w:space="0" w:color="FFFFFF"/></w:tblBorders>`
}

func pageNumberParagraphXML() string {
	return `<w:p><w:pPr><w:jc w:val="right"/><w:shd w:val="clear" w:color="auto" w:fill="FFFFFF"/></w:pPr><w:r><w:rPr><w:shd w:val="clear" w:color="auto" w:fill="FFFFFF"/></w:rPr><w:t xml:space="preserve">Page </w:t></w:r><w:fldSimple w:instr=" PAGE "><w:r><w:rPr><w:shd w:val="clear" w:color="auto" w:fill="FFFFFF"/></w:rPr><w:t>1</w:t></w:r></w:fldSimple></w:p>`
}

func tableCellMarginsXML() string {
	return `<w:tblCellMar>
<w:top w:w="100" w:type="dxa"/>
<w:left w:w="120" w:type="dxa"/>
<w:bottom w:w="100" w:type="dxa"/>
<w:right w:w="120" w:type="dxa"/>
</w:tblCellMar>`
}

func buildSummaryRows(doc models.InvoicePDFData) []summaryRow {
	rows := []summaryRow{
		{label: "Subtotal", value: formatMoney(doc.Totals.SubtotalMinor, doc.Currency)},
		{label: "VAT", value: formatMoney(doc.Totals.VatAmountMinor, doc.Currency)},
		{label: "Total", value: formatMoney(doc.Totals.TotalMinor, doc.Currency)},
	}

	if doc.Totals.DiscountMinor > 0 {
		rows = append(rows[:1], append([]summaryRow{
			{label: "Discount", value: formatMoney(-doc.Totals.DiscountMinor, doc.Currency)},
		}, rows[1:]...)...)
	}
	if doc.Totals.DepositMinor > 0 {
		rows = append(rows, summaryRow{label: "Deposit", value: formatMoney(-doc.Totals.DepositMinor, doc.Currency)})
	}
	if doc.Totals.PaidMinor > 0 {
		rows = append(rows, summaryRow{label: "Paid", value: formatMoney(-doc.Totals.PaidMinor, doc.Currency)})
	}

	rows = append(rows, summaryRow{
		label:     "Balance Due",
		value:     formatMoney(doc.Totals.BalanceDue, doc.Currency),
		highlight: true,
	})

	return rows
}

func groupInvoiceLines(lines []models.InvoicePDFItem) []itemGroup {
	sorted := append([]models.InvoicePDFItem(nil), lines...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].SortOrder < sorted[j].SortOrder
	})

	groups := []itemGroup{
		{Title: "Styles"},
		{Title: "Samples"},
		{Title: "Other Items"},
	}

	for _, line := range sorted {
		switch strings.TrimSpace(strings.ToLower(line.LineType)) {
		case "style":
			groups[0].Lines = append(groups[0].Lines, line)
		case "sample":
			groups[1].Lines = append(groups[1].Lines, line)
		default:
			groups[2].Lines = append(groups[2].Lines, line)
		}
	}

	return groups
}

func loadEmbeddedLogo(logoURL string) *embeddedImage {
	logoPath, ok := resolveLocalLogoPath(logoURL)
	if !ok {
		return nil
	}

	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(logoPath)), ".")
	contentType, ok := imageContentType(ext)
	if !ok {
		return nil
	}

	data, err := os.ReadFile(logoPath)
	if err != nil {
		return nil
	}

	widthEMU, heightEMU := imageSizeEMU(data)
	return &embeddedImage{
		archivePath: "word/media/" + defaultLogoName + "." + ext,
		fileName:    defaultLogoName + "." + ext,
		extension:   ext,
		contentType: contentType,
		data:        data,
		widthEMU:    widthEMU,
		heightEMU:   heightEMU,
	}
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

func imageContentType(ext string) (string, bool) {
	switch ext {
	case "png":
		return "image/png", true
	case "jpg", "jpeg":
		return "image/jpeg", true
	case "gif":
		return "image/gif", true
	default:
		return "", false
	}
}

func imageSizeEMU(data []byte) (int64, int64) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width < 1 || cfg.Height < 1 {
		return defaultLogoWidthEMU, defaultLogoHeightEMU
	}

	widthEMU := int64(cfg.Width) * emuPerPixel
	heightEMU := int64(cfg.Height) * emuPerPixel

	if widthEMU <= maxLogoWidthEMU && heightEMU <= maxLogoHeightEMU {
		return widthEMU, heightEMU
	}

	widthRatio := float64(maxLogoWidthEMU) / float64(widthEMU)
	heightRatio := float64(maxLogoHeightEMU) / float64(heightEMU)
	scale := minFloat(widthRatio, heightRatio)
	if scale <= 0 {
		return defaultLogoWidthEMU, defaultLogoHeightEMU
	}

	scaledWidth := int64(float64(widthEMU) * scale)
	scaledHeight := int64(float64(heightEMU) * scale)
	if scaledWidth < 1 || scaledHeight < 1 {
		return defaultLogoWidthEMU, defaultLogoHeightEMU
	}

	return scaledWidth, scaledHeight
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func metaLine(label, value string) string {
	value = clean(value)
	if value == "" {
		return ""
	}
	return label + " " + value
}

func nonEmptyLines(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = clean(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func linesOf(v string) []string {
	v = strings.ReplaceAll(v, "\r\n", "\n")
	v = strings.ReplaceAll(v, "\r", "\n")
	v = strings.ReplaceAll(v, "\\n", "\n")

	parts := strings.Split(strings.TrimSpace(v), "\n")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
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

func defaultText(v string) string {
	if v == "" {
		return " "
	}
	return v
}

func clean(v string) string {
	return strings.TrimSpace(v)
}

func cleanPtr(v *string) string {
	if v == nil {
		return ""
	}
	return clean(*v)
}

func escapeXML(v string) string {
	var b bytes.Buffer
	_ = xml.EscapeText(&b, []byte(v))
	return b.String()
}

func formatMoney(minorUnits int64, currency string) string {
	sign := ""
	if minorUnits < 0 {
		sign = "-"
		minorUnits = -minorUnits
	}

	symbol := currencySymbol(currency)
	major := minorUnits / 100
	minor := minorUnits % 100

	return fmt.Sprintf("%s%s%d.%02d", sign, symbol, major, minor)
}

func currencySymbol(code string) string {
	switch code {
	case "EUR":
		return "€"
	case "USD":
		return "$"
	default:
		return "£"
	}
}
