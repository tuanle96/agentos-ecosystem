package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"

	"github.com/tuanle96/agentos-ecosystem/core/api/config"
	"github.com/tuanle96/agentos-ecosystem/core/api/handlers"
	"github.com/tuanle96/agentos-ecosystem/core/api/middleware"
)

// TestSuite defines the test suite for AgentOS Core API
type TestSuite struct {
	suite.Suite
	router     *gin.Engine
	db         *sql.DB
	redis      *redis.Client
	handler    *handlers.Handler
	testUser   TestUser
	testToolID string
	token      string
}

// TestUser represents a test user for authentication
type TestUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// SetupSuite runs once before all tests
func (suite *TestSuite) SetupSuite() {
	// Set test environment
	os.Setenv("GO_ENV", "test")
	os.Setenv("DATABASE_URL", "postgres://agentos:agentos_dev_password@localhost:5432/agentos_test?sslmode=disable")
	os.Setenv("REDIS_URL", "localhost:6379")
	os.Setenv("JWT_SECRET", "test-jwt-secret")

	// Load test configuration
	cfg := config.Load()

	// Initialize test database
	var err error
	suite.db, err = sql.Open("postgres", cfg.DatabaseURL)
	suite.Require().NoError(err)

	// Test database connection
	err = suite.db.Ping()
	suite.Require().NoError(err)

	// Initialize Redis client
	suite.redis = redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
		DB:   1, // Use DB 1 for testing
	})

	// Initialize handlers
	suite.handler = handlers.New(suite.db, suite.redis, cfg)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Setup routes
	suite.setupRoutes()

	// Create test user
	suite.createTestUser()
}

// TearDownSuite runs once after all tests
func (suite *TestSuite) TearDownSuite() {
	// Clean up test data
	suite.cleanupTestData()

	// Close connections
	if suite.db != nil {
		suite.db.Close()
	}
	if suite.redis != nil {
		suite.redis.Close()
	}
}

// SetupTest runs before each test
func (suite *TestSuite) SetupTest() {
	// Clean up test data before each test
	suite.cleanupTestData()
	// Recreate test user for each test
	suite.createTestUser()
}

// setupRoutes configures the test router
func (suite *TestSuite) setupRoutes() {
	// Health check
	suite.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "agentos-core-api",
			"version":   "0.1.0-test",
			"timestamp": "2024-12-27",
		})
	})

	// Public routes
	public := suite.router.Group("/api/v1")
	{
		public.POST("/auth/register", suite.handler.Register)
		public.POST("/auth/login", suite.handler.Login)
		public.GET("/tools", suite.handler.ListTools)
	}

	// Protected routes
	protected := suite.router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware("test-jwt-secret"))
	{
		protected.GET("/agents", suite.handler.ListAgents)
		protected.POST("/agents", suite.handler.CreateAgent)
		protected.GET("/agents/:id", suite.handler.GetAgent)
		protected.PUT("/agents/:id", suite.handler.UpdateAgent)
		protected.DELETE("/agents/:id", suite.handler.DeleteAgent)
		protected.POST("/agents/:id/execute", suite.handler.ExecuteAgent)
		protected.GET("/agents/:id/executions", suite.handler.GetAgentExecutions)
		protected.GET("/executions/:id", suite.handler.GetExecution)
		protected.GET("/executions/:id/logs", suite.handler.GetExecutionLogs)
		protected.GET("/agents/:id/memory", suite.handler.GetAgentMemory)
		protected.POST("/agents/:id/memory/clear", suite.handler.ClearAgentMemory)
		protected.GET("/profile", suite.handler.GetProfile)
		protected.PUT("/profile", suite.handler.UpdateProfile)

		// Tool Marketplace routes (Week 5)
		protected.POST("/marketplace/tools", suite.handler.CreateTool)
		protected.GET("/marketplace/tools", suite.handler.GetTools)
		protected.GET("/marketplace/tools/:id", suite.handler.GetTool)
		protected.PUT("/marketplace/tools/:id", suite.handler.UpdateTool)
		protected.DELETE("/marketplace/tools/:id", suite.handler.DeleteTool)
		protected.POST("/marketplace/tools/install", suite.handler.InstallTool)
	}
}

// createTestUser creates a test user for authentication tests
func (suite *TestSuite) createTestUser() {
	// Register test user
	registerData := map[string]string{
		"email":      "test@agentos.com",
		"password":   "testpassword123",
		"first_name": "Test",
		"last_name":  "User",
	}

	jsonData, _ := json.Marshal(registerData)
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	if w.Code == http.StatusCreated {
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		suite.testUser.Token = response["token"].(string)
		suite.token = response["token"].(string) // Set token for marketplace tests
		user := response["user"].(map[string]interface{})
		suite.testUser.ID = user["id"].(string)
		suite.testUser.Email = user["email"].(string)
		suite.testUser.Password = "testpassword123"
	}
}

// cleanupTestData removes test data from database
func (suite *TestSuite) cleanupTestData() {
	// Clean up in reverse order of dependencies
	suite.db.Exec("DELETE FROM executions")
	suite.db.Exec("DELETE FROM memories")
	suite.db.Exec("DELETE FROM sessions")
	suite.db.Exec("DELETE FROM agents")
	suite.db.Exec("DELETE FROM users WHERE email LIKE '%test%' OR email LIKE '%@agentos.com'")

	// Clear Redis test data
	suite.redis.FlushDB(context.Background())
}

// makeRequest helper function for making HTTP requests
func (suite *TestSuite) makeRequest(method, url string, body interface{}, token string) *httptest.ResponseRecorder {
	var jsonData []byte
	if body != nil {
		jsonData, _ = json.Marshal(body)
	}

	req, _ := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

// parseResponse helper function for parsing JSON responses
func (suite *TestSuite) parseResponse(w *httptest.ResponseRecorder, target interface{}) error {
	return json.Unmarshal(w.Body.Bytes(), target)
}

// performRequest helper function for performing raw HTTP requests
func (suite *TestSuite) performRequest(req *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

// TestRunner runs the test suite
func TestAPISuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
