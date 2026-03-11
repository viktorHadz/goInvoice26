package logging

import (
	"net/http"

	"log/slog"

	"github.com/go-chi/httplog/v3"
	"github.com/go-chi/traceid"
)

func TraceIDToHTTPLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if traceID := traceid.FromContext(r.Context()); traceID != "" {
			httplog.SetAttrs(r.Context(), slog.String("traceId", traceID))
		}
		next.ServeHTTP(w, r)
	})
}
