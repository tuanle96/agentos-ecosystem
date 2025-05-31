package main

import (
	"database/sql"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/tuanle96/agentos-ecosystem/core/api/config"
	"github.com/tuanle96/agentos-ecosystem/core/api/handlers"
	"github.com/tuanle96/agentos-ecosystem/core/api/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database connection
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Connected to PostgreSQL database")

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
		DB:   0,
	})

	// Initialize Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "agentos-core-api",
			"version":   "0.1.0-mvp",
			"timestamp": "2024-12-27",
		})
	})

	// Performance profiling endpoints (Week 6 - Development only)
	if cfg.Environment != "production" {
		// Add pprof endpoints for performance profiling
		r.GET("/debug/pprof/*any", gin.WrapH(http.DefaultServeMux))
		log.Println("Performance profiling endpoints enabled at /debug/pprof/")
	}

	// Initialize handlers with dependencies
	h := handlers.New(db, rdb, cfg)

	// Public routes (no authentication required)
	public := r.Group("/api/v1")
	{
		// Authentication
		public.POST("/auth/register", h.Register)
		public.POST("/auth/login", h.Login)
		public.GET("/tools", h.ListTools) // Public tool listing
	}

	// Protected routes (authentication required)
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Agent management
		protected.GET("/agents", h.ListAgents)
		protected.POST("/agents", h.CreateAgent)
		protected.GET("/agents/:id", h.GetAgent)
		protected.PUT("/agents/:id", h.UpdateAgent)
		protected.DELETE("/agents/:id", h.DeleteAgent)

		// Agent Factory (Week 2 Enhancement)
		protected.GET("/agents/:agent_id/recommendations", h.GetCapabilityRecommendations)
		protected.POST("/agents/validate-capabilities", h.ValidateCapabilities)

		// Tool Execution (Week 2 Enhancement)
		protected.GET("/tools/definitions", h.GetToolDefinitions)
		protected.POST("/tools/execute", h.ExecuteTool)
		protected.GET("/tools/executions/:execution_id", h.GetToolExecution)

		// Agent execution
		protected.POST("/agents/:id/execute", h.ExecuteAgent)
		protected.GET("/agents/:id/executions", h.GetAgentExecutions)

		// Execution management
		protected.GET("/executions/:id", h.GetExecution)
		protected.GET("/executions/:id/logs", h.GetExecutionLogs)
		protected.GET("/executions/agents/:agent_id", h.GetAgentExecutions)

		// Memory management (Week 2 Enhancement)
		protected.GET("/agents/:id/memory", h.GetAgentMemory)
		protected.DELETE("/agents/:id/memory", h.ClearAgentMemory)

		// Enhanced Memory management
		protected.GET("/memory/agents/:agent_id", h.GetAgentMemoryEnhanced)
		protected.DELETE("/memory/agents/:agent_id", h.ClearAgentMemoryEnhanced)
		protected.POST("/memory/working-sessions", h.CreateWorkingMemorySession)
		protected.PUT("/memory/working-sessions/:session_id", h.UpdateWorkingMemory)

		// User profile
		protected.GET("/profile", h.GetProfile)
		protected.PUT("/profile", h.UpdateProfile)

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

	// Start server
	port := os.Getenv("CORE_API_PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Starting AgentOS Core API on port %s", port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Fatal(r.Run(":" + port))
}
