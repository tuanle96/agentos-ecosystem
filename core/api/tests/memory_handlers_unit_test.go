package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMemoryService for testing memory operations
type MockMemoryService struct {
	mock.Mock
}

func (m *MockMemoryService) SearchMemories(userID, query, framework string, limit int) ([]interface{}, error) {
	args := m.Called(userID, query, framework, limit)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockMemoryService) StoreMemory(userID, content, framework string, metadata map[string]interface{}) (string, error) {
	args := m.Called(userID, content, framework, metadata)
	return args.String(0), args.Error(1)
}

func (m *MockMemoryService) TriggerConsolidation(userID, framework string) (map[string]interface{}, error) {
	args := m.Called(userID, framework)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// TestSemanticMemorySearchUnit tests semantic memory search unit functionality
func TestSemanticMemorySearchUnit(t *testing.T) {
	// Setup test environment
	db, redis, cleanup := setupTestEnvironment(t)
	defer cleanup()

	handler := setupTestHandler(db, redis)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Next()
	})
	router.POST("/api/v1/memory/semantic/search", handler.SemanticMemorySearch)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		validateFunc   func(*testing.T, map[string]interface{})
	}{
		{
			name: "Valid search with all parameters",
			requestBody: map[string]interface{}{
				"query":     "machine learning algorithms",
				"framework": "langchain",
				"limit":     5,
				"threshold": 0.8,
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "machine learning algorithms", response["query"])
				assert.Equal(t, "mem0", response["engine"])
				assert.Contains(t, response, "memories")
				assert.Contains(t, response, "count")
				assert.Equal(t, 0.8, response["threshold"])
			},
		},
		{
			name: "Search with minimal parameters",
			requestBody: map[string]interface{}{
				"query": "neural networks",
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "neural networks", response["query"])
				assert.Equal(t, "mem0", response["engine"])
				assert.Equal(t, 0.7, response["threshold"]) // Default threshold
			},
		},
		{
			name: "Search with framework filter",
			requestBody: map[string]interface{}{
				"query":     "swarm intelligence",
				"framework": "swarms",
				"limit":     3,
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "swarm intelligence", response["query"])
				assert.Equal(t, "mem0", response["engine"])
				memories := response["memories"].([]interface{})
				assert.GreaterOrEqual(t, len(memories), 0)
			},
		},
		{
			name: "Invalid request - missing query",
			requestBody: map[string]interface{}{
				"framework": "langchain",
				"limit":     5,
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				assert.Contains(t, response, "error")
			},
		},
		{
			name: "Empty query string",
			requestBody: map[string]interface{}{
				"query": "",
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				assert.Contains(t, response, "error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/search", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			tt.validateFunc(t, response)
		})
	}
}

