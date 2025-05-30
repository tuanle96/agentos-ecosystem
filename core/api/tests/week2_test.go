package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Week 2 Test Suite: Agent Factory, LangChain Integration, Memory System, Tool Execution

func TestAgentFactoryCapabilityValidation(t *testing.T) {
	// Test capability validation endpoint
	payload := map[string]interface{}{
		"capabilities": []string{"web_search", "calculations"},
	}
	
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/capabilities/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.True(t, response["valid"].(bool))
	assert.Equal(t, "langchain", response["optimal_framework"].(string))
	assert.NotNil(t, response["resource_cost"])
}

func TestAgentFactoryCapabilityConflicts(t *testing.T) {
	// Test capability conflict detection
	payload := map[string]interface{}{
		"capabilities": []string{"file_operations", "api_calls"}, // These should conflict
	}
	
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/capabilities/validate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.False(t, response["valid"].(bool))
	assert.Contains(t, response["error"].(string), "conflicts")
}

func TestAgentFactoryRecommendations(t *testing.T) {
	// Test capability recommendations
	req := httptest.NewRequest("GET", "/api/v1/capabilities/recommendations?capabilities=web_search", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.NotNil(t, response["recommendations"])
	assert.NotNil(t, response["resource_usage"])
	assert.Equal(t, float64(6), response["resource_limit"].(float64))
}

func TestWorkingMemorySession(t *testing.T) {
	// Create test agent first
	agentID := createTestAgent(t)
	
	// Create working memory session
	req := httptest.NewRequest("POST", "/api/v1/agents/"+agentID+"/memory/session", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.NotNil(t, response["session_id"])
	assert.Equal(t, agentID, response["agent_id"].(string))
	assert.NotNil(t, response["expires_at"])
}

func TestWorkingMemoryUpdate(t *testing.T) {
	// Create test agent first
	agentID := createTestAgent(t)
	
	// Update working memory
	payload := map[string]interface{}{
		"variables": map[string]interface{}{
			"current_task": "test task",
			"step_count":   1,
		},
		"context": map[string]interface{}{
			"user_intent": "testing",
		},
	}
	
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("PUT", "/api/v1/agents/"+agentID+"/memory/working", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.Equal(t, agentID, response["agent_id"].(string))
	assert.NotNil(t, response["session_id"])
	assert.Equal(t, float64(1), response["variables_count"].(float64))
}

func TestToolDefinitions(t *testing.T) {
	// Test tool definitions endpoint
	req := httptest.NewRequest("GET", "/api/v1/tools/definitions", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	tools := response["tools"].([]interface{})
	assert.GreaterOrEqual(t, len(tools), 5) // Should have at least 5 tools
	assert.Equal(t, float64(len(tools)), response["count"].(float64))
	
	// Check first tool structure
	firstTool := tools[0].(map[string]interface{})
	assert.NotNil(t, firstTool["name"])
	assert.NotNil(t, firstTool["description"])
	assert.NotNil(t, firstTool["category"])
	assert.NotNil(t, firstTool["security"])
}

func TestToolExecution(t *testing.T) {
	// Test tool execution
	payload := map[string]interface{}{
		"tool_name": "calculations",
		"parameters": map[string]interface{}{
			"expression": "2+2",
		},
		"timeout": 10,
	}
	
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/tools/execute", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.NotNil(t, response["execution_id"])
	assert.Equal(t, "calculations", response["tool_name"].(string))
	assert.Equal(t, "completed", response["status"].(string))
	assert.NotNil(t, response["result"])
	assert.NotNil(t, response["execution_time"])
}

func TestToolExecutionInvalidTool(t *testing.T) {
	// Test invalid tool execution
	payload := map[string]interface{}{
		"tool_name": "invalid_tool",
		"parameters": map[string]interface{}{
			"test": "value",
		},
	}
	
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/tools/execute", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.Contains(t, response["error"].(string), "Unknown tool")
}

func TestTextProcessingTool(t *testing.T) {
	// Test text processing tool
	payload := map[string]interface{}{
		"tool_name": "text_processing",
		"parameters": map[string]interface{}{
			"text":      "Hello World",
			"operation": "lowercase",
		},
	}
	
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/tools/execute", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.Equal(t, "completed", response["status"].(string))
	
	result := response["result"].(map[string]interface{})
	assert.Equal(t, "Hello World", result["original"].(string))
	assert.Equal(t, "hello world", result["processed"].(string))
	assert.Equal(t, "lowercase", result["operation"].(string))
}

func TestWebSearchTool(t *testing.T) {
	// Test web search tool
	payload := map[string]interface{}{
		"tool_name": "web_search",
		"parameters": map[string]interface{}{
			"query":       "AgentOS test",
			"max_results": 3,
		},
	}
	
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/tools/execute", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.Equal(t, "completed", response["status"].(string))
	
	result := response["result"].(map[string]interface{})
	assert.Equal(t, "AgentOS test", result["query"].(string))
	assert.NotNil(t, result["results"])
	assert.NotNil(t, result["count"])
}

func TestAgentMemoryRetrieval(t *testing.T) {
	// Create test agent first
	agentID := createTestAgent(t)
	
	// Get agent memory
	req := httptest.NewRequest("GET", "/api/v1/agents/"+agentID+"/memory", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.Equal(t, agentID, response["agent_id"].(string))
	assert.NotNil(t, response["working_memory"])
	assert.NotNil(t, response["episodic_memories"])
	assert.NotNil(t, response["memory_stats"])
}

func TestAgentMemoryClear(t *testing.T) {
	// Create test agent first
	agentID := createTestAgent(t)
	
	// Clear agent memory
	req := httptest.NewRequest("POST", "/api/v1/agents/"+agentID+"/memory/clear", nil)
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	assert.Equal(t, agentID, response["agent_id"].(string))
	assert.Contains(t, response["message"].(string), "cleared successfully")
	assert.NotNil(t, response["cleared_at"])
}

// Helper function to create test agent for memory and tool tests
func createTestAgent(t *testing.T) string {
	payload := map[string]interface{}{
		"name":        "Test Agent Week 2",
		"description": "Test agent for Week 2 features",
		"capabilities": []string{"web_search", "calculations"},
		"personality": map[string]interface{}{
			"style": "helpful",
		},
		"framework_preference": "langchain",
	}
	
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/agents", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+testToken)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	
	return response["id"].(string)
}
