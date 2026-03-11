package invoice

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/service/pdf"
)

func GeneratePDF(a *app.App) http.HandlerFunc {
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

		doc, err := pdf.BuildInvoicePDFData(r.Context(), a.DB, clientID, baseNumber, revisionNo)
		if err != nil {
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

		pdfBytes, err := pdf.GenerateInvoicePDF(r.Context(), &pdf.MarotoRenderer{}, doc)
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
		w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filename))
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write(pdfBytes); err != nil {
			slog.ErrorContext(r.Context(),
				"write pdf response failed",
				"err", err,
			)
		}
	}
}