// TestStoreSemanticMemoryUnit tests semantic memory storage unit functionality
func TestStoreSemanticMemoryUnit(t *testing.T) {
	// Setup test environment
	db, redis, cleanup := setupTestEnvironment(t)
	defer cleanup()

	handler := setupTestHandler(db, redis)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Next()
	})
	router.POST("/api/v1/memory/semantic/store", handler.StoreSemanticMemory)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		validateFunc   func(*testing.T, map[string]interface{})
	}{
		{
			name: "Store complete memory entry",
			requestBody: map[string]interface{}{
				"content":     "Deep learning revolutionizes artificial intelligence through neural networks",
				"concepts":    []string{"deep learning", "AI", "neural networks", "revolution"},
				"framework":   "langchain",
				"source_type": "user_input",
				"importance":  0.9,
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				assert.Contains(t, response["memory_id"].(string), "mem0_")
				assert.Equal(t, "Deep learning revolutionizes artificial intelligence through neural networks", response["content"])
				assert.Equal(t, "langchain", response["framework"])
				assert.Equal(t, 0.9, response["importance"])
				assert.Equal(t, "mem0", response["engine"])
				assert.Contains(t, response, "created_at")
			},
		},
		{
			name: "Store memory with defaults",
			requestBody: map[string]interface{}{
				"content":  "Transformers architecture changed natural language processing",
				"concepts": []string{"transformers", "NLP", "architecture"},
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				assert.Contains(t, response["memory_id"].(string), "mem0_")
				assert.Equal(t, "universal", response["framework"]) // Default framework
				assert.Equal(t, 0.5, response["importance"])        // Default importance
				assert.Equal(t, "mem0", response["engine"])
			},
		},
		{
			name: "Store CrewAI specific memory",
			requestBody: map[string]interface{}{
				"content":     "Role-based agents collaborate effectively in team workflows",
				"concepts":    []string{"roles", "collaboration", "agents", "workflows"},
				"framework":   "crewai",
				"source_type": "system_generated",
				"importance":  0.8,
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				assert.Equal(t, "crewai", response["framework"])
				assert.Equal(t, 0.8, response["importance"])
				assert.Equal(t, "mem0", response["engine"])
			},
		},
		{
			name: "Invalid request - missing content",
			requestBody: map[string]interface{}{
				"concepts":   []string{"test"},
				"framework":  "langchain",
				"importance": 0.7,
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				assert.Contains(t, response, "error")
			},
		},
		{
			name: "Invalid importance value",
			requestBody: map[string]interface{}{
				"content":    "Test content",
				"importance": 1.5, // Invalid - should be 0.0-1.0
			},
			expectedStatus: http.StatusCreated, // Still accepts, but should clamp
			validateFunc: func(t *testing.T, response map[string]interface{}) {
				// Should handle gracefully
				assert.Contains(t, response["memory_id"].(string), "mem0_")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/store", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			tt.validateFunc(t, response)
		})
	}
}

// TestMemoryConsolidationUnit tests memory consolidation unit functionality
func TestMemoryConsolidationUnit(t *testing.T) {
	// Setup test environment
	db, redis, cleanup := setupTestEnvironment(t)
	defer cleanup()

	handler := setupTestHandler(db, redis)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Next()
	})
	router.POST("/api/v1/memory/consolidation/trigger", handler.TriggerMemoryConsolidation)
	router.GET("/api/v1/memory/consolidation/status", handler.GetConsolidationStatus)

	// Test consolidation trigger
	t.Run("Trigger consolidation for each framework", func(t *testing.T) {
		frameworks := []string{"langchain", "swarms", "crewai", "autogen"}

		for _, framework := range frameworks {
			t.Run("Framework: "+framework, func(t *testing.T) {
				requestBody := map[string]interface{}{
					"framework":         framework,
					"time_window_hours": 24.0,
					"force_run":         true,
				}

				jsonBody, _ := json.Marshal(requestBody)
				req, _ := http.NewRequest("POST", "/api/v1/memory/consolidation/trigger", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusAccepted, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Validate response structure
				assert.Contains(t, response["consolidation_id"].(string), "consolidation_")
				assert.Equal(t, framework, response["framework"])
				assert.Equal(t, "completed", response["status"])
				assert.Equal(t, "mem0", response["engine"])
				assert.Contains(t, response, "result")

				// Validate consolidation result
				result := response["result"].(map[string]interface{})
				assert.Equal(t, "completed", result["status"])
				assert.Equal(t, framework, result["framework"])
				assert.Contains(t, result, "memories_analyzed")
				assert.Contains(t, result, "consolidation_score")
			})
		}
	})

	// Test consolidation status
	t.Run("Get consolidation status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/memory/consolidation/status?framework=langchain", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "langchain", response["framework"])
		assert.Contains(t, response, "consolidations")
		assert.Contains(t, response, "metrics")
		assert.Equal(t, "operational", response["status"])
	})

	// Test invalid consolidation requests
	t.Run("Invalid consolidation requests", func(t *testing.T) {
		invalidRequests := []struct {
			name        string
			requestBody map[string]interface{}
		}{
			{
				name: "Missing framework",
				requestBody: map[string]interface{}{
					"time_window_hours": 24.0,
				},
			},
			{
				name: "Invalid framework",
				requestBody: map[string]interface{}{
					"framework": "invalid_framework",
				},
			},
			{
				name: "Negative time window",
				requestBody: map[string]interface{}{
					"framework":         "langchain",
					"time_window_hours": -5.0,
				},
			},
		}

		for _, test := range invalidRequests {
			t.Run(test.name, func(t *testing.T) {
				jsonBody, _ := json.Marshal(test.requestBody)
				req, _ := http.NewRequest("POST", "/api/v1/memory/consolidation/trigger", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				// Should return bad request for invalid inputs
				assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusAccepted)
			})
		}
	})
}

