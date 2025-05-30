package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAPIResponseTime tests that API responses are under 15ms
func (suite *TestSuite) TestAPIResponseTime() {
	endpoints := []struct {
		method string
		url    string
		token  bool
	}{
		{"GET", "/health", false},
		{"GET", "/api/v1/tools", false},
		{"GET", "/api/v1/agents", true},
		{"GET", "/api/v1/profile", true},
	}

	for _, endpoint := range endpoints {
		suite.T().Run(fmt.Sprintf("%s %s", endpoint.method, endpoint.url), func(t *testing.T) {
			token := ""
			if endpoint.token {
				token = suite.testUser.Token
			}

			// Measure response time
			start := time.Now()
			w := suite.makeRequest(endpoint.method, endpoint.url, nil, token)
			duration := time.Since(start)

			// Assert response is successful
			if endpoint.token {
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusCreated)
			}

			// Assert response time is under 15ms
			assert.Less(t, duration.Milliseconds(), int64(15),
				"Response time should be under 15ms, got %dms", duration.Milliseconds())

			t.Logf("Endpoint %s %s responded in %dms", endpoint.method, endpoint.url, duration.Milliseconds())
		})
	}
}

// TestConcurrentRequests tests handling multiple concurrent requests
func (suite *TestSuite) TestConcurrentRequests() {
	const numRequests = 100
	const maxResponseTime = 50 // Allow higher threshold for concurrent testing

	var wg sync.WaitGroup
	results := make(chan time.Duration, numRequests)
	errors := make(chan error, numRequests)

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			start := time.Now()
			w := suite.makeRequest("GET", "/health", nil, "")
			duration := time.Since(start)

			if w.Code != http.StatusOK {
				errors <- fmt.Errorf("request %d failed with status %d", id, w.Code)
				return
			}

			results <- duration
		}(i)
	}

	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	errorCount := 0
	for err := range errors {
		suite.T().Logf("Error: %v", err)
		errorCount++
	}

	assert.Equal(suite.T(), 0, errorCount, "No requests should fail")

	// Analyze response times
	var totalDuration time.Duration
	var maxDuration time.Duration
	var minDuration time.Duration = time.Hour // Initialize to high value
	count := 0

	for duration := range results {
		totalDuration += duration
		count++

		if duration > maxDuration {
			maxDuration = duration
		}
		if duration < minDuration {
			minDuration = duration
		}
	}

	avgDuration := totalDuration / time.Duration(count)

	suite.T().Logf("Concurrent requests stats:")
	suite.T().Logf("  Total requests: %d", count)
	suite.T().Logf("  Average response time: %dms", avgDuration.Milliseconds())
	suite.T().Logf("  Min response time: %dms", minDuration.Milliseconds())
	suite.T().Logf("  Max response time: %dms", maxDuration.Milliseconds())

	// Assert performance criteria
	assert.Equal(suite.T(), numRequests, count, "All requests should complete")
	assert.Less(suite.T(), avgDuration.Milliseconds(), int64(maxResponseTime),
		"Average response time should be under %dms", maxResponseTime)
}

// TestDatabasePerformance tests database query performance
func (suite *TestSuite) TestDatabasePerformance() {
	// Create multiple agents for testing
	const numAgents = 50

	for i := 0; i < numAgents; i++ {
		agentData := map[string]interface{}{
			"name":         fmt.Sprintf("Performance Test Agent %d", i),
			"capabilities": []string{"web_search"},
		}

		w := suite.makeRequest("POST", "/api/v1/agents", agentData, suite.testUser.Token)
		assert.Equal(suite.T(), http.StatusCreated, w.Code)
	}

	// Test listing performance
	start := time.Now()
	w := suite.makeRequest("GET", "/api/v1/agents", nil, suite.testUser.Token)
	duration := time.Since(start)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	agents := response["agents"].([]interface{})
	assert.Equal(suite.T(), numAgents, len(agents))

	// Assert database query performance
	assert.Less(suite.T(), duration.Milliseconds(), int64(10),
		"Database query should be under 10ms, got %dms", duration.Milliseconds())

	suite.T().Logf("Listed %d agents in %dms", numAgents, duration.Milliseconds())
}

// TestMemoryUsage tests basic memory usage patterns
func (suite *TestSuite) TestMemoryUsage() {
	// This is a basic test - in production you'd use more sophisticated memory profiling

	// Create and delete agents to test memory cleanup
	const cycles = 10
	const agentsPerCycle = 20

	for cycle := 0; cycle < cycles; cycle++ {
		agentIDs := make([]string, 0, agentsPerCycle)

		// Create agents
		for i := 0; i < agentsPerCycle; i++ {
			agentData := map[string]interface{}{
				"name":         fmt.Sprintf("Memory Test Agent %d-%d", cycle, i),
				"capabilities": []string{"web_search"},
			}

			w := suite.makeRequest("POST", "/api/v1/agents", agentData, suite.testUser.Token)
			assert.Equal(suite.T(), http.StatusCreated, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			agentIDs = append(agentIDs, response["id"].(string))
		}

		// Delete agents
		for _, agentID := range agentIDs {
			w := suite.makeRequest("DELETE", fmt.Sprintf("/api/v1/agents/%s", agentID), nil, suite.testUser.Token)
			assert.Equal(suite.T(), http.StatusOK, w.Code)
		}

		// Verify cleanup
		w := suite.makeRequest("GET", "/api/v1/agents", nil, suite.testUser.Token)
		assert.Equal(suite.T(), http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// Handle case where agents might be nil after deletion
		if response["agents"] != nil {
			agents := response["agents"].([]interface{})
			assert.Equal(suite.T(), 0, len(agents), "All agents should be deleted")
		} else {
			// If agents is nil, that's also acceptable for empty list
			assert.Nil(suite.T(), response["agents"], "Agents should be nil or empty after deletion")
		}
	}

	suite.T().Logf("Completed %d memory test cycles", cycles)
}

// TestErrorHandling tests error handling performance
func (suite *TestSuite) TestErrorHandling() {
	errorTests := []struct {
		name           string
		method         string
		url            string
		data           interface{}
		token          string
		expectedStatus int
	}{
		{
			name:           "Invalid JSON",
			method:         "POST",
			url:            "/api/v1/agents",
			data:           "invalid json",
			token:          suite.testUser.Token,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing Auth",
			method:         "GET",
			url:            "/api/v1/agents",
			data:           nil,
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Agent ID",
			method:         "GET",
			url:            "/api/v1/agents/invalid-uuid",
			data:           nil,
			token:          suite.testUser.Token,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Invalid Capabilities",
			method: "POST",
			url:    "/api/v1/agents",
			data: map[string]interface{}{
				"name":         "Invalid Agent",
				"capabilities": []string{"invalid_capability"},
			},
			token:          suite.testUser.Token,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range errorTests {
		suite.T().Run(test.name, func(t *testing.T) {
			start := time.Now()
			w := suite.makeRequest(test.method, test.url, test.data, test.token)
			duration := time.Since(start)

			assert.Equal(t, test.expectedStatus, w.Code)

			// Error responses should also be fast
			assert.Less(t, duration.Milliseconds(), int64(15),
				"Error response should be under 15ms, got %dms", duration.Milliseconds())

			// Verify error response format
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Contains(t, response, "error")

			t.Logf("Error test '%s' responded in %dms", test.name, duration.Milliseconds())
		})
	}
}

// BenchmarkHealthCheck benchmarks the health check endpoint
func BenchmarkHealthCheck(b *testing.B) {
	// This would need to be run separately with `go test -bench=.`
	// For now, we'll skip it in the main test suite
	b.Skip("Benchmark test - run separately with go test -bench=.")
}
