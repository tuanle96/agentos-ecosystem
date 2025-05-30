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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Week2PerformanceTest tests performance characteristics of Week 2 features
func TestWeek2Performance(t *testing.T) {
	setupTestDB(t)
	defer cleanupTestDB(t)
	
	insertTestTools(t)
	token := createTestUserAndGetToken(t)
	
	t.Run("API Response Time Performance", func(t *testing.T) {
		testAPIResponseTimes(t, token)
	})
	
	t.Run("Tool Execution Performance", func(t *testing.T) {
		testToolExecutionPerformance(t, token)
	})
	
	t.Run("Memory System Performance", func(t *testing.T) {
		testMemorySystemPerformance(t, token)
	})
	
	t.Run("Concurrent Load Testing", func(t *testing.T) {
		testConcurrentLoad(t, token)
	})
	
	t.Run("Capability Validation Performance", func(t *testing.T) {
		testCapabilityValidationPerformance(t, token)
	})
}

// testAPIResponseTimes measures API response times
func testAPIResponseTimes(t *testing.T, token string) {
	endpoints := []struct {
		name   string
		method string
		path   string
		body   map[string]interface{}
		target time.Duration // Target response time
	}{
		{
			name:   "Health Check",
			method: "GET",
			path:   "/health",
			target: 1 * time.Millisecond,
		},
		{
			name:   "List Tools",
			method: "GET", 
			path:   "/api/v1/tools",
			target: 5 * time.Millisecond,
		},
		{
			name:   "Tool Definitions",
			method: "GET",
			path:   "/api/v1/tools/definitions",
			target: 10 * time.Millisecond,
		},
		{
			name:   "Capability Validation",
			method: "POST",
			path:   "/api/v1/capabilities/validate",
			body: map[string]interface{}{
				"capabilities": []string{"web_search", "calculations"},
			},
			target: 15 * time.Millisecond,
		},
		{
			name:   "Tool Execution - Calculations",
			method: "POST",
			path:   "/api/v1/tools/execute",
			body: map[string]interface{}{
				"tool_name": "calculations",
				"parameters": map[string]interface{}{
					"expression": "2+2",
				},
			},
			target: 50 * time.Millisecond, // Tool execution can be slower
		},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.name, func(t *testing.T) {
			// Warm up
			for i := 0; i < 3; i++ {
				makeRequest(t, endpoint.method, endpoint.path, endpoint.body, token)
			}
			
			// Measure performance over multiple requests
			var totalDuration time.Duration
			iterations := 10
			
			for i := 0; i < iterations; i++ {
				start := time.Now()
				resp := makeRequest(t, endpoint.method, endpoint.path, endpoint.body, token)
				duration := time.Since(start)
				
				totalDuration += duration
				
				// Ensure successful response
				assert.True(t, resp.Code < 400, "Request should be successful")
			}
			
			avgDuration := totalDuration / time.Duration(iterations)
			
			t.Logf("%s - Average response time: %v (target: %v)", 
				endpoint.name, avgDuration, endpoint.target)
			
			// Performance assertion - allow 2x target for CI environments
			maxAllowed := endpoint.target * 2
			assert.True(t, avgDuration <= maxAllowed, 
				"Average response time %v should be <= %v", avgDuration, maxAllowed)
		})
	}
}

// testToolExecutionPerformance measures tool execution performance
func testToolExecutionPerformance(t *testing.T, token string) {
	tools := []struct {
		name       string
		toolName   string
		parameters map[string]interface{}
		target     time.Duration
	}{
		{
			name:     "Fast Calculation",
			toolName: "calculations",
			parameters: map[string]interface{}{
				"expression": "1+1",
			},
			target: 10 * time.Millisecond,
		},
		{
			name:     "Complex Calculation", 
			toolName: "calculations",
			parameters: map[string]interface{}{
				"expression": "sqrt(16) + pow(2,3) - 5",
			},
			target: 20 * time.Millisecond,
		},
		{
			name:     "Text Processing - Simple",
			toolName: "text_processing",
			parameters: map[string]interface{}{
				"text":      "hello",
				"operation": "uppercase",
			},
			target: 15 * time.Millisecond,
		},
		{
			name:     "Text Processing - Complex",
			toolName: "text_processing",
			parameters: map[string]interface{}{
				"text":      "This is a longer text that needs to be processed for word counting and analysis",
				"operation": "word_count",
			},
			target: 25 * time.Millisecond,
		},
	}

	for _, tool := range tools {
		t.Run(tool.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"tool_name":  tool.toolName,
				"parameters": tool.parameters,
			}

			// Warm up
			for i := 0; i < 3; i++ {
				makeRequest(t, "POST", "/api/v1/tools/execute", payload, token)
			}

			// Measure execution time
			var totalDuration time.Duration
			var totalExecutionTime float64
			iterations := 5

			for i := 0; i < iterations; i++ {
				start := time.Now()
				resp := makeRequest(t, "POST", "/api/v1/tools/execute", payload, token)
				apiDuration := time.Since(start)
				
				totalDuration += apiDuration

				// Parse response to get tool execution time
				var response map[string]interface{}
				err := json.Unmarshal(resp.Body.Bytes(), &response)
				require.NoError(t, err)
				
				if execTime, ok := response["execution_time"].(float64); ok {
					totalExecutionTime += execTime
				}
			}

			avgAPIDuration := totalDuration / time.Duration(iterations)
			avgToolExecution := time.Duration(totalExecutionTime/float64(iterations) * float64(time.Second))

			t.Logf("%s - API Duration: %v, Tool Execution: %v (target: %v)", 
				tool.name, avgAPIDuration, avgToolExecution, tool.target)

			// Tool execution should be very fast
			assert.True(t, avgToolExecution <= tool.target,
				"Tool execution time %v should be <= %v", avgToolExecution, tool.target)
		})
	}
}

