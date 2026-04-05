package products

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/clientsTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/productsTx"
	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

const (
	maxProductImportFileSize = 50 * 1024
	maxProductImportRows     = 400
	maxImportKindFieldBytes  = 64
)

type productImportColumn struct {
	Field string
	Label string
}

type productImportKindSpec struct {
	Kind        string
	ProductType string
	PricingMode string
	Headers     []string
	Columns     map[string]productImportColumn
}

type productImportUpload struct {
	Kind                string
	FileName            string
	DeclaredContentType string
	Data                []byte
}

type productImportSummary struct {
	CreatedCount int    `json:"createdCount"`
	ClientID     int64  `json:"clientId"`
	ImportKind   string `json:"importKind"`
}

var productImportSpecs = map[string]productImportKindSpec{
	"style": {
		Kind:        "style",
		ProductType: "style",
		PricingMode: "flat",
		Headers:     []string{"name", "unit price"},
		Columns: map[string]productImportColumn{
			"productName": {Field: "name", Label: "name"},
			"flatPrice":   {Field: "unitPrice", Label: "unit price"},
		},
	},
	"sample_flat": {
		Kind:        "sample_flat",
		ProductType: "sample",
		PricingMode: "flat",
		Headers:     []string{"name", "unit price"},
		Columns: map[string]productImportColumn{
			"productName": {Field: "name", Label: "name"},
			"flatPrice":   {Field: "unitPrice", Label: "unit price"},
		},
	},
	"sample_hourly": {
		Kind:        "sample_hourly",
		ProductType: "sample",
		PricingMode: "hourly",
		Headers:     []string{"name", "time to produce (in minutes)", "unit price"},
		Columns: map[string]productImportColumn{
			"productName":   {Field: "name", Label: "name"},
			"minutesWorked": {Field: "timeToProduceMinutes", Label: "time to produce (in minutes)"},
			"hourlyRate":    {Field: "unitPrice", Label: "unit price"},
		},
	},
}

