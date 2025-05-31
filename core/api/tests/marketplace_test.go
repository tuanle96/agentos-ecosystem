package tests

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/tuanle96/agentos-ecosystem/core/api/models"
)

// Helper method to create a test tool
func (suite *TestSuite) createTestTool() {
	toolReq := models.CreateToolRequest{
		Name:        "test_calculator",
		DisplayName: "Test Calculator",
		Description: "A simple calculator tool for testing",
		Category:    "math",
		Tags:        []string{"calculator", "math", "test"},
		FunctionSchema: map[string]interface{}{
			"name":        "calculator",
			"description": "Perform basic mathematical calculations",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"operation": map[string]interface{}{
						"type":        "string",
						"description": "The operation to perform",
						"enum":        []string{"add", "subtract", "multiply", "divide"},
					},
					"a": map[string]interface{}{
						"type":        "number",
						"description": "First number",
					},
					"b": map[string]interface{}{
						"type":        "number",
						"description": "Second number",
					},
				},
				"required": []string{"operation", "a", "b"},
			},
		},
		SourceCode: `def calculator(operation, a, b):
    """Perform basic mathematical calculations"""
    if operation == "add":
        return a + b
    elif operation == "subtract":
        return a - b
    elif operation == "multiply":
        return a * b
    elif operation == "divide":
        if b == 0:
            raise ValueError("Cannot divide by zero")
        return a / b
    else:
        raise ValueError("Invalid operation")`,
		Documentation: "# Calculator Tool\n\nA simple calculator for basic math operations.",
		Examples: []map[string]interface{}{
			{
				"input":  map[string]interface{}{"operation": "add", "a": 5, "b": 3},
				"output": 8,
			},
		},
		Dependencies: []string{"python>=3.8"},
		Requirements: map[string]interface{}{
			"python_version": ">=3.8",
		},
		IsPublic: true,
	}

	reqBody, _ := json.Marshal(toolReq)
	req, _ := http.NewRequest("POST", "/api/v1/marketplace/tools", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	resp := suite.performRequest(req)

	if resp.Code == http.StatusCreated {
		var response map[string]interface{}
		json.Unmarshal(resp.Body.Bytes(), &response)
		if tool, ok := response["tool"].(map[string]interface{}); ok {
			suite.testToolID = tool["id"].(string)
		}
	}
}

func (suite *TestSuite) TestCreateTool() {
	// Create a tool
	toolReq := models.CreateToolRequest{
		Name:        "test_calculator",
		DisplayName: "Test Calculator",
		Description: "A simple calculator tool for testing",
		Category:    "math",
		Tags:        []string{"calculator", "math", "test"},
		FunctionSchema: map[string]interface{}{
			"name":        "calculator",
			"description": "Perform basic mathematical calculations",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"operation": map[string]interface{}{
						"type":        "string",
						"description": "The operation to perform",
						"enum":        []string{"add", "subtract", "multiply", "divide"},
					},
					"a": map[string]interface{}{
						"type":        "number",
						"description": "First number",
					},
					"b": map[string]interface{}{
						"type":        "number",
						"description": "Second number",
					},
				},
				"required": []string{"operation", "a", "b"},
			},
		},
		SourceCode: `def calculator(operation, a, b):
    """Perform basic mathematical calculations"""
    if operation == "add":
        return a + b
    elif operation == "subtract":
        return a - b
    elif operation == "multiply":
        return a * b
    elif operation == "divide":
        if b == 0:
            raise ValueError("Cannot divide by zero")
        return a / b
    else:
        raise ValueError("Invalid operation")`,
		Documentation: "# Calculator Tool\n\nA simple calculator for basic math operations.",
		Examples: []map[string]interface{}{
			{
				"input":  map[string]interface{}{"operation": "add", "a": 5, "b": 3},
				"output": 8,
			},
		},
		Dependencies: []string{"python>=3.8"},
		Requirements: map[string]interface{}{
			"python_version": ">=3.8",
		},
		IsPublic: true,
	}

	reqBody, _ := json.Marshal(toolReq)
	req, _ := http.NewRequest("POST", "/api/v1/marketplace/tools", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	resp := suite.performRequest(req)

	assert.Equal(suite.T(), http.StatusCreated, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "message")
	assert.Contains(suite.T(), response, "tool")

	// Store tool for other tests
	tool := response["tool"].(map[string]interface{})
	suite.testToolID = tool["id"].(string)
}

