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

	// Phase 4 Routes -> Problems API
	protectedMux.HandleFunc("/api/problems/", func(w http.ResponseWriter, r *http.Request) {
		// Extremely simple router without 3rd party frameworks like Gin or standard Go 1.22 methods
		if r.URL.Path == "/api/problems" || r.URL.Path == "/api/problems/" {
			handlers.GetProblemsHandler(w, r)
			return
		}
		// Otherwise it's /api/problems/123
		handlers.GetProblemDetailHandler(w, r)
	})

	// Phase 4 Routes -> Submissions API
	protectedMux.HandleFunc("/api/submissions", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/submissions" || r.URL.Path == "/api/submissions/" {
			if r.Method == http.MethodGet {
				handlers.GetUserSubmissionsHandler(w, r)
			} else if r.Method == http.MethodPost {
				handlers.CreateSubmissionHandler(w, r)
			} else {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
			return
		}
	})

	// SSE stream specifically mounts over submissions prefix
	protectedMux.HandleFunc("/api/submissions/", func(w http.ResponseWriter, r *http.Request) {
		// Need to differentiate between exact root and child paths
		if r.URL.Path == "/api/submissions" || r.URL.Path == "/api/submissions/" {
			// Defer to above logic
			if r.Method == http.MethodGet {
				handlers.GetUserSubmissionsHandler(w, r)
			} else if r.Method == http.MethodPost {
				handlers.CreateSubmissionHandler(w, r)
			} else {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// Otherwise check if it ends with /stream
		if len(r.URL.Path) > 7 && r.URL.Path[len(r.URL.Path)-7:] == "/stream" {
			handlers.SSEStreamHandler(w, r)
			return
		}

		http.Error(w, "Endpoint Not Found", http.StatusNotFound)
	})

	// Mount the protected sub-router under /api/ with RequireAuth middleware
	// We map the root prefixes explicitly because Go's default Mux matches trailing slashes to children
	mux.Handle("/api/problems", middleware.RequireAuth(protectedMux))
	mux.Handle("/api/problems/", middleware.RequireAuth(protectedMux))
	mux.Handle("/api/submissions", middleware.RequireAuth(protectedMux))
	mux.Handle("/api/submissions/", middleware.RequireAuth(protectedMux))

	log.Println("API Router configured successfully")
	return mux
}
