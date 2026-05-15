package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"bookrank/internal/api"
	"bookrank/internal/auth"
	"bookrank/internal/config"
	"bookrank/internal/models"
	"bookrank/pkg/cache"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// Only log if it's not "file not found" - allow running without .env
		if !os.IsNotExist(err) {
			log.Printf("Warning: Error loading .env file: %v", err)
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Setup logger
	logger := setupLogger(cfg.Logging)
	logger.Info("Starting BookRank server", "version", "1.0.0", "environment", os.Getenv("ENVIRONMENT"))

	// Connect to database
	db, err := setupDatabase(cfg, logger)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return err
	}

	// Run database migrations
	if err := runMigrations(db, logger); err != nil {
		logger.Error("Failed to run database migrations", "error", err)
		return err
	}

	// Initialize Redis cache
	redisCache, err := setupRedisCache(cfg, logger)
	if err != nil {
		logger.Error("Failed to setup Redis cache", "error", err)
		return err
	}

	// Initialize services
	authService := auth.NewAuthService(cfg.Auth.JWTSecret, logger)

	// Setup server using NewServer constructor pattern
	httpHandler := api.NewServer(db, authService, logger, redisCache)

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      httpHandler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Server starting", "address", cfg.GetServerAddress())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed to start", "error", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Server shutting down...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	logger.Info("Server exited")
	return nil
}

// setupLogger configures the structured logger
func setupLogger(cfg config.LoggingConfig) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	opts := &slog.HandlerOptions{Level: level}

	if cfg.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// setupDatabase connects to the database and configures connection pooling
func setupDatabase(cfg *config.Config, appLogger *slog.Logger) (*gorm.DB, error) {
	// Configure GORM logger
	var gormLogLevel logger.LogLevel
	switch cfg.Logging.Level {
	case "debug":
		gormLogLevel = logger.Info
	case "info":
		gormLogLevel = logger.Warn
	default:
		gormLogLevel = logger.Error
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.GetDatabaseDSN()), gormConfig)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	appLogger.Info("Database connected successfully",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"database", cfg.Database.DBName,
	)

	return db, nil
}

// runMigrations runs database migrations
func runMigrations(db *gorm.DB, logger *slog.Logger) error {
	logger.Info("Running database migrations...")

	err := models.AutoMigrate(db)
	if err != nil {
		return err
	}

	// Create additional indexes for performance
	err = models.CreateIndexes(db)
	if err != nil {
		logger.Warn("Failed to create some database indexes", "error", err)
		// Don't fail the startup for index creation issues
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

// setupRedisCache initializes Redis cache connection
func setupRedisCache(cfg *config.Config, logger *slog.Logger) (cache.Cache, error) {
	redisCache, err := cache.NewRedisCacheFromURL(cfg.Redis.URL)
	if err != nil {
		return nil, err
	}

	// Test the connection
	ctx := context.Background()
	if err := redisCache.Ping(ctx); err != nil {
		logger.Warn("Redis connection failed, continuing without cache", "error", err)
		// Return a no-op cache or continue without cache
		return nil, nil
	}

	logger.Info("Redis cache connected successfully", "url", cfg.Redis.URL)
	return redisCache, nil
}