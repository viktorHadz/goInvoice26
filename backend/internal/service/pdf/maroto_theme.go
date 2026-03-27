package pdf

import (
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/props"
)

var invoiceTheme = newMarotoTheme()

type marotoTheme struct {
	text struct {
		title        props.Text
		documentNo   props.Text
		meta         props.Text
		sectionLabel props.Text
		partyLabel   props.Text
		partyName    props.Text
		partyBody    props.Text
		tableHeader  props.Text
		tableCell    props.Text
		emptyState   props.Text
		noteLabel    props.Text
		noteBody     props.Text
		totalLabel   props.Text
		totalValue   props.Text
		balanceLabel props.Text
		balanceValue props.Text
		paymentLabel props.Text
		paymentBody  props.Text
		footer       props.Text
	}
	line struct {
		soft    props.Line
		divider props.Line
		accent  props.Line
	}
	cell struct {
		party       *props.Cell
		tableHeader *props.Cell
		total       *props.Cell
		balance     *props.Cell
		note        *props.Cell
		payment     *props.Cell
	}
	space struct {
		xxs float64
		xs  float64
		sm  float64
		md  float64
		lg  float64
		xl  float64
	}
	row struct {
		headerLogo   float64
		headerText   float64
		headerMeta   float64
		sectionLabel float64
		tableHeader  float64
		groupLabel   float64
		footerRule   float64
	}
}

