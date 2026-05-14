package auth

import (
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTService_GenerateToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	jwtService := NewJWTService("test-secret", logger)

	tests := map[string]struct {
		userID   uint
		username string
		wantErr  bool
	}{
		"valid_user":           {userID: 1, username: "testuser", wantErr: false},
		"valid_user_long_id":   {userID: 9999999, username: "user", wantErr: false},
		"valid_user_long_name": {userID: 1, username: "verylongusernamethatisvalid", wantErr: false},
		"zero_user_id":         {userID: 0, username: "testuser", wantErr: false},
		"empty_username":       {userID: 1, username: "", wantErr: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			token, err := jwtService.GenerateToken(tt.userID, tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Token should be a valid JWT format (3 parts separated by dots)
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					t.Errorf("GenerateToken() returned invalid JWT format, got %d parts, want 3", len(parts))
				}

				// Should not be empty
				if token == "" {
					t.Errorf("GenerateToken() returned empty token")
				}
			}
		})
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	jwtService := NewJWTService("test-secret", logger)

	// Generate a valid token
	validToken, err := jwtService.GenerateToken(123, "testuser")
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Generate token with different secret
	otherService := NewJWTService("different-secret", logger)
	wrongSecretToken, err := otherService.GenerateToken(123, "testuser")
	if err != nil {
		t.Fatalf("Failed to generate wrong secret token: %v", err)
	}

	// Generate expired token (mock)
	expiredClaims := &Claims{
		UserID:   123,
		Username: "testuser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired 1 hour ago
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    "bookrank",
			Subject:   "user:123",
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}

	tests := map[string]struct {
		token       string
		wantErr     bool
		expectedErr error
	}{
		"valid_token":        {token: validToken, wantErr: false, expectedErr: nil},
		"wrong_secret_token": {token: wrongSecretToken, wantErr: true, expectedErr: ErrInvalidToken},
		"expired_token":      {token: expiredTokenString, wantErr: true, expectedErr: ErrExpiredToken},
		"empty_token":        {token: "", wantErr: true, expectedErr: ErrInvalidToken},
		"malformed_token":    {token: "invalid.jwt.token", wantErr: true, expectedErr: ErrInvalidToken},
		"random_string":      {token: "randomstring", wantErr: true, expectedErr: ErrInvalidToken},
		"partial_jwt":        {token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid", wantErr: true, expectedErr: ErrInvalidToken},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			claims, err := jwtService.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedErr != nil {
				if err != tt.expectedErr {
					t.Errorf("ValidateToken() error = %v, expectedErr %v", err, tt.expectedErr)
				}
			}

			if !tt.wantErr {
				if claims == nil {
					t.Errorf("ValidateToken() returned nil claims for valid token")
					return
				}

				// Validate claims content
				if claims.UserID != 123 {
					t.Errorf("ValidateToken() userID = %v, want %v", claims.UserID, 123)
				}
				if claims.Username != "testuser" {
					t.Errorf("ValidateToken() username = %v, want %v", claims.Username, "testuser")
				}
				if claims.Issuer != "bookrank" {
					t.Errorf("ValidateToken() issuer = %v, want %v", claims.Issuer, "bookrank")
				}
			}
		})
	}
}

func TestJWTService_ValidateToken_InvalidClaims(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	jwtService := NewJWTService("test-secret", logger)

	// Generate token with invalid claims (missing user info)
	invalidClaims := &Claims{
		UserID:   0, // Invalid user ID
		Username: "", // Invalid username
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "bookrank",
			Subject:   "user:0",
		},
	}
	invalidClaimsToken := jwt.NewWithClaims(jwt.SigningMethodHS256, invalidClaims)
	invalidClaimsTokenString, err := invalidClaimsToken.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("Failed to generate invalid claims token: %v", err)
	}

	claims, err := jwtService.ValidateToken(invalidClaimsTokenString)
	if err != ErrInvalidToken {
		t.Errorf("ValidateToken() with invalid claims should return ErrInvalidToken, got %v", err)
	}
	if claims != nil {
		t.Errorf("ValidateToken() with invalid claims should return nil claims, got %v", claims)
	}
}

func TestJWTService_TokenRoundTrip(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	jwtService := NewJWTService("my-secret-key", logger)

	testCases := []struct {
		userID   uint
		username string
	}{
		{1, "alice"},
		{999, "bob"},
		{42, "charlie"},
		{1000000, "verylongusername"},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			// Generate token
			token, err := jwtService.GenerateToken(tc.userID, tc.username)
			if err != nil {
				t.Fatalf("GenerateToken() failed: %v", err)
			}

			// Validate token
			claims, err := jwtService.ValidateToken(token)
			if err != nil {
				t.Fatalf("ValidateToken() failed: %v", err)
			}

			// Check that claims match input
			if claims.UserID != tc.userID {
				t.Errorf("Round trip userID = %v, want %v", claims.UserID, tc.userID)
			}
			if claims.Username != tc.username {
				t.Errorf("Round trip username = %v, want %v", claims.Username, tc.username)
			}

			// Check token expiration is in the future
			if claims.ExpiresAt.Time.Before(time.Now()) {
				t.Errorf("Token should not be expired")
			}

			// Check token expiration is approximately correct
			expectedExpiry := time.Now().Add(TokenExpiration)
			timeDiff := claims.ExpiresAt.Time.Sub(expectedExpiry)
			if timeDiff > time.Minute || timeDiff < -time.Minute {
				t.Errorf("Token expiration time diff is too large: %v", timeDiff)
			}
		})
	}
}

func TestJWTService_DifferentSecrets(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	service1 := NewJWTService("secret1", logger)
	service2 := NewJWTService("secret2", logger)

	// Generate token with service1
	token, err := service1.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("GenerateToken() failed: %v", err)
	}

	// Try to validate with service2 (should fail)
	_, err = service2.ValidateToken(token)
	if err != ErrInvalidToken {
		t.Errorf("Cross-service validation should fail with ErrInvalidToken, got %v", err)
	}

	// Validate with same service (should succeed)
	claims, err := service1.ValidateToken(token)
	if err != nil {
		t.Errorf("Same-service validation should succeed, got %v", err)
	}
	if claims == nil {
		t.Errorf("Same-service validation should return claims")
	}
}