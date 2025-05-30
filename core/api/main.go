package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/tuanle96/agentos-ecosystem/core/api/handlers"
	"github.com/tuanle96/agentos-ecosystem/core/api/config"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Initialize Gin router
	r := gin.Default()
	
	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"service": "agentos-core-api",
			"version": "0.1.0",
		})
	})
	
	// API routes
	api := r.Group("/api/v1")
	{
		// Agent management
		api.GET("/agents", handlers.ListAgents)
		api.POST("/agents", handlers.CreateAgent)
		api.GET("/agents/:id", handlers.GetAgent)
		api.PUT("/agents/:id", handlers.UpdateAgent)
		api.DELETE("/agents/:id", handlers.DeleteAgent)
		
		// Task management
		api.GET("/tasks", handlers.ListTasks)
		api.POST("/tasks", handlers.CreateTask)
		api.GET("/tasks/:id", handlers.GetTask)
		api.PUT("/tasks/:id", handlers.UpdateTask)
		api.DELETE("/tasks/:id", handlers.DeleteTask)
		
		// Execution
		api.POST("/execute", handlers.ExecuteTask)
		api.GET("/executions/:id", handlers.GetExecution)
		api.GET("/executions/:id/logs", handlers.GetExecutionLogs)
	}
	
	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Port
	}
	
	log.Printf("Starting AgentOS Core API on port %s", port)
	log.Fatal(r.Run(":" + port))
}
