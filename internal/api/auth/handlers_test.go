package auth

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"bookrank/internal/auth"
	"bookrank/internal/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto migrate tables
	if err := db.AutoMigrate(&models.User{}); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func setupTestHandler(t *testing.T) (*AuthHandler, *gorm.DB) {
	db := setupTestDB(t)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	authService := auth.NewAuthService("test-secret", logger)
	handler := NewAuthHandler(db, authService, logger)
	return handler, db
}

func TestAuthHandler_Register(t *testing.T) {
	handler, _ := setupTestHandler(t)

	tests := map[string]struct {
		requestBody  interface{}
		wantStatus   int
		checkDB      bool
	}{
		"valid_registration": {
			requestBody: models.UserRegisterRequest{
				Username: "testuser",
				Password: "password123!",
			},
			wantStatus: http.StatusCreated,
			checkDB:    true,
		},
		"invalid_username_too_short": {
			requestBody: models.UserRegisterRequest{
				Username: "ab",
				Password: "password123!",
			},
			wantStatus: http.StatusBadRequest,
			checkDB:    false,
		},
		"invalid_password_no_special": {
			requestBody: models.UserRegisterRequest{
				Username: "testuser2",
				Password: "password123",
			},
			wantStatus: http.StatusBadRequest,
			checkDB:    false,
		},
		"invalid_json": {
			requestBody: "invalid json",
			wantStatus:  http.StatusBadRequest,
			checkDB:     false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Register(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Register() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create test user
	hashedPassword, _ := handler.authService.HashPassword("password123!")
	testUser := models.User{
		Username:     "testuser",
		PasswordHash: hashedPassword,
	}
	db.Create(&testUser)

	tests := map[string]struct {
		requestBody interface{}
		wantStatus  int
	}{
		"valid_login": {
			requestBody: models.UserLoginRequest{
				Username: "testuser",
				Password: "password123!",
			},
			wantStatus: http.StatusOK,
		},
		"invalid_password": {
			requestBody: models.UserLoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			wantStatus: http.StatusUnauthorized,
		},
		"nonexistent_user": {
			requestBody: models.UserLoginRequest{
				Username: "nonexistent",
				Password: "password123!",
			},
			wantStatus: http.StatusUnauthorized,
		},
		"invalid_json": {
			requestBody: "invalid json",
			wantStatus:  http.StatusBadRequest,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Login() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}