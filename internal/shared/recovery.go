package shared

import (
	"net/http"

	"github.com/mobentum/kern"
)

func Recovery() kern.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if l := kern.LoggerFromContext(r.Context()); l != nil {
						l.Error("panic recovered", "error", err)
					}
					WriteJSONError(w, r, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
