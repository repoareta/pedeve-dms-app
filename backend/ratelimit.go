package main

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
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

	// Bersihkan visitor lama setiap menit
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
	// General API rate limiter: 100 requests per second, burst of 50 (ditingkatkan untuk development)
	generalRateLimiter = NewRateLimiter(100, 50)

	// Auth rate limiter: 5 requests per minute, burst of 5 (to prevent brute force)
	authRateLimiter = NewRateLimiter(rate.Every(time.Minute/5), 5)

	// Strict rate limiter: 10 requests per minute, burst of 5
	strictRateLimiter = NewRateLimiter(rate.Every(time.Minute/10), 5)
)

// RateLimitMiddleware applies rate limiting based on client IP (untuk Fiber)
func RateLimitMiddleware(limiter *RateLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get client IP
		ip := getClientIP(c)

		// Ambil rate limiter untuk IP ini
		visitorLimiter := limiter.GetVisitor(ip)

		// Cek apakah request diizinkan
		if !visitorLimiter.Allow() {
			return c.Status(fiber.StatusTooManyRequests).JSON(ErrorResponse{
				Error:   "rate_limit_exceeded",
				Message: "Too many requests. Please try again later.",
			})
		}

		// Request allowed, continue
		return c.Next()
	}
}

// getClientIP extracts client IP from request (untuk Fiber)
func getClientIP(c *fiber.Ctx) string {
	// Check X-Forwarded-For header first (bisa mengandung multiple IPs, ambil yang pertama)
	xff := c.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For bisa mengandung multiple IPs dipisahkan koma
		// Ambil IP pertama (client asli)
		if idx := 0; idx < len(xff) {
			for i, char := range xff {
				if char == ',' {
					return xff[:i]
				}
			}
			return xff
		}
	}

	// Check X-Real-IP header
	xri := c.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fallback to IP() method dari Fiber (handles all cases including ::1)
	ip := c.IP()
	if ip == "" || ip == "::1" {
		// Untuk localhost/development, gunakan IP yang konsisten
		return "127.0.0.1"
	}
	return ip
}

// AuthRateLimitMiddleware applies stricter rate limiting for auth endpoints (untuk Fiber)
func AuthRateLimitMiddleware(c *fiber.Ctx) error {
	return RateLimitMiddleware(authRateLimiter)(c)
}

// StrictRateLimitMiddleware applies strict rate limiting (untuk Fiber)
func StrictRateLimitMiddleware(c *fiber.Ctx) error {
	return RateLimitMiddleware(strictRateLimiter)(c)
}