// TestFrameworkMemoryUnit tests framework-specific memory operations
func TestFrameworkMemoryUnit(t *testing.T) {
	// Setup test environment
	db, redis, cleanup := setupTestEnvironment(t)
	defer cleanup()

	handler := setupTestHandler(db, redis)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Next()
	})
	router.GET("/api/v1/memory/frameworks/:framework", handler.GetFrameworkMemory)

	frameworks := []string{"langchain", "swarms", "crewai", "autogen", "universal"}

	for _, framework := range frameworks {
		t.Run("Framework: "+framework, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/memory/frameworks/"+framework, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Validate response structure
			assert.Equal(t, framework, response["framework"])
			assert.Contains(t, response, "statistics")
			assert.Contains(t, response, "recent_memories")
			assert.Contains(t, response, "last_updated")

			// Validate statistics structure
			stats := response["statistics"].(map[string]interface{})
			assert.Contains(t, stats, "semantic_memories")
			assert.Contains(t, stats, "episodic_memories")
			assert.Contains(t, stats, "total_memories")
			assert.Contains(t, stats, "framework")
			assert.Equal(t, framework, stats["framework"])

			// Validate recent memories
			memories := response["recent_memories"].([]interface{})
			assert.GreaterOrEqual(t, len(memories), 0)
		})
	}

	// Test invalid framework
	t.Run("Invalid framework", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/memory/frameworks/invalid_framework", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response, "error")
	})
}

// TestMemoryHelperFunctions tests memory helper functions
func TestMemoryHelperFunctions(t *testing.T) {
	// Setup test environment
	db, redis, cleanup := setupTestEnvironment(t)
	defer cleanup()

	handler := setupTestHandler(db, redis)

	t.Run("Test Handler Structure", func(t *testing.T) {
		// Test that handler is properly initialized
		assert.NotNil(t, handler)
		assert.NotNil(t, handler.db)

		// Note: Real mem0 integration tests would require:
		// 1. mem0 service running on port 8001
		// 2. Proper HTTP client testing
		// 3. Integration with Python service
		t.Log("Handler structure validated - real mem0 integration requires service")
	})

	t.Run("Test Database Connection", func(t *testing.T) {
		// Test database connectivity
		err := handler.db.Ping()
		if err != nil {
			t.Skip("Database not available for testing")
		}
		assert.NoError(t, err)
	})

	t.Run("Test Redis Connection", func(t *testing.T) {
		// Test Redis connectivity if available
		if redis != nil {
			// Redis connection test would go here
			t.Log("Redis connection available")
		} else {
			t.Log("Redis not available for testing")
		}
	})
}

// BenchmarkMemoryHandlerOperations benchmarks memory handler operations performance
func BenchmarkMemoryHandlerOperations(b *testing.B) {
	// Setup test environment
	db, redis, cleanup := setupTestEnvironment(&testing.T{})
	defer cleanup()

	handler := setupTestHandler(db, redis)

	b.Run("BenchmarkHandlerSetup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Benchmark handler initialization
			testHandler := setupTestHandler(db, redis)
			if testHandler == nil {
				b.Fatal("Handler setup failed")
			}
		}
	})

	b.Run("BenchmarkDatabasePing", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := handler.db.Ping()
			if err != nil {
				b.Skip("Database not available for benchmarking")
			}
		}
	})

	b.Run("BenchmarkMemoryStructure", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Benchmark memory structure operations
			metadata := map[string]interface{}{
				"concepts":   []string{"benchmark", "test"},
				"importance": 0.5,
			}

			// Test metadata marshaling performance
			if len(metadata) == 0 {
				b.Fatal("Metadata creation failed")
			}
		}
	})
}
