package midware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

func TestLimitByAuthenticatedUser_LimitsPerUser(t *testing.T) {
	limiter := LimitByAuthenticatedUser(2, time.Hour, "Too many revision saves")
	handler := limiter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	first := doLimitedRequest(t, handler, 1)
	if first.Code != http.StatusNoContent {
		t.Fatalf("first status = %d, want %d", first.Code, http.StatusNoContent)
	}

	second := doLimitedRequest(t, handler, 1)
	if second.Code != http.StatusNoContent {
		t.Fatalf("second status = %d, want %d", second.Code, http.StatusNoContent)
	}

	third := doLimitedRequest(t, handler, 1)
	if third.Code != http.StatusTooManyRequests {
		t.Fatalf("third status = %d, want %d", third.Code, http.StatusTooManyRequests)
	}
	if !strings.Contains(third.Body.String(), "RATE_LIMITED") {
		t.Fatalf("third body = %q, want RATE_LIMITED", third.Body.String())
	}

	otherUser := doLimitedRequest(t, handler, 2)
	if otherUser.Code != http.StatusNoContent {
		t.Fatalf("other user status = %d, want %d", otherUser.Code, http.StatusNoContent)
	}
}

func doLimitedRequest(t *testing.T, handler http.Handler, userID int64) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/clients/1/invoice/10/revisions", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	req = req.WithContext(userscope.WithPrincipal(req.Context(), userscope.Principal{
		UserID:    userID,
		AccountID: 77,
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}
