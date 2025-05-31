package tests

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
)

// TestValidateCapabilities tests capability validation endpoint
func (suite *TestSuite) TestValidateCapabilities() {
	payload := map[string]interface{}{
		"capabilities": []string{"web_search", "calculations", "text_processing"},
	}

	w := suite.makeRequest("POST", "/api/v1/capabilities/validate", payload, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "valid")
	assert.Contains(suite.T(), response, "capabilities")
	assert.Contains(suite.T(), response, "conflicts")
	assert.Contains(suite.T(), response, "recommendations")
	assert.Equal(suite.T(), true, response["valid"])
}

// TestGetCapabilityRecommendations tests capability recommendations endpoint
func (suite *TestSuite) TestGetCapabilityRecommendations() {
	w := suite.makeRequest("GET", "/api/v1/capabilities/recommendations?task_description=data_analysis&framework=langchain", nil, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "recommendations")
	assert.Contains(suite.T(), response, "confidence_score")
	assert.Contains(suite.T(), response, "analysis")

	recommendations := response["recommendations"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(recommendations), 1)
}

// TestGetCapabilityRecommendationsComplex tests complex task recommendations
func (suite *TestSuite) TestGetCapabilityRecommendationsComplex() {
	w := suite.makeRequest("GET", "/api/v1/capabilities/recommendations?task_description=complex_ai_workflow&current_capabilities=web_search,calculations", nil, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "recommendations")
	assert.Contains(suite.T(), response, "confidence_score")

	confidenceScore := response["confidence_score"].(float64)
	assert.GreaterOrEqual(suite.T(), confidenceScore, 0.8)
}

// TestGetCapabilityRecommendationsInvalidTask tests invalid task handling
func (suite *TestSuite) TestGetCapabilityRecommendationsInvalidTask() {
	w := suite.makeRequest("GET", "/api/v1/capabilities/recommendations", nil, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "recommendations")
	assert.Contains(suite.T(), response, "confidence_score")
}

// TestValidateCapabilitiesWithConflicts tests capability validation with conflicts
func (suite *TestSuite) TestValidateCapabilitiesWithConflicts() {
	payload := map[string]interface{}{
		"capabilities": []string{"web_search", "file_operations", "api_calls", "calculations", "text_processing"},
	}

	w := suite.makeRequest("POST", "/api/v1/capabilities/validate", payload, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "valid")
	assert.Contains(suite.T(), response, "capabilities")
	assert.Contains(suite.T(), response, "conflicts")
	assert.Contains(suite.T(), response, "resource_cost")

	// Should detect conflicts due to too many capabilities
	conflicts := response["conflicts"].([]interface{})
	if len(conflicts) > 0 {
		firstConflict := conflicts[0].(map[string]interface{})
		assert.Contains(suite.T(), firstConflict, "type")
		assert.Contains(suite.T(), firstConflict, "capabilities")
		assert.Contains(suite.T(), firstConflict, "severity")
	}
}

// TestValidateCapabilitiesInvalid tests validation with invalid capabilities
func (suite *TestSuite) TestValidateCapabilitiesInvalid() {
	payload := map[string]interface{}{
		"capabilities": []string{"invalid_capability", "another_invalid"},
	}

	w := suite.makeRequest("POST", "/api/v1/capabilities/validate", payload, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "error")
	assert.Contains(suite.T(), response["error"].(string), "invalid capability")
}

// TestGetCapabilityRecommendationsAdvanced tests advanced capability recommendations
func (suite *TestSuite) TestGetCapabilityRecommendationsAdvanced() {
	payload := map[string]interface{}{
		"task_description": "I need an agent that can search the web and process text data",
		"domain":           "research",
		"complexity":       "medium",
	}

	w := suite.makeRequest("POST", "/api/v1/capabilities/recommendations", payload, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "recommendations")
	assert.Contains(suite.T(), response, "task_analysis")
	assert.Contains(suite.T(), response, "suggested_framework")

	recommendations := response["recommendations"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(recommendations), 1)

	// Check recommendation structure
	firstRec := recommendations[0].(map[string]interface{})
	assert.Contains(suite.T(), firstRec, "capability")
	assert.Contains(suite.T(), firstRec, "confidence")
	assert.Contains(suite.T(), firstRec, "reason")
}

// TestGetCapabilityRecommendationsComplexScenario tests recommendations for complex tasks
func (suite *TestSuite) TestGetCapabilityRecommendationsComplexScenario() {
	payload := map[string]interface{}{
		"task_description": "Create a comprehensive research agent that can search multiple sources, analyze data, generate reports, and interact with APIs",
		"domain":           "data_analysis",
		"complexity":       "high",
		"constraints": map[string]interface{}{
			"max_capabilities": 6,
			"security_level":   "high",
		},
	}

	w := suite.makeRequest("POST", "/api/v1/capabilities/recommendations", payload, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "recommendations")
	assert.Contains(suite.T(), response, "task_analysis")
	assert.Contains(suite.T(), response, "suggested_framework")
	assert.Contains(suite.T(), response, "resource_estimate")

	taskAnalysis := response["task_analysis"].(map[string]interface{})
	assert.Contains(suite.T(), taskAnalysis, "complexity")
	assert.Contains(suite.T(), taskAnalysis, "domain")
	assert.Contains(suite.T(), taskAnalysis, "key_requirements")
}

