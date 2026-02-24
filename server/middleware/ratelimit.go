package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// IPRateLimiter holds ratelimiters mapped by IP address
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter creates a new limiter for IPs
// r: tokens replenished per second
// b: maximum burst size
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}

	// Periodically cleanup IPs that haven't been seen (simple wipe to prevent OOM)
	// In production, you'd want a more sophisticated LRU cache.
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			limiter.mu.Lock()
			// Clear all memory to prevent unbounded growth from unique IPs over days
			limiter.ips = make(map[string]*rate.Limiter)
			limiter.mu.Unlock()
			log.Println("[RateLimiter] Flushed IP cache to free memory")
		}
	}()

	return limiter
}

// GetLimiter returns the rate limiter for the provided IP address
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		i.mu.Lock()
		limiter, exists = i.ips[ip] // Double check after obtaining write lock
		if !exists {
			limiter = rate.NewLimiter(i.r, i.b)
			i.ips[ip] = limiter
		}
		i.mu.Unlock()
	}

	return limiter
}

// LimitMiddleware creates an HTTP middleware that throttles requests based on IP.
func (i *IPRateLimiter) LimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Basic IP extraction. In production behind Nginx/ALB, use X-Forwarded-For
		ip := r.RemoteAddr

		limiter := i.GetLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "429 Too Many Requests - Please slow down", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
