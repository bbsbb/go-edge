package middlewarefx

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	transporthttp "github.com/bbsbb/go-edge/core/transport/http"
)

var errPanic = errors.New("internal error")

// Recovery catches panics in downstream handlers, logs the stack trace,
// and returns a 500 problem details response.
func Recovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rv := recover(); rv != nil {
					logger.ErrorContext(r.Context(), "panic recovered",
						"panic", fmt.Sprint(rv),
						"stack", string(debug.Stack()),
					)
					transporthttp.WriteError(w, r, errPanic, logger)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
