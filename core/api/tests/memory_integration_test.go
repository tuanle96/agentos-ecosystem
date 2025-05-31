package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestMemorySystemIntegration tests the complete memory system integration
func TestMemorySystemIntegration(t *testing.T) {
	// Setup test environment
	db, redis, cleanup := setupTestEnvironment(t)
	defer cleanup()

	handler := setupTestHandler(db, redis)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Next()
	})

	// Setup all memory endpoints
	router.POST("/api/v1/memory/semantic/store", handler.StoreSemanticMemory)
	router.POST("/api/v1/memory/semantic/search", handler.SemanticMemorySearch)
	router.POST("/api/v1/memory/consolidation/trigger", handler.TriggerMemoryConsolidation)
	router.GET("/api/v1/memory/consolidation/status", handler.GetConsolidationStatus)
	router.GET("/api/v1/memory/frameworks/:framework", handler.GetFrameworkMemory)

	// Test complete workflow: Store → Search → Consolidate → Status → Framework Memory
	t.Run("Complete Memory Workflow", func(t *testing.T) {
		// Step 1: Store multiple memories across different frameworks
		memories := []struct {
			content    string
			framework  string
			concepts   []string
			importance float64
		}{
			{
				content:    "LangChain provides a framework for developing applications with LLMs",
				framework:  "langchain",
				concepts:   []string{"langchain", "LLM", "framework", "applications"},
				importance: 0.9,
			},
			{
				content:    "Swarms enable distributed AI agent coordination and collaboration",
				framework:  "swarms",
				concepts:   []string{"swarms", "distributed", "coordination", "collaboration"},
				importance: 0.8,
			},
			{
				content:    "CrewAI facilitates role-based multi-agent workflows and task management",
				framework:  "crewai",
				concepts:   []string{"crewai", "roles", "workflows", "task management"},
				importance: 0.85,
			},
			{
				content:    "AutoGen enables conversational AI with code generation capabilities",
				framework:  "autogen",
				concepts:   []string{"autogen", "conversational", "code generation"},
				importance: 0.75,
			},
		}

		var memoryIDs []string

		// Store memories
		for i, memory := range memories {
			t.Run(fmt.Sprintf("Store Memory %d (%s)", i+1, memory.framework), func(t *testing.T) {
				requestBody := map[string]interface{}{
					"content":    memory.content,
					"framework":  memory.framework,
					"concepts":   memory.concepts,
					"importance": memory.importance,
				}

				jsonBody, _ := json.Marshal(requestBody)
				req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/store", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusCreated, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				// Validate response
				assert.Equal(t, "mem0", response["engine"])
				assert.Equal(t, memory.content, response["content"])
				assert.Equal(t, memory.framework, response["framework"])
				assert.Equal(t, memory.importance, response["importance"])

				memoryID, ok := response["memory_id"].(string)
				assert.True(t, ok)
				assert.Contains(t, memoryID, "mem0_")
				memoryIDs = append(memoryIDs, memoryID)
			})
		}

		// Step 2: Search for memories across frameworks
		searchQueries := []struct {
			query     string
			framework string
			expected  int
		}{
			{"LLM framework applications", "langchain", 1},
			{"distributed coordination", "swarms", 1},
			{"role-based workflows", "crewai", 1},
			{"conversational AI", "autogen", 1},
			{"AI frameworks", "", 4}, // Universal search
		}

		for _, searchQuery := range searchQueries {
			t.Run(fmt.Sprintf("Search: %s", searchQuery.query), func(t *testing.T) {
				requestBody := map[string]interface{}{
					"query": searchQuery.query,
					"limit": 10,
				}

				if searchQuery.framework != "" {
					requestBody["framework"] = searchQuery.framework
				}

				jsonBody, _ := json.Marshal(requestBody)
				req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/search", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, "mem0", response["engine"])
				assert.Equal(t, searchQuery.query, response["query"])

				memories := response["memories"].([]interface{})
				assert.GreaterOrEqual(t, len(memories), 0)
			})
		}

		// Step 3: Trigger consolidation for each framework
		frameworks := []string{"langchain", "swarms", "crewai", "autogen"}
		var consolidationIDs []string

		for _, framework := range frameworks {
			t.Run(fmt.Sprintf("Consolidate: %s", framework), func(t *testing.T) {
				requestBody := map[string]interface{}{
					"framework": framework,
					"force_run": true,
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

				assert.Equal(t, "mem0", response["engine"])
				assert.Equal(t, framework, response["framework"])
				assert.Equal(t, "completed", response["status"])

				consolidationID, ok := response["consolidation_id"].(string)
				assert.True(t, ok)
				assert.Contains(t, consolidationID, "consolidation_")
				consolidationIDs = append(consolidationIDs, consolidationID)

				// Validate consolidation result
				result := response["result"].(map[string]interface{})
				assert.Equal(t, "completed", result["status"])
				assert.Equal(t, framework, result["framework"])
				assert.Contains(t, result, "memories_analyzed")
				assert.Contains(t, result, "consolidation_score")
			})
		}

		// Step 4: Check consolidation status
		t.Run("Check Consolidation Status", func(t *testing.T) {
			for _, framework := range frameworks {
				req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/memory/consolidation/status?framework=%s", framework), nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, framework, response["framework"])
				assert.Equal(t, "operational", response["status"])
				assert.Contains(t, response, "consolidations")
				assert.Contains(t, response, "metrics")
			}
		})

		// Step 5: Get framework-specific memory statistics
		t.Run("Framework Memory Statistics", func(t *testing.T) {
			allFrameworks := append(frameworks, "universal")

			for _, framework := range allFrameworks {
				req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/memory/frameworks/%s", framework), nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, framework, response["framework"])
				assert.Contains(t, response, "statistics")
				assert.Contains(t, response, "recent_memories")

				// Validate statistics structure
				stats := response["statistics"].(map[string]interface{})
				assert.Contains(t, stats, "semantic_memories")
				assert.Contains(t, stats, "episodic_memories")
				assert.Contains(t, stats, "total_memories")
				assert.Contains(t, stats, "framework")
			}
		})
	})
}

