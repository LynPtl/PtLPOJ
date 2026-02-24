package api

import (
	"log"
	"net/http"
	"pt_lpoj/api/handlers"
	"pt_lpoj/middleware"

	"golang.org/x/time/rate"
)

// SetupRouter initializes and returns a configured HTTP multiplexer
func SetupRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// 1. Initialize Middlewares
	// General rate limiter for login endpoints: 1 request per second, burst of 3
	loginLimiter := middleware.NewIPRateLimiter(rate.Limit(1), 3)

	// 2. Public Endpoints (Auth)
	// Apply rate limiting specifically to Auth endpoints to prevent OTP bombing
	mux.Handle("/api/auth/login", loginLimiter.LimitMiddleware(http.HandlerFunc(handlers.RequestOTPHandler)))
	mux.Handle("/api/auth/verify", loginLimiter.LimitMiddleware(http.HandlerFunc(handlers.VerifyOTPHandler)))

	// 3. Protected Endpoints (Requires JWT)
	protectedMux := http.NewServeMux()

	// Example stub route (to be fleshed out in Phase 4)
	protectedMux.HandleFunc("/api/problems", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status": "protected route reached"}`))
	})

	// Mount the protected sub-router under /api/ with RequireAuth middleware
	// Since ServeMux doesn't support sub-routing easily before Go 1.22 perfectly,
	// we will apply it directly.
	mux.Handle("/api/problems", middleware.RequireAuth(protectedMux))

	log.Println("API Router configured successfully")
	return mux
}
