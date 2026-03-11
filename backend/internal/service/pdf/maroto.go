package pdf

import (
	"context"
	"fmt"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/image"
	"github.com/johnfercher/maroto/v2/pkg/components/line"
	"github.com/johnfercher/maroto/v2/pkg/components/list"
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

	addHeader(mr)
	addInvoiceDetails(mr, doc)
	if err := addItemList(mr, doc.Lines); err != nil {
		return nil, fmt.Errorf("pdf generation failed: %w", err)
	}
	addFooter(mr)

	d, err := mr.Generate()
	if err != nil {
		return nil, fmt.Errorf("maroto generate: %w", err)
	}
	return d.GetBytes(), nil
}

func addHeader(mr core.Maroto) {
	mr.AddRow(50,
		image.NewFromFileCol(12, "assets/goInvoicerMascot.png",
			props.Rect{
				Center:  true,
				Percent: 75,
			},
		),
	)

	mr.AddRow(20,
		text.NewCol(12, "S.A.M.",
			props.Text{
				Top:   5,
				Style: fontstyle.Bold,
				Align: align.Center,
				Size:  16,
			},
		),
	)

	mr.AddRow(20,
		text.NewCol(12, "Invoice", props.Text{
			Top:   5,
			Style: fontstyle.Bold,
			Align: align.Center,
			Size:  12,
		}),
	)
}

func addInvoiceDetails(mr core.Maroto, doc models.InvoicePDFData) {
	invoiceLabel := fmt.Sprintf("Invoice #%d.%s", doc.BaseNumber, doc.RevisionNumber)

	mr.AddRow(10,
		text.NewCol(6, "Date: "+doc.IssueAt, props.Text{
			Align: align.Left,
			Size:  10,
		}),
		text.NewCol(6, invoiceLabel, props.Text{
			Align: align.Right,
			Size:  10,
		}),
	)
	mr.AddRow(40, line.NewCol(12))
}

// Maroto-specific adapter - embeds the canonical model to satisfy list.Listable.
type itemLine struct {
	models.InvoicePDFItem
}

func (o itemLine) GetHeader() core.Row {
	return row.New(10).Add(
		text.NewCol(3, "Name", props.Text{Style: fontstyle.Bold}),
		text.NewCol(3, "Type", props.Text{Style: fontstyle.Bold}),
		text.NewCol(2, "Quantity", props.Text{Style: fontstyle.Bold}),
		text.NewCol(2, "Unit Price", props.Text{Style: fontstyle.Bold}),
		text.NewCol(2, "Total", props.Text{Style: fontstyle.Bold}),
	)
}

func (o itemLine) GetContent(i int) core.Row {
	r := row.New(5).Add(
		text.NewCol(3, o.Name),
		text.NewCol(3, o.LineType),
		text.NewCol(2, o.Quantity),
		text.NewCol(2, o.ItemPrice),
		text.NewCol(2, o.ItemTotal),
	)
	if i%2 != 0 {
		r.WithStyle(&props.Cell{
			BackgroundColor: &props.Color{Red: 240, Green: 240, Blue: 240},
		})
	}
	return r
}

func addItemList(mr core.Maroto, lines []models.InvoicePDFItem) error {
	if len(lines) == 0 {
		return fmt.Errorf("invoice has no line items")
	}

	adapted := make([]itemLine, len(lines))
	for i, l := range lines {
		adapted[i] = itemLine{InvoicePDFItem: l}
	}

	rows, err := list.Build(adapted)
	if err != nil {
		return fmt.Errorf("render item rows: %w", err)
	}

	mr.AddRows(rows...)
	return nil
}

func addFooter(mr core.Maroto) {
	mr.AddRow(15,
		text.NewCol(6, "Subtotal", props.Text{
			Align: align.Center,
			Size:  10,
		}),
		text.NewCol(6, "VAT", props.Text{
			Align: align.Center,
			Size:  10,
		}),
		text.NewCol(6, "Total", props.Text{
			Align: align.Center,
			Size:  10,
		}),
	)
	mr.AddRow(15, line.NewCol(12))
	mr.AddRow(15,
		text.NewCol(6, "Thank you for your business!", props.Text{
			Align: align.Center,
			Size:  10,
		}),
	)

}
