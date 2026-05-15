package api

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"bookrank/internal/auth"
	"bookrank/pkg/cache"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestNewServer tests the NewServer constructor pattern
func TestNewServer(t *testing.T) {
	// Create test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Create test logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	// Create test auth service
	authService := auth.NewAuthService("test-secret", logger)

	// Create test cache (nil is acceptable for testing)
	var testCache cache.Cache

	// Test NewServer constructor
	handler := NewServer(db, authService, logger, testCache)

	// Verify handler is not nil
	if handler == nil {
		t.Fatal("NewServer returned nil handler")
	}

	// Test that the handler responds to basic requests
	tests := map[string]struct {
		method         string
		path           string
		expectedStatus int
	}{
		"root endpoint": {
			method:         "GET",
			path:           "/",
			expectedStatus: http.StatusOK,
		},
		"health endpoint": {
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK, // Should work with in-memory DB
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d for %s %s",
					tt.expectedStatus, recorder.Code, tt.method, tt.path)
			}
		})
	}
}

// TestHandlerConstructorPattern tests that individual handlers follow the constructor pattern
func TestHandlerConstructorPattern(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	tests := map[string]struct {
		handlerFunc func() http.Handler
		testPath    string
		testMethod  string
	}{
		"rootHandler": {
			handlerFunc: func() http.Handler { return rootHandler(logger) },
			testPath:    "/",
			testMethod:  "GET",
		},
		"notImplementedHandler": {
			handlerFunc: func() http.Handler { return notImplementedHandler(logger) },
			testPath:    "/test",
			testMethod:  "GET",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			handler := tt.handlerFunc()
			if handler == nil {
				t.Fatalf("%s returned nil handler", name)
			}

			req := httptest.NewRequest(tt.testMethod, tt.testPath, nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			// Just verify the handler responds (status codes vary by handler)
			if recorder.Code == 0 {
				t.Errorf("%s did not set a status code", name)
			}
		})
	}
}