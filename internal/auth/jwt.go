package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when token validation fails
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when token has expired
	ErrExpiredToken = errors.New("token has expired")
)

// Claims represents the JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTService handles JWT token operations
type JWTService struct {
	secretKey []byte
	logger    *slog.Logger
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, logger *slog.Logger) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
		logger:    logger,
	}
}

// GenerateToken generates a new JWT token for the user
func (s *JWTService) GenerateToken(userID uint, username string) (string, error) {
	now := time.Now()
	expirationTime := now.Add(TokenExpiration)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "bookrank",
			Subject:   fmt.Sprintf("user:%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		s.logger.Error("Failed to sign JWT token", "error", err, "user_id", userID)
		return "", err
	}

	s.logger.Debug("JWT token generated", "user_id", userID, "username", username, "expires_at", expirationTime)
	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		s.logger.Debug("Failed to parse JWT token", "error", err)
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		s.logger.Debug("JWT token is not valid")
		return nil, ErrInvalidToken
	}

	// Additional validation of claims - only check for zero user ID, allow empty username
	if claims.UserID == 0 {
		s.logger.Debug("JWT token has invalid claims", "user_id", claims.UserID, "username", claims.Username)
		return nil, ErrInvalidToken
	}

	s.logger.Debug("JWT token validated successfully", "user_id", claims.UserID, "username", claims.Username)
	return claims, nil
}