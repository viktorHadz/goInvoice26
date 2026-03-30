package invoice

import (
	"fmt"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/service/pdf"
	"github.com/viktorHadz/goInvoice26/internal/transaction/settingsTx"
)

func GeneratePDFHandler(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}

		baseNumber, ok := params.ValidateParam(w, r, "baseNumber")
		if !ok {
			return
		}

		revisionNo, ok := params.ValidateParam(w, r, "revisionNo")
		if !ok {
			return
		}

		builder := func() (models.InvoicePDFData, error) {
			return pdf.BuildInvoiceFromDB(r.Context(), a.DB, clientID, baseNumber, revisionNo)
		}

		handleInvoiceFileGeneration(
			w,
			r,
			clientID,
			baseNumber,
			revisionNo,
			"pdf",
			"application/pdf",
			buildPDFFilename(baseNumber, revisionNo),
			builder,
			func(doc models.InvoicePDFData) ([]byte, error) {
				return pdf.RenderPDF(r.Context(), &pdf.MarotoRenderer{}, doc)
			},
		)
	}
}

func QuickPDFHandler(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}

		baseNumber, ok := params.ValidateParam(w, r, "baseNumber")
		if !ok {
			return
		}

		revisionNo, ok := params.ValidateParam(w, r, "revisionNo")
		if !ok {
			return
		}

		var dtoInvoice models.FEInvoiceIn
		if ok := res.DecodeJSON(w, r, &dtoInvoice); !ok {
			return
		}

		var routeErrs []res.FieldError
		if dtoInvoice.Overview.ClientID != clientID {
			routeErrs = append(routeErrs, res.Invalid("clientId", "does not match route parameter"))
		}
		if dtoInvoice.Overview.BaseNumber != baseNumber {
			routeErrs = append(routeErrs, res.Invalid("baseNumber", "does not match route parameter"))
		}
		if len(routeErrs) > 0 {
			res.Validation(w, routeErrs...)
			return
		}

		canonical := RecalcInvoice(dtoInvoice)
		if errs := verifyTotalsMatch(dtoInvoice.Totals, canonical.Totals); len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}
		if errs := ValidatePaidVsDepositTotal(canonical.Totals); len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}

		builder := func() (models.InvoicePDFData, error) {
			settings, err := settingsTx.Get(r.Context(), a.DB, accountscope.AccountID(r.Context()))
			if err != nil {
				return models.InvoicePDFData{}, fmt.Errorf("get settings: %w", err)
			}

			return pdf.BuildQuickInvoice(canonical, settings, revisionNo), nil
		}

		handleInvoiceFileGeneration(
			w,
			r,
			clientID,
			baseNumber,
			revisionNo,
			"pdf",
			"application/pdf",
			buildPDFFilename(baseNumber, revisionNo),
			builder,
			func(doc models.InvoicePDFData) ([]byte, error) {
				return pdf.RenderPDF(r.Context(), &pdf.MarotoRenderer{}, doc)
			},
		)
	}
}

func buildPDFFilename(baseNumber int64, revisionNo int64) string {
	return buildDocumentFilename(baseNumber, revisionNo, "pdf")
}
