package products

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/service/productimport"
	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

type importResponse struct {
	CreatedCount int    `json:"createdCount"`
	ClientID     int64  `json:"clientId"`
	ImportKind   string `json:"importKind"`
}

func newImportApp(t *testing.T) (*app.App, func()) {
	t.Helper()

	conn, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "products-import.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	return &app.App{
			DB:             conn,
			ProductImports: productimport.NewCoordinator(),
		}, func() {
			_ = conn.Close()
		}
}

func insertImportClient(t *testing.T, a *app.App, accountID int64, name string) int64 {
	t.Helper()

	res, err := a.DB.Exec(`INSERT INTO clients (account_id, name) VALUES (?, ?)`, accountID, name)
	if err != nil {
		t.Fatalf("insert client: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("client lastInsertId: %v", err)
	}

	return id
}

func newImportRequest(
	t *testing.T,
	method string,
	target string,
	body *bytes.Buffer,
	contentType string,
	accountID int64,
	userID int64,
) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, body)
	req.Header.Set("Content-Type", contentType)
	ctx := accountscope.WithAccountID(req.Context(), accountID)
	ctx = userscope.WithPrincipal(ctx, userscope.Principal{
		UserID:               userID,
		AccountID:            accountID,
		BillingAccessGranted: true,
		Role:                 "owner",
	})

	return req.WithContext(ctx)
}

func withClientIDParam(req *http.Request, clientID int64) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("clientID", fmt.Sprintf("%d", clientID))
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
}

func makeImportMultipart(
	t *testing.T,
	kind *string,
	fileName string,
	fileContent []byte,
	fileContentType string,
	extraFields map[string]string,
) (*bytes.Buffer, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if kind != nil {
		if err := writer.WriteField("kind", *kind); err != nil {
			t.Fatalf("write kind field: %v", err)
		}
	}

	for key, value := range extraFields {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("write extra field %q: %v", key, err)
		}
	}

	if fileName != "" {
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileName))
		if fileContentType != "" {
			header.Set("Content-Type", fileContentType)
		}

		part, err := writer.CreatePart(header)
		if err != nil {
			t.Fatalf("create file part: %v", err)
		}
		if _, err := part.Write(fileContent); err != nil {
			t.Fatalf("write file part: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	return body, writer.FormDataContentType()
}

func productCountForClient(t *testing.T, a *app.App, accountID int64, clientID int64) int {
	t.Helper()

	var count int
	if err := a.DB.QueryRow(
		`SELECT COUNT(*) FROM products WHERE account_id = ? AND client_id = ?`,
		accountID,
		clientID,
	).Scan(&count); err != nil {
		t.Fatalf("count products: %v", err)
	}

	return count
}

func TestImportProducts_ImportsStyleCSV(t *testing.T) {
	a, cleanup := newImportApp(t)
	defer cleanup()

	clientID := insertImportClient(t, a, accountscope.DefaultAccountID, "Import Client")
	kind := "style"
	body, contentType := makeImportMultipart(
		t,
		&kind,
		"styles.csv",
		[]byte("name,unit price\nHemline,12.50\nPocket,8.00\n"),
		"text/csv",
		nil,
	)

	req := withClientIDParam(
		newImportRequest(t, http.MethodPost, "/api/clients/1/products/import", body, contentType, accountscope.DefaultAccountID, 42),
		clientID,
	)
	rec := httptest.NewRecorder()

	ImportProducts(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var got importResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.CreatedCount != 2 {
		t.Fatalf("createdCount = %d, want 2", got.CreatedCount)
	}
	if got.ClientID != clientID {
		t.Fatalf("clientId = %d, want %d", got.ClientID, clientID)
	}
	if got.ImportKind != "style" {
		t.Fatalf("importKind = %q, want %q", got.ImportKind, "style")
	}

	if count := productCountForClient(t, a, accountscope.DefaultAccountID, clientID); count != 2 {
		t.Fatalf("product count = %d, want 2", count)
	}
}

func TestImportProducts_RejectsMissingKind(t *testing.T) {
	a, cleanup := newImportApp(t)
	defer cleanup()

	clientID := insertImportClient(t, a, accountscope.DefaultAccountID, "Import Client")
	body, contentType := makeImportMultipart(
		t,
		nil,
		"styles.csv",
		[]byte("name,unit price\nHemline,12.50\n"),
		"text/csv",
		nil,
	)

	req := withClientIDParam(
		newImportRequest(t, http.MethodPost, "/api/clients/1/products/import", body, contentType, accountscope.DefaultAccountID, 42),
		clientID,
	)
	rec := httptest.NewRecorder()

	ImportProducts(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), `"field":"kind"`) {
		t.Fatalf("body = %q, want missing kind error", rec.Body.String())
	}
	if count := productCountForClient(t, a, accountscope.DefaultAccountID, clientID); count != 0 {
		t.Fatalf("product count = %d, want 0", count)
	}
}

func TestImportProducts_RejectsOversizeFile(t *testing.T) {
	a, cleanup := newImportApp(t)
	defer cleanup()

	clientID := insertImportClient(t, a, accountscope.DefaultAccountID, "Import Client")
	kind := "style"
	body, contentType := makeImportMultipart(
		t,
		&kind,
		"styles.csv",
		bytes.Repeat([]byte("a"), maxProductImportFileSize+1),
		"text/csv",
		nil,
	)

	req := withClientIDParam(
		newImportRequest(t, http.MethodPost, "/api/clients/1/products/import", body, contentType, accountscope.DefaultAccountID, 42),
		clientID,
	)
	rec := httptest.NewRecorder()

	ImportProducts(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "50KB or smaller") {
		t.Fatalf("body = %q, want file size error", rec.Body.String())
	}
}

func TestImportProducts_RejectsUnknownHeader(t *testing.T) {
	a, cleanup := newImportApp(t)
	defer cleanup()

	clientID := insertImportClient(t, a, accountscope.DefaultAccountID, "Import Client")
	kind := "style"
	body, contentType := makeImportMultipart(
		t,
		&kind,
		"styles.csv",
		[]byte("name,price\nHemline,12.50\n"),
		"text/csv",
		nil,
	)

	req := withClientIDParam(
		newImportRequest(t, http.MethodPost, "/api/clients/1/products/import", body, contentType, accountscope.DefaultAccountID, 42),
		clientID,
	)
	rec := httptest.NewRecorder()

	ImportProducts(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), "unknown column") {
		t.Fatalf("body = %q, want header error", rec.Body.String())
	}
	if count := productCountForClient(t, a, accountscope.DefaultAccountID, clientID); count != 0 {
		t.Fatalf("product count = %d, want 0", count)
	}
}

