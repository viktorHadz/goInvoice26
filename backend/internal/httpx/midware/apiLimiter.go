package midware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

// LimitBodyMaxSize returns middleware that caps the HTTP request body size to n bytes
// using http.MaxBytesReader. Requests exceeding this limit will cause the handler
// to receive an error on read (e.g., http.ErrBodyReadAfterClose).
//
// Usage example:
//
//	mux := http.NewServeMux()
//	mux.Handle("/upload", uploadHandler)
//	limited := LimitBodyMaxSize(1 << 20)(mux) // limit body to 1MB = 1 << 20 | 2MB = 2 << 20
//	http.ListenAndServe(":8080", limited)
func LimitBodyMaxSize(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}

func LimitByAuthenticatedUser(requestLimit int, windowLength time.Duration, message string) func(http.Handler) http.Handler {
	if message == "" {
		message = "Too many requests"
	}

	return httprate.Limit(
		requestLimit,
		windowLength,
		httprate.WithKeyFuncs(keyByAuthenticatedUser),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			res.Error(w, http.StatusTooManyRequests, "RATE_LIMITED", message)
		}),
	)
}

func LimitInvoiceRevisionCreateByUser() func(http.Handler) http.Handler {
	return LimitByAuthenticatedUser(
		10,
		time.Minute,
		"Too many revision saves. Please wait a moment and try again.",
	)
}

func keyByAuthenticatedUser(r *http.Request) (string, error) {
	principal, ok := userscope.PrincipalFromContext(r.Context())
	if !ok || principal.UserID <= 0 {
		return httprate.KeyByIP(r)
	}

	return fmt.Sprintf("%d:%d", principal.AccountID, principal.UserID), nil
}