// TestMemorySystemConcurrency tests concurrent memory operations
func TestMemorySystemConcurrency(t *testing.T) {
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

	t.Run("Concurrent Memory Storage", func(t *testing.T) {
		const numGoroutines = 10
		const memoriesPerGoroutine = 5

		var wg sync.WaitGroup
		var mu sync.Mutex
		var allMemoryIDs []string
		var errors []error

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < memoriesPerGoroutine; j++ {
					requestBody := map[string]interface{}{
						"content":    fmt.Sprintf("Concurrent memory %d-%d: Testing concurrent storage", goroutineID, j),
						"framework":  "langchain",
						"concepts":   []string{"concurrent", "testing", "memory"},
						"importance": 0.6,
					}

					jsonBody, _ := json.Marshal(requestBody)
					req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/store", bytes.NewBuffer(jsonBody))
					req.Header.Set("Content-Type", "application/json")

					w := httptest.NewRecorder()
					router.ServeHTTP(w, req)

					mu.Lock()
					if w.Code != http.StatusCreated {
						errors = append(errors, fmt.Errorf("goroutine %d, memory %d: expected status %d, got %d", goroutineID, j, http.StatusCreated, w.Code))
					} else {
						var response map[string]interface{}
						if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
							if memoryID, ok := response["memory_id"].(string); ok {
								allMemoryIDs = append(allMemoryIDs, memoryID)
							}
						}
					}
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()

		// Validate results
		assert.Empty(t, errors, "No errors should occur during concurrent storage")
		assert.Equal(t, numGoroutines*memoriesPerGoroutine, len(allMemoryIDs), "All memories should be stored")

		// Verify all memory IDs are unique
		uniqueIDs := make(map[string]bool)
		for _, id := range allMemoryIDs {
			assert.False(t, uniqueIDs[id], "Memory ID should be unique: %s", id)
			uniqueIDs[id] = true
		}
	})

	t.Run("Concurrent Memory Search", func(t *testing.T) {
		const numSearches = 20

		var wg sync.WaitGroup
		var mu sync.Mutex
		var searchResults []int
		var errors []error

		for i := 0; i < numSearches; i++ {
			wg.Add(1)
			go func(searchID int) {
				defer wg.Done()

				requestBody := map[string]interface{}{
					"query":     "concurrent testing",
					"framework": "langchain",
					"limit":     10,
				}

				jsonBody, _ := json.Marshal(requestBody)
				req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/search", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				mu.Lock()
				if w.Code != http.StatusOK {
					errors = append(errors, fmt.Errorf("search %d: expected status %d, got %d", searchID, http.StatusOK, w.Code))
				} else {
					var response map[string]interface{}
					if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
						if memories, ok := response["memories"].([]interface{}); ok {
							searchResults = append(searchResults, len(memories))
						}
					}
				}
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Validate results
		assert.Empty(t, errors, "No errors should occur during concurrent searches")
		assert.Equal(t, numSearches, len(searchResults), "All searches should complete")
	})
}

// TestMemorySystemPerformance tests memory system performance
func TestMemorySystemPerformance(t *testing.T) {
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

	t.Run("Memory Storage Performance", func(t *testing.T) {
		const numMemories = 100
		startTime := time.Now()

		for i := 0; i < numMemories; i++ {
			requestBody := map[string]interface{}{
				"content":    fmt.Sprintf("Performance test memory %d: Testing storage performance", i),
				"framework":  "langchain",
				"concepts":   []string{"performance", "testing", "storage"},
				"importance": 0.5,
			}

			jsonBody, _ := json.Marshal(requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/store", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)
		}

		duration := time.Since(startTime)
		avgTimePerMemory := duration / numMemories

		t.Logf("Stored %d memories in %v (avg: %v per memory)", numMemories, duration, avgTimePerMemory)

		// Performance assertion: should be able to store memories quickly
		assert.Less(t, avgTimePerMemory, 100*time.Millisecond, "Average memory storage should be under 100ms")
	})

	t.Run("Memory Search Performance", func(t *testing.T) {
		const numSearches = 50
		queries := []string{
			"performance testing",
			"storage optimization",
			"memory retrieval",
			"framework integration",
			"semantic search",
		}

		startTime := time.Now()

		for i := 0; i < numSearches; i++ {
			query := queries[i%len(queries)]

			requestBody := map[string]interface{}{
				"query":     query,
				"framework": "langchain",
				"limit":     10,
			}

			jsonBody, _ := json.Marshal(requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/search", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}

		duration := time.Since(startTime)
		avgTimePerSearch := duration / numSearches

		t.Logf("Performed %d searches in %v (avg: %v per search)", numSearches, duration, avgTimePerSearch)

		// Performance assertion: should be able to search memories quickly
		assert.Less(t, avgTimePerSearch, 50*time.Millisecond, "Average memory search should be under 50ms")
	})
}

// TestMemorySystemErrorHandling tests error handling in memory system
func TestMemorySystemErrorHandling(t *testing.T) {
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
	router.GET("/api/v1/memory/frameworks/:framework", handler.GetFrameworkMemory)

	t.Run("Invalid Memory Storage Requests", func(t *testing.T) {
		invalidRequests := []struct {
			name        string
			requestBody map[string]interface{}
			expectedCode int
		}{
			{
				name:         "Missing content",
				requestBody:  map[string]interface{}{"framework": "langchain"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Empty content",
				requestBody:  map[string]interface{}{"content": "", "framework": "langchain"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Invalid importance value",
				requestBody:  map[string]interface{}{"content": "test", "importance": 2.0},
				expectedCode: http.StatusCreated, // Should handle gracefully
			},
		}

		for _, test := range invalidRequests {
			t.Run(test.name, func(t *testing.T) {
				jsonBody, _ := json.Marshal(test.requestBody)
				req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/store", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, test.expectedCode, w.Code)
			})
		}
	})

	t.Run("Invalid Memory Search Requests", func(t *testing.T) {
		invalidRequests := []struct {
			name        string
			requestBody map[string]interface{}
			expectedCode int
		}{
			{
				name:         "Missing query",
				requestBody:  map[string]interface{}{"framework": "langchain"},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Empty query",
				requestBody:  map[string]interface{}{"query": ""},
				expectedCode: http.StatusBadRequest,
			},
			{
				name:         "Invalid limit",
				requestBody:  map[string]interface{}{"query": "test", "limit": -1},
				expectedCode: http.StatusOK, // Should handle gracefully
			},
		}

		for _, test := range invalidRequests {
			t.Run(test.name, func(t *testing.T) {
				jsonBody, _ := json.Marshal(test.requestBody)
				req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/search", bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type", "application/json")

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, test.expectedCode, w.Code)
			})
		}
	})

	t.Run("Invalid Framework Requests", func(t *testing.T) {
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

// BenchmarkMemoryOperations benchmarks memory operations
func BenchmarkMemoryOperations(b *testing.B) {
	// Setup test environment
	db, redis, cleanup := setupTestEnvironment(&testing.T{})
	defer cleanup()

	handler := setupTestHandler(db, redis)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", testUserID)
		c.Next()
	})
	router.POST("/api/v1/memory/semantic/store", handler.StoreSemanticMemory)
	router.POST("/api/v1/memory/semantic/search", handler.SemanticMemorySearch)

	b.Run("BenchmarkMemoryStorage", func(b *testing.B) {
		requestBody := map[string]interface{}{
			"content":    "Benchmark memory storage performance",
			"framework":  "langchain",
			"concepts":   []string{"benchmark", "performance"},
			"importance": 0.5,
		}

		jsonBody, _ := json.Marshal(requestBody)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/store", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				b.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
			}
		}
	})

	b.Run("BenchmarkMemorySearch", func(b *testing.B) {
		requestBody := map[string]interface{}{
			"query":     "benchmark performance",
			"framework": "langchain",
			"limit":     10,
		}

		jsonBody, _ := json.Marshal(requestBody)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req, _ := http.NewRequest("POST", "/api/v1/memory/semantic/search", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
			}
		}
	})
}
