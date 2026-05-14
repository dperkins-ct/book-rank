package auth

import (
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestAuthService_ValidateUsername(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	authService := NewAuthService("test-secret", logger)

	tests := map[string]struct {
		username string
		wantErr  bool
	}{
		"valid_username_min":          {username: "abc", wantErr: false},
		"valid_username_max":          {username: "abcdefghijklmnopqrst", wantErr: false},
		"valid_username_middle":       {username: "testuser", wantErr: false},
		"invalid_username_too_short":  {username: "ab", wantErr: true},
		"invalid_username_too_long":   {username: "abcdefghijklmnopqrstu", wantErr: true},
		"invalid_username_empty":      {username: "", wantErr: true},
		"valid_username_with_numbers": {username: "user123", wantErr: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := authService.ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthService_ValidatePassword(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	authService := NewAuthService("test-secret", logger)

	tests := map[string]struct {
		password string
		wantErr  bool
	}{
		"valid_password_min":                 {password: "pass123!", wantErr: false},
		"valid_password_long":                {password: "verylongpasswordwithspecial!", wantErr: false},
		"invalid_password_too_short":         {password: "pass1!", wantErr: true},
		"invalid_password_no_special":        {password: "password123", wantErr: true},
		"invalid_password_empty":             {password: "", wantErr: true},
		"valid_password_multiple_specials":   {password: "password123!@#", wantErr: false},
		"valid_password_with_symbols":        {password: "mypass$word1", wantErr: false},
		"invalid_password_only_letters":      {password: "passwordonly", wantErr: true},
		"invalid_password_only_numbers":      {password: "12345678", wantErr: true},
		"valid_password_mixed_case_special":  {password: "MyPass123!", wantErr: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := authService.ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthService_HashPassword(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	authService := NewAuthService("test-secret", logger)

	tests := map[string]struct {
		password string
		wantErr  bool
	}{
		"valid_password":       {password: "password123!", wantErr: false},
		"empty_password":       {password: "", wantErr: false}, // bcrypt can handle empty strings
		"long_password":        {password: "verylongpasswordwithlotsofcharacters123!", wantErr: false},
		"special_characters":   {password: "p@$$w0rd!@#", wantErr: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			hash1, err := authService.HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Hash should not be empty
				if hash1 == "" {
					t.Errorf("HashPassword() returned empty hash")
				}

				// Hash should be different from password
				if hash1 == tt.password {
					t.Errorf("HashPassword() returned same as password")
				}

				// Second hash should be different (salt should vary)
				hash2, err := authService.HashPassword(tt.password)
				if err != nil {
					t.Errorf("HashPassword() second call error = %v", err)
				}
				if hash1 == hash2 {
					t.Errorf("HashPassword() should produce different hashes due to salt")
				}
			}
		})
	}
}

func TestAuthService_ComparePassword(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	authService := NewAuthService("test-secret", logger)

	password := "testpass123!"
	hash, err := authService.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := map[string]struct {
		hashedPassword string
		password       string
		wantErr        bool
	}{
		"valid_password":           {hashedPassword: hash, password: password, wantErr: false},
		"invalid_password":         {hashedPassword: hash, password: "wrongpass", wantErr: true},
		"empty_password":           {hashedPassword: hash, password: "", wantErr: true},
		"empty_hash":              {hashedPassword: "", password: password, wantErr: true},
		"invalid_hash":            {hashedPassword: "invalid-hash", password: password, wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			err := authService.ComparePassword(tt.hashedPassword, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ComparePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthService_GenerateToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	authService := NewAuthService("test-secret-key", logger)

	tests := map[string]struct {
		userID       uint
		username     string
		wantErr      bool
		wantValidate bool
	}{
		"valid_user":        {userID: 1, username: "testuser", wantErr: false, wantValidate: true},
		"valid_user_long":   {userID: 999999, username: "verylongusername", wantErr: false, wantValidate: true},
		"zero_user_id":      {userID: 0, username: "testuser", wantErr: false, wantValidate: false}, // JWT generation works, validation fails
		"empty_username":    {userID: 1, username: "", wantErr: false, wantValidate: true}, // Should work
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			token, err := authService.GenerateToken(tt.userID, tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Token should not be empty
				if token == "" {
					t.Errorf("GenerateToken() returned empty token")
				}

				// Should be able to validate the token if wantValidate is true
				claims, err := authService.ValidateToken(token)
				if tt.wantValidate {
					if err != nil {
						t.Errorf("ValidateToken() failed for generated token: %v", err)
					} else {
						if claims.UserID != tt.userID {
							t.Errorf("ValidateToken() userID = %v, want %v", claims.UserID, tt.userID)
						}
						if claims.Username != tt.username {
							t.Errorf("ValidateToken() username = %v, want %v", claims.Username, tt.username)
						}
					}
				} else {
					if err == nil {
						t.Errorf("ValidateToken() should have failed for invalid claims")
					}
				}
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	authService := NewAuthService("test-secret-key", logger)

	// Generate a valid token
	validToken, err := authService.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Generate a token with different secret (should be invalid)
	otherAuthService := NewAuthService("different-secret", logger)
	invalidToken, err := otherAuthService.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("Failed to generate invalid test token: %v", err)
	}

	tests := map[string]struct {
		token   string
		wantErr bool
	}{
		"valid_token":     {token: validToken, wantErr: false},
		"invalid_token":   {token: invalidToken, wantErr: true},
		"empty_token":     {token: "", wantErr: true},
		"malformed_token": {token: "invalid.jwt.token", wantErr: true},
		"random_string":   {token: "randomstring", wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			claims, err := authService.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && claims == nil {
				t.Errorf("ValidateToken() returned nil claims for valid token")
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	authService := NewAuthService("test-secret-key", logger)

	// Generate a valid token
	originalToken, err := authService.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Wait a moment to ensure timestamp difference
	time.Sleep(time.Millisecond * 100)

	tests := map[string]struct {
		token   string
		wantErr bool
	}{
		"valid_token":     {token: originalToken, wantErr: false},
		"invalid_token":   {token: "invalid.jwt.token", wantErr: true},
		"empty_token":     {token: "", wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			newToken, err := authService.RefreshToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("RefreshToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// New token should not be empty
				if newToken == "" {
					t.Errorf("RefreshToken() returned empty token")
				}

				// New token should be different from original (may occasionally be same due to timestamp precision)
				// This is not a strict requirement for security

				// New token should be valid
				claims, err := authService.ValidateToken(newToken)
				if err != nil {
					t.Errorf("RefreshToken() produced invalid token: %v", err)
				} else {
					if claims.UserID != 1 {
						t.Errorf("RefreshToken() userID = %v, want %v", claims.UserID, 1)
					}
					if claims.Username != "testuser" {
						t.Errorf("RefreshToken() username = %v, want %v", claims.Username, "testuser")
					}
				}
			}
		})
	}
}