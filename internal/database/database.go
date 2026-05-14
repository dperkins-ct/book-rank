package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"bookrank/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is a global database instance
var DB *gorm.DB

// Initialize initializes the database connection with the given configuration
func Initialize(config *Config) (*gorm.DB, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid database configuration: %w", err)
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(config.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pool configuration
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), config.QueryTimeout)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Database connection established successfully")

	// Auto-migrate schemas
	if err := models.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to auto-migrate schemas: %w", err)
	}

	slog.Info("Database schemas migrated successfully")

	// Create custom indexes
	if err := models.CreateIndexes(db); err != nil {
		return nil, fmt.Errorf("failed to create custom indexes: %w", err)
	}

	slog.Info("Database indexes created successfully")

	// Set global DB instance
	DB = db

	return db, nil
}

// Close closes the database connection
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	slog.Info("Database connection closed")
	return nil
}

// GetDB returns the global database instance
func GetDB() *gorm.DB {
	return DB
}

// HealthCheck performs a database health check
func HealthCheck(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}

// GetStats returns database connection statistics
func GetStats() sql.DBStats {
	if DB == nil {
		return sql.DBStats{}
	}

	sqlDB, err := DB.DB()
	if err != nil {
		slog.Error("Failed to get underlying sql.DB for stats", "error", err)
		return sql.DBStats{}
	}

	return sqlDB.Stats()
}

// WithTimeout creates a new database context with timeout
func WithTimeout(timeout time.Duration) (*gorm.DB, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return DB.WithContext(ctx), cancel
}

// WithTransaction executes a function within a database transaction
func WithTransaction(fn func(*gorm.DB) error) error {
	return DB.Transaction(fn)
}