package database

import (
	"context"
	"os"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func TestNewConfig(t *testing.T) {
	// Save original env vars
	originalValues := map[string]string{
		"DB_HOST":             os.Getenv("DB_HOST"),
		"DB_PORT":             os.Getenv("DB_PORT"),
		"DB_USER":             os.Getenv("DB_USER"),
		"DB_PASSWORD":         os.Getenv("DB_PASSWORD"),
		"DB_NAME":             os.Getenv("DB_NAME"),
		"DB_SSL_MODE":         os.Getenv("DB_SSL_MODE"),
		"DB_MAX_OPEN_CONNS":   os.Getenv("DB_MAX_OPEN_CONNS"),
		"DB_MAX_IDLE_CONNS":   os.Getenv("DB_MAX_IDLE_CONNS"),
		"DB_CONN_MAX_LIFETIME": os.Getenv("DB_CONN_MAX_LIFETIME"),
		"DB_CONN_MAX_IDLE_TIME": os.Getenv("DB_CONN_MAX_IDLE_TIME"),
		"DB_QUERY_TIMEOUT":    os.Getenv("DB_QUERY_TIMEOUT"),
	}

	// Clean up environment
	defer func() {
		for key, value := range originalValues {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("default values", func(t *testing.T) {
		// Clear all env vars
		for key := range originalValues {
			os.Unsetenv(key)
		}

		config := NewConfig()

		if config.Host != "localhost" {
			t.Errorf("Host = %v, want 'localhost'", config.Host)
		}
		if config.Port != 5432 {
			t.Errorf("Port = %v, want 5432", config.Port)
		}
		if config.User != "postgres" {
			t.Errorf("User = %v, want 'postgres'", config.User)
		}
		if config.Database != "bookrank" {
			t.Errorf("Database = %v, want 'bookrank'", config.Database)
		}
		if config.MaxOpenConns != 25 {
			t.Errorf("MaxOpenConns = %v, want 25", config.MaxOpenConns)
		}
		if config.MaxIdleConns != 5 {
			t.Errorf("MaxIdleConns = %v, want 5", config.MaxIdleConns)
		}
	})

	t.Run("environment overrides", func(t *testing.T) {
		os.Setenv("DB_HOST", "testhost")
		os.Setenv("DB_PORT", "5433")
		os.Setenv("DB_USER", "testuser")
		os.Setenv("DB_PASSWORD", "testpass")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_MAX_OPEN_CONNS", "50")
		os.Setenv("DB_MAX_IDLE_CONNS", "10")

		config := NewConfig()

		if config.Host != "testhost" {
			t.Errorf("Host = %v, want 'testhost'", config.Host)
		}
		if config.Port != 5433 {
			t.Errorf("Port = %v, want 5433", config.Port)
		}
		if config.User != "testuser" {
			t.Errorf("User = %v, want 'testuser'", config.User)
		}
		if config.Password != "testpass" {
			t.Errorf("Password = %v, want 'testpass'", config.Password)
		}
		if config.Database != "testdb" {
			t.Errorf("Database = %v, want 'testdb'", config.Database)
		}
		if config.MaxOpenConns != 50 {
			t.Errorf("MaxOpenConns = %v, want 50", config.MaxOpenConns)
		}
		if config.MaxIdleConns != 10 {
			t.Errorf("MaxIdleConns = %v, want 10", config.MaxIdleConns)
		}
	})
}

func TestConfigValidate(t *testing.T) {
	tests := map[string]struct {
		config  Config
		wantErr bool
	}{
		"valid config": {
			config: Config{
				Host:         "localhost",
				User:         "postgres",
				Database:     "bookrank",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantErr: false,
		},
		"empty host": {
			config: Config{
				Host:         "",
				User:         "postgres",
				Database:     "bookrank",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantErr: true,
		},
		"empty user": {
			config: Config{
				Host:         "localhost",
				User:         "",
				Database:     "bookrank",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantErr: true,
		},
		"empty database": {
			config: Config{
				Host:         "localhost",
				User:         "postgres",
				Database:     "",
				MaxOpenConns: 25,
				MaxIdleConns: 5,
			},
			wantErr: true,
		},
		"zero max open conns": {
			config: Config{
				Host:         "localhost",
				User:         "postgres",
				Database:     "bookrank",
				MaxOpenConns: 0,
				MaxIdleConns: 5,
			},
			wantErr: true,
		},
		"zero max idle conns": {
			config: Config{
				Host:         "localhost",
				User:         "postgres",
				Database:     "bookrank",
				MaxOpenConns: 25,
				MaxIdleConns: 0,
			},
			wantErr: true,
		},
		"idle conns > open conns": {
			config: Config{
				Host:         "localhost",
				User:         "postgres",
				Database:     "bookrank",
				MaxOpenConns: 5,
				MaxIdleConns: 25,
			},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := tc.config.Validate()
			if (err != nil) != tc.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestConfigDSN(t *testing.T) {
	config := Config{
		Host:     "testhost",
		Port:     5433,
		User:     "testuser",
		Password: "testpass",
		Database: "testdb",
		SSLMode:  "require",
	}

	expected := "host=testhost port=5433 user=testuser password=testpass dbname=testdb sslmode=require"
	got := config.DSN()

	if got != expected {
		t.Errorf("DSN() = %v, want %v", got, expected)
	}
}

func TestHealthCheck(t *testing.T) {
	// Test with no DB initialized
	ctx := context.Background()
	err := HealthCheck(ctx)
	if err == nil {
		t.Error("HealthCheck should fail when DB is not initialized")
	}

	// Test with mock database (would need actual database connection for full test)
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Set global DB for testing
	DB = db

	err = HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck should pass with valid database: %v", err)
	}

	// Test with timeout context
	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// This might pass quickly with SQLite, so we'll just verify it doesn't panic
	_ = HealthCheck(ctxWithTimeout)
}

func TestWithTimeout(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	DB = db

	dbWithTimeout, cancel := WithTimeout(30 * time.Second)
	defer cancel()

	if dbWithTimeout == nil {
		t.Error("WithTimeout should return a database instance")
	}

	// Verify we can use the database with timeout
	var count int64
	err = dbWithTimeout.Raw("SELECT 1").Count(&count).Error
	if err != nil {
		t.Errorf("Should be able to query with timeout database: %v", err)
	}
}

func TestWithTransaction(t *testing.T) {
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	DB = db

	// Test successful transaction
	err = WithTransaction(func(tx *gorm.DB) error {
		// Do nothing, just test the transaction wrapper
		return nil
	})

	if err != nil {
		t.Errorf("WithTransaction should succeed: %v", err)
	}

	// Test failing transaction
	testErr := gorm.ErrInvalidTransaction
	err = WithTransaction(func(tx *gorm.DB) error {
		return testErr
	})

	if err != testErr {
		t.Errorf("WithTransaction should return the error from the function: got %v, want %v", err, testErr)
	}
}

func TestGetStats(t *testing.T) {
	// Test with no DB
	stats := GetStats()
	// Should return empty stats without panicking
	if stats.OpenConnections < 0 {
		t.Error("Stats should not have negative values")
	}

	// Test with database
	db, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	DB = db

	stats = GetStats()
	if stats.OpenConnections < 0 {
		t.Error("Stats should not have negative values")
	}
}

func TestGetDB(t *testing.T) {
	// Test with no DB
	DB = nil
	db := GetDB()
	if db != nil {
		t.Error("GetDB should return nil when no database is initialized")
	}

	// Test with DB
	testDB, err := setupTestDatabase()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	DB = testDB
	db = GetDB()
	if db != testDB {
		t.Error("GetDB should return the initialized database")
	}
}