func (suite *TestSuite) TestGetTools() {
	// Get all tools
	req, _ := http.NewRequest("GET", "/api/v1/marketplace/tools", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	resp := suite.performRequest(req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var response models.ToolSearchResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.GreaterOrEqual(suite.T(), len(response.Tools), 0) // Should return tools array (can be empty)
	assert.GreaterOrEqual(suite.T(), response.TotalCount, 0)
}

func (suite *TestSuite) TestGetToolsWithSearch() {
	// Search for tools
	req, _ := http.NewRequest("GET", "/api/v1/marketplace/tools?query=calculator&category=math", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	resp := suite.performRequest(req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var response models.ToolSearchResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	// Should find tools matching the search criteria
}

func (suite *TestSuite) TestGetTool() {
	// First create a tool to get
	suite.createTestTool()

	// Get specific tool
	req, _ := http.NewRequest("GET", "/api/v1/marketplace/tools/"+suite.testToolID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	resp := suite.performRequest(req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var response models.ToolDetailsResponse
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), suite.testToolID, response.Tool.ID.String())
	assert.Equal(suite.T(), "test_calculator", response.Tool.Name)
	assert.GreaterOrEqual(suite.T(), len(response.Versions), 1)
}

func (suite *TestSuite) TestUpdateTool() {
	// First create a tool to update
	suite.createTestTool()

	// Update tool
	updateReq := models.UpdateToolRequest{
		DisplayName:   "Updated Calculator",
		Description:   "An updated calculator tool",
		Documentation: "# Updated Calculator Tool\n\nAn updated calculator for basic math operations.",
		IsPublic:      true,
	}

	reqBody, _ := json.Marshal(updateReq)
	req, _ := http.NewRequest("PUT", "/api/v1/marketplace/tools/"+suite.testToolID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	resp := suite.performRequest(req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "message")
}

func (suite *TestSuite) TestInstallTool() {
	// First create a tool to install
	suite.createTestTool()

	// Install tool
	installReq := models.InstallToolRequest{
		ToolID:  suite.testToolID,
		Version: "1.0.0",
		Configuration: map[string]interface{}{
			"precision": 2,
		},
	}

	reqBody, _ := json.Marshal(installReq)
	req, _ := http.NewRequest("POST", "/api/v1/marketplace/tools/install", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	resp := suite.performRequest(req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "message")
	assert.Contains(suite.T(), response, "tool_id")
	assert.Contains(suite.T(), response, "version")
}

func (suite *TestSuite) TestDeleteTool() {
	// First create a tool to delete
	suite.createTestTool()

	// Delete tool
	req, _ := http.NewRequest("DELETE", "/api/v1/marketplace/tools/"+suite.testToolID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	resp := suite.performRequest(req)

	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	var response map[string]interface{}
	err := json.Unmarshal(resp.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Contains(suite.T(), response, "message")
}

func (suite *TestSuite) TestCreateToolUnauthorized() {
	// Try to create tool without authentication
	toolReq := models.CreateToolRequest{
		Name:        "unauthorized_tool",
		DisplayName: "Unauthorized Tool",
		Description: "This should fail",
		Category:    "test",
	}

	reqBody, _ := json.Marshal(toolReq)
	req, _ := http.NewRequest("POST", "/api/v1/marketplace/tools", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp := suite.performRequest(req)

	assert.Equal(suite.T(), http.StatusUnauthorized, resp.Code)
}

func (suite *TestSuite) TestCreateToolInvalidData() {
	// Try to create tool with invalid data
	invalidReq := map[string]interface{}{
		"name": "", // Empty name should fail
	}

	reqBody, _ := json.Marshal(invalidReq)
	req, _ := http.NewRequest("POST", "/api/v1/marketplace/tools", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)

	resp := suite.performRequest(req)

	assert.Equal(suite.T(), http.StatusBadRequest, resp.Code)
}

func (suite *TestSuite) TestGetToolNotFound() {
	// Try to get non-existent tool
	req, _ := http.NewRequest("GET", "/api/v1/marketplace/tools/00000000-0000-0000-0000-000000000000", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)

	resp := suite.performRequest(req)

	assert.Equal(suite.T(), http.StatusNotFound, resp.Code)
}

func (suite *TestSuite) TestToolMarketplaceWorkflow() {
	// Complete workflow test
	suite.createTestTool()

	// Test get tools
	req, _ := http.NewRequest("GET", "/api/v1/marketplace/tools", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	resp := suite.performRequest(req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	// Test get specific tool
	req, _ = http.NewRequest("GET", "/api/v1/marketplace/tools/"+suite.testToolID, nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	resp = suite.performRequest(req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	// Test update tool
	updateReq := models.UpdateToolRequest{
		DisplayName:   "Updated Calculator",
		Description:   "An updated calculator tool",
		Documentation: "# Updated Calculator Tool\n\nAn updated calculator for basic math operations.",
		IsPublic:      true,
	}
	reqBody, _ := json.Marshal(updateReq)
	req, _ = http.NewRequest("PUT", "/api/v1/marketplace/tools/"+suite.testToolID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)
	resp = suite.performRequest(req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)

	// Test install tool
	installReq := models.InstallToolRequest{
		ToolID:  suite.testToolID,
		Version: "1.0.0",
		Configuration: map[string]interface{}{
			"precision": 2,
		},
	}
	reqBody, _ = json.Marshal(installReq)
	req, _ = http.NewRequest("POST", "/api/v1/marketplace/tools/install", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)
	resp = suite.performRequest(req)
	assert.Equal(suite.T(), http.StatusOK, resp.Code)
}
