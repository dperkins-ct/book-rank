package middleware

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"bookrank/internal/auth"
)

const (
	// AuthorizationHeader is the header name for authorization
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for bearer tokens
	BearerPrefix = "Bearer "
	// UserContextKey is the context key for storing user information
	UserContextKey = "user"
)

// ErrorResponse represents a JSON error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	authService *auth.AuthService
	logger      *slog.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *auth.AuthService, logger *slog.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth middleware that requires valid JWT authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := m.extractToken(r)
		if token == "" {
			m.logger.Debug("Missing authorization token", "path", r.URL.Path, "method", r.Method)
			m.sendErrorResponse(w, "Authorization token required", http.StatusUnauthorized)
			return
		}

		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			m.logger.Debug("Invalid authorization token", "error", err, "path", r.URL.Path, "method", r.Method)

			var statusCode int
			var message string

			switch err {
			case auth.ErrExpiredToken:
				statusCode = http.StatusUnauthorized
				message = "Token has expired"
			case auth.ErrInvalidToken:
				statusCode = http.StatusUnauthorized
				message = "Invalid token"
			default:
				statusCode = http.StatusInternalServerError
				message = "Token validation failed"
			}

			m.sendErrorResponse(w, message, statusCode)
			return
		}

		// Add user information to request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth middleware that optionally validates JWT authentication
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := m.extractToken(r)
		if token != "" {
			claims, err := m.authService.ValidateToken(token)
			if err == nil {
				// Add user information to request context if token is valid
				ctx := context.WithValue(r.Context(), UserContextKey, claims)
				r = r.WithContext(ctx)
			} else {
				m.logger.Debug("Optional auth failed", "error", err, "path", r.URL.Path)
			}
		}

		next.ServeHTTP(w, r)
	})
}

// extractToken extracts the JWT token from the Authorization header
func (m *AuthMiddleware) extractToken(r *http.Request) string {
	authHeader := r.Header.Get(AuthorizationHeader)
	if authHeader == "" {
		return ""
	}

	if !strings.HasPrefix(authHeader, BearerPrefix) {
		return ""
	}

	return strings.TrimPrefix(authHeader, BearerPrefix)
}

// sendErrorResponse sends a JSON error response
func (m *AuthMiddleware) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    statusCode,
	}

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		m.logger.Error("Failed to encode error response", "error", err)
	}
}

// GetUserFromContext extracts user claims from request context
func GetUserFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*auth.Claims)
	return claims, ok
}