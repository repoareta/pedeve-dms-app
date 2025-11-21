package middleware

import (
	"sync"
	"time"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
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
	// General API rate limiter: 200 requests per second, burst of 200 (sangat longgar untuk development)
	// Burst tinggi untuk mengakomodasi frontend yang melakukan polling/auto-refresh
	// Di production, bisa dikurangi ke 100 req/s, burst 100
	GeneralRateLimiter = NewRateLimiter(200, 200)

	// Auth rate limiter: 5 requests per minute, burst of 5 (to prevent brute force)
	// Hanya untuk public auth endpoints (login)
	AuthRateLimiter = NewRateLimiter(rate.Every(time.Minute/5), 5)

	// Strict rate limiter: 20 requests per minute, burst of 20 (ditingkatkan untuk development)
	StrictRateLimiter = NewRateLimiter(rate.Every(time.Minute/20), 20)
)

// RateLimitMiddleware applies rate limiting based on client IP (untuk Fiber)
func RateLimitMiddleware(limiter *RateLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get client IP
		ip := getClientIP(c)

		// Get rate limiter for this IP
		visitorLimiter := limiter.GetVisitor(ip)

		// Check if request is allowed
		if !visitorLimiter.Allow() {
			return c.Status(fiber.StatusTooManyRequests).JSON(domain.ErrorResponse{
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
	// Normalisasi IP untuk konsistensi rate limiting
	normalizeIP := func(ip string) string {
		// Normalisasi IPv6 localhost ke IPv4 untuk konsistensi
		if ip == "::1" || ip == "::ffff:127.0.0.1" || ip == "[::1]" {
			return "127.0.0.1"
		}
		return ip
	}

	// Check X-Forwarded-For header first (bisa mengandung multiple IPs, ambil yang pertama)
	xff := c.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For bisa mengandung multiple IPs dipisahkan koma
		// Ambil IP pertama (client asli)
		for i, char := range xff {
			if char == ',' {
				return normalizeIP(xff[:i])
			}
		}
		return normalizeIP(xff)
	}

	// Check X-Real-IP header
	xri := c.Get("X-Real-IP")
	if xri != "" {
		return normalizeIP(xri)
	}

	// Fallback to IP() method dari Fiber (handles all cases including ::1)
	ip := c.IP()
	if ip == "" {
		return "127.0.0.1"
	}
	return normalizeIP(ip)
}

// AuthRateLimitMiddleware applies stricter rate limiting for auth endpoints (untuk Fiber)
func AuthRateLimitMiddleware(c *fiber.Ctx) error {
	return RateLimitMiddleware(AuthRateLimiter)(c)
}

// StrictRateLimitMiddleware applies strict rate limiting (untuk Fiber)
func StrictRateLimitMiddleware(c *fiber.Ctx) error {
	return RateLimitMiddleware(StrictRateLimiter)(c)
}
