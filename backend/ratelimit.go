package main

import (
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter holds rate limiters for different endpoints
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps rate.Limit, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rps,
		burst:    burst,
	}

	// Clean up old visitors every minute
	go rl.cleanupVisitors()

	return rl
}

// GetVisitor gets or creates a visitor limiter
func (rl *RateLimiter) GetVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// cleanupVisitors removes old visitors
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Global rate limiters
var (
	// General API rate limiter: 100 requests per second, burst of 10
	generalRateLimiter = NewRateLimiter(100, 10)

	// Auth rate limiter: 5 requests per minute, burst of 2 (to prevent brute force)
	authRateLimiter = NewRateLimiter(rate.Every(time.Minute/5), 2)

	// Strict rate limiter: 10 requests per minute, burst of 3
	strictRateLimiter = NewRateLimiter(rate.Every(time.Minute/10), 3)
)

// RateLimitMiddleware applies rate limiting based on client IP
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := getClientIP(r)

			// Get rate limiter for this IP
			visitorLimiter := limiter.GetVisitor(ip)

			// Check if request is allowed
			if !visitorLimiter.Allow() {
				http.Error(w, "Too many requests. Please try again later.", http.StatusTooManyRequests)
				return
			}

			// Request allowed, continue
			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	ip := r.RemoteAddr
	if ip == "" {
		return "unknown"
	}
	return ip
}

// AuthRateLimitMiddleware applies stricter rate limiting for auth endpoints
func AuthRateLimitMiddleware(next http.Handler) http.Handler {
	return RateLimitMiddleware(authRateLimiter)(next)
}

// StrictRateLimitMiddleware applies strict rate limiting
func StrictRateLimitMiddleware(next http.Handler) http.Handler {
	return RateLimitMiddleware(strictRateLimiter)(next)
}

