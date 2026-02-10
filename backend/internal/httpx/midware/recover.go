package midware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
)

func PanicRecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ws := &writerState{ResponseWriter: w}

			defer func() {
				if rec := recover(); rec != nil {
					slog.ErrorContext(
						r.Context(),
						"panic recovered",
						"panic", rec,
						"method", r.Method,
						"path", r.URL.Path,
						"stack", string(debug.Stack()),
					)

					if ws.wroteHeader {
						// Too late to safely write a JSON error
						return
					}

					res.Error(ws, res.Internal())
				}
			}()

			next.ServeHTTP(ws, r)
		})
	}
}

type writerState struct {
	http.ResponseWriter
	wroteHeader bool
}

func (w *writerState) WriteHeader(statusCode int) {
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *writerState) Write(b []byte) (int, error) {
	w.wroteHeader = true
	return w.ResponseWriter.Write(b)
}
