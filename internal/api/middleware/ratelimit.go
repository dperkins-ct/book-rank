package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

const (
	// DefaultRateLimit is the default number of requests per minute
	DefaultRateLimit = 100
	// RateLimitCleanupInterval is how often we clean up old rate limiters
	RateLimitCleanupInterval = 5 * time.Minute
)

// RateLimitMiddleware provides rate limiting functionality
type RateLimitMiddleware struct {
	limiters    map[string]*rateLimiter
	mu          sync.RWMutex
	rps         rate.Limit
	burst       int
	cleanupTick *time.Ticker
	logger      *slog.Logger
}

// rateLimiter wraps rate.Limiter with last access time
type rateLimiter struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(requestsPerMinute int, logger *slog.Logger) *RateLimitMiddleware {
	if requestsPerMinute <= 0 {
		requestsPerMinute = DefaultRateLimit
	}

	// Convert requests per minute to requests per second
	rps := rate.Limit(float64(requestsPerMinute) / 60.0)
	burst := requestsPerMinute / 10 // Allow bursts of 10% of the rate limit

	if burst < 1 {
		burst = 1
	}

	rlm := &RateLimitMiddleware{
		limiters:    make(map[string]*rateLimiter),
		rps:         rps,
		burst:       burst,
		cleanupTick: time.NewTicker(RateLimitCleanupInterval),
		logger:      logger,
	}

	// Start cleanup goroutine
	go rlm.cleanup()

	return rlm
}

// Limit middleware that applies rate limiting per IP address
func (rlm *RateLimitMiddleware) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP address
		ip := getClientIP(r)

		// Get or create rate limiter for this IP
		limiter := rlm.getLimiter(ip)

		if !limiter.Allow() {
			rlm.logger.Debug("Rate limit exceeded", "ip", ip, "path", r.URL.Path, "method", r.Method)
			rlm.sendRateLimitResponse(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LimitByUser middleware that applies rate limiting per authenticated user
func (rlm *RateLimitMiddleware) LimitByUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get user from context
		var key string
		if claims, ok := GetUserFromContext(r.Context()); ok {
			key = fmt.Sprintf("user:%d", claims.UserID)
		} else {
			// Fallback to IP-based limiting for unauthenticated requests
			key = getClientIP(r)
		}

		// Get or create rate limiter for this key
		limiter := rlm.getLimiter(key)

		if !limiter.Allow() {
			rlm.logger.Debug("Rate limit exceeded", "key", key, "path", r.URL.Path, "method", r.Method)
			rlm.sendRateLimitResponse(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getLimiter gets or creates a rate limiter for the given key
func (rlm *RateLimitMiddleware) getLimiter(key string) *rate.Limiter {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	limiter, exists := rlm.limiters[key]
	if !exists {
		limiter = &rateLimiter{
			limiter:    rate.NewLimiter(rlm.rps, rlm.burst),
			lastAccess: time.Now(),
		}
		rlm.limiters[key] = limiter
	} else {
		limiter.lastAccess = time.Now()
	}

	return limiter.limiter
}

// cleanup removes old rate limiters to prevent memory leaks
func (rlm *RateLimitMiddleware) cleanup() {
	for range rlm.cleanupTick.C {
		rlm.mu.Lock()
		cutoff := time.Now().Add(-RateLimitCleanupInterval)

		for key, limiter := range rlm.limiters {
			if limiter.lastAccess.Before(cutoff) {
				delete(rlm.limiters, key)
			}
		}

		rlm.logger.Debug("Rate limiter cleanup completed", "active_limiters", len(rlm.limiters))
		rlm.mu.Unlock()
	}
}

// sendRateLimitResponse sends a rate limit exceeded response
func (rlm *RateLimitMiddleware) sendRateLimitResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "60") // Suggest retry after 1 minute
	w.WriteHeader(http.StatusTooManyRequests)

	errorResp := ErrorResponse{
		Error:   "Rate Limit Exceeded",
		Message: "Too many requests. Please try again later.",
		Code:    http.StatusTooManyRequests,
	}

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		rlm.logger.Error("Failed to encode rate limit response", "error", err)
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (for load balancers/proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// Take the first IP in the list
		if idx := len(xff); idx > 0 {
			if comma := strings.Index(xff, ","); comma > 0 {
				return strings.TrimSpace(xff[:comma])
			}
			return strings.TrimSpace(xff)
		}
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}