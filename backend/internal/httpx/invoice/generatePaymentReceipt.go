package invoice

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/service/docx"
	"github.com/viktorHadz/goInvoice26/internal/service/pdf"
	"github.com/viktorHadz/goInvoice26/internal/transaction/invoiceTx"
)

func GeneratePaymentReceiptPDFHandler(a *app.App) http.HandlerFunc {
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
		receiptNo, ok := params.ValidateParam(w, r, "receiptNo")
		if !ok {
			return
		}

		doc, err := pdf.BuildPaymentReceiptFromDB(r.Context(), a.DB, clientID, baseNumber, revisionNo, receiptNo)
		if err != nil {
			handlePaymentReceiptDocumentBuildError(w, r, clientID, baseNumber, revisionNo, receiptNo, "pdf", err)
			return
		}

		fileBytes, err := pdf.RenderPDF(r.Context(), &pdf.MarotoRenderer{}, doc)
		if err != nil {
			handlePaymentReceiptDocumentRenderError(w, r, clientID, baseNumber, revisionNo, receiptNo, "PDF", err)
			return
		}

		writeGeneratedDocument(w, "application/pdf", buildPaymentReceiptFilename(baseNumber, revisionNo, receiptNo, "pdf"), fileBytes)
	}
}

func GeneratePaymentReceiptDOCXHandler(a *app.App) http.HandlerFunc {
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
		receiptNo, ok := params.ValidateParam(w, r, "receiptNo")
		if !ok {
			return
		}

		doc, err := pdf.BuildPaymentReceiptFromDB(r.Context(), a.DB, clientID, baseNumber, revisionNo, receiptNo)
		if err != nil {
			handlePaymentReceiptDocumentBuildError(w, r, clientID, baseNumber, revisionNo, receiptNo, "docx", err)
			return
		}

		fileBytes, err := docx.RenderDOCX(doc)
		if err != nil {
			handlePaymentReceiptDocumentRenderError(w, r, clientID, baseNumber, revisionNo, receiptNo, "DOCX", err)
			return
		}

		writeGeneratedDocument(w, docxContentType, buildPaymentReceiptFilename(baseNumber, revisionNo, receiptNo, "docx"), fileBytes)
	}
}

func handlePaymentReceiptDocumentBuildError(
	w http.ResponseWriter,
	r *http.Request,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
	receiptNo int64,
	format string,
	err error,
) {
	if errors.Is(err, invoiceTx.ErrPaymentReceiptNotFound) {
		res.Error(w, http.StatusNotFound, "PAYMENT_RECEIPT_NOT_FOUND", "Payment receipt not found")
		return
	}

	slog.ErrorContext(r.Context(),
		"build payment receipt download data failed",
		"format", format,
		"client_id", clientID,
		"base_number", baseNumber,
		"revision_no", revisionNo,
		"receipt_no", receiptNo,
		"err", err,
	)
	res.Error(w, http.StatusInternalServerError, "INTERNAL", "Internal server error")
}

func handlePaymentReceiptDocumentRenderError(
	w http.ResponseWriter,
	r *http.Request,
	clientID int64,
	baseNumber int64,
	revisionNo int64,
	receiptNo int64,
	formatUpper string,
	err error,
) {
	slog.ErrorContext(r.Context(),
		"generate payment receipt file failed",
		"format", formatUpper,
		"client_id", clientID,
		"base_number", baseNumber,
		"revision_no", revisionNo,
		"receipt_no", receiptNo,
		"err", err,
	)
	res.Error(
		w,
		http.StatusInternalServerError,
		formatUpper+"_GENERATION_FAILED",
		fmt.Sprintf("Failed to generate %s", formatUpper),
	)
}

func writeGeneratedDocument(w http.ResponseWriter, contentType string, filename string, fileBytes []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(fileBytes)
}