// testMemorySystemPerformance measures memory system performance
func testMemorySystemPerformance(t *testing.T, token string) {
	// Create test agent
	agentID := createTestAgentForPerformance(t, token)
	
	t.Run("Memory Session Creation", func(t *testing.T) {
		var totalDuration time.Duration
		iterations := 10
		
		for i := 0; i < iterations; i++ {
			start := time.Now()
			resp := makeRequest(t, "POST", fmt.Sprintf("/api/v1/agents/%s/memory/session", agentID), nil, token)
			duration := time.Since(start)
			
			totalDuration += duration
			assert.Equal(t, http.StatusCreated, resp.Code)
		}
		
		avgDuration := totalDuration / time.Duration(iterations)
		target := 20 * time.Millisecond
		
		t.Logf("Memory session creation - Average: %v (target: %v)", avgDuration, target)
		assert.True(t, avgDuration <= target*2, "Memory session creation should be fast")
	})
	
	t.Run("Memory Update Performance", func(t *testing.T) {
		payload := map[string]interface{}{
			"variables": map[string]interface{}{
				"test_var": "test_value",
				"counter":  1,
			},
			"context": map[string]interface{}{
				"session": "performance_test",
			},
		}
		
		var totalDuration time.Duration
		iterations := 10
		
		for i := 0; i < iterations; i++ {
			// Update counter for each iteration
			payload["variables"].(map[string]interface{})["counter"] = i
			
			start := time.Now()
			resp := makeRequest(t, "PUT", fmt.Sprintf("/api/v1/agents/%s/memory/working", agentID), payload, token)
			duration := time.Since(start)
			
			totalDuration += duration
			assert.Equal(t, http.StatusOK, resp.Code)
		}
		
		avgDuration := totalDuration / time.Duration(iterations)
		target := 25 * time.Millisecond
		
		t.Logf("Memory update - Average: %v (target: %v)", avgDuration, target)
		assert.True(t, avgDuration <= target*2, "Memory update should be fast")
	})
}

// testConcurrentLoad tests system under concurrent load
func testConcurrentLoad(t *testing.T, token string) {
	t.Run("Concurrent Tool Executions", func(t *testing.T) {
		concurrency := 10
		iterations := 5
		
		var wg sync.WaitGroup
		var mu sync.Mutex
		var totalDuration time.Duration
		var successCount int
		
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				for j := 0; j < iterations; j++ {
					payload := map[string]interface{}{
						"tool_name": "calculations",
						"parameters": map[string]interface{}{
							"expression": fmt.Sprintf("%d + %d", workerID, j),
						},
					}
					
					start := time.Now()
					resp := makeRequest(t, "POST", "/api/v1/tools/execute", payload, token)
					duration := time.Since(start)
					
					mu.Lock()
					totalDuration += duration
					if resp.Code == http.StatusOK {
						successCount++
					}
					mu.Unlock()
				}
			}(i)
		}
		
		wg.Wait()
		
		totalRequests := concurrency * iterations
		avgDuration := totalDuration / time.Duration(totalRequests)
		successRate := float64(successCount) / float64(totalRequests) * 100
		
		t.Logf("Concurrent load test - %d workers, %d iterations each", concurrency, iterations)
		t.Logf("Average response time: %v", avgDuration)
		t.Logf("Success rate: %.1f%% (%d/%d)", successRate, successCount, totalRequests)
		
		// Assertions
		assert.True(t, successRate >= 95.0, "Success rate should be >= 95%%")
		assert.True(t, avgDuration <= 100*time.Millisecond, "Average response time should be reasonable under load")
	})
}

// testCapabilityValidationPerformance measures capability validation performance
func testCapabilityValidationPerformance(t *testing.T, token string) {
	testCases := []struct {
		name         string
		capabilities []string
	}{
		{"Single Capability", []string{"web_search"}},
		{"Two Capabilities", []string{"web_search", "calculations"}},
		{"Three Capabilities", []string{"web_search", "calculations", "text_processing"}},
		{"Maximum Valid", []string{"web_search", "api_calls", "text_processing"}},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"capabilities": tc.capabilities,
			}
			
			var totalDuration time.Duration
			iterations := 20
			
			for i := 0; i < iterations; i++ {
				start := time.Now()
				resp := makeRequest(t, "POST", "/api/v1/capabilities/validate", payload, token)
				duration := time.Since(start)
				
				totalDuration += duration
				assert.True(t, resp.Code < 400, "Request should be successful")
			}
			
			avgDuration := totalDuration / time.Duration(iterations)
			target := 10 * time.Millisecond
			
			t.Logf("%s - Average: %v (target: %v)", tc.name, avgDuration, target)
			assert.True(t, avgDuration <= target*2, "Capability validation should be fast")
		})
	}
}

// Helper function to make HTTP requests
func makeRequest(t *testing.T, method, path string, body map[string]interface{}, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	
	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)
	
	return w
}

// Helper function to create test agent for performance testing
func createTestAgentForPerformance(t *testing.T, token string) string {
	payload := map[string]interface{}{
		"name":        "Performance Test Agent",
		"description": "Agent for performance testing",
		"capabilities": []string{"web_search", "calculations"},
		"personality": map[string]interface{}{
			"style": "efficient",
		},
	}
	
	resp := makeRequest(t, "POST", "/api/v1/agents", payload, token)
	require.Equal(t, http.StatusCreated, resp.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	require.NoError(t, err)
	
	return response["id"].(string)
}
