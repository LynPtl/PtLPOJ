package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"pt_lpoj/auth"
	"pt_lpoj/models"
	"pt_lpoj/storage"
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

// RequireAdminToken protects paths that should only be hit by an admin with the global token
func RequireAdminToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedToken := os.Getenv("PTLPOJ_ADMIN_TOKEN")
		if expectedToken == "" {
			expectedToken = "ptlpoj_default_admin" // Fallback default admin token
		}

		adminToken := r.Header.Get("X-Admin-Token")
		if adminToken != expectedToken {
			http.Error(w, `{"error": "Invalid Admin Environment Token"}`, http.StatusUnauthorized)
			return
		}

		// Now also require valid JWT
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Missing Authorization Header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error": "Invalid Authorization format"}`, http.StatusUnauthorized)
			return
		}

		userID, err := auth.ValidateJWT(parts[1])
		if err != nil {
			http.Error(w, `{"error": "Invalid or Expired JWT"}`, http.StatusUnauthorized)
			return
		}

		user, err := storage.GetUserByID(userID)
		if err != nil || user == nil {
			http.Error(w, `{"error": "User does not exist"}`, http.StatusUnauthorized)
			return
		}

		if user.Role != models.RoleAdmin {
			http.Error(w, `{"error": "Insufficient permissions: Non-Admin Role"}`, http.StatusForbidden) // 403 Forbidden
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
