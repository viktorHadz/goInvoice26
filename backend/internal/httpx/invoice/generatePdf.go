package invoice

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/service/pdf"
)

type pdfBuilder func() (models.InvoicePDFData, error)

func handlePdfGeneration(
	w http.ResponseWriter,
	r *http.Request,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
	builder pdfBuilder,
) {
	doc, err := builder()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			res.Error(w, http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice revision not found")
			return
		}

		slog.ErrorContext(r.Context(),
			"build invoice pdf data failed",
			"client_id", clientID,
			"base_number", baseNumber,
			"revision_no", revisionNo,
			"err", err,
		)

		res.Error(w, http.StatusInternalServerError, "INTERNAL", "Internal server error")
		return
	}

	pdfBytes, err := pdf.RenderPDF(r.Context(), &pdf.MarotoRenderer{}, doc)
	if err != nil {
		slog.ErrorContext(r.Context(),
			"generate invoice pdf failed",
			"client_id", clientID,
			"base_number", baseNumber,
			"revision_no", revisionNo,
			"err", err,
		)

		res.Error(w, http.StatusInternalServerError, "PDF_GENERATION_FAILED", "Failed to generate PDF")
		return
	}

	filename := fmt.Sprintf("invoice-%d.%s.pdf", doc.BaseNumber, doc.RevisionNumber)

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(pdfBytes); err != nil {
		slog.ErrorContext(r.Context(),
			"write pdf response failed",
			"err", err,
		)
	}
}

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

		handlePdfGeneration(w, r, clientID, baseNumber, revisionNo, builder)
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

		// Verify route params match invoice data
		var routeErrs []res.FieldError
		if dtoInvoice.Overview.ClientID != clientID {
			routeErrs = append(routeErrs, res.Invalid("clientId", "does not match route param"))
		}
		if dtoInvoice.Overview.BaseNumber != baseNumber {
			routeErrs = append(routeErrs, res.Invalid("baseNumber", "does not match route param"))
		}
		if len(routeErrs) > 0 {
			res.Validation(w, routeErrs...)
			return
		}

		// Recalculate to verify totals match
		canonical := RecalcInvoice(dtoInvoice)
		if errs := verifyTotalsMatch(dtoInvoice.Totals, canonical.Totals); len(errs) > 0 {
			res.Validation(w, errs...)
			return
		}

		builder := func() (models.InvoicePDFData, error) {
			return pdf.BuildQuickInvoice(canonical, revisionNo), nil
		}

		handlePdfGeneration(w, r, clientID, baseNumber, revisionNo, builder)
	}
}
