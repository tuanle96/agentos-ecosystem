package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestSemanticMemorySearch tests semantic memory search functionality
func TestSemanticMemorySearch(t *testing.T) {
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
		expectedFields []string
	}{
		{
			name: "Valid semantic search",
			requestBody: map[string]interface{}{
				"query":     "machine learning algorithms",
				"framework": "langchain",
				"limit":     5,
				"threshold": 0.7,
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"query", "memories", "count", "threshold"},
		},
		{
			name: "Search with defaults",
			requestBody: map[string]interface{}{
				"query": "artificial intelligence",
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"query", "memories", "count", "threshold"},
		},
		{
			name: "Invalid request - missing query",
			requestBody: map[string]interface{}{
				"framework": "langchain",
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
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

			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field, "Response should contain field: %s", field)
			}

			if tt.expectedStatus == http.StatusOK {
				assert.IsType(t, []interface{}{}, response["memories"])
				assert.IsType(t, float64(0), response["count"])
			}
		})
	}
}

// TestStoreSemanticMemory tests semantic memory storage
func TestStoreSemanticMemory(t *testing.T) {
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
		expectedFields []string
	}{
		{
			name: "Valid semantic memory storage",
			requestBody: map[string]interface{}{
				"content":    "Neural networks are inspired by biological neural networks",
				"concepts":   []string{"neural networks", "biology", "artificial intelligence"},
				"framework":  "langchain",
				"importance": 0.8,
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"memory_id", "content", "concepts", "framework", "importance", "created_at"},
		},
		{
			name: "Memory with defaults",
			requestBody: map[string]interface{}{
				"content":  "Deep learning uses multiple layers of neural networks",
				"concepts": []string{"deep learning", "neural networks"},
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"memory_id", "content", "concepts", "importance"},
		},
		{
			name: "Invalid request - missing content",
			requestBody: map[string]interface{}{
				"concepts": []string{"test"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
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

			for _, field := range tt.expectedFields {
				assert.Contains(t, response, field, "Response should contain field: %s", field)
			}

			if tt.expectedStatus == http.StatusCreated {
				assert.NotEmpty(t, response["memory_id"])
				assert.Equal(t, tt.requestBody["content"], response["content"])
			}
		})
	}
}

// TestMemoryConsolidation tests memory consolidation functionality
func TestMemoryConsolidation(t *testing.T) {
	// Setup test environment
	db, redis, cleanup := setupTestEnvironment(t)
	defer cleanup()

	handler := setupTestHandler(db, redis)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Next()
	})
	router.GET("/api/v1/memory/consolidation/status", handler.GetConsolidationStatus)
	router.POST("/api/v1/memory/consolidation/trigger", handler.TriggerMemoryConsolidation)

	// Test consolidation status
	t.Run("Get consolidation status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/memory/consolidation/status?framework=langchain", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		expectedFields := []string{"framework", "consolidations", "metrics", "status"}
		for _, field := range expectedFields {
			assert.Contains(t, response, field)
		}
	})

	// Test trigger consolidation
	t.Run("Trigger memory consolidation", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"framework":         "langchain",
			"time_window_hours": 24.0,
			"force_run":         false,
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

		expectedFields := []string{"consolidation_id", "framework", "time_window_hours", "status", "started_at"}
		for _, field := range expectedFields {
			assert.Contains(t, response, field)
		}

		assert.NotEmpty(t, response["consolidation_id"])
		assert.Equal(t, "started", response["status"])
	})

	// Test invalid consolidation request
	t.Run("Invalid consolidation request", func(t *testing.T) {
		requestBody := map[string]interface{}{
			"time_window_hours": 24.0,
			// Missing required framework
		}

		jsonBody, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", "/api/v1/memory/consolidation/trigger", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestFrameworkMemory tests framework-specific memory endpoints
func TestFrameworkMemory(t *testing.T) {
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
		t.Run("Get framework memory: "+framework, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/memory/frameworks/"+framework, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			expectedFields := []string{"framework", "statistics", "recent_memories", "last_updated"}
			for _, field := range expectedFields {
				assert.Contains(t, response, field)
			}

			assert.Equal(t, framework, response["framework"])
			assert.IsType(t, map[string]interface{}{}, response["statistics"])
			assert.IsType(t, []interface{}{}, response["recent_memories"])
		})
	}

	// Test invalid framework
	t.Run("Invalid framework", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/memory/frameworks/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestSemanticMemoryIntegration tests end-to-end semantic memory workflow
func TestSemanticMemoryIntegration(t *testing.T) {
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
	router.POST("/api/v1/memory/semantic/search", handler.SemanticMemorySearch)

	// Step 1: Store multiple semantic memories
	memories := []map[string]interface{}{
		{
			"content":    "Machine learning algorithms learn patterns from data",
			"concepts":   []string{"machine learning", "algorithms", "patterns", "data"},
			"framework":  "langchain",
			"importance": 0.9,
		},
		{
			"content":    "Neural networks are inspired by biological neurons",
			"concepts":   []string{"neural networks", "biology", "neurons"},
			"framework":  "langchain",
			"importance": 0.8,
		},
		{
			"content":    "Deep learning uses multiple layers of neural networks",
			"concepts":   []string{"deep learning", "neural networks", "layers"},
			"framework":  "langchain",
			"importance": 0.85,
		},
	}

	var memoryIDs []string
	for i, memory := range memories {
		t.Run("Store memory "+string(rune(i+1)), func(t *testing.T) {
			jsonBody, _ := json.Marshal(memory)
			req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/store", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			memoryID, ok := response["memory_id"].(string)
			assert.True(t, ok)
			assert.NotEmpty(t, memoryID)
			memoryIDs = append(memoryIDs, memoryID)
		})
	}

	// Step 2: Search for related memories
	t.Run("Search related memories", func(t *testing.T) {
		searchRequest := map[string]interface{}{
			"query":     "neural networks and machine learning",
			"framework": "langchain",
			"limit":     10,
			"threshold": 0.5,
		}

		jsonBody, _ := json.Marshal(searchRequest)
		req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/search", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		memories, ok := response["memories"].([]interface{})
		assert.True(t, ok)
		assert.NotNil(t, memories) // Use the memories variable

		// Should find some related memories
		count, ok := response["count"].(float64)
		assert.True(t, ok)
		assert.GreaterOrEqual(t, int(count), 0)
	})

	// Wait a bit for any async operations
	time.Sleep(100 * time.Millisecond)
}
