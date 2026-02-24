package middleware

import (
	"context"
	"net/http"
	"strings"

	"pt_lpoj/auth"
)

// UserContextKey is a custom type to prevent context key collisions
type contextKey string

const UserContextKey contextKey = "userID"

// RequireAuth is a standard net/http middleware that enforces JWT authentication
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		// Expect format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization Header Format", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]
		userID, err := auth.ValidateJWT(tokenStr)
		if err != nil {
			http.Error(w, "Invalid or Expired Token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Attach the UserID to the request context
		ctx := context.WithValue(r.Context(), UserContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