func newMarotoTheme() marotoTheme {
	t := marotoTheme{}

	inkStrong := &props.Color{Red: 18, Green: 18, Blue: 18}
	inkBody := &props.Color{Red: 48, Green: 48, Blue: 48}
	inkMuted := &props.Color{Red: 80, Green: 80, Blue: 80}
	ruleSoft := &props.Color{Red: 216, Green: 216, Blue: 216}
	ruleStrong := &props.Color{Red: 188, Green: 188, Blue: 188}
	panelSoft := &props.Color{Red: 244, Green: 244, Blue: 244}
	panelStrong := &props.Color{Red: 235, Green: 235, Blue: 235}

	t.text.title = props.Text{
		Size:  25,
		Style: fontstyle.Bold,
		Color: inkStrong,
		Top:   1.2,
	}
	t.text.documentNo = props.Text{
		Size:  12,
		Color: inkBody,
		Style: fontstyle.Bold,
		Top:   0.8,
	}
	t.text.meta = props.Text{
		Size:  8.8,
		Color: inkBody,
		Top:   0.7,
	}
	t.text.sectionLabel = props.Text{
		Size:  7.8,
		Style: fontstyle.Bold,
		Color: inkBody,
		Top:   0.5,
	}
	t.text.partyLabel = props.Text{
		Size:   7.2,
		Style:  fontstyle.Bold,
		Color:  inkMuted,
		Top:    2.2,
		Bottom: 0.4,
		Left:   3,
		Right:  3,
	}
	t.text.partyName = props.Text{
		Size:   11,
		Style:  fontstyle.Bold,
		Color:  inkStrong,
		Top:    0.4,
		Bottom: 0.7,
		Left:   3,
		Right:  3,
	}
	t.text.partyBody = props.Text{
		Size:   8.5,
		Color:  inkBody,
		Top:    0.6,
		Bottom: 0.8,
		Left:   3,
		Right:  3,
	}
	t.text.tableHeader = props.Text{
		Size:   7.7,
		Style:  fontstyle.Bold,
		Color:  inkBody,
		Top:    2,
		Bottom: 1.3,
	}
	t.text.tableCell = props.Text{
		Size:   9.1,
		Color:  inkStrong,
		Top:    1.8,
		Bottom: 1.4,
	}
	t.text.emptyState = props.Text{
		Size:   9,
		Color:  inkMuted,
		Top:    2.5,
		Bottom: 2.2,
		Align:  align.Center,
	}
	t.text.noteLabel = props.Text{
		Size:   7.4,
		Style:  fontstyle.Bold,
		Color:  inkBody,
		Top:    0,
		Bottom: 1,
		Align:  align.Left,
	}
	t.text.noteBody = props.Text{
		Size:   8.8,
		Color:  inkBody,
		Top:    0.8,
		Bottom: 0.9,
		Left:   3,
		Right:  3,
	}
	t.text.totalLabel = props.Text{
		Size:   8.8,
		Color:  inkBody,
		Top:    1.4,
		Bottom: 1.2,
		Left:   1.5,
		Right:  1.5,
	}
	t.text.totalValue = props.Text{
		Size:   8.8,
		Color:  inkStrong,
		Top:    1.4,
		Bottom: 1.2,
		Left:   1,
		Right:  2.4,
	}
	t.text.balanceLabel = props.Text{
		Size:   10.4,
		Style:  fontstyle.Bold,
		Color:  inkStrong,
		Top:    1.8,
		Bottom: 1.8,
		Left:   1.5,
		Right:  1.5,
	}
	t.text.balanceValue = props.Text{
		Size:   10.4,
		Style:  fontstyle.Bold,
		Color:  inkStrong,
		Top:    1.8,
		Bottom: 1.8,
		Left:   1,
		Right:  2.4,
	}
	t.text.paymentLabel = props.Text{
		Size:   7.4,
		Style:  fontstyle.Bold,
		Color:  inkBody,
		Top:    2.1,
		Bottom: 0.4,
		Left:   3,
		Right:  3,
	}
	t.text.paymentBody = props.Text{
		Size:   8.8,
		Color:  inkBody,
		Top:    0.8,
		Bottom: 0.9,
		Left:   3,
		Right:  3,
	}
	t.text.footer = props.Text{
		Size:  7.3,
		Align: align.Center,
		Color: inkBody,
		Top:   0.9,
	}

	t.line.soft = props.Line{Color: ruleSoft, Thickness: 0.14, OffsetPercent: 50, SizePercent: 100}
	t.line.divider = props.Line{Color: ruleSoft, Thickness: 0.24, OffsetPercent: 50, SizePercent: 100}
	t.line.accent = props.Line{Color: ruleStrong, Thickness: 0.4, OffsetPercent: 50, SizePercent: 100}

	t.cell.party = &props.Cell{BackgroundColor: panelSoft}
	t.cell.tableHeader = &props.Cell{BackgroundColor: panelSoft}
	t.cell.total = &props.Cell{BackgroundColor: panelSoft}
	t.cell.balance = &props.Cell{BackgroundColor: panelStrong}
	t.cell.note = &props.Cell{BackgroundColor: panelSoft}
	t.cell.payment = &props.Cell{BackgroundColor: panelSoft}

	t.space.xxs = 1.2
	t.space.xs = 2
	t.space.sm = 3
	t.space.md = 4
	t.space.lg = 6
	t.space.xl = 8

	t.row.headerLogo = 24
	t.row.headerText = 5.4
	t.row.headerMeta = 4.8
	t.row.sectionLabel = 4.5
	t.row.tableHeader = 7.5
	t.row.groupLabel = 5
	t.row.footerRule = 1.6

	return t
}

func (t marotoTheme) titleText(a align.Type) props.Text {
	s := t.text.title
	s.Align = a
	return s
}

func (t marotoTheme) documentNoText(a align.Type) props.Text {
	s := t.text.documentNo
	s.Align = a
	return s
}

func (t marotoTheme) metaText(a align.Type) props.Text {
	s := t.text.meta
	s.Align = a
	return s
}

func (t marotoTheme) sectionLabelText(a align.Type) props.Text {
	s := t.text.sectionLabel
	s.Align = a
	return s
}

func (t marotoTheme) tableHeaderText(a align.Type) props.Text {
	s := t.text.tableHeader
	s.Align = a
	return s
}

func (t marotoTheme) tableCellText(a align.Type) props.Text {
	s := t.text.tableCell
	s.Align = a
	return s
}

func (t marotoTheme) totalLabelText() props.Text {
	s := t.text.totalLabel
	s.Align = align.Right
	return s
}

func (t marotoTheme) totalValueText() props.Text {
	s := t.text.totalValue
	s.Align = align.Right
	return s
}

func (t marotoTheme) balanceLabelText() props.Text {
	s := t.text.balanceLabel
	s.Align = align.Right
	return s
}

func (t marotoTheme) balanceValueText() props.Text {
	s := t.text.balanceValue
	s.Align = align.Right
	return s
}
