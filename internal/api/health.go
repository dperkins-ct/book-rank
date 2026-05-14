package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"
	"bookrank/pkg/cache"
)

// HealthStatus represents the overall health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status      HealthStatus           `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Checks      map[string]HealthCheck `json:"checks"`
	Uptime      string                 `json:"uptime"`
}

// HealthCheck represents an individual health check
type HealthCheck struct {
	Status      HealthStatus `json:"status"`
	Message     string       `json:"message,omitempty"`
	Duration    string       `json:"duration"`
	LastChecked time.Time    `json:"last_checked"`
}

// HealthChecker handles health checks
type HealthChecker struct {
	db        *gorm.DB
	cache     cache.Cache
	startTime time.Time
	version   string
	env       string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(db *gorm.DB, cache cache.Cache, version, env string) *HealthChecker {
	return &HealthChecker{
		db:        db,
		cache:     cache,
		startTime: time.Now(),
		version:   version,
		env:       env,
	}
}

// CheckHealth performs comprehensive health checks
func (h *HealthChecker) CheckHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	checks := make(map[string]HealthCheck)
	overallStatus := HealthStatusHealthy

	// Database health check
	dbCheck := h.checkDatabase(ctx)
	checks["database"] = dbCheck
	if dbCheck.Status != HealthStatusHealthy {
		overallStatus = HealthStatusDegraded
	}

	// Cache health check (optional - don't fail if cache is down)
	if h.cache != nil {
		cacheCheck := h.checkCache(ctx)
		checks["cache"] = cacheCheck
		// Cache failure only degrades performance, doesn't make service unhealthy
		if cacheCheck.Status != HealthStatusHealthy && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	// Memory check
	memCheck := h.checkMemory()
	checks["memory"] = memCheck
	if memCheck.Status == HealthStatusUnhealthy {
		overallStatus = HealthStatusUnhealthy
	}

	// Disk check
	diskCheck := h.checkDisk()
	checks["disk"] = diskCheck
	if diskCheck.Status == HealthStatusUnhealthy {
		overallStatus = HealthStatusUnhealthy
	}

	response := HealthResponse{
		Status:      overallStatus,
		Timestamp:   time.Now(),
		Version:     h.version,
		Environment: h.env,
		Checks:      checks,
		Uptime:      time.Since(h.startTime).String(),
	}

	// Set appropriate HTTP status code
	statusCode := http.StatusOK
	if overallStatus == HealthStatusDegraded {
		statusCode = http.StatusOK // 200 for degraded (still serving traffic)
	} else if overallStatus == HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable // 503 for unhealthy
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// checkDatabase verifies database connectivity and basic functionality
func (h *HealthChecker) checkDatabase(ctx context.Context) HealthCheck {
	start := time.Now()

	// Test database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		return HealthCheck{
			Status:      HealthStatusUnhealthy,
			Message:     "Failed to get database connection: " + err.Error(),
			Duration:    time.Since(start).String(),
			LastChecked: time.Now(),
		}
	}

	// Ping database
	if err := sqlDB.PingContext(ctx); err != nil {
		return HealthCheck{
			Status:      HealthStatusUnhealthy,
			Message:     "Database ping failed: " + err.Error(),
			Duration:    time.Since(start).String(),
			LastChecked: time.Now(),
		}
	}

	// Check database stats
	stats := sqlDB.Stats()
	if stats.OpenConnections >= stats.MaxOpenConnections-5 {
		return HealthCheck{
			Status:      HealthStatusDegraded,
			Message:     "Database connection pool nearly exhausted",
			Duration:    time.Since(start).String(),
			LastChecked: time.Now(),
		}
	}

	// Test a simple query
	var result int
	err = h.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		return HealthCheck{
			Status:      HealthStatusUnhealthy,
			Message:     "Database query failed: " + err.Error(),
			Duration:    time.Since(start).String(),
			LastChecked: time.Now(),
		}
	}

	return HealthCheck{
		Status:      HealthStatusHealthy,
		Duration:    time.Since(start).String(),
		LastChecked: time.Now(),
	}
}

// checkCache verifies cache connectivity
func (h *HealthChecker) checkCache(ctx context.Context) HealthCheck {
	start := time.Now()

	// Test cache connectivity
	if err := h.cache.Ping(ctx); err != nil {
		return HealthCheck{
			Status:      HealthStatusUnhealthy,
			Message:     "Cache ping failed: " + err.Error(),
			Duration:    time.Since(start).String(),
			LastChecked: time.Now(),
		}
	}

	// Test cache operations
	testKey := "health_check_" + time.Now().Format("20060102150405")
	testValue := "test"

	// Test set
	if err := h.cache.Set(ctx, testKey, testValue, 30*time.Second); err != nil {
		return HealthCheck{
			Status:      HealthStatusDegraded,
			Message:     "Cache set operation failed: " + err.Error(),
			Duration:    time.Since(start).String(),
			LastChecked: time.Now(),
		}
	}

	// Test get
	var retrieved string
	if err := h.cache.Get(ctx, testKey, &retrieved); err != nil {
		return HealthCheck{
			Status:      HealthStatusDegraded,
			Message:     "Cache get operation failed: " + err.Error(),
			Duration:    time.Since(start).String(),
			LastChecked: time.Now(),
		}
	}

	// Test delete
	if err := h.cache.Delete(ctx, testKey); err != nil {
		// Not critical if delete fails
	}

	return HealthCheck{
		Status:      HealthStatusHealthy,
		Duration:    time.Since(start).String(),
		LastChecked: time.Now(),
	}
}

// checkMemory checks memory usage (simplified version)
func (h *HealthChecker) checkMemory() HealthCheck {
	start := time.Now()

	// In a real implementation, you would check actual memory usage
	// For now, we'll just return healthy
	// You could use runtime.MemStats for basic Go memory statistics

	return HealthCheck{
		Status:      HealthStatusHealthy,
		Duration:    time.Since(start).String(),
		LastChecked: time.Now(),
	}
}

// checkDisk checks disk space (simplified version)
func (h *HealthChecker) checkDisk() HealthCheck {
	start := time.Now()

	// In a real implementation, you would check actual disk usage
	// For now, we'll just return healthy
	// You could use syscall.Statfs or similar for disk space checks

	return HealthCheck{
		Status:      HealthStatusHealthy,
		Duration:    time.Since(start).String(),
		LastChecked: time.Now(),
	}
}

// ReadinessCheck provides a simpler readiness check for Kubernetes-style probes
func (h *HealthChecker) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Only check critical components for readiness
	dbCheck := h.checkDatabase(ctx)

	if dbCheck.Status == HealthStatusUnhealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"ready": false,
			"error": dbCheck.Message,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"ready": true,
	})
}

// LivenessCheck provides a simple liveness check
func (h *HealthChecker) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"alive": true,
		"uptime": time.Since(h.startTime).String(),
	})
}