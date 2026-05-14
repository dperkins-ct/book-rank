package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	Port        int
	DatabaseURL string
	JWTSecret   string
	Environment string
	Redis       RedisConfig
	Cache       CacheConfig
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	URL      string
	Password string
	DB       int
}

// CacheConfig holds cache-related configuration
type CacheConfig struct {
	RecommendationTTL int // TTL in seconds for recommendation cache
	DefaultTTL        int // Default TTL in seconds
}

// Load reads configuration from environment variables with defaults
func Load() (*Config, error) {
	cfg := &Config{
		Port:        getEnvAsInt("PORT", 8080),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://bookrank:bookrank@localhost:5432/bookrank?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "your-secret-key"),
		Environment: getEnv("ENVIRONMENT", "development"),
		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "redis://localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Cache: CacheConfig{
			RecommendationTTL: getEnvAsInt("CACHE_RECOMMENDATION_TTL", 3600), // 1 hour default
			DefaultTTL:        getEnvAsInt("CACHE_DEFAULT_TTL", 1800),        // 30 minutes default
		},
	}

	return cfg, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}