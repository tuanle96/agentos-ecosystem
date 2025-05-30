package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Week2TestSuite comprehensive test suite for Week 2 features
type Week2TestSuite struct {
	suite.Suite
	token   string
	userID  string
	agentID string
}

// SetupSuite runs before all tests
func (suite *Week2TestSuite) SetupSuite() {
	setupTestDB(suite.T())
	insertTestTools(suite.T())
	suite.token = createTestUserAndGetToken(suite.T())
	suite.agentID = suite.createTestAgent()
}

// TearDownSuite runs after all tests
func (suite *Week2TestSuite) TearDownSuite() {
	cleanupTestDB(suite.T())
}

// TestAgentFactoryCapabilityValidation tests capability validation system
func (suite *Week2TestSuite) TestAgentFactoryCapabilityValidation() {
	tests := []struct {
		name           string
		capabilities   []string
		expectedValid  bool
		expectedCost   float64
		expectedFramework string
		expectError    bool
	}{
		{
			name:              "Valid Basic Combination",
			capabilities:      []string{"web_search", "calculations"},
			expectedValid:     true,
			expectedCost:      3.0, // web_search(2) + calculations(1)
			expectedFramework: "langchain",
			expectError:       false,
		},
		{
			name:              "Single Capability",
			capabilities:      []string{"text_processing"},
			expectedValid:     true,
			expectedCost:      1.0,
			expectedFramework: "langchain",
			expectError:       false,
		},
		{
			name:              "Maximum Resource Usage",
			capabilities:      []string{"web_search", "api_calls", "text_processing"},
			expectedValid:     true,
			expectedCost:      6.0, // web_search(2) + api_calls(3) + text_processing(1)
			expectedFramework: "langchain",
			expectError:       false,
		},
		{
			name:         "Conflicting Capabilities",
			capabilities: []string{"file_operations", "api_calls"},
			expectedValid: false,
			expectError:  true,
		},
		{
			name:         "Over Resource Limit",
			capabilities: []string{"web_search", "api_calls", "file_operations", "text_processing"},
			expectedValid: false,
			expectError:  true,
		},
		{
			name:         "Empty Capabilities",
			capabilities: []string{},
			expectedValid: false,
			expectError:  true,
		},
		{
			name:         "Unknown Capability",
			capabilities: []string{"unknown_capability"},
			expectedValid: false,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			payload := map[string]interface{}{
				"capabilities": tt.capabilities,
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/api/v1/capabilities/validate", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.token)

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			if tt.expectError {
				assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
				
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(suite.T(), err)
				
				assert.False(suite.T(), response["valid"].(bool))
				assert.NotEmpty(suite.T(), response["error"])
			} else {
				assert.Equal(suite.T(), http.StatusOK, w.Code)
				
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(suite.T(), err)
				
				assert.Equal(suite.T(), tt.expectedValid, response["valid"].(bool))
				assert.Equal(suite.T(), tt.expectedCost, response["resource_cost"].(float64))
				assert.Equal(suite.T(), tt.expectedFramework, response["optimal_framework"].(string))
				
				// Check recommendations exist
				if recommendations, ok := response["recommendations"]; ok {
					assert.IsType(suite.T(), []interface{}{}, recommendations)
				}
			}
		})
	}
}

// TestToolExecutionSystem tests comprehensive tool execution
func (suite *Week2TestSuite) TestToolExecutionSystem() {
	tests := []struct {
		name           string
		toolName       string
		parameters     map[string]interface{}
		expectedStatus string
		expectError    bool
		validateResult func(result map[string]interface{}) bool
	}{
		{
			name:     "Calculations Tool",
			toolName: "calculations",
			parameters: map[string]interface{}{
				"expression": "10 + 5 * 2",
			},
			expectedStatus: "completed",
			expectError:    false,
			validateResult: func(result map[string]interface{}) bool {
				return result["result"].(float64) == 20.0
			},
		},
		{
			name:     "Text Processing - Uppercase",
			toolName: "text_processing",
			parameters: map[string]interface{}{
				"text":      "hello world",
				"operation": "uppercase",
			},
			expectedStatus: "completed",
			expectError:    false,
			validateResult: func(result map[string]interface{}) bool {
				return result["processed"].(string) == "HELLO WORLD"
			},
		},
		{
			name:     "Text Processing - Word Count",
			toolName: "text_processing",
			parameters: map[string]interface{}{
				"text":      "AgentOS is the future of AI",
				"operation": "word_count",
			},
			expectedStatus: "completed",
			expectError:    false,
			validateResult: func(result map[string]interface{}) bool {
				return result["word_count"].(float64) == 6.0
			},
		},
		{
			name:     "Web Search Tool",
			toolName: "web_search",
			parameters: map[string]interface{}{
				"query":       "artificial intelligence",
				"max_results": 3,
			},
			expectedStatus: "completed",
			expectError:    false,
			validateResult: func(result map[string]interface{}) bool {
				return result["query"].(string) == "artificial intelligence"
			},
		},
		{
			name:     "Invalid Tool",
			toolName: "nonexistent_tool",
			parameters: map[string]interface{}{
				"test": "value",
			},
			expectError: true,
		},
		{
			name:     "Invalid Parameters",
			toolName: "calculations",
			parameters: map[string]interface{}{
				"invalid_param": "value",
			},
			expectError: true,
		},
		{
			name:     "Missing Required Parameters",
			toolName: "text_processing",
			parameters: map[string]interface{}{
				"operation": "uppercase",
				// missing "text" parameter
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			payload := map[string]interface{}{
				"tool_name":  tt.toolName,
				"parameters": tt.parameters,
				"timeout":    30,
			}

			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/api/v1/tools/execute", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+suite.token)

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			if tt.expectError {
				assert.True(suite.T(), w.Code >= 400)
			} else {
				assert.Equal(suite.T(), http.StatusOK, w.Code)
				
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(suite.T(), err)
				
				assert.Equal(suite.T(), tt.expectedStatus, response["status"].(string))
				assert.Equal(suite.T(), tt.toolName, response["tool_name"].(string))
				assert.NotEmpty(suite.T(), response["execution_id"])
				assert.NotNil(suite.T(), response["execution_time"])
				assert.NotNil(suite.T(), response["started_at"])
				assert.NotNil(suite.T(), response["completed_at"])
				
				// Validate result if validator provided
				if tt.validateResult != nil {
					result := response["result"].(map[string]interface{})
					assert.True(suite.T(), tt.validateResult(result))
				}
				
				// Test execution retrieval
				executionID := response["execution_id"].(string)
				suite.testToolExecutionRetrieval(executionID)
			}
		})
	}
}

