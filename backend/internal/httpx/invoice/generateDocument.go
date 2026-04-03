package invoice

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
)

func validateQuickDocumentInvoice(
	dtoInvoice models.FEInvoiceIn,
	clientID int64,
	baseNumber int64,
) (models.FEInvoiceIn, []res.FieldError) {
	var routeErrs []res.FieldError
	if dtoInvoice.Overview.ClientID != clientID {
		routeErrs = append(routeErrs, res.Invalid("clientId", "does not match route parameter"))
	}
	if dtoInvoice.Overview.BaseNumber != baseNumber {
		routeErrs = append(routeErrs, res.Invalid("baseNumber", "does not match route parameter"))
	}
	if len(routeErrs) > 0 {
		return models.FEInvoiceIn{}, routeErrs
	}

	validInvoice, errs := ValidateInvoiceCreate(dtoInvoice)
	if len(errs) > 0 {
		return models.FEInvoiceIn{}, errs
	}

	canonical := RecalcInvoice(validInvoice)
	if errs := verifyTotalsMatch(validInvoice.Totals, canonical.Totals); len(errs) > 0 {
		return models.FEInvoiceIn{}, errs
	}
	if errs := ValidatePaidVsDepositTotal(canonical.Totals); len(errs) > 0 {
		return models.FEInvoiceIn{}, errs
	}

	return canonical, nil
}

type invoiceDocumentBuilder func() (models.InvoicePDFData, error)
type invoiceDocumentRenderer func(models.InvoicePDFData) ([]byte, error)

func handleInvoiceFileGeneration(
	w http.ResponseWriter,
	r *http.Request,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
	format string,
	contentType string,
	filename string,
	builder invoiceDocumentBuilder,
	renderer invoiceDocumentRenderer,
) {
	doc, err := builder()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			res.Error(w, http.StatusNotFound, "INVOICE_NOT_FOUND", "Invoice revision not found")
			return
		}

		slog.ErrorContext(r.Context(),
			"build invoice download data failed",
			"format", format,
			"client_id", clientID,
			"base_number", baseNumber,
			"revision_no", revisionNo,
			"err", err,
		)

		res.Error(w, http.StatusInternalServerError, "INTERNAL", "Internal server error")
		return
	}

	fileBytes, err := renderer(doc)
	if err != nil {
		slog.ErrorContext(r.Context(),
			"generate invoice file failed",
			"format", format,
			"client_id", clientID,
			"base_number", baseNumber,
			"revision_no", revisionNo,
			"err", err,
		)

		formatUpper := strings.ToUpper(format)
		res.Error(
			w,
			http.StatusInternalServerError,
			formatUpper+"_GENERATION_FAILED",
			fmt.Sprintf("Failed to generate %s", formatUpper),
		)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(fileBytes); err != nil {
		slog.ErrorContext(r.Context(),
			"write invoice file response failed",
			"format", format,
			"err", err,
		)
	}
}

func buildDocumentFilename(baseNumber int64, revisionNo int64, ext string) string {
	ext = strings.TrimPrefix(strings.TrimSpace(ext), ".")
	if ext == "" {
		ext = "bin"
	}

	if baseNumber < 1 {
		return "Invoice." + ext
	}
	if revisionNo <= 1 {
		return fmt.Sprintf("Invoice-%d.%s", baseNumber, ext)
	}

	return fmt.Sprintf("Invoice-%d-Rev-%d.%s", baseNumber, revisionNo-1, ext)
}
