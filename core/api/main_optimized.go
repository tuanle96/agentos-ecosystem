package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/tuanle96/agentos-ecosystem/core/api/cache"
	"github.com/tuanle96/agentos-ecosystem/core/api/config"
	"github.com/tuanle96/agentos-ecosystem/core/api/database"
	"github.com/tuanle96/agentos-ecosystem/core/api/handlers"
	"github.com/tuanle96/agentos-ecosystem/core/api/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize optimized database connection pool
	dbConfig := database.DefaultPoolConfig()
	dbConfig.MaxOpenConns = 200 // Increased for high concurrency
	dbConfig.MaxIdleConns = 50  // Optimized idle connections
	dbConfig.ConnMaxLifetime = time.Hour * 2
	dbConfig.ConnMaxIdleTime = time.Minute * 30
	dbConfig.QueryTimeout = time.Second * 10

	optimizedDB, err := database.NewOptimizedDB(cfg.DatabaseURL, dbConfig)
	if err != nil {
		log.Fatal("Failed to connect to optimized database:", err)
	}
	defer optimizedDB.Close()

	// Initialize optimized Redis cache
	cacheConfig := cache.DefaultCacheConfig()
	cacheConfig.PoolSize = 200    // Increased pool size
	cacheConfig.MinIdleConns = 20 // More idle connections
	cacheConfig.LocalCacheEnabled = true
	cacheConfig.LocalCacheTTL = time.Minute * 10
	cacheConfig.CompressionEnabled = true
	cacheConfig.PipelineEnabled = true

	optimizedCache, err := cache.NewOptimizedRedisCache(cfg.RedisURL, "", cacheConfig)
	if err != nil {
		log.Fatal("Failed to connect to optimized Redis cache:", err)
	}
	defer optimizedCache.Close()

	// Create legacy database connection for compatibility
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Configure legacy database connection
	db.SetMaxOpenConns(200)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(time.Hour * 2)
	db.SetConnMaxIdleTime(time.Minute * 30)

	// Create legacy Redis client for compatibility
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisURL,
		Password:     "",
		DB:           0,
		PoolSize:     200,
		MinIdleConns: 20,
		DialTimeout:  time.Second * 5,
		ReadTimeout:  time.Second * 3,
		WriteTimeout: time.Second * 3,
	})
	defer rdb.Close()

	// Test connections
	if err := optimizedDB.Ping(); err != nil {
		log.Fatal("Failed to ping optimized database:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Failed to ping Redis:", err)
	}

	// Initialize handlers with both optimized and legacy connections
	h := handlers.New(db, rdb, cfg)
	h.OptimizedDB = optimizedDB
	h.OptimizedCache = optimizedCache

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.New()

	// Add performance middleware first (highest priority)
	r.Use(middleware.PerformanceMiddleware())

	// Add compression middleware
	r.Use(middleware.CompressionMiddleware())

	// Add intelligent caching middleware
	r.Use(middleware.CacheMiddleware(time.Minute * 5))

	// Add rate limiting middleware
	r.Use(middleware.RateLimitMiddleware(1000)) // 1000 requests per second

	// Add connection pool middleware
	r.Use(middleware.ConnectionPoolMiddleware())

	// Add recovery middleware
	r.Use(gin.Recovery())

	// Add logger middleware (after performance middleware to avoid double logging)
	r.Use(gin.Logger())

	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Response-Time", "X-Response-Time-Ms", "X-Cache", "X-RateLimit-Remaining"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint with performance metrics
	r.GET("/health", func(c *gin.Context) {
		start := time.Now()

		// Quick health checks
		dbHealthy := optimizedDB.Ping() == nil
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
				"connection_pool":        true,
				"redis_cache":            true,
				"compression":            true,
				"rate_limiting":          true,
				"performance_monitoring": true,
			},
		}

		if status == "healthy" {
			c.JSON(http.StatusOK, response)
		} else {
			c.JSON(http.StatusServiceUnavailable, response)
		}
	})

	// Performance profiling endpoints (Week 6 - Development only)
	if cfg.Environment != "production" {
		// Add pprof endpoints for performance profiling
		r.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
		log.Println("Performance profiling endpoints enabled at /debug/pprof/")
	}

	// API routes with authentication middleware
	api := r.Group("/api/v1")
	{
		// Authentication routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", h.Register)
			auth.POST("/login", h.Login)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User management
			protected.GET("/users/profile", h.GetProfile)
			protected.PUT("/users/profile", h.UpdateProfile)

			// Agent management
			protected.POST("/agents", h.CreateAgent)
			protected.GET("/agents", h.GetAgents)
			protected.GET("/agents/:id", h.GetAgent)
			protected.PUT("/agents/:id", h.UpdateAgent)
			protected.DELETE("/agents/:id", h.DeleteAgent)
			protected.POST("/agents/:id/execute", h.ExecuteAgent)

			// Tool management
			protected.GET("/tools", h.GetTools)
			protected.GET("/tools/definitions", h.GetToolDefinitions)
			protected.POST("/tools/execute", h.ExecuteTool)

			// Memory management
			protected.POST("/memory/store", h.StoreMemory)
			protected.POST("/memory/search", h.SearchMemory)
			protected.GET("/memory/agents/:agent_id", h.GetAgentMemory)

			// Tool Marketplace (Week 5)
			protected.POST("/marketplace/tools", h.CreateTool)
			protected.GET("/marketplace/tools", h.GetTools)
			protected.GET("/marketplace/tools/:id", h.GetTool)
			protected.PUT("/marketplace/tools/:id", h.UpdateTool)
			protected.DELETE("/marketplace/tools/:id", h.DeleteTool)
			protected.POST("/marketplace/tools/install", h.InstallTool)

			// Performance Monitoring (Week 6)
			if cfg.Environment != "production" {
				protected.GET("/performance/metrics", h.GetPerformanceMetrics)
				protected.GET("/performance/health", h.GetSystemHealth)
				protected.GET("/performance/benchmark", h.GetPerformanceBenchmark)
			}
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

	log.Printf("🚀 AgentOS Core API (Optimized) starting on port %s", cfg.Port)
	log.Printf("📊 Performance optimizations enabled:")
	log.Printf("   - Database pool: %d max connections", dbConfig.MaxOpenConns)
	log.Printf("   - Redis pool: %d connections", cacheConfig.PoolSize)
	log.Printf("   - Local cache: %v", cacheConfig.LocalCacheEnabled)
	log.Printf("   - Compression: %v", cacheConfig.CompressionEnabled)
	log.Printf("   - Rate limiting: 1000 req/s")
	log.Printf("   - Performance monitoring: enabled")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}
