package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/tuanle96/agentos-ecosystem/core/api/config"
	"github.com/tuanle96/agentos-ecosystem/core/api/handlers"
	"github.com/tuanle96/agentos-ecosystem/core/api/middleware"
)

func mainOptimized() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize optimized database connection
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Configure optimized database connection pool
	db.SetMaxOpenConns(200) // Increased for high concurrency
	db.SetMaxIdleConns(50)  // Optimized idle connections
	db.SetConnMaxLifetime(time.Hour * 2)
	db.SetConnMaxIdleTime(time.Minute * 30)

	// Initialize optimized Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisURL,
		Password:     "",
		DB:           0,
		PoolSize:     200, // Increased pool size
		MinIdleConns: 20,  // More idle connections
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
	})
	defer rdb.Close()

	// Test connections
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to ping Redis:", err)
	}

	// Initialize handlers
	h := handlers.New(db, rdb, cfg)

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.New()

	// Add optimized middleware stack
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Add CORS middleware with optimized settings
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Response-Time", "X-Response-Time-Ms", "X-Cache", "X-RateLimit-Remaining"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Optimized health check endpoint
	r.GET("/health", func(c *gin.Context) {
		start := time.Now()

		// Quick health checks
		dbHealthy := db.Ping() == nil
		redisHealthy := rdb.Ping(c.Request.Context()).Err() == nil

		status := "healthy"
		if !dbHealthy || !redisHealthy {
			status = "degraded"
		}

		response := gin.H{
			"status":           status,
			"service":          "agentos-core-api-optimized",
			"version":          "0.1.0-mvp-week6-day2",
			"timestamp":        time.Now().Format(time.RFC3339),
			"response_time_ms": float64(time.Since(start).Nanoseconds()) / 1e6,
			"database":         dbHealthy,
			"redis":            redisHealthy,
			"optimizations": gin.H{
				"connection_pool": gin.H{
					"max_open_conns": 200,
					"max_idle_conns": 50,
				},
				"redis_pool": gin.H{
					"pool_size":      200,
					"min_idle_conns": 20,
				},
				"features": []string{
					"optimized_connection_pool",
					"redis_optimization",
					"performance_monitoring",
				},
			},
		}

		if status == "healthy" {
			c.JSON(http.StatusOK, response)
		} else {
			c.JSON(http.StatusServiceUnavailable, response)
		}
	})

	// Performance monitoring endpoints
	r.GET("/performance/metrics", func(c *gin.Context) {
		start := time.Now()

		dbStats := db.Stats()
		redisPoolStats := rdb.PoolStats()

		metrics := gin.H{
			"timestamp": time.Now().Format(time.RFC3339),
			"service":   "agentos-core-api-optimized",
			"database": gin.H{
				"open_connections": dbStats.OpenConnections,
				"idle_connections": dbStats.Idle,
				"max_open_conns":   dbStats.MaxOpenConnections,
				"in_use":           dbStats.InUse,
			},
			"redis": gin.H{
				"total_conns": redisPoolStats.TotalConns,
				"idle_conns":  redisPoolStats.IdleConns,
				"stale_conns": redisPoolStats.StaleConns,
			},
			"response_time_ms": float64(time.Since(start).Nanoseconds()) / 1e6,
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   metrics,
		})
	})

	// Performance benchmark endpoint
	r.GET("/performance/benchmark", func(c *gin.Context) {
		start := time.Now()

		// Database benchmark
		dbStart := time.Now()
		db.Ping()
		dbDuration := time.Since(dbStart)

		// Redis benchmark
		redisStart := time.Now()
		rdb.Ping(c.Request.Context())
		redisDuration := time.Since(redisStart)

		benchmark := gin.H{
			"timestamp": time.Now().Format(time.RFC3339),
			"tests": gin.H{
				"database_ping": gin.H{
					"duration_ms": float64(dbDuration.Nanoseconds()) / 1e6,
					"status":      "success",
				},
				"redis_ping": gin.H{
					"duration_ms": float64(redisDuration.Nanoseconds()) / 1e6,
					"status":      "success",
				},
			},
			"total_duration_ms": float64(time.Since(start).Nanoseconds()) / 1e6,
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   benchmark,
		})
	})

	// API routes
	api := r.Group("/api/v1")
	{
		// Authentication routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
		}

		// Protected routes (simplified for testing)
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			// User management
			protected.GET("/users/profile", h.GetProfile)
			protected.PUT("/users/profile", h.UpdateProfile)

			// Basic agent management (only implemented methods)
			protected.POST("/agents", h.CreateAgent)
			protected.GET("/agents/:id", h.GetAgent)
			protected.PUT("/agents/:id", h.UpdateAgent)
			protected.DELETE("/agents/:id", h.DeleteAgent)
			protected.POST("/agents/:id/execute", h.ExecuteAgent)

			// Tool management
			protected.GET("/tools", h.GetTools)
			protected.GET("/tools/definitions", h.GetToolDefinitions)
			protected.POST("/tools/execute", h.ExecuteTool)
		}
	}

	// Start server with optimized settings
	server := &http.Server{
		Addr:           ":" + cfg.Port,
		Handler:        r,
		ReadTimeout:    time.Second * 15,
		WriteTimeout:   time.Second * 15,
		IdleTimeout:    time.Second * 60,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	log.Printf("🚀 AgentOS Core API (Week 6 Optimized) starting on port %s", cfg.Port)
	log.Printf("📊 Performance optimizations enabled:")
	log.Printf("   - Database pool: 200 max connections, 50 idle")
	log.Printf("   - Redis pool: 200 connections, 20 idle")
	log.Printf("   - Optimized timeouts and connection management")
	log.Printf("   - Performance monitoring endpoints available")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}

// Entry point for optimized version
func main() {
	mainOptimized()
}