func TestImportProducts_RejectsBinaryContent(t *testing.T) {
	a, cleanup := newImportApp(t)
	defer cleanup()

	clientID := insertImportClient(t, a, accountscope.DefaultAccountID, "Import Client")
	kind := "style"
	body, contentType := makeImportMultipart(
		t,
		&kind,
		"styles.csv",
		[]byte{0x89, 0x50, 0x4e, 0x47, 0x00, 0x01},
		"image/png",
		nil,
	)

	req := withClientIDParam(
		newImportRequest(t, http.MethodPost, "/api/clients/1/products/import", body, contentType, accountscope.DefaultAccountID, 42),
		clientID,
	)
	rec := httptest.NewRecorder()

	ImportProducts(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "upload a CSV file") {
		t.Fatalf("body = %q, want csv error", rec.Body.String())
	}
}

func TestImportProducts_RejectsInvalidUTF8(t *testing.T) {
	a, cleanup := newImportApp(t)
	defer cleanup()

	clientID := insertImportClient(t, a, accountscope.DefaultAccountID, "Import Client")
	kind := "style"
	body, contentType := makeImportMultipart(
		t,
		&kind,
		"styles.csv",
		[]byte{0xff, 0xfe, 0xfd},
		"application/octet-stream",
		nil,
	)

	req := withClientIDParam(
		newImportRequest(t, http.MethodPost, "/api/clients/1/products/import", body, contentType, accountscope.DefaultAccountID, 42),
		clientID,
	)
	rec := httptest.NewRecorder()

	ImportProducts(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "valid UTF-8") {
		t.Fatalf("body = %q, want utf-8 error", rec.Body.String())
	}
}