// TestMemorySystem tests working memory system
func (suite *Week2TestSuite) TestMemorySystem() {
	// Test memory session creation
	suite.Run("Create Memory Session", func() {
		req := httptest.NewRequest("POST", "/api/v1/agents/"+suite.agentID+"/memory/session", nil)
		req.Header.Set("Authorization", "Bearer "+suite.token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(suite.T(), err)
		
		assert.Equal(suite.T(), suite.agentID, response["agent_id"].(string))
		assert.NotEmpty(suite.T(), response["session_id"])
		assert.NotNil(suite.T(), response["expires_at"])
		
		// Verify expiration is 24 hours from now
		expiresAt := response["expires_at"].(string)
		expTime, err := time.Parse(time.RFC3339, expiresAt)
		require.NoError(suite.T(), err)
		
		expectedExpiry := time.Now().Add(24 * time.Hour)
		assert.WithinDuration(suite.T(), expectedExpiry, expTime, 5*time.Minute)
	})

	// Test memory update
	suite.Run("Update Working Memory", func() {
		payload := map[string]interface{}{
			"variables": map[string]interface{}{
				"current_task":    "testing memory system",
				"step_count":      5,
				"user_preference": "detailed responses",
			},
			"context": map[string]interface{}{
				"conversation_id": "test-conv-123",
				"user_intent":     "testing",
				"session_start":   time.Now().Format(time.RFC3339),
			},
		}

		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("PUT", "/api/v1/agents/"+suite.agentID+"/memory/working", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(suite.T(), err)
		
		assert.Equal(suite.T(), suite.agentID, response["agent_id"].(string))
		assert.NotEmpty(suite.T(), response["session_id"])
		assert.Equal(suite.T(), float64(3), response["variables_count"].(float64))
		assert.NotNil(suite.T(), response["updated_at"])
	})

	// Test memory retrieval
	suite.Run("Retrieve Agent Memory", func() {
		req := httptest.NewRequest("GET", "/api/v1/agents/"+suite.agentID+"/memory", nil)
		req.Header.Set("Authorization", "Bearer "+suite.token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(suite.T(), err)
		
		assert.Equal(suite.T(), suite.agentID, response["agent_id"].(string))
		assert.NotNil(suite.T(), response["working_memory"])
		assert.NotNil(suite.T(), response["episodic_memories"])
		assert.NotNil(suite.T(), response["memory_stats"])
		
		// Check working memory contains our test data
		workingMemory := response["working_memory"].(map[string]interface{})
		if variables, ok := workingMemory["variables"]; ok {
			vars := variables.(map[string]interface{})
			assert.Equal(suite.T(), "testing memory system", vars["current_task"])
		}
	})

	// Test memory clearing
	suite.Run("Clear Agent Memory", func() {
		req := httptest.NewRequest("POST", "/api/v1/agents/"+suite.agentID+"/memory/clear", nil)
		req.Header.Set("Authorization", "Bearer "+suite.token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(suite.T(), err)
		
		assert.Equal(suite.T(), suite.agentID, response["agent_id"].(string))
		assert.Contains(suite.T(), response["message"].(string), "cleared successfully")
		assert.NotNil(suite.T(), response["cleared_at"])
	})
}

// Helper method to test tool execution retrieval
func (suite *Week2TestSuite) testToolExecutionRetrieval(executionID string) {
	req := httptest.NewRequest("GET", "/api/v1/tools/executions/"+executionID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	assert.Equal(suite.T(), executionID, response["execution_id"].(string))
	assert.NotNil(suite.T(), response["tool_name"])
	assert.NotNil(suite.T(), response["status"])
	assert.NotNil(suite.T(), response["result"])
}

// Helper method to create test agent
func (suite *Week2TestSuite) createTestAgent() string {
	payload := map[string]interface{}{
		"name":        "Week 2 Test Agent",
		"description": "Comprehensive test agent for Week 2 features",
		"capabilities": []string{"web_search", "calculations", "text_processing"},
		"personality": map[string]interface{}{
			"style":       "helpful",
			"tone":        "professional",
			"verbosity":   "detailed",
		},
		"framework_preference": "langchain",
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/agents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	require.Equal(suite.T(), http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)
	
	return response["id"].(string)
}

// TestWeek2Comprehensive runs the comprehensive test suite
func TestWeek2Comprehensive(t *testing.T) {
	suite.Run(t, new(Week2TestSuite))
}
