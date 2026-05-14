package auth

import (
	"errors"
	"log/slog"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	// BcryptCost is the cost factor for bcrypt hashing
	BcryptCost = 12
	// TokenExpiration is the duration for which JWT tokens are valid
	TokenExpiration = 24 * time.Hour
	// MinUsernameLength is the minimum allowed username length
	MinUsernameLength = 3
	// MaxUsernameLength is the maximum allowed username length
	MaxUsernameLength = 20
	// MinPasswordLength is the minimum allowed password length
	MinPasswordLength = 8
)

var (
	// ErrInvalidUsername is returned when username validation fails
	ErrInvalidUsername = errors.New("username must be 3-20 characters long")
	// ErrInvalidPassword is returned when password validation fails
	ErrInvalidPassword = errors.New("password must be at least 8 characters long and contain special characters")
	// ErrInvalidCredentials is returned when login credentials are invalid
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrUserAlreadyExists is returned when trying to register an existing user
	ErrUserAlreadyExists = errors.New("user already exists")
)

// AuthService handles authentication operations
type AuthService struct {
	jwtService *JWTService
	logger     *slog.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtSecret string, logger *slog.Logger) *AuthService {
	return &AuthService{
		jwtService: NewJWTService(jwtSecret, logger),
		logger:     logger,
	}
}

// ValidateUsername validates username according to business rules
func (s *AuthService) ValidateUsername(username string) error {
	if len(username) < MinUsernameLength || len(username) > MaxUsernameLength {
		s.logger.Debug("Username validation failed", "username", username, "length", len(username))
		return ErrInvalidUsername
	}
	return nil
}

// ValidatePassword validates password according to business rules
func (s *AuthService) ValidatePassword(password string) error {
	if password == "" {
		s.logger.Debug("Password validation failed", "reason", "empty")
		return ErrInvalidPassword
	}

	if len(password) < MinPasswordLength {
		s.logger.Debug("Password validation failed", "reason", "too_short", "length", len(password))
		return ErrInvalidPassword
	}

	// Check if password is only letters
	onlyLetters := regexp.MustCompile(`^[a-zA-Z]+$`).MatchString(password)
	if onlyLetters {
		s.logger.Debug("Password validation failed", "reason", "only_letters")
		return ErrInvalidPassword
	}

	// Check if password is only numbers
	onlyNumbers := regexp.MustCompile(`^[0-9]+$`).MatchString(password)
	if onlyNumbers {
		s.logger.Debug("Password validation failed", "reason", "only_numbers")
		return ErrInvalidPassword
	}

	// Check for at least one special character
	hasSpecial := regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(password)
	if !hasSpecial {
		s.logger.Debug("Password validation failed", "reason", "no_special_chars")
		return ErrInvalidPassword
	}

	return nil
}

// HashPassword hashes a password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return "", err
	}
	return string(bytes), nil
}

// ComparePassword compares a password with its hash
func (s *AuthService) ComparePassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}
		s.logger.Error("Failed to compare password", "error", err)
		return err
	}
	return nil
}

// GenerateToken generates a JWT token for the user
func (s *AuthService) GenerateToken(userID uint, username string) (string, error) {
	return s.jwtService.GenerateToken(userID, username)
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	return s.jwtService.ValidateToken(tokenString)
}

// RefreshToken generates a new token if the current one is valid
func (s *AuthService) RefreshToken(tokenString string) (string, error) {
	claims, err := s.jwtService.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Generate new token with the same user information
	return s.jwtService.GenerateToken(claims.UserID, claims.Username)
}