func TestImportProducts_RejectsInvalidRowAndWritesNothing(t *testing.T) {
	a, cleanup := newImportApp(t)
	defer cleanup()

	clientID := insertImportClient(t, a, accountscope.DefaultAccountID, "Import Client")
	kind := "sample_hourly"
	body, contentType := makeImportMultipart(
		t,
		&kind,
		"samples.csv",
		[]byte("name,time to produce (in minutes),unit price\nFitting,,22.00\n"),
		"text/csv",
		nil,
	)

	req := withClientIDParam(
		newImportRequest(t, http.MethodPost, "/api/clients/1/products/import", body, contentType, accountscope.DefaultAccountID, 42),
		clientID,
	)
	rec := httptest.NewRecorder()

	ImportProducts(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Row 2") {
		t.Fatalf("body = %q, want row error", rec.Body.String())
	}
	if count := productCountForClient(t, a, accountscope.DefaultAccountID, clientID); count != 0 {
		t.Fatalf("product count = %d, want 0", count)
	}
}

func TestImportProducts_RejectsTooManyRows(t *testing.T) {
	a, cleanup := newImportApp(t)
	defer cleanup()

	clientID := insertImportClient(t, a, accountscope.DefaultAccountID, "Import Client")
	var builder strings.Builder
	builder.WriteString("name,unit price\n")
	for i := 0; i < maxProductImportRows+1; i++ {
		builder.WriteString(fmt.Sprintf("Style %d,1.00\n", i+1))
	}

	kind := "style"
	body, contentType := makeImportMultipart(
		t,
		&kind,
		"styles.csv",
		[]byte(builder.String()),
		"text/csv",
		nil,
	)

	req := withClientIDParam(
		newImportRequest(t, http.MethodPost, "/api/clients/1/products/import", body, contentType, accountscope.DefaultAccountID, 42),
		clientID,
	)
	rec := httptest.NewRecorder()

	ImportProducts(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "at most 400 data rows") {
		t.Fatalf("body = %q, want row count error", rec.Body.String())
	}
}

func TestImportProducts_AllowsDuplicateNames(t *testing.T) {
	a, cleanup := newImportApp(t)
	defer cleanup()

	clientID := insertImportClient(t, a, accountscope.DefaultAccountID, "Import Client")
	kind := "style"
	body, contentType := makeImportMultipart(
		t,
		&kind,
		"styles.csv",
		[]byte("name,unit price\nHemline,12.50\nHemline,12.50\n"),
		"text/csv",
		nil,
	)

	req := withClientIDParam(
		newImportRequest(t, http.MethodPost, "/api/clients/1/products/import", body, contentType, accountscope.DefaultAccountID, 42),
		clientID,
	)
	rec := httptest.NewRecorder()

	ImportProducts(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if count := productCountForClient(t, a, accountscope.DefaultAccountID, clientID); count != 2 {
		t.Fatalf("product count = %d, want 2", count)
	}
}

func TestImportProducts_RejectsConcurrentImport(t *testing.T) {
	a, cleanup := newImportApp(t)
	defer cleanup()

	clientID := insertImportClient(t, a, accountscope.DefaultAccountID, "Import Client")
	if ok := a.ProductImports.Acquire(accountscope.DefaultAccountID); !ok {
		t.Fatal("expected to acquire import coordinator lock")
	}
	defer a.ProductImports.Release(accountscope.DefaultAccountID)

	kind := "style"
	body, contentType := makeImportMultipart(
		t,
		&kind,
		"styles.csv",
		[]byte("name,unit price\nHemline,12.50\n"),
		"text/csv",
		nil,
	)

	req := withClientIDParam(
		newImportRequest(t, http.MethodPost, "/api/clients/1/products/import", body, contentType, accountscope.DefaultAccountID, 42),
		clientID,
	)
	rec := httptest.NewRecorder()

	ImportProducts(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "PRODUCT_IMPORT_IN_PROGRESS") {
		t.Fatalf("body = %q, want import-in-progress code", rec.Body.String())
	}
}

func TestImportProducts_RejectsClientOutsideAccountScope(t *testing.T) {
	a, cleanup := newImportApp(t)
	defer cleanup()

	if _, err := a.DB.Exec(`INSERT INTO accounts (id, name) VALUES (2, 'Second Account')`); err != nil {
		t.Fatalf("insert second account: %v", err)
	}
	clientID := insertImportClient(t, a, 2, "Second Account Client")

	kind := "style"
	body, contentType := makeImportMultipart(
		t,
		&kind,
		"styles.csv",
		[]byte("name,unit price\nHemline,12.50\n"),
		"text/csv",
		nil,
	)

	req := withClientIDParam(
		newImportRequest(t, http.MethodPost, "/api/clients/1/products/import", body, contentType, accountscope.DefaultAccountID, 42),
		clientID,
	)
	rec := httptest.NewRecorder()

	ImportProducts(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}
