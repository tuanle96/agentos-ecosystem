package handlers

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// PerformanceMetrics represents system performance metrics
type PerformanceMetrics struct {
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`

	// Go Runtime Metrics
	GoVersion   string `json:"go_version"`
	Goroutines  int    `json:"goroutines"`
	MemoryAlloc uint64 `json:"memory_alloc_bytes"`
	MemorySys   uint64 `json:"memory_sys_bytes"`
	MemoryHeap  uint64 `json:"memory_heap_bytes"`
	GCCycles    uint32 `json:"gc_cycles"`

	// Database Metrics
	DBConnections     int `json:"db_connections"`
	DBMaxConnections  int `json:"db_max_connections"`
	DBIdleConnections int `json:"db_idle_connections"`

	// Redis Metrics
	RedisConnections int    `json:"redis_connections"`
	RedisMemoryUsage string `json:"redis_memory_usage"`
	RedisHitRate     string `json:"redis_hit_rate"`

	// Custom Application Metrics
	ActiveAgents     int `json:"active_agents"`
	TotalExecutions  int `json:"total_executions"`
	MemoryOperations int `json:"memory_operations"`
}

// GetPerformanceMetrics returns comprehensive system performance metrics
func (h *Handler) GetPerformanceMetrics(c *gin.Context) {
	ctx := context.Background()

	// Get Go runtime metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Get database metrics
	dbStats := h.db.Stats()

	// Get Redis metrics
	redisInfo, err := h.redis.Info(ctx, "memory", "stats").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get Redis metrics",
			"details": err.Error(),
		})
		return
	}

	// Get application-specific metrics
	activeAgents, err := h.getActiveAgentsCount()
	if err != nil {
		activeAgents = -1 // Indicate error
	}

	totalExecutions, err := h.getTotalExecutionsCount()
	if err != nil {
		totalExecutions = -1 // Indicate error
	}

	memoryOperations, err := h.getMemoryOperationsCount()
	if err != nil {
		memoryOperations = -1 // Indicate error
	}

	metrics := PerformanceMetrics{
		Timestamp: time.Now(),
		Service:   "agentos-core-api",
		Version:   "0.1.0-mvp-week6",

		// Go Runtime
		GoVersion:   runtime.Version(),
		Goroutines:  runtime.NumGoroutine(),
		MemoryAlloc: m.Alloc,
		MemorySys:   m.Sys,
		MemoryHeap:  m.HeapAlloc,
		GCCycles:    m.NumGC,

		// Database
		DBConnections:     dbStats.OpenConnections,
		DBMaxConnections:  dbStats.MaxOpenConnections,
		DBIdleConnections: dbStats.Idle,

		// Redis
		RedisConnections: int(h.redis.PoolStats().TotalConns),
		RedisMemoryUsage: "N/A", // Will be parsed from redisInfo
		RedisHitRate:     "N/A", // Will be parsed from redisInfo

		// Application
		ActiveAgents:     activeAgents,
		TotalExecutions:  totalExecutions,
		MemoryOperations: memoryOperations,
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     "success",
		"data":       metrics,
		"redis_info": redisInfo, // Include raw Redis info for detailed analysis
	})
}

// GetSystemHealth returns detailed system health information
func (h *Handler) GetSystemHealth(c *gin.Context) {
	ctx := context.Background()

	health := gin.H{
		"timestamp": time.Now(),
		"service":   "agentos-core-api",
		"version":   "0.1.0-mvp-week6",
		"status":    "healthy",
		"checks":    gin.H{},
	}

	checks := health["checks"].(gin.H)

	// Database health check
	if err := h.db.Ping(); err != nil {
		checks["database"] = gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		health["status"] = "degraded"
	} else {
		checks["database"] = gin.H{
			"status":      "healthy",
			"connections": h.db.Stats().OpenConnections,
		}
	}

	// Redis health check
	if err := h.redis.Ping(ctx).Err(); err != nil {
		checks["redis"] = gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		health["status"] = "degraded"
	} else {
		checks["redis"] = gin.H{
			"status":      "healthy",
			"connections": int(h.redis.PoolStats().TotalConns),
		}
	}

	// Memory health check
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memoryHealthy := true
	memoryStatus := "healthy"

	// Check if memory usage is concerning (>1GB allocated)
	if m.Alloc > 1024*1024*1024 {
		memoryHealthy = false
		memoryStatus = "warning"
	}

	checks["memory"] = gin.H{
		"status":     memoryStatus,
		"alloc_mb":   m.Alloc / 1024 / 1024,
		"sys_mb":     m.Sys / 1024 / 1024,
		"heap_mb":    m.HeapAlloc / 1024 / 1024,
		"goroutines": runtime.NumGoroutine(),
		"gc_cycles":  m.NumGC,
	}

	if !memoryHealthy && health["status"] == "healthy" {
		health["status"] = "warning"
	}

	c.JSON(http.StatusOK, health)
}

// Helper functions for application metrics

func (h *Handler) getActiveAgentsCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM agents WHERE status = 'active'`
	err := h.db.QueryRow(query).Scan(&count)
	return count, err
}

