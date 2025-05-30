package tests

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
)

// TestHealthCheck tests the health check endpoint
func (suite *TestSuite) TestHealthCheck() {
	w := suite.makeRequest("GET", "/health", nil, "")

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "healthy", response["status"])
	assert.Equal(suite.T(), "agentos-core-api", response["service"])
	assert.Equal(suite.T(), "0.1.0-test", response["version"])
}

// TestUserRegistration tests user registration
func (suite *TestSuite) TestUserRegistration() {
	// Test successful registration
	registerData := map[string]string{
		"email":      "newuser@test.com",
		"password":   "password123",
		"first_name": "New",
		"last_name":  "User",
	}

	w := suite.makeRequest("POST", "/api/v1/auth/register", registerData, "")

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "token")
	assert.Contains(suite.T(), response, "user")

	user := response["user"].(map[string]interface{})
	assert.Equal(suite.T(), "newuser@test.com", user["email"])
	assert.Equal(suite.T(), "New", user["first_name"])
}

// TestUserRegistrationDuplicate tests duplicate email registration
func (suite *TestSuite) TestUserRegistrationDuplicate() {
	registerData := map[string]string{
		"email":    suite.testUser.Email,
		"password": "password123",
	}

	w := suite.makeRequest("POST", "/api/v1/auth/register", registerData, "")

	assert.Equal(suite.T(), http.StatusConflict, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response["error"], "already exists")
}

// TestUserLogin tests user login
func (suite *TestSuite) TestUserLogin() {
	// Test successful login
	loginData := map[string]string{
		"email":    suite.testUser.Email,
		"password": suite.testUser.Password,
	}

	w := suite.makeRequest("POST", "/api/v1/auth/login", loginData, "")

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "token")
	assert.Contains(suite.T(), response, "user")
}

// TestUserLoginInvalid tests invalid login credentials
func (suite *TestSuite) TestUserLoginInvalid() {
	loginData := map[string]string{
		"email":    suite.testUser.Email,
		"password": "wrongpassword",
	}

	w := suite.makeRequest("POST", "/api/v1/auth/login", loginData, "")

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response["error"], "Invalid email or password")
}

// TestListTools tests the tools listing endpoint
func (suite *TestSuite) TestListTools() {
	w := suite.makeRequest("GET", "/api/v1/tools", nil, "")

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "tools")
	assert.Contains(suite.T(), response, "count")
	assert.Contains(suite.T(), response, "tools_by_category")

	tools := response["tools"].([]interface{})
	assert.Equal(suite.T(), 5, len(tools)) // Should have 5 tools

	// Check first tool structure
	tool := tools[0].(map[string]interface{})
	assert.Contains(suite.T(), tool, "id")
	assert.Contains(suite.T(), tool, "name")
	assert.Contains(suite.T(), tool, "description")
	assert.Contains(suite.T(), tool, "category")
	assert.Contains(suite.T(), tool, "function_schema")
}

// TestAuthenticationRequired tests that protected endpoints require authentication
func (suite *TestSuite) TestAuthenticationRequired() {
	// Test without token
	w := suite.makeRequest("GET", "/api/v1/agents", nil, "")
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	// Test with invalid token
	w = suite.makeRequest("GET", "/api/v1/agents", nil, "invalid-token")
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// TestGetProfile tests user profile retrieval
func (suite *TestSuite) TestGetProfile() {
	w := suite.makeRequest("GET", "/api/v1/profile", nil, suite.testUser.Token)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), suite.testUser.Email, response["email"])
	assert.Equal(suite.T(), suite.testUser.ID, response["id"])
}

// TestUpdateProfile tests user profile update
func (suite *TestSuite) TestUpdateProfile() {
	updateData := map[string]string{
		"first_name": "Updated",
		"last_name":  "Name",
	}

	w := suite.makeRequest("PUT", "/api/v1/profile", updateData, suite.testUser.Token)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Updated", response["first_name"])
	assert.Equal(suite.T(), "Name", response["last_name"])
}

// TestListAgentsEmpty tests listing agents when none exist
func (suite *TestSuite) TestListAgentsEmpty() {
	w := suite.makeRequest("GET", "/api/v1/agents", nil, suite.testUser.Token)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "agents")
	assert.Contains(suite.T(), response, "count")

	// Handle case where agents might be nil
	if response["agents"] != nil {
		agents := response["agents"].([]interface{})
		assert.Equal(suite.T(), 0, len(agents))
	} else {
		// If agents is nil, that's also acceptable for empty list
		assert.Nil(suite.T(), response["agents"])
	}
	assert.Equal(suite.T(), float64(0), response["count"])
}