func ImportProducts(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}

		accountID, err := accountscope.Require(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "product import missing account scope", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to import products")
			return
		}

		principal, _ := userscope.PrincipalFromContext(r.Context())
		slog.InfoContext(r.Context(),
			"product import started",
			"account_id", accountID,
			"client_id", clientID,
			"user_id", principal.UserID,
		)

		if err := clientsTx.VerifyClientID(r.Context(), a, clientID); err != nil {
			if errors.Is(err, clientsTx.ErrClientNotFound) {
				slog.WarnContext(r.Context(),
					"product import rejected because client was not found",
					"account_id", accountID,
					"client_id", clientID,
					"user_id", principal.UserID,
				)
				res.NotFound(w, "client not found")
				return
			}

			slog.ErrorContext(r.Context(),
				"product import verify client failed",
				"account_id", accountID,
				"client_id", clientID,
				"user_id", principal.UserID,
				"err", err,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		if a.ProductImports == nil {
			slog.ErrorContext(r.Context(),
				"product import coordinator missing",
				"account_id", accountID,
				"client_id", clientID,
				"user_id", principal.UserID,
			)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Product import is unavailable")
			return
		}

		if !a.ProductImports.Acquire(accountID) {
			slog.WarnContext(r.Context(),
				"product import rejected because another import is active",
				"account_id", accountID,
				"client_id", clientID,
				"user_id", principal.UserID,
			)
			res.Error(
				w,
				http.StatusConflict,
				"PRODUCT_IMPORT_IN_PROGRESS",
				"Another product import is already running for this workspace. Please wait for it to finish.",
			)
			return
		}
		defer a.ProductImports.Release(accountID)

		upload, fieldErrs, err := readProductImportUpload(r)
		if err != nil {
			slog.ErrorContext(r.Context(),
				"product import request parse failed",
				"account_id", accountID,
				"client_id", clientID,
				"user_id", principal.UserID,
				"err", err,
			)
			res.Error(w, http.StatusBadRequest, "BAD_DATA", "Invalid multipart form data")
			return
		}
		if len(fieldErrs) > 0 {
			slog.WarnContext(r.Context(),
				"product import rejected during upload validation",
				"account_id", accountID,
				"client_id", clientID,
				"user_id", principal.UserID,
				"kind", upload.Kind,
				"field_error_count", len(fieldErrs),
			)
			res.Validation(w, fieldErrs...)
			return
		}

		spec, ok := productImportSpecs[upload.Kind]
		if !ok {
			errs := []res.FieldError{res.Invalid("kind", "must be style, sample_flat, or sample_hourly")}
			slog.WarnContext(r.Context(),
				"product import rejected because kind was invalid",
				"account_id", accountID,
				"client_id", clientID,
				"user_id", principal.UserID,
				"kind", upload.Kind,
			)
			res.Validation(w, errs...)
			return
		}

		rows, fieldErrs := parseImportedProducts(upload, clientID, spec)
		if len(fieldErrs) > 0 {
			slog.WarnContext(r.Context(),
				"product import rejected during csv validation",
				"account_id", accountID,
				"client_id", clientID,
				"user_id", principal.UserID,
				"kind", spec.Kind,
				"field_error_count", len(fieldErrs),
			)
			res.Validation(w, fieldErrs...)
			return
		}

		createdCount, err := productsTx.BulkInsertTx(a, r.Context(), rows)
		if err != nil {
			slog.ErrorContext(r.Context(),
				"product import bulk insert failed",
				"account_id", accountID,
				"client_id", clientID,
				"user_id", principal.UserID,
				"kind", spec.Kind,
				"row_count", len(rows),
				"err", err,
			)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		slog.InfoContext(r.Context(),
			"product import completed",
			"account_id", accountID,
			"client_id", clientID,
			"user_id", principal.UserID,
			"kind", spec.Kind,
			"row_count", len(rows),
			"created_count", createdCount,
		)

		res.JSON(w, http.StatusCreated, productImportSummary{
			CreatedCount: createdCount,
			ClientID:     clientID,
			ImportKind:   spec.Kind,
		})
	}
}

func readProductImportUpload(r *http.Request) (productImportUpload, []res.FieldError, error) {
	contentType := r.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil || mediaType != "multipart/form-data" {
		return productImportUpload{}, []res.FieldError{
			res.Invalid("file", "expected a multipart CSV upload"),
		}, nil
	}

	reader, err := r.MultipartReader()
	if err != nil {
		return productImportUpload{}, nil, err
	}

	var upload productImportUpload
	var errs []res.FieldError
	seenFields := map[string]int{}

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return productImportUpload{}, nil, err
		}

		field := strings.TrimSpace(part.FormName())
		seenFields[field]++

		switch field {
		case "kind":
			value, readErr := io.ReadAll(io.LimitReader(part, maxImportKindFieldBytes+1))
			if readErr != nil {
				return productImportUpload{}, nil, readErr
			}
			if len(value) > maxImportKindFieldBytes {
				errs = append(errs, res.Invalid("kind", "kind is too long"))
				continue
			}
			if seenFields[field] > 1 {
				errs = append(errs, res.Invalid("kind", "only one kind field is allowed"))
				continue
			}

			upload.Kind = strings.TrimSpace(strings.ToLower(string(value)))

		case "file":
			data, readErr := io.ReadAll(io.LimitReader(part, maxProductImportFileSize+1))
			if readErr != nil {
				return productImportUpload{}, nil, readErr
			}
			if seenFields[field] > 1 {
				errs = append(errs, res.Invalid("file", "only one file upload is allowed"))
				continue
			}
			if len(data) > maxProductImportFileSize {
				errs = append(errs, res.Invalid("file", "file must be 50KB or smaller"))
				continue
			}

			upload.FileName = strings.TrimSpace(part.FileName())
			upload.DeclaredContentType = strings.TrimSpace(part.Header.Get("Content-Type"))
			upload.Data = data

		case "":
			_, _ = io.Copy(io.Discard, part)

		default:
			_, _ = io.Copy(io.Discard, part)
			errs = append(errs, res.Invalid("request", fmt.Sprintf("unexpected multipart field %q", field)))
		}
	}

	if seenFields["kind"] == 0 {
		errs = append(errs, res.Required("kind"))
	}
	if seenFields["file"] == 0 {
		errs = append(errs, res.Required("file"))
	}
	if seenFields["kind"] == 1 && upload.Kind == "" {
		errs = append(errs, res.Required("kind"))
	}
	if seenFields["file"] == 1 && len(upload.Data) == 0 {
		errs = append(errs, res.Invalid("file", "uploaded CSV is empty"))
	}

	if upload.DeclaredContentType != "" && isClearlyUnsupportedImportMIME(upload.DeclaredContentType) {
		errs = append(errs, res.Invalid("file", "upload a CSV file"))
	}

	return upload, errs, nil
}

