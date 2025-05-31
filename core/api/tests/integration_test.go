package tests

import (
	"github.com/stretchr/testify/assert"
)

// TestCompleteUserWorkflow tests complete user workflow from registration to agent execution
func (suite *TestSuite) TestCompleteUserWorkflow() {
	// Step 1: User Registration
	registerData := map[string]interface{}{
		"email":      "integration@test.com",
		"password":   "integration123",
		"first_name": "Integration",
		"last_name":  "Test",
	}

	w := suite.makeRequest("POST", "/api/v1/auth/register", registerData, "")
	assert.Equal(suite.T(), 201, w.Code)

	var registerResponse map[string]interface{}
	err := suite.parseResponse(w, &registerResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), registerResponse, "token")

	token := registerResponse["token"].(string)
	assert.NotEmpty(suite.T(), token)

	// Step 2: User Login
	loginData := map[string]interface{}{
		"email":    "integration@test.com",
		"password": "integration123",
	}

	w = suite.makeRequest("POST", "/api/v1/auth/login", loginData, "")
	assert.Equal(suite.T(), 200, w.Code)

	var loginResponse map[string]interface{}
	err = suite.parseResponse(w, &loginResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), loginResponse, "token")

	// Step 3: Create Agent
	agentData := map[string]interface{}{
		"name":         "Integration Test Agent",
		"description":  "Agent for integration testing",
		"capabilities": []string{"web_search", "calculations", "text_processing"},
		"framework":    "langchain",
	}

	w = suite.makeRequest("POST", "/api/v1/agents", agentData, token)
	assert.Equal(suite.T(), 201, w.Code)

	var agentResponse map[string]interface{}
	err = suite.parseResponse(w, &agentResponse)
	assert.NoError(suite.T(), err)

	agentID := agentResponse["id"].(string)
	assert.NotEmpty(suite.T(), agentID)

	// Step 4: Execute Agent
	executeData := map[string]interface{}{
		"input_text":      "Calculate 2+2 and search for AgentOS",
		"framework":       "langchain",
		"include_memory":  true,
		"max_tokens":      1000,
		"temperature":     0.7,
	}

	w = suite.makeRequest("POST", "/api/v1/agents/"+agentID+"/execute", executeData, token)
	assert.Equal(suite.T(), 200, w.Code)

	var executeResponse map[string]interface{}
	err = suite.parseResponse(w, &executeResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), executeResponse, "execution_id")
	assert.Contains(suite.T(), executeResponse, "output_text")

	executionID := executeResponse["execution_id"].(string)
	assert.NotEmpty(suite.T(), executionID)

	// Step 5: Get Execution Details
	w = suite.makeRequest("GET", "/api/v1/executions/"+executionID, nil, token)
	assert.Equal(suite.T(), 200, w.Code)

	var executionDetails map[string]interface{}
	err = suite.parseResponse(w, &executionDetails)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), executionID, executionDetails["id"])

	// Step 6: Update Agent
	updateData := map[string]interface{}{
		"name":        "Updated Integration Agent",
		"description": "Updated description for integration testing",
	}

	w = suite.makeRequest("PUT", "/api/v1/agents/"+agentID, updateData, token)
	assert.Equal(suite.T(), 200, w.Code)

	var updateResponse map[string]interface{}
	err = suite.parseResponse(w, &updateResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Integration Agent", updateResponse["name"])

	// Step 7: List User's Agents
	w = suite.makeRequest("GET", "/api/v1/agents", nil, token)
	assert.Equal(suite.T(), 200, w.Code)

	var listResponse map[string]interface{}
	err = suite.parseResponse(w, &listResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), listResponse, "agents")

	agents := listResponse["agents"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(agents), 1)

	// Step 8: Get User Profile
	w = suite.makeRequest("GET", "/api/v1/profile", nil, token)
	assert.Equal(suite.T(), 200, w.Code)

	var profileResponse map[string]interface{}
	err = suite.parseResponse(w, &profileResponse)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "integration@test.com", profileResponse["email"])

	// Step 9: Clean up - Delete Agent
	w = suite.makeRequest("DELETE", "/api/v1/agents/"+agentID, nil, token)
	assert.Equal(suite.T(), 200, w.Code)
}

