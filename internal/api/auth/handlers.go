package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"gorm.io/gorm"

	"bookrank/internal/auth"
	"bookrank/internal/models"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	db          *gorm.DB
	authService *auth.AuthService
	logger      *slog.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *gorm.DB, authService *auth.AuthService, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		db:          db,
		authService: authService,
		logger:      logger,
	}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.UserRegisterRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("Failed to parse registration request", "error", err)
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate username
	if err := h.authService.ValidateUsername(req.Username); err != nil {
		h.logger.Debug("Username validation failed", "username", req.Username, "error", err)
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate password
	if err := h.authService.ValidatePassword(req.Password); err != nil {
		h.logger.Debug("Password validation failed", "error", err)
		h.sendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := h.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		h.logger.Debug("User already exists", "username", req.Username)
		h.sendErrorResponse(w, "User already exists", http.StatusConflict)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		h.logger.Error("Database error checking existing user", "error", err)
		h.sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Hash password
	hashedPassword, err := h.authService.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("Failed to hash password", "error", err)
		h.sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	user := models.User{
		Username:     req.Username,
		PasswordHash: hashedPassword,
	}

	if err := h.db.Create(&user).Error; err != nil {
		h.logger.Error("Failed to create user", "error", err, "username", req.Username)
		h.sendErrorResponse(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	token, err := h.authService.GenerateToken(user.ID, user.Username)
	if err != nil {
		h.logger.Error("Failed to generate token", "error", err, "user_id", user.ID)
		h.sendErrorResponse(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Send response
	response := struct {
		User  models.UserResponse  `json:"user"`
		Token models.TokenResponse `json:"token"`
	}{
		User: models.UserResponse{
			ID:       user.ID,
			Username: user.Username,
		},
		Token: models.TokenResponse{
			Token:     token,
			ExpiresAt: time.Now().Add(auth.TokenExpiration).Unix(),
		},
	}

	h.logger.Info("User registered successfully", "user_id", user.ID, "username", user.Username)
	h.sendJSONResponse(w, response, http.StatusCreated)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.UserLoginRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("Failed to parse login request", "error", err)
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find user by username
	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		h.logger.Debug("User not found", "username", req.Username)
		h.sendErrorResponse(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Compare password
	if err := h.authService.ComparePassword(user.PasswordHash, req.Password); err != nil {
		h.logger.Debug("Password comparison failed", "username", req.Username, "error", err)
		h.sendErrorResponse(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := h.authService.GenerateToken(user.ID, user.Username)
	if err != nil {
		h.logger.Error("Failed to generate token", "error", err, "user_id", user.ID)
		h.sendErrorResponse(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Send response
	response := struct {
		User  models.UserResponse  `json:"user"`
		Token models.TokenResponse `json:"token"`
	}{
		User: models.UserResponse{
			ID:       user.ID,
			Username: user.Username,
		},
		Token: models.TokenResponse{
			Token:     token,
			ExpiresAt: time.Now().Add(auth.TokenExpiration).Unix(),
		},
	}

	h.logger.Info("User logged in successfully", "user_id", user.ID, "username", user.Username)
	h.sendJSONResponse(w, response, http.StatusOK)
}

// Refresh handles token refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Debug("Failed to parse refresh request", "error", err)
		h.sendErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Refresh token
	newToken, err := h.authService.RefreshToken(req.Token)
	if err != nil {
		h.logger.Debug("Token refresh failed", "error", err)

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
			message = "Token refresh failed"
		}

		h.sendErrorResponse(w, message, statusCode)
		return
	}

	// Send response
	response := models.TokenResponse{
		Token:     newToken,
		ExpiresAt: time.Now().Add(auth.TokenExpiration).Unix(),
	}

	h.logger.Info("Token refreshed successfully")
	h.sendJSONResponse(w, response, http.StatusOK)
}

// Me handles getting current user information
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// This endpoint requires authentication, so user should be in context
	claims, ok := r.Context().Value("user").(*auth.Claims)
	if !ok {
		h.logger.Error("User not found in context")
		h.sendErrorResponse(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get user from database to ensure they still exist
	var user models.User
	if err := h.db.Where("id = ?", claims.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.logger.Debug("User not found in database", "user_id", claims.UserID)
			h.sendErrorResponse(w, "User not found", http.StatusNotFound)
		} else {
			h.logger.Error("Database error getting user", "error", err, "user_id", claims.UserID)
			h.sendErrorResponse(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	response := models.UserResponse{
		ID:       user.ID,
		Username: user.Username,
	}

	h.sendJSONResponse(w, response, http.StatusOK)
}

// sendJSONResponse sends a JSON response
func (h *AuthHandler) sendJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
	}
}

// sendErrorResponse sends a JSON error response
func (h *AuthHandler) sendErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := struct {
		Error   string `json:"error"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	}{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    statusCode,
	}

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		h.logger.Error("Failed to encode error response", "error", err)
	}
}