func parseImportedProducts(upload productImportUpload, clientID int64, spec productImportKindSpec) ([]models.ProductCreate, []res.FieldError) {
	if len(upload.Data) == 0 {
		return nil, []res.FieldError{res.Invalid("file", "uploaded CSV is empty")}
	}
	if !utf8.Valid(upload.Data) {
		return nil, []res.FieldError{res.Invalid("file", "CSV must be valid UTF-8 text")}
	}
	if containsDisallowedBinary(upload.Data) {
		return nil, []res.FieldError{res.Invalid("file", "CSV contains unsupported binary data")}
	}

	sniffedType := http.DetectContentType(upload.Data[:min(len(upload.Data), 512)])
	if isClearlyUnsupportedImportMIME(sniffedType) {
		return nil, []res.FieldError{res.Invalid("file", "upload a CSV file")}
	}

	reader := csv.NewReader(bytes.NewReader(upload.Data))
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if errors.Is(err, io.EOF) {
		return nil, []res.FieldError{res.Invalid("file", "CSV must include a header row and at least one data row")}
	}
	if err != nil {
		return nil, []res.FieldError{csvParseError(err)}
	}

	headerErrs := validateImportHeader(header, spec.Headers)
	if len(headerErrs) > 0 {
		return nil, headerErrs
	}

	rows := make([]models.ProductCreate, 0, maxProductImportRows)
	dataRowCount := 0

	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, []res.FieldError{csvParseError(err)}
		}

		dataRowCount++
		csvRowNumber := dataRowCount + 1
		if dataRowCount > maxProductImportRows {
			return nil, []res.FieldError{
				res.Invalid("file", fmt.Sprintf("CSV can contain at most %d data rows", maxProductImportRows)),
			}
		}
		if len(record) != len(spec.Headers) {
			return nil, []res.FieldError{
				rowGeneralError(csvRowNumber, fmt.Sprintf("expected %d columns but found %d", len(spec.Headers), len(record))),
			}
		}

		in := buildImportRowInput(spec, record)
		product, errs := ValidateCreate(in, clientID)
		if len(errs) > 0 {
			return nil, decorateImportRowErrors(csvRowNumber, spec, errs)
		}

		rows = append(rows, product)
	}

	if dataRowCount == 0 {
		return nil, []res.FieldError{res.Invalid("file", "CSV must include at least one data row")}
	}

	return rows, nil
}

func validateImportHeader(header []string, expected []string) []res.FieldError {
	normalized := make([]string, 0, len(header))
	seen := make(map[string]int, len(header))
	for idx, value := range header {
		name := normalizeImportHeader(value, idx == 0)
		normalized = append(normalized, name)
		if name != "" {
			seen[name]++
		}
	}

	var errs []res.FieldError
	for name, count := range seen {
		if count > 1 {
			errs = append(errs, res.Invalid("header", fmt.Sprintf("duplicate column %q", name)))
		}
	}

	expectedSet := make(map[string]struct{}, len(expected))
	for _, name := range expected {
		expectedSet[name] = struct{}{}
	}

	for _, name := range normalized {
		if _, ok := expectedSet[name]; !ok {
			errs = append(errs, res.Invalid("header", fmt.Sprintf("unknown column %q", name)))
		}
	}

	normalizedSet := make(map[string]struct{}, len(normalized))
	for _, name := range normalized {
		normalizedSet[name] = struct{}{}
	}
	for _, name := range expected {
		if _, ok := normalizedSet[name]; !ok {
			errs = append(errs, res.Invalid("header", fmt.Sprintf("missing column %q", name)))
		}
	}

	if len(normalized) != len(expected) || !equalStrings(normalized, expected) {
		errs = append(errs, res.Invalid("header", fmt.Sprintf("columns must exactly be: %s", strings.Join(expected, ", "))))
	}

	return errs
}