func (h *Handler) getTotalExecutionsCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM executions`
	err := h.db.QueryRow(query).Scan(&count)
	return count, err
}

func (h *Handler) getMemoryOperationsCount() (int, error) {
	// This would count memory operations from a hypothetical memory_operations table
	// For now, return a placeholder
	return 0, nil
}

// GetPerformanceBenchmark runs a quick performance benchmark
func (h *Handler) GetPerformanceBenchmark(c *gin.Context) {
	ctx := context.Background()

	benchmarks := gin.H{
		"timestamp": time.Now(),
		"service":   "agentos-core-api",
		"tests":     gin.H{},
	}

	tests := benchmarks["tests"].(gin.H)

	// Database benchmark
	start := time.Now()
	if err := h.db.Ping(); err != nil {
		tests["database_ping"] = gin.H{
			"status":   "failed",
			"error":    err.Error(),
			"duration": time.Since(start).Milliseconds(),
		}
	} else {
		tests["database_ping"] = gin.H{
			"status":   "success",
			"duration": time.Since(start).Milliseconds(),
		}
	}

	// Redis benchmark
	start = time.Now()
	if err := h.redis.Ping(ctx).Err(); err != nil {
		tests["redis_ping"] = gin.H{
			"status":   "failed",
			"error":    err.Error(),
			"duration": time.Since(start).Milliseconds(),
		}
	} else {
		tests["redis_ping"] = gin.H{
			"status":   "success",
			"duration": time.Since(start).Milliseconds(),
		}
	}

	// Simple query benchmark
	start = time.Now()
	var count int
	if err := h.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		tests["simple_query"] = gin.H{
			"status":   "failed",
			"error":    err.Error(),
			"duration": time.Since(start).Milliseconds(),
		}
	} else {
		tests["simple_query"] = gin.H{
			"status":   "success",
			"duration": time.Since(start).Milliseconds(),
			"result":   count,
		}
	}

	// Redis set/get benchmark
	start = time.Now()
	testKey := "benchmark_test"
	testValue := "benchmark_value"

	if err := h.redis.Set(ctx, testKey, testValue, time.Minute).Err(); err != nil {
		tests["redis_set"] = gin.H{
			"status":   "failed",
			"error":    err.Error(),
			"duration": time.Since(start).Milliseconds(),
		}
	} else {
		setDuration := time.Since(start).Milliseconds()

		start = time.Now()
		if val, err := h.redis.Get(ctx, testKey).Result(); err != nil {
			tests["redis_get"] = gin.H{
				"status":   "failed",
				"error":    err.Error(),
				"duration": time.Since(start).Milliseconds(),
			}
		} else {
			tests["redis_set"] = gin.H{
				"status":   "success",
				"duration": setDuration,
			}
			tests["redis_get"] = gin.H{
				"status":   "success",
				"duration": time.Since(start).Milliseconds(),
				"value":    val,
			}
		}

		// Cleanup
		h.redis.Del(ctx, testKey)
	}

	c.JSON(http.StatusOK, benchmarks)
}
