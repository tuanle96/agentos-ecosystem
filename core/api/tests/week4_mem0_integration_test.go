package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestMem0SemanticMemorySearch tests mem0-powered semantic search
func TestMem0SemanticMemorySearch(t *testing.T) {
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
			name: "Valid mem0 semantic search",
			requestBody: map[string]interface{}{
				"query":     "machine learning",
				"framework": "langchain",
				"limit":     5,
				"threshold": 0.7,
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"query", "memories", "count", "threshold", "engine"},
		},
		{
			name: "Search with universal framework",
			requestBody: map[string]interface{}{
				"query": "neural networks",
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"query", "memories", "count", "engine"},
		},
		{
			name: "Search for patterns",
			requestBody: map[string]interface{}{
				"query":     "patterns",
				"framework": "swarms",
				"limit":     3,
			},
			expectedStatus: http.StatusOK,
			expectedFields: []string{"query", "memories", "count", "engine"},
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
				assert.Equal(t, "mem0", response["engine"])
				assert.IsType(t, []interface{}{}, response["memories"])
				assert.IsType(t, float64(0), response["count"])
			}
		})
	}
}

// TestMem0StoreSemanticMemory tests mem0-powered memory storage
func TestMem0StoreSemanticMemory(t *testing.T) {
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
			name: "Store memory with mem0",
			requestBody: map[string]interface{}{
				"content":    "Deep learning revolutionizes artificial intelligence",
				"concepts":   []string{"deep learning", "AI", "revolution"},
				"framework":  "langchain",
				"importance": 0.9,
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"memory_id", "content", "concepts", "framework", "importance", "engine", "created_at"},
		},
		{
			name: "Store memory with defaults",
			requestBody: map[string]interface{}{
				"content": "Transformers changed natural language processing",
				"concepts": []string{"transformers", "NLP"},
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"memory_id", "content", "concepts", "framework", "engine"},
		},
		{
			name: "Store CrewAI memory",
			requestBody: map[string]interface{}{
				"content":    "Role-based agents collaborate effectively",
				"concepts":   []string{"roles", "collaboration", "agents"},
				"framework":  "crewai",
				"importance": 0.8,
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"memory_id", "content", "framework", "engine"},
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
				assert.Equal(t, "mem0", response["engine"])
				assert.NotEmpty(t, response["memory_id"])
				assert.Equal(t, tt.requestBody["content"], response["content"])
				
				// Check memory ID format (should start with "mem0_")
				memoryID, ok := response["memory_id"].(string)
				assert.True(t, ok)
				assert.Contains(t, memoryID, "mem0_")
			}
		})
	}
}

// TestMem0MemoryConsolidation tests mem0-powered consolidation
func TestMem0MemoryConsolidation(t *testing.T) {
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

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "Trigger mem0 consolidation for LangChain",
			requestBody: map[string]interface{}{
				"framework":         "langchain",
				"time_window_hours": 24.0,
				"force_run":         true,
			},
			expectedStatus: http.StatusAccepted,
			expectedFields: []string{"consolidation_id", "framework", "status", "engine", "result"},
		},
		{
			name: "Trigger mem0 consolidation for Swarms",
			requestBody: map[string]interface{}{
				"framework": "swarms",
				"force_run": true,
			},
			expectedStatus: http.StatusAccepted,
			expectedFields: []string{"consolidation_id", "framework", "status", "engine", "result"},
		},
		{
			name: "Trigger mem0 consolidation for CrewAI",
			requestBody: map[string]interface{}{
				"framework": "crewai",
				"force_run": true,
			},
			expectedStatus: http.StatusAccepted,
			expectedFields: []string{"consolidation_id", "framework", "status", "engine", "result"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/memory/consolidation/trigger", bytes.NewBuffer(jsonBody))
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

			if tt.expectedStatus == http.StatusAccepted {
				assert.Equal(t, "mem0", response["engine"])
				assert.Equal(t, "completed", response["status"])
				assert.NotEmpty(t, response["consolidation_id"])
				
				// Check consolidation result
				result, ok := response["result"].(map[string]interface{})
				assert.True(t, ok)
				assert.Equal(t, "completed", result["status"])
				assert.Equal(t, tt.requestBody["framework"], result["framework"])
				assert.Contains(t, result, "memories_analyzed")
				assert.Contains(t, result, "consolidation_score")
			}
		})
	}
}

// TestMem0FrameworkMemory tests framework-specific memory with mem0
func TestMem0FrameworkMemory(t *testing.T) {
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
		t.Run("Get mem0 framework memory: "+framework, func(t *testing.T) {
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
			
			// Check statistics structure
			stats, ok := response["statistics"].(map[string]interface{})
			assert.True(t, ok)
			assert.Contains(t, stats, "semantic_memories")
			assert.Contains(t, stats, "episodic_memories")
			assert.Contains(t, stats, "total_memories")
			assert.Contains(t, stats, "framework")
			
			// Check recent memories
			memories, ok := response["recent_memories"].([]interface{})
			assert.True(t, ok)
			assert.GreaterOrEqual(t, len(memories), 0)
		})
	}
}

// TestMem0IntegrationWorkflow tests end-to-end mem0 workflow
func TestMem0IntegrationWorkflow(t *testing.T) {
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
	router.POST("/api/v1/memory/consolidation/trigger", handler.TriggerMemoryConsolidation)

	// Step 1: Store multiple memories using mem0
	memories := []map[string]interface{}{
		{
			"content":    "mem0 provides intelligent memory management for AI agents",
			"concepts":   []string{"mem0", "memory management", "AI agents"},
			"framework":  "langchain",
			"importance": 0.9,
		},
		{
			"content":    "Semantic search enables finding relevant memories by meaning",
			"concepts":   []string{"semantic search", "relevance", "meaning"},
			"framework":  "langchain",
			"importance": 0.8,
		},
	}

	var memoryIDs []string
	for i, memory := range memories {
		t.Run("Store mem0 memory "+string(rune(i+1)), func(t *testing.T) {
			jsonBody, _ := json.Marshal(memory)
			req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/store", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, "mem0", response["engine"])
			memoryID, ok := response["memory_id"].(string)
			assert.True(t, ok)
			assert.Contains(t, memoryID, "mem0_")
			memoryIDs = append(memoryIDs, memoryID)
		})
	}

	// Step 2: Search for memories using mem0
	t.Run("Search mem0 memories", func(t *testing.T) {
		searchRequest := map[string]interface{}{
			"query":     "memory management",
			"framework": "langchain",
			"limit":     10,
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

		assert.Equal(t, "mem0", response["engine"])
		memories, ok := response["memories"].([]interface{})
		assert.True(t, ok)
		assert.GreaterOrEqual(t, len(memories), 0)
	})

	// Step 3: Trigger mem0 consolidation
	t.Run("Trigger mem0 consolidation", func(t *testing.T) {
		consolidationRequest := map[string]interface{}{
			"framework": "langchain",
			"force_run": true,
		}

		jsonBody, _ := json.Marshal(consolidationRequest)
		req, _ := http.NewRequest("POST", "/api/v1/memory/consolidation/trigger", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Equal(t, "mem0", response["engine"])
		assert.Equal(t, "completed", response["status"])
		
		result, ok := response["result"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "completed", result["status"])
		assert.Equal(t, "langchain", result["framework"])
	})
}
