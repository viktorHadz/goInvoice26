package limiter

import "net/http"

// LimitBodyMaxSize returns middleware that caps the HTTP request body size to n bytes
// using http.MaxBytesReader. Requests exceeding this limit will cause the handler
// to receive an error on read (e.g., http.ErrBodyReadAfterClose).
//
// Usage example:
//
//	mux := http.NewServeMux()
//	mux.Handle("/upload", uploadHandler)
//	limited := LimitBodyMaxSize(1 << 20)(mux) // limit body to 1MB= 1 << 20 | 2MB= 2 << 20
//	http.ListenAndServe(":8080", limited)
func LimitBodyMaxSize(n int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, n)
			next.ServeHTTP(w, r)
		})
	}
}

/*
10/2 - 0
5/2 - 1
2/2 - 0
1/2 - 1



18/2 - 0
9/2 - 1
4/2 - 0
2/2 - 0
1/2 - 1


Bit [0] - 0
Bit [1]-  1
Bit [2] - 0
Bit [3] - 0
Bit [4] - 1
Result 10010

 4  3  2  1  0
[1][0][0][1][0]

1 << 2 - first bit move left 2 times
--1st time - i assume the 0 gets shifted into ints place
 4  3  2  1  0
[1][0][1][0][0]

--2nd time
 4  3  2  1  0
[1][1][0][0][0]

11000
to decimal that is
1x2^4 + 1x2^3 + 0x2^2 + 0x2^1 + 0x2^0
  16  +   8   +  0    +   0   +   0
*/
