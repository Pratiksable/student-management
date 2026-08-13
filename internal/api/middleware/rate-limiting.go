package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu        sync.Mutex
	visitors  map[string]int
	limit     int
	resetTime time.Duration
}

func NewRateLimiter(limit int, resetTime time.Duration) *rateLimiter {
	r1 := &rateLimiter{
		visitors:  make(map[string]int),
		limit:     limit,
		resetTime: resetTime,
	}
	go r1.resetVisitorCount()

	return r1
}

func (r1 *rateLimiter) resetVisitorCount() {
	for {
		time.Sleep(r1.resetTime)
		r1.mu.Lock()
		r1.visitors = make(map[string]int)
		r1.mu.Unlock()
	}
}

func (r1 *rateLimiter) RateLimitingMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		visitorIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			visitorIP = r.RemoteAddr
		}

		r1.mu.Lock()
		r1.visitors[visitorIP]++
		requestCount := r1.visitors[visitorIP]
		r1.mu.Unlock()

		if requestCount > r1.limit {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
