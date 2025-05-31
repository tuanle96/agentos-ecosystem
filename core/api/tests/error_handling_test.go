package tests

import (
	"github.com/stretchr/testify/assert"
)

// TestErrorHandling tests various error scenarios
func (suite *TestSuite) TestErrorHandlingScenarios() {
	// Test invalid agent ID format
	w := suite.makeRequest("GET", "/api/v1/agents/invalid-uuid-format", nil, suite.testUser.Token)
	assert.Equal(suite.T(), 404, w.Code)

	// Test missing required fields
	payload := map[string]interface{}{
		"name": "", // Empty name should cause error
	}
	w = suite.makeRequest("POST", "/api/v1/agents", payload, suite.testUser.Token)
	assert.Equal(suite.T(), 400, w.Code)

	// Test unauthorized access
	w = suite.makeRequest("GET", "/api/v1/agents", nil, "invalid-token")
	assert.Equal(suite.T(), 401, w.Code)
}

// TestMemoryEndpointsErrorHandling tests memory endpoint error scenarios
func (suite *TestSuite) TestMemoryEndpointsErrorHandling() {
	// Test missing agent ID
	w := suite.makeRequest("GET", "/api/v1/memory/agents/", nil, suite.testUser.Token)
	assert.Equal(suite.T(), 404, w.Code)

	// Test invalid session creation
	payload := map[string]interface{}{
		"session_name": "", // Empty session name
	}
	w = suite.makeRequest("POST", "/api/v1/memory/working-sessions", payload, suite.testUser.Token)
	assert.Equal(suite.T(), 404, w.Code) // Route not implemented yet
}

// TestJWTErrorHandling tests JWT token error scenarios
func (suite *TestSuite) TestJWTErrorHandling() {
	// Test malformed token
	w := suite.makeRequest("GET", "/api/v1/profile", nil, "malformed.jwt.token")
	assert.Equal(suite.T(), 401, w.Code)

	// Test expired token (mock)
	w = suite.makeRequest("GET", "/api/v1/profile", nil, "expired.token.here")
	assert.Equal(suite.T(), 401, w.Code)
}
