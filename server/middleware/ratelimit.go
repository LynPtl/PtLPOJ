package middleware

import (
	"log"
	"net/http"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"golang.org/x/time/rate"
)

// IPRateLimiter holds an LRU cache of rate limiters mapped by IP address
type IPRateLimiter struct {
	lru *lru.ARCCache
	mu  sync.Mutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter creates a new limiter for IPs
// r: tokens replenished per second
// b: maximum burst size
// maxIPs: maximum number of IPs to track before evicting least recently used
func NewIPRateLimiter(r rate.Limit, b int, maxIPs int) *IPRateLimiter {
	cache, err := lru.NewARC(maxIPs)
	if err != nil {
		log.Printf("[RateLimiter] Warning: failed to create LRU cache: %v, falling back to unlimited", err)
		cache, _ = lru.NewARC(0)
	}

	limiter := &IPRateLimiter{
		lru: cache,
		r:   r,
		b:   b,
	}

	return limiter
}

// GetLimiter returns the rate limiter for the provided IP address
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	if val, ok := i.lru.Get(ip); ok {
		return val.(*rate.Limiter)
	}

	limiter := rate.NewLimiter(i.r, i.b)
	i.lru.Add(ip, limiter)
	return limiter
}

// LimitMiddleware creates an HTTP middleware that throttles requests based on IP.
func (i *IPRateLimiter) LimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		limiter := i.GetLimiter(ip)
		if !limiter.Allow() {
			http.Error(w, "429 Too Many Requests - Please slow down", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
