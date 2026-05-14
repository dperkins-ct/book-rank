package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// LoggingMiddleware provides request logging functionality
type LoggingMiddleware struct {
	logger *slog.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *slog.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// LogRequests middleware that logs HTTP requests
func (lm *LoggingMiddleware) LogRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response wrapper to capture status code and size
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			size:          0,
		}

		// Process the request
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Extract user information if available
		var userID *uint
		var username *string
		if claims, ok := GetUserFromContext(r.Context()); ok {
			userID = &claims.UserID
			username = &claims.Username
		}

		// Log the request
		lm.logger.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"status", wrapped.statusCode,
			"size", wrapped.size,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", getClientIP(r),
			"user_agent", r.Header.Get("User-Agent"),
			"user_id", userID,
			"username", username,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}