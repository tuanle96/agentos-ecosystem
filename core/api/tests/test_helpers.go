package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/tuanle96/agentos-ecosystem/core/api/config"
	"github.com/tuanle96/agentos-ecosystem/core/api/handlers"
	"github.com/tuanle96/agentos-ecosystem/core/api/middleware"
	"github.com/tuanle96/agentos-ecosystem/core/api/models"
)

var (
	testDB     *sql.DB
	testRouter *gin.Engine
	testToken  string
	testUserID string
)

// setupTestDB initializes test database
func setupTestDB(t *testing.T) {
	// Set test environment
	os.Setenv("GO_ENV", "test")
	os.Setenv("DATABASE_URL", "postgres://agentos:agentos_dev_password@localhost:5432/agentos_test?sslmode=disable")

	// Initialize database connection
	cfg := config.Load()
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	require.NoError(t, err)

	testDB = db

	// Create test tables if they don't exist
	createTestTables(t)

	// Setup router
	setupTestRouter(t)
}

// cleanupTestDB cleans up test database
func cleanupTestDB(t *testing.T) {
	if testDB != nil {
		// Clean up test data
		cleanupTestData(t)
		testDB.Close()
	}
}

// createTestTables creates necessary test tables
func createTestTables(t *testing.T) {
	// Users table
	_, err := testDB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			first_name VARCHAR(100),
			last_name VARCHAR(100),
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	// Agents table
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS agents (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			capabilities JSONB,
			personality JSONB,
			framework_preference VARCHAR(50) DEFAULT 'langchain',
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	// Tools table
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS tools (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) UNIQUE NOT NULL,
			description TEXT,
			category VARCHAR(100),
			function_schema JSONB,
			implementation_code TEXT,
			is_active BOOLEAN DEFAULT true,
			version VARCHAR(20) DEFAULT '1.0.0',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	// Tool executions table (Week 2)
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS tool_executions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			tool_name VARCHAR(255) NOT NULL,
			user_id UUID REFERENCES users(id),
			agent_id UUID REFERENCES agents(id),
			parameters JSONB,
			result JSONB,
			status VARCHAR(50) DEFAULT 'pending',
			execution_time FLOAT,
			started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			completed_at TIMESTAMP,
			error_message TEXT
		)
	`)
	require.NoError(t, err)

	// Working memory sessions table (Week 2)
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS working_memory_sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
			session_id VARCHAR(255) UNIQUE NOT NULL,
			variables JSONB,
			context JSONB,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)
}

// setupTestRouter initializes test router
func setupTestRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(gin.Recovery())

	// Initialize handler with test database
	cfg := &config.Config{
		JWTSecret: "test-secret-key",
	}
	handler := handlers.New(testDB, nil, cfg)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "agentos-core-api",
			"status":    "healthy",
			"timestamp": time.Now().Format("2006-01-02"),
			"version":   "0.1.0-mvp",
		})
	})

	// Auth routes
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
	}

	// Public routes
	router.GET("/api/v1/tools", handler.ListTools)

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		// Agent routes
		agents := protected.Group("/agents")
		{
			agents.GET("", handler.ListAgents)
			agents.POST("", handler.CreateAgent)
			agents.GET("/:id", handler.GetAgent)
			agents.PUT("/:id", handler.UpdateAgent)
			agents.DELETE("/:id", handler.DeleteAgent)
			agents.POST("/:id/execute", handler.ExecuteAgent)
			agents.GET("/:id/executions", handler.GetAgentExecutions)
			agents.GET("/:id/memory", handler.GetAgentMemory)
			agents.POST("/:id/memory/clear", handler.ClearAgentMemory)
			agents.POST("/:id/memory/session", handler.CreateWorkingMemorySession)
			agents.PUT("/:id/memory/working", handler.UpdateWorkingMemory)
		}

		// Week 2: Agent Factory routes
		capabilities := protected.Group("/capabilities")
		{
			capabilities.GET("/recommendations", handler.GetCapabilityRecommendations)
			capabilities.POST("/validate", handler.ValidateCapabilities)
		}

		// Week 2: Tool Execution routes
		tools := protected.Group("/tools")
		{
			tools.GET("/definitions", handler.GetToolDefinitions)
			tools.POST("/execute", handler.ExecuteTool)
			tools.GET("/executions/:execution_id", handler.GetToolExecution)
		}

		// Execution routes
		executions := protected.Group("/executions")
		{
			executions.GET("/:id", handler.GetExecution)
			executions.GET("/:id/logs", handler.GetExecutionLogs)
		}

		// Profile routes
		protected.GET("/profile", handler.GetProfile)
		protected.PUT("/profile", handler.UpdateProfile)
	}

	testRouter = router
}

// createTestUserAndGetToken creates a test user and returns JWT token
func createTestUserAndGetToken(t *testing.T) string {
	// Register test user
	payload := map[string]interface{}{
		"email":    fmt.Sprintf("test_%d@agentos.dev", time.Now().Unix()),
		"password": "testpass123",
		"name":     "Test User",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	token := response["token"].(string)
	user := response["user"].(map[string]interface{})
	testUserID = user["id"].(string)

	return token
}

// cleanupTestData removes test data
func cleanupTestData(t *testing.T) {
	tables := []string{
		"tool_executions",
		"working_memory_sessions",
		"agents",
		"users",
	}

	for _, table := range tables {
		_, err := testDB.Exec(fmt.Sprintf("DELETE FROM %s WHERE created_at > NOW() - INTERVAL '1 hour'", table))
		if err != nil {
			t.Logf("Warning: Could not clean up table %s: %v", table, err)
		}
	}
}

// insertTestTools inserts test tools for testing
func insertTestTools(t *testing.T) {
	tools := []struct {
		Name           string
		Description    string
		Category       string
		FunctionSchema models.JSONB
		IsActive       bool
		Version        string
	}{
		{
			Name:        "web_search",
			Description: "Search the web for information using DuckDuckGo",
			Category:    "search",
			FunctionSchema: models.JSONB{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type": "string",
					},
					"max_results": map[string]interface{}{
						"type":    "integer",
						"default": 5,
					},
				},
			},
			IsActive: true,
			Version:  "1.0.0",
		},
		{
			Name:        "calculations",
			Description: "Perform mathematical calculations",
			Category:    "math",
			FunctionSchema: models.JSONB{
				"type": "object",
				"properties": map[string]interface{}{
					"expression": map[string]interface{}{
						"type": "string",
					},
				},
			},
			IsActive: true,
			Version:  "1.0.0",
		},
		{
			Name:        "text_processing",
			Description: "Process and analyze text content",
			Category:    "text",
			FunctionSchema: models.JSONB{
				"type": "object",
				"properties": map[string]interface{}{
					"text": map[string]interface{}{
						"type": "string",
					},
					"operation": map[string]interface{}{
						"type": "string",
						"enum": []string{"summarize", "analyze", "extract"},
					},
				},
			},
			IsActive: true,
			Version:  "1.0.0",
		},
	}

	for _, tool := range tools {
		schemaJSON, _ := json.Marshal(tool.FunctionSchema)
		_, err := testDB.Exec(`
			INSERT INTO tools (name, description, category, function_schema, is_active, version)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (name) DO NOTHING
		`, tool.Name, &tool.Description, &tool.Category, schemaJSON, tool.IsActive, tool.Version)
		require.NoError(t, err)
	}
}
