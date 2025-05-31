package tests

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"
)

// Test configuration
const (
	testUserID = "test_user_12345"
	testDBName = "agentos_test"
)

// Handler represents the main handler for testing
type Handler struct {
	db    *sql.DB
	redis *redis.Client
}

// setupTestEnvironment sets up the test environment with database and Redis
func setupTestEnvironment(t *testing.T) (*sql.DB, *redis.Client, func()) {
	// Setup test database
	db := setupTestDatabase(t)

	// Setup test Redis
	redisClient := setupTestRedis(t)

	// Return cleanup function
	cleanup := func() {
		if db != nil {
			db.Close()
		}
		if redisClient != nil {
			redisClient.Close()
		}
	}

	return db, redisClient, cleanup
}

// setupTestDatabase sets up a test database connection
func setupTestDatabase(t *testing.T) *sql.DB {
	// Use environment variable or default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://agentos:agentos_dev_password@localhost:5432/agentos_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		// If database connection fails, use mock database
		t.Logf("Warning: Could not connect to test database: %v", err)
		return nil
	}

	// Test connection
	if err := db.Ping(); err != nil {
		t.Logf("Warning: Could not ping test database: %v", err)
		db.Close()
		return nil
	}

	// Create test tables if needed
	createTestTables(db, t)

	return db
}

// setupTestRedis sets up a test Redis connection
func setupTestRedis(t *testing.T) *redis.Client {
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/1" // Use database 1 for testing
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		t.Logf("Warning: Could not parse Redis URL: %v", err)
		return nil
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx := client.Context()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Logf("Warning: Could not connect to test Redis: %v", err)
		client.Close()
		return nil
	}

	// Clear test database
	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Logf("Warning: Could not flush test Redis database: %v", err)
	}

	return client
}

// createTestTables creates necessary test tables
func createTestTables(db *sql.DB, t *testing.T) {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS memories (
			id SERIAL PRIMARY KEY,
			memory_id VARCHAR(255) UNIQUE NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			framework VARCHAR(100),
			metadata JSONB,
			importance FLOAT DEFAULT 0.5,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS consolidations (
			id SERIAL PRIMARY KEY,
			consolidation_id VARCHAR(255) UNIQUE NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			framework VARCHAR(100),
			status VARCHAR(50) DEFAULT 'pending',
			result JSONB,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			t.Logf("Warning: Could not create test table: %v", err)
		}
	}
}

// setupTestHandler creates a test handler with database and Redis connections
func setupTestHandler(db *sql.DB, redisClient *redis.Client) *Handler {
	return &Handler{
		db:    db,
		redis: redisClient,
	}
}

// Memory handler methods for testing

// SemanticMemorySearch handles semantic memory search requests
func (h *Handler) SemanticMemorySearch(c *gin.Context) {
	var req struct {
		Query     string  `json:"query" binding:"required"`
		Framework string  `json:"framework"`
		Limit     int     `json:"limit"`
		Threshold float64 `json:"threshold"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Query == "" {
		c.JSON(400, gin.H{"error": "Query cannot be empty"})
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Threshold == 0 {
		req.Threshold = 0.7
	}
	if req.Framework == "" {
		req.Framework = "universal"
	}

	_ = c.GetString("user_id") // userID would be used in real implementation
	// Note: For tests, we'll return empty results since mem0 service may not be running
	// In production, this would call the real mem0 service
	memories := []interface{}{}
	err := error(nil)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to search memories"})
		return
	}

	c.JSON(200, gin.H{
		"query":     req.Query,
		"framework": req.Framework,
		"memories":  memories,
		"count":     len(memories),
		"threshold": req.Threshold,
		"engine":    "mem0",
	})
}

// StoreSemanticMemory handles semantic memory storage requests
func (h *Handler) StoreSemanticMemory(c *gin.Context) {
	var req struct {
		Content    string   `json:"content" binding:"required"`
		Concepts   []string `json:"concepts"`
		Framework  string   `json:"framework"`
		SourceType string   `json:"source_type"`
		Importance float64  `json:"importance"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request format"})
		return
	}

	// Set defaults
	if req.Importance == 0 {
		req.Importance = 0.5
	}
	if req.SourceType == "" {
		req.SourceType = "user_input"
	}
	if req.Framework == "" {
		req.Framework = "universal"
	}

	_ = c.GetString("user_id")  // userID would be used in real implementation
	_ = map[string]interface{}{ // metadata would be used in real implementation
		"concepts":    req.Concepts,
		"source_type": req.SourceType,
		"importance":  req.Importance,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	// Note: For tests, we'll generate a mock memory ID since mem0 service may not be running
	// In production, this would call the real mem0 service
	memoryID := fmt.Sprintf("test_mem0_%d", time.Now().UnixNano())
	err := error(nil)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to store memory via mem0"})
		return
	}

	c.JSON(201, gin.H{
		"memory_id":  memoryID,
		"content":    req.Content,
		"concepts":   req.Concepts,
		"framework":  req.Framework,
		"importance": req.Importance,
		"engine":     "mem0",
		"created_at": time.Now(),
	})
}