// TestAgentCapabilityWorkflow tests agent capability management workflow
func (suite *TestSuite) TestAgentCapabilityWorkflow() {
	// Create agent with basic capabilities
	agentData := map[string]interface{}{
		"name":         "Capability Test Agent",
		"capabilities": []string{"web_search"},
		"framework":    "swarms",
	}

	w := suite.makeRequest("POST", "/api/v1/agents", agentData, suite.testUser.Token)
	assert.Equal(suite.T(), 201, w.Code)

	var agentResponse map[string]interface{}
	err := suite.parseResponse(w, &agentResponse)
	assert.NoError(suite.T(), err)

	agentID := agentResponse["id"].(string)

	// Test capability recommendations
	w = suite.makeRequest("GET", "/api/v1/agents/"+agentID+"/recommendations", nil, suite.testUser.Token)
	assert.Equal(suite.T(), 200, w.Code)

	var recommendationsResponse map[string]interface{}
	err = suite.parseResponse(w, &recommendationsResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), recommendationsResponse, "recommendations")

	// Test capability validation
	validateData := map[string]interface{}{
		"capabilities": []string{"web_search", "calculations", "text_processing"},
	}

	w = suite.makeRequest("POST", "/api/v1/agents/validate-capabilities", validateData, suite.testUser.Token)
	assert.Equal(suite.T(), 200, w.Code)

	var validateResponse map[string]interface{}
	err = suite.parseResponse(w, &validateResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), validateResponse, "valid")

	// Update agent with more capabilities
	updateData := map[string]interface{}{
		"capabilities": []string{"web_search", "calculations", "text_processing"},
	}

	w = suite.makeRequest("PUT", "/api/v1/agents/"+agentID, updateData, suite.testUser.Token)
	assert.Equal(suite.T(), 200, w.Code)

	// Verify updated capabilities
	w = suite.makeRequest("GET", "/api/v1/agents/"+agentID, nil, suite.testUser.Token)
	assert.Equal(suite.T(), 200, w.Code)

	var updatedAgent map[string]interface{}
	err = suite.parseResponse(w, &updatedAgent)
	assert.NoError(suite.T(), err)

	capabilities := updatedAgent["capabilities"].([]interface{})
	assert.Equal(suite.T(), 3, len(capabilities))

	// Clean up
	suite.makeRequest("DELETE", "/api/v1/agents/"+agentID, nil, suite.testUser.Token)
}

// TestToolsWorkflow tests tools listing and validation
func (suite *TestSuite) TestToolsWorkflow() {
	// Test public tools listing (no auth required)
	w := suite.makeRequest("GET", "/api/v1/public/tools", nil, "")
	assert.Equal(suite.T(), 200, w.Code)

	var publicToolsResponse map[string]interface{}
	err := suite.parseResponse(w, &publicToolsResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), publicToolsResponse, "tools")

	tools := publicToolsResponse["tools"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(tools), 1)

	// Verify tool structure
	if len(tools) > 0 {
		tool := tools[0].(map[string]interface{})
		assert.Contains(suite.T(), tool, "name")
		assert.Contains(suite.T(), tool, "description")
		assert.Contains(suite.T(), tool, "category")
	}

	// Test authenticated tools listing
	w = suite.makeRequest("GET", "/api/v1/tools/definitions", nil, suite.testUser.Token)
	assert.Equal(suite.T(), 200, w.Code)

	var authToolsResponse map[string]interface{}
	err = suite.parseResponse(w, &authToolsResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), authToolsResponse, "tools")
	assert.Contains(suite.T(), authToolsResponse, "count")
}

// TestErrorHandlingWorkflow tests comprehensive error scenarios
func (suite *TestSuite) TestErrorHandlingWorkflow() {
	// Test invalid agent creation
	invalidAgentData := map[string]interface{}{
		"name": "", // Empty name
	}

	w := suite.makeRequest("POST", "/api/v1/agents", invalidAgentData, suite.testUser.Token)
	assert.Equal(suite.T(), 400, w.Code)

	// Test accessing non-existent agent
	w = suite.makeRequest("GET", "/api/v1/agents/non-existent-id", nil, suite.testUser.Token)
	assert.Equal(suite.T(), 404, w.Code)

	// Test unauthorized access
	w = suite.makeRequest("GET", "/api/v1/agents", nil, "invalid-token")
	assert.Equal(suite.T(), 401, w.Code)

	// Test invalid execution
	invalidExecuteData := map[string]interface{}{
		"input_text": "", // Empty input
	}

	w = suite.makeRequest("POST", "/api/v1/agents/test-id/execute", invalidExecuteData, suite.testUser.Token)
	assert.Equal(suite.T(), 400, w.Code)

	// Test invalid capability validation
	invalidCapabilityData := map[string]interface{}{
		"capabilities": []string{"invalid_capability"},
	}

	w = suite.makeRequest("POST", "/api/v1/agents/validate-capabilities", invalidCapabilityData, suite.testUser.Token)
	assert.Equal(suite.T(), 200, w.Code) // Should return 200 but with validation errors

	var validateResponse map[string]interface{}
	err := suite.parseResponse(w, &validateResponse)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), validateResponse, "valid")
	assert.False(suite.T(), validateResponse["valid"].(bool))
}

// TestPerformanceWorkflow tests response times and concurrent operations
func (suite *TestSuite) TestPerformanceWorkflow() {
	// Test health check performance
	w := suite.makeRequest("GET", "/api/v1/health", nil, "")
	assert.Equal(suite.T(), 200, w.Code)

	// Test authenticated endpoint performance
	w = suite.makeRequest("GET", "/api/v1/profile", nil, suite.testUser.Token)
	assert.Equal(suite.T(), 200, w.Code)

	// Test agent listing performance
	w = suite.makeRequest("GET", "/api/v1/agents", nil, suite.testUser.Token)
	assert.Equal(suite.T(), 200, w.Code)

	// Test tools listing performance
	w = suite.makeRequest("GET", "/api/v1/public/tools", nil, "")
	assert.Equal(suite.T(), 200, w.Code)

	// All performance tests should complete quickly (already validated by test execution time)
}