// TestSelectOptimalFramework tests framework selection
func (suite *TestSuite) TestSelectOptimalFramework() {
	payload := map[string]interface{}{
		"capabilities": []string{"web_search", "calculations", "text_processing"},
		"task_type":    "research",
		"performance_requirements": map[string]interface{}{
			"response_time": "fast",
			"accuracy":      "high",
			"scalability":   "medium",
		},
	}

	w := suite.makeRequest("POST", "/api/v1/framework/select", payload, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "selected_framework")
	assert.Contains(suite.T(), response, "confidence")
	assert.Contains(suite.T(), response, "reasoning")
	assert.Contains(suite.T(), response, "alternatives")

	selectedFramework := response["selected_framework"].(map[string]interface{})
	assert.Contains(suite.T(), selectedFramework, "name")
	assert.Contains(suite.T(), selectedFramework, "version")
	assert.Contains(suite.T(), selectedFramework, "strengths")
}

// TestResolveCapabilityConflicts tests conflict resolution
func (suite *TestSuite) TestResolveCapabilityConflicts() {
	payload := map[string]interface{}{
		"capabilities":        []string{"web_search", "file_operations", "api_calls", "calculations", "text_processing", "data_analysis"},
		"resolution_strategy": "optimize_performance",
		"constraints": map[string]interface{}{
			"max_capabilities": 4,
			"priority_order":   []string{"web_search", "calculations", "text_processing", "api_calls"},
		},
	}

	w := suite.makeRequest("POST", "/api/v1/capabilities/resolve-conflicts", payload, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "resolved_capabilities")
	assert.Contains(suite.T(), response, "removed_capabilities")
	assert.Contains(suite.T(), response, "resolution_reasoning")
	assert.Contains(suite.T(), response, "performance_impact")

	resolvedCaps := response["resolved_capabilities"].([]interface{})
	assert.LessOrEqual(suite.T(), len(resolvedCaps), 4) // Should respect max constraint
}

// TestCalculateResourceCost tests resource cost calculation
func (suite *TestSuite) TestCalculateResourceCost() {
	payload := map[string]interface{}{
		"capabilities": []string{"web_search", "calculations", "text_processing"},
		"expected_usage": map[string]interface{}{
			"requests_per_hour": 100,
			"concurrent_users":  10,
			"data_volume":       "medium",
		},
	}

	w := suite.makeRequest("POST", "/api/v1/capabilities/resource-cost", payload, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "total_cost")
	assert.Contains(suite.T(), response, "cost_breakdown")
	assert.Contains(suite.T(), response, "resource_requirements")
	assert.Contains(suite.T(), response, "optimization_suggestions")

	costBreakdown := response["cost_breakdown"].(map[string]interface{})
	assert.Contains(suite.T(), costBreakdown, "cpu_cost")
	assert.Contains(suite.T(), costBreakdown, "memory_cost")
	assert.Contains(suite.T(), costBreakdown, "network_cost")
}

// TestValidateCapabilitiesEmpty tests validation with empty capabilities
func (suite *TestSuite) TestValidateCapabilitiesEmpty() {
	payload := map[string]interface{}{
		"capabilities": []string{},
	}

	w := suite.makeRequest("POST", "/api/v1/capabilities/validate", payload, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "error")
}

// TestValidateCapabilitiesMissingPayload tests validation without payload
func (suite *TestSuite) TestValidateCapabilitiesMissingPayload() {
	w := suite.makeRequest("POST", "/api/v1/capabilities/validate", nil, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestGetCapabilityRecommendationsInvalidTaskScenario tests recommendations with invalid task
func (suite *TestSuite) TestGetCapabilityRecommendationsInvalidTaskScenario() {
	payload := map[string]interface{}{
		"task_description": "", // Empty task description
		"domain":           "unknown_domain",
	}

	w := suite.makeRequest("POST", "/api/v1/capabilities/recommendations", payload, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "error")
}

// TestFrameworkSelectionInvalidPayload tests framework selection with invalid data
func (suite *TestSuite) TestFrameworkSelectionInvalidPayload() {
	payload := map[string]interface{}{
		"capabilities": []string{}, // Empty capabilities
		"task_type":    "",         // Empty task type
	}

	w := suite.makeRequest("POST", "/api/v1/framework/select", payload, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestConflictResolutionInvalidStrategy tests conflict resolution with invalid strategy
func (suite *TestSuite) TestConflictResolutionInvalidStrategy() {
	payload := map[string]interface{}{
		"capabilities":        []string{"web_search", "calculations"},
		"resolution_strategy": "invalid_strategy",
	}

	w := suite.makeRequest("POST", "/api/v1/capabilities/resolve-conflicts", payload, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "error")
}

// TestResourceCostInvalidUsage tests resource cost with invalid usage data
func (suite *TestSuite) TestResourceCostInvalidUsage() {
	payload := map[string]interface{}{
		"capabilities": []string{"web_search"},
		"expected_usage": map[string]interface{}{
			"requests_per_hour": -1, // Invalid negative value
		},
	}

	w := suite.makeRequest("POST", "/api/v1/capabilities/resource-cost", payload, suite.testUser.Token)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestAgentFactoryUnauthorized tests agent factory operations without authentication
func (suite *TestSuite) TestAgentFactoryUnauthorized() {
	payload := map[string]interface{}{
		"capabilities": []string{"web_search", "calculations"},
	}

	// Test validation without auth
	w := suite.makeRequest("POST", "/api/v1/capabilities/validate", payload, "")
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	// Test recommendations without auth
	recPayload := map[string]interface{}{
		"task_description": "Test task",
	}

	w = suite.makeRequest("POST", "/api/v1/capabilities/recommendations", recPayload, "")
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	// Test framework selection without auth
	w = suite.makeRequest("POST", "/api/v1/framework/select", payload, "")
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}
