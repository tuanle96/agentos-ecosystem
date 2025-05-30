package tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
)

// TestCreateAgent tests agent creation
func (suite *TestSuite) TestCreateAgent() {
	agentData := map[string]interface{}{
		"name":                 "Test Agent",
		"description":          "A test agent for testing",
		"capabilities":         []string{"web_search", "text_processing"},
		"framework_preference": "auto",
		"personality": map[string]interface{}{
			"tone":  "friendly",
			"style": "professional",
		},
	}

	w := suite.makeRequest("POST", "/api/v1/agents", agentData, suite.testUser.Token)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "id")
	assert.Equal(suite.T(), "Test Agent", response["name"])
	assert.Equal(suite.T(), "A test agent for testing", response["description"])
	assert.Equal(suite.T(), "auto", response["framework_preference"])
	assert.Equal(suite.T(), "active", response["status"])
	assert.Equal(suite.T(), suite.testUser.ID, response["user_id"])
}

// TestCreateAgentInvalidCapabilities tests agent creation with invalid capabilities
func (suite *TestSuite) TestCreateAgentInvalidCapabilities() {
	agentData := map[string]interface{}{
		"name":         "Invalid Agent",
		"capabilities": []string{"invalid_capability"},
	}

	w := suite.makeRequest("POST", "/api/v1/agents", agentData, suite.testUser.Token)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response["error"], "Invalid capability")
}

// TestCreateAgentTooManyCapabilities tests agent creation with too many capabilities
func (suite *TestSuite) TestCreateAgentTooManyCapabilities() {
	agentData := map[string]interface{}{
		"name": "Overloaded Agent",
		"capabilities": []string{
			"web_search", "text_processing", "file_operations", "calculations",
		}, // 4 capabilities, max is 3 for MVP
	}

	w := suite.makeRequest("POST", "/api/v1/agents", agentData, suite.testUser.Token)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// The validation happens in handler logic, not Gin binding
	assert.Contains(suite.T(), response["error"], "Maximum 3 capabilities")
}

// TestGetAgent tests retrieving a specific agent
func (suite *TestSuite) TestGetAgent() {
	// First create an agent
	agentData := map[string]interface{}{
		"name":         "Get Test Agent",
		"description":  "Agent for get testing",
		"capabilities": []string{"web_search"},
	}

	w := suite.makeRequest("POST", "/api/v1/agents", agentData, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	agentID := createResponse["id"].(string)

	// Now get the agent
	w = suite.makeRequest("GET", fmt.Sprintf("/api/v1/agents/%s", agentID), nil, suite.testUser.Token)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), agentID, response["id"])
	assert.Equal(suite.T(), "Get Test Agent", response["name"])
	assert.Equal(suite.T(), "Agent for get testing", response["description"])
}

// TestGetAgentNotFound tests retrieving a non-existent agent
func (suite *TestSuite) TestGetAgentNotFound() {
	w := suite.makeRequest("GET", "/api/v1/agents/00000000-0000-0000-0000-000000000000", nil, suite.testUser.Token)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response["error"], "not found")
}

// TestUpdateAgent tests agent update
func (suite *TestSuite) TestUpdateAgent() {
	// First create an agent
	agentData := map[string]interface{}{
		"name":         "Update Test Agent",
		"capabilities": []string{"web_search"},
	}

	w := suite.makeRequest("POST", "/api/v1/agents", agentData, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	agentID := createResponse["id"].(string)

	// Update the agent
	updateData := map[string]interface{}{
		"name":                 "Updated Agent Name",
		"description":          "Updated description",
		"capabilities":         []string{"web_search", "text_processing"},
		"framework_preference": "langchain",
	}

	w = suite.makeRequest("PUT", fmt.Sprintf("/api/v1/agents/%s", agentID), updateData, suite.testUser.Token)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Updated Agent Name", response["name"])
	assert.Equal(suite.T(), "Updated description", response["description"])
	assert.Equal(suite.T(), "langchain", response["framework_preference"])
}

// TestDeleteAgent tests agent deletion
func (suite *TestSuite) TestDeleteAgent() {
	// First create an agent
	agentData := map[string]interface{}{
		"name":         "Delete Test Agent",
		"capabilities": []string{"web_search"},
	}

	w := suite.makeRequest("POST", "/api/v1/agents", agentData, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	agentID := createResponse["id"].(string)

	// Delete the agent
	w = suite.makeRequest("DELETE", fmt.Sprintf("/api/v1/agents/%s", agentID), nil, suite.testUser.Token)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response["message"], "deleted successfully")

	// Verify agent is deleted
	w = suite.makeRequest("GET", fmt.Sprintf("/api/v1/agents/%s", agentID), nil, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)
}

// TestListAgentsWithData tests listing agents when some exist
func (suite *TestSuite) TestListAgentsWithData() {
	// Create multiple agents
	agents := []map[string]interface{}{
		{
			"name":         "Agent 1",
			"capabilities": []string{"web_search"},
		},
		{
			"name":         "Agent 2",
			"capabilities": []string{"text_processing"},
		},
		{
			"name":         "Agent 3",
			"capabilities": []string{"calculations"},
		},
	}

	for _, agentData := range agents {
		w := suite.makeRequest("POST", "/api/v1/agents", agentData, suite.testUser.Token)
		assert.Equal(suite.T(), http.StatusCreated, w.Code)
	}

	// List agents
	w := suite.makeRequest("GET", "/api/v1/agents", nil, suite.testUser.Token)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "agents")
	assert.Contains(suite.T(), response, "count")

	agentsList := response["agents"].([]interface{})
	assert.Equal(suite.T(), 3, len(agentsList))
	assert.Equal(suite.T(), float64(3), response["count"])

	// Check agent structure
	agent := agentsList[0].(map[string]interface{})
	assert.Contains(suite.T(), agent, "id")
	assert.Contains(suite.T(), agent, "name")
	assert.Contains(suite.T(), agent, "user_id")
	assert.Contains(suite.T(), agent, "status")
	assert.Contains(suite.T(), agent, "created_at")
}

// TestAgentExecution tests agent execution
func (suite *TestSuite) TestAgentExecution() {
	// First create an agent
	agentData := map[string]interface{}{
		"name":         "Execution Test Agent",
		"capabilities": []string{"web_search", "text_processing"},
	}

	w := suite.makeRequest("POST", "/api/v1/agents", agentData, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	agentID := createResponse["id"].(string)

	// Execute the agent
	executeData := map[string]interface{}{
		"input_text": "Hello, can you help me search for information about AI?",
	}

	w = suite.makeRequest("POST", fmt.Sprintf("/api/v1/agents/%s/execute", agentID), executeData, suite.testUser.Token)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "execution_id")
	assert.Contains(suite.T(), response, "output_text")
	assert.Contains(suite.T(), response, "status")
	assert.Contains(suite.T(), response, "framework_used")
	assert.Contains(suite.T(), response, "execution_time_ms")

	assert.Equal(suite.T(), "completed", response["status"])
	assert.Equal(suite.T(), "simulated", response["framework_used"])
	assert.Contains(suite.T(), response["output_text"], "AgentOS MVP agent")
}
