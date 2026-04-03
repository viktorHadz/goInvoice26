package settings

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/db"
	"github.com/viktorHadz/goInvoice26/internal/service/logo"
	"github.com/viktorHadz/goInvoice26/internal/service/storage"
	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

type settingsResponse struct {
	ReadOnly bool `json:"readOnly"`
}

func newSettingsApp(t *testing.T) (*app.App, func()) {
	t.Helper()

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "settings-http.sqlite")
	uploadRoot := filepath.Join(dir, "uploads")

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	store := storage.NewLocalStore(uploadRoot)
	return &app.App{
			DB:    conn,
			Logos: logo.NewService(conn, store),
		}, func() {
			_ = conn.Close()
		}
}

func memberRequest(t *testing.T, method, target string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, strings.NewReader(`{}`))
	ctx := accountscope.WithAccountID(req.Context(), accountscope.DefaultAccountID)
	ctx = userscope.WithPrincipal(ctx, userscope.Principal{
		UserID:               2,
		AccountID:            accountscope.DefaultAccountID,
		Role:                 "member",
		BillingAccessGranted: true,
	})
	return req.WithContext(ctx)
}

func ownerRequest(t *testing.T, method, target string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, strings.NewReader(`{}`))
	ctx := accountscope.WithAccountID(req.Context(), accountscope.DefaultAccountID)
	ctx = userscope.WithPrincipal(ctx, userscope.Principal{
		UserID:               1,
		AccountID:            accountscope.DefaultAccountID,
		Role:                 "owner",
		BillingAccessGranted: true,
	})
	return req.WithContext(ctx)
}

func TestGet_MarksMembersReadOnly(t *testing.T) {
	a, cleanup := newSettingsApp(t)
	defer cleanup()

	req := memberRequest(t, http.MethodGet, "/api/settings")
	rec := httptest.NewRecorder()

	Get(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got settingsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !got.ReadOnly {
		t.Fatalf("readOnly = false, want true")
	}
}

func TestGet_MarksOwnersEditable(t *testing.T) {
	a, cleanup := newSettingsApp(t)
	defer cleanup()

	req := ownerRequest(t, http.MethodGet, "/api/settings")
	rec := httptest.NewRecorder()

	Get(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var got settingsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ReadOnly {
		t.Fatalf("readOnly = true, want false")
	}
}

func TestPut_RejectsMemberEdits(t *testing.T) {
	a, cleanup := newSettingsApp(t)
	defer cleanup()

	req := memberRequest(t, http.MethodPut, "/api/settings")
	rec := httptest.NewRecorder()

	Put(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if !strings.Contains(rec.Body.String(), "SETTINGS_OWNER_ONLY") {
		t.Fatalf("body = %q, want SETTINGS_OWNER_ONLY", rec.Body.String())
	}
}

func TestPutLogo_RejectsMemberEdits(t *testing.T) {
	a, cleanup := newSettingsApp(t)
	defer cleanup()

	req := memberRequest(t, http.MethodPut, "/api/settings/logo")
	rec := httptest.NewRecorder()

	PutLogo(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if !strings.Contains(rec.Body.String(), "SETTINGS_OWNER_ONLY") {
		t.Fatalf("body = %q, want SETTINGS_OWNER_ONLY", rec.Body.String())
	}
}

func TestDeleteLogo_RejectsMemberEdits(t *testing.T) {
	a, cleanup := newSettingsApp(t)
	defer cleanup()

	req := memberRequest(t, http.MethodDelete, "/api/settings/logo")
	rec := httptest.NewRecorder()

	DeleteLogo(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if !strings.Contains(rec.Body.String(), "SETTINGS_OWNER_ONLY") {
		t.Fatalf("body = %q, want SETTINGS_OWNER_ONLY", rec.Body.String())
	}
}
