package shared

import (
	"context"
	"net/http"

	"github.com/mobentum/kern"
)

type TokenValidator interface {
	ValidateToken(tokenStr string) (uint, error)
	IsActive(ctx context.Context, userID uint) (bool, error)
}

type userIDKey struct{}

func UserIDFromCtx(ctx context.Context) (uint, bool) {
	v, ok := ctx.Value(userIDKey{}).(uint)
	return v, ok
}

func JWTAuth(validator TokenValidator) kern.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractToken(r)
			if tokenStr == "" {
				WriteJSONError(w, r, "missing or malformed token", http.StatusUnauthorized)
				return
			}

			userID, err := validator.ValidateToken(tokenStr)
			if err != nil {
				if l := kern.LoggerFromContext(r.Context()); l != nil {
					l.Error("token validation failed", "error", err)
				}
				WriteJSONError(w, r, "invalid or expired token", http.StatusUnauthorized)
				return
			}

			active, err := validator.IsActive(r.Context(), userID)
			if err != nil {
				if l := kern.LoggerFromContext(r.Context()); l != nil {
					l.Error("failed to check account status", "error", err)
				}
				WriteJSONError(w, r, "account is suspended", http.StatusForbidden)
				return
			}
			if !active {
				WriteJSONError(w, r, "account is suspended", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	c, err := r.Cookie("token")
	if err != nil || c.Value == "" {
		return ""
	}
	return c.Value
}
