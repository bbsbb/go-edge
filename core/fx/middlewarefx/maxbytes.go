package middlewarefx

import (
	"net/http"
)

const defaultMaxBytes int64 = 1 << 20 // 1 MB

// MaxBytes limits the size of incoming request bodies. Requests exceeding
// the limit receive a 400 response when the handler reads the body.
func MaxBytes(maxBytes int64) func(http.Handler) http.Handler {
	if maxBytes <= 0 {
		maxBytes = defaultMaxBytes
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