func buildImportRowInput(spec productImportKindSpec, record []string) models.ProductCreateIn {
	productType := spec.ProductType
	pricingMode := spec.PricingMode
	name := record[0]

	out := models.ProductCreateIn{
		ProductType: &productType,
		PricingMode: &pricingMode,
		ProductName: &name,
	}

	switch spec.Kind {
	case "style", "sample_flat":
		price := json.Number(record[1])
		out.FlatPrice = &price
	case "sample_hourly":
		minutes := json.Number(record[1])
		price := json.Number(record[2])
		out.MinutesWorked = &minutes
		out.HourlyRate = &price
	}

	return out
}

func decorateImportRowErrors(row int, spec productImportKindSpec, errs []res.FieldError) []res.FieldError {
	out := make([]res.FieldError, 0, len(errs))
	for _, fieldErr := range errs {
		column, ok := spec.Columns[fieldErr.Field]
		fieldName := fieldErr.Field
		columnLabel := fieldErr.Field
		if ok {
			fieldName = column.Field
			columnLabel = column.Label
		}

		meta := map[string]any{
			"row":    row,
			"column": columnLabel,
		}
		for key, value := range fieldErr.Meta {
			meta[key] = value
		}

		out = append(out, res.FieldError{
			Field:   fmt.Sprintf("rows[%d].%s", row, fieldName),
			Code:    fieldErr.Code,
			Message: fmt.Sprintf("Row %d: %s %s", row, columnLabel, fieldErr.Message),
			Meta:    meta,
		})
	}
	return out
}

func rowGeneralError(row int, message string) res.FieldError {
	return res.FieldError{
		Field:   fmt.Sprintf("rows[%d]", row),
		Code:    "INVALID",
		Message: fmt.Sprintf("Row %d: %s", row, message),
		Meta: map[string]any{
			"row": row,
		},
	}
}

func csvParseError(err error) res.FieldError {
	var parseErr *csv.ParseError
	if errors.As(err, &parseErr) && parseErr.Line > 0 {
		return res.FieldError{
			Field:   fmt.Sprintf("rows[%d]", parseErr.Line),
			Code:    "INVALID",
			Message: fmt.Sprintf("Row %d: invalid CSV format", parseErr.Line),
			Meta: map[string]any{
				"row": parseErr.Line,
			},
		}
	}

	return res.Invalid("file", "invalid CSV format")
}

func normalizeImportHeader(value string, trimBOM bool) string {
	value = strings.TrimSpace(value)
	if trimBOM {
		value = strings.TrimPrefix(value, "\ufeff")
	}
	return strings.ToLower(value)
}

func containsDisallowedBinary(data []byte) bool {
	for _, b := range data {
		if b == 0x00 {
			return true
		}
		if b < 0x20 && b != '\n' && b != '\r' && b != '\t' {
			return true
		}
	}
	return false
}

func isClearlyUnsupportedImportMIME(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	switch {
	case contentType == "":
		return false
	case strings.HasPrefix(contentType, "text/"):
		return false
	case strings.HasPrefix(contentType, "application/octet-stream"):
		return false
	case strings.HasPrefix(contentType, "application/csv"):
		return false
	case strings.HasPrefix(contentType, "application/vnd.ms-excel"):
		return false
	default:
		return true
	}
}

func equalStrings(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