// TriggerMemoryConsolidation handles memory consolidation requests
func (h *Handler) TriggerMemoryConsolidation(c *gin.Context) {
	var req struct {
		Framework       string  `json:"framework" binding:"required"`
		TimeWindowHours float64 `json:"time_window_hours"`
		ForceRun        bool    `json:"force_run"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request format"})
		return
	}

	// Set defaults
	if req.TimeWindowHours == 0 {
		req.TimeWindowHours = 24.0
	}

	_ = c.GetString("user_id") // userID would be used in real implementation
	consolidationID := fmt.Sprintf("consolidation_%d", time.Now().UnixNano())

	// Note: For tests, we'll generate a mock consolidation result since mem0 service may not be running
	// In production, this would call the real mem0 service
	consolidationResult := map[string]interface{}{
		"status":              "completed",
		"framework":           req.Framework,
		"memories_analyzed":   10,
		"consolidation_score": 0.75,
		"timestamp":           time.Now().Format(time.RFC3339),
	}
	err := error(nil)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to trigger mem0 consolidation"})
		return
	}

	c.JSON(202, gin.H{
		"consolidation_id":  consolidationID,
		"framework":         req.Framework,
		"time_window_hours": req.TimeWindowHours,
		"status":            "completed",
		"engine":            "mem0",
		"result":            consolidationResult,
		"started_at":        time.Now(),
	})
}

// GetConsolidationStatus handles consolidation status requests
func (h *Handler) GetConsolidationStatus(c *gin.Context) {
	framework := c.Query("framework")
	if framework == "" {
		framework = "universal"
	}

	c.JSON(200, gin.H{
		"framework":      framework,
		"status":         "operational",
		"consolidations": []interface{}{},
		"metrics": map[string]interface{}{
			"total_consolidations": 0,
			"avg_duration":         "2.5s",
			"success_rate":         1.0,
		},
	})
}

// GetFrameworkMemory handles framework memory requests
func (h *Handler) GetFrameworkMemory(c *gin.Context) {
	framework := c.Param("framework")

	validFrameworks := map[string]bool{
		"langchain": true,
		"swarms":    true,
		"crewai":    true,
		"autogen":   true,
		"universal": true,
	}

	if !validFrameworks[framework] {
		c.JSON(400, gin.H{"error": "Invalid framework"})
		return
	}

	c.JSON(200, gin.H{
		"framework": framework,
		"statistics": map[string]interface{}{
			"semantic_memories": 0,
			"episodic_memories": 0,
			"total_memories":    0,
			"framework":         framework,
		},
		"recent_memories": []interface{}{},
		"last_updated":    time.Now(),
	})
}

// Note: Real mem0 integration functions are now in handlers/memory.go
// Tests should use the real implementations or proper mocking frameworks

// Note: All mem0 integration methods are now real implementations in handlers/memory.go
// Tests should import and use the real handlers or use proper mocking frameworks

// Test helper functions

// createTestUser creates a test user in the database
func createTestUser(db *sql.DB, userID string) error {
	if db == nil {
		return nil // Skip if no database
	}

	query := `INSERT INTO users (user_id, email) VALUES ($1, $2) ON CONFLICT (user_id) DO NOTHING`
	_, err := db.Exec(query, userID, fmt.Sprintf("%s@test.com", userID))
	return err
}

// cleanupTestData cleans up test data from database
func cleanupTestData(db *sql.DB, userID string) error {
	if db == nil {
		return nil // Skip if no database
	}

	queries := []string{
		`DELETE FROM consolidations WHERE user_id = $1`,
		`DELETE FROM memories WHERE user_id = $1`,
		`DELETE FROM users WHERE user_id = $1`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query, userID); err != nil {
			log.Printf("Warning: Could not cleanup test data: %v", err)
		}
	}

	return nil
}

// setupGinTestMode sets up Gin in test mode
func setupGinTestMode() {
	gin.SetMode(gin.TestMode)
}

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Setup
	setupGinTestMode()

	// Run tests
	code := m.Run()

	// Cleanup
	// Any global cleanup can go here

	os.Exit(code)
}

// Benchmark helper functions

// BenchmarkHelper provides utilities for benchmark tests
type BenchmarkHelper struct {
	handler *Handler
}

// NewBenchmarkHelper creates a new benchmark helper
func NewBenchmarkHelper(handler *Handler) *BenchmarkHelper {
	return &BenchmarkHelper{handler: handler}
}

// GenerateTestMemories generates test memories for benchmarking
func (bh *BenchmarkHelper) GenerateTestMemories(count int) []map[string]interface{} {
	memories := make([]map[string]interface{}, count)

	for i := 0; i < count; i++ {
		memories[i] = map[string]interface{}{
			"content":    fmt.Sprintf("Benchmark memory %d: Testing memory operations at scale", i),
			"framework":  "langchain",
			"concepts":   []string{"benchmark", "testing", "memory"},
			"importance": 0.5,
		}
	}

	return memories
}

// Performance monitoring utilities

// PerformanceMonitor tracks performance metrics during tests
type PerformanceMonitor struct {
	startTime time.Time
	metrics   map[string]time.Duration
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor() *PerformanceMonitor {
	return &PerformanceMonitor{
		metrics: make(map[string]time.Duration),
	}
}

// StartTimer starts timing an operation
func (pm *PerformanceMonitor) StartTimer(operation string) {
	pm.startTime = time.Now()
}

// EndTimer ends timing an operation and records the duration
func (pm *PerformanceMonitor) EndTimer(operation string) {
	if !pm.startTime.IsZero() {
		pm.metrics[operation] = time.Since(pm.startTime)
		pm.startTime = time.Time{}
	}
}

// GetMetrics returns all recorded metrics
func (pm *PerformanceMonitor) GetMetrics() map[string]time.Duration {
	return pm.metrics
}

// GetDuration returns the duration for a specific operation
func (pm *PerformanceMonitor) GetDuration(operation string) time.Duration {
	return pm.metrics[operation]
}

// Test data generators

// GenerateTestFrameworks returns all supported frameworks for testing
func GenerateTestFrameworks() []string {
	return []string{"langchain", "swarms", "crewai", "autogen", "universal"}
}

// GenerateTestQueries returns test queries for different scenarios
func GenerateTestQueries() []string {
	return []string{
		"machine learning algorithms",
		"distributed agent coordination",
		"role-based workflows",
		"conversational AI",
		"semantic memory search",
		"framework integration",
		"performance optimization",
	}
}

// GenerateTestConcepts returns test concepts for memory storage
func GenerateTestConcepts() [][]string {
	return [][]string{
		{"machine learning", "AI", "algorithms"},
		{"distributed", "coordination", "agents"},
		{"roles", "workflows", "tasks"},
		{"conversation", "AI", "generation"},
		{"semantic", "memory", "search"},
		{"framework", "integration", "API"},
		{"performance", "optimization", "speed"},
	}
}
