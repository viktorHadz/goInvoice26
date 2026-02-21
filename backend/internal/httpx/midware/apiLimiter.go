package midware

import "net/http"

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
