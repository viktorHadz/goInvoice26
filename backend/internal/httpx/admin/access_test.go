package admin

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/db"
	authsvc "github.com/viktorHadz/goInvoice26/internal/service/auth"
	"github.com/viktorHadz/goInvoice26/internal/transaction/accessTx"
	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

func newAdminApp(t *testing.T) (*app.App, func()) {
	t.Helper()

	conn, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "admin-http.sqlite"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(context.Background(), conn); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	return &app.App{
			DB: conn,
			Auth: authsvc.NewService(conn, authsvc.Config{
				PlatformAdminEmail: "vikecah@gmail.com",
			}),
		}, func() {
			_ = conn.Close()
		}
}

func platformRequest(t *testing.T, email string, body string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/admin/access/grants", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := accountscope.WithAccountID(req.Context(), accountscope.DefaultAccountID)
	ctx = userscope.WithPrincipal(ctx, userscope.Principal{
		UserID:               1,
		AccountID:            accountscope.DefaultAccountID,
		Email:                email,
		Role:                 "owner",
		BillingAccessGranted: true,
	})

	return req.WithContext(ctx)
}

func TestCreateDirectAccessGrant_RejectsNonPlatformAdmin(t *testing.T) {
	a, cleanup := newAdminApp(t)
	defer cleanup()

	rec := httptest.NewRecorder()
	req := platformRequest(t, "someone@example.com", `{"email":"trusted@example.com"}`)

	CreateDirectAccessGrant(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestCreateDirectAccessGrant_CreatesGrantForPlatformAdmin(t *testing.T) {
	a, cleanup := newAdminApp(t)
	defer cleanup()

	rec := httptest.NewRecorder()
	req := platformRequest(t, "vikecah@gmail.com", `{"email":"trusted@example.com","note":"beta"}`)

	CreateDirectAccessGrant(a).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "trusted@example.com") {
		t.Fatalf("body = %q, want created email", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"plan":"single"`) {
		t.Fatalf("body = %q, want default single plan", rec.Body.String())
	}
}

func TestPlatformAccessHandlers_RejectNonPlatformAdmin(t *testing.T) {
	a, cleanup := newAdminApp(t)
	defer cleanup()

	ctx := context.Background()
	grant, err := accessTx.CreateDirectAccessGrant(ctx, a.DB, "trusted@example.com", "single", "beta", 1)
	if err != nil {
		t.Fatalf("CreateDirectAccessGrant seed: %v", err)
	}
	promo, err := accessTx.CreatePromoCode(ctx, a.DB, "PROMO14", 14, 1)
	if err != nil {
		t.Fatalf("CreatePromoCode seed: %v", err)
	}

	testCases := []struct {
		name    string
		method  string
		target  string
		body    string
		handler http.HandlerFunc
	}{
		{
			name:    "overview",
			method:  http.MethodGet,
			target:  "/api/admin/access",
			handler: Overview(a),
		},
		{
			name:    "delete grant",
			method:  http.MethodDelete,
			target:  "/api/admin/access/grants/1",
			handler: DeleteDirectAccessGrant(a),
		},
		{
			name:    "create promo",
			method:  http.MethodPost,
			target:  "/api/admin/access/promo-codes",
			body:    `{"code":"PROMO30","durationDays":30}`,
			handler: CreatePromoCode(a),
		},
		{
			name:    "update promo",
			method:  http.MethodPatch,
			target:  "/api/admin/access/promo-codes/1",
			body:    `{"active":false}`,
			handler: UpdatePromoCodeStatus(a),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var bodyReader *strings.Reader
			if tc.body != "" {
				bodyReader = strings.NewReader(tc.body)
			} else {
				bodyReader = strings.NewReader("")
			}
			req := httptest.NewRequest(tc.method, tc.target, bodyReader)
			req.Header.Set("Content-Type", "application/json")
			ctx := accountscope.WithAccountID(req.Context(), accountscope.DefaultAccountID)
			ctx = userscope.WithPrincipal(ctx, userscope.Principal{
				UserID:               2,
				AccountID:            accountscope.DefaultAccountID,
				Email:                "someone@example.com",
				Role:                 "owner",
				BillingAccessGranted: true,
			})
			req = req.WithContext(ctx)

			if strings.Contains(tc.target, "/grants/") {
				req.SetPathValue("grantID", strconv.FormatInt(grant.ID, 10))
			}
			if strings.Contains(tc.target, "/promo-codes/") && tc.method == http.MethodPatch {
				req.SetPathValue("promoCodeID", strconv.FormatInt(promo.ID, 10))
			}

			rec := httptest.NewRecorder()
			tc.handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
			}
		})
	}
}
