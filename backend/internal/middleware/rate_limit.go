package middleware

import (
	"os"
	"sync"
	"time"

	"github.com/Fajarriswandi/dms-app/backend/internal/domain"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/config"
	"github.com/Fajarriswandi/dms-app/backend/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
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

// Global rate limiters (will be initialized from config)
var (
	GeneralRateLimiter *RateLimiter
	AuthRateLimiter    *RateLimiter
	StrictRateLimiter  *RateLimiter
)

// InitRateLimiters initializes rate limiters from configuration
func InitRateLimiters() {
	zapLog := logger.GetLogger()

	// Load config
	cfg := config.GetConfig()

	// Initialize General rate limiter
	generalRate, generalBurst := cfg.RateLimit.GetGeneralRateLimit()
	GeneralRateLimiter = NewRateLimiter(generalRate, generalBurst)
	zapLog.Info("General rate limiter initialized",
		zap.Float64("rps", cfg.RateLimit.General.RPS),
		zap.Int("burst", cfg.RateLimit.General.Burst),
	)

	// Initialize Auth rate limiter
	authRate, authBurst := cfg.RateLimit.GetAuthRateLimit()
	AuthRateLimiter = NewRateLimiter(authRate, authBurst)
	zapLog.Info("Auth rate limiter initialized",
		zap.Int("rpm", cfg.RateLimit.Auth.RPM),
		zap.Int("burst", cfg.RateLimit.Auth.Burst),
	)

	// Initialize Strict rate limiter
	strictRate, strictBurst := cfg.RateLimit.GetStrictRateLimit()
	StrictRateLimiter = NewRateLimiter(strictRate, strictBurst)
	zapLog.Info("Strict rate limiter initialized",
		zap.Int("rpm", cfg.RateLimit.Strict.RPM),
		zap.Int("burst", cfg.RateLimit.Strict.Burst),
	)
}

// RateLimitMiddleware applies rate limiting based on client IP (untuk Fiber)
func RateLimitMiddleware(limiter *RateLimiter) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Disable rate limiting sepenuhnya untuk development
		env := os.Getenv("ENV")
		disableRateLimit := os.Getenv("DISABLE_RATE_LIMIT") == "true"

		// Bypass rate limiting jika:
		// 1. ENV tidak "production" (development/staging)
		// 2. DISABLE_RATE_LIMIT=true
		if env != "production" || disableRateLimit {
			// Development: bypass rate limiting sepenuhnya
			// Log hanya sekali untuk menghindari spam log
			if disableRateLimit {
				zapLog := logger.GetLogger()
				zapLog.Debug("Rate limit bypassed",
					zap.String("reason", "DISABLE_RATE_LIMIT=true"),
					zap.String("path", c.Path()),
					zap.String("method", c.Method()),
				)
			}
			return c.Next()
		}

		// Production: apply rate limiting
		ip := getClientIP(c)
		visitorLimiter := limiter.GetVisitor(ip)

		// Check if request is allowed
		if !visitorLimiter.Allow() {
			zapLog := logger.GetLogger()
			zapLog.Warn("Rate limit exceeded",
				zap.String("ip", ip),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.String("env", env),
			)
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
	if ip == "" || ip == "::1" || ip == "[::1]" {
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
