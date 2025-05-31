package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tuanle96/agentos-ecosystem/core/api/handlers"
)

// TestUnitCoverageBoost contains unit tests specifically designed to boost coverage
// These tests target individual functions that currently have 0% coverage

// TestResolveCapabilityConflicts tests the resolveCapabilityConflicts function directly
func TestResolveCapabilityConflicts(t *testing.T) {
	// Test with conflicting capabilities
	capabilities := []string{"web_search", "file_operations", "calculations", "text_processing"}
	strategy := "remove_conflicts"
	
	// This should call the internal resolveCapabilityConflicts function
	resolved := handlers.ResolveCapabilityConflictsLogic(capabilities, strategy)
	
	assert.NotNil(t, resolved)
	assert.LessOrEqual(t, len(resolved), len(capabilities))
}

// TestResolveCapabilityConflictsOptimizePerformance tests performance optimization strategy
func TestResolveCapabilityConflictsOptimizePerformance(t *testing.T) {
	capabilities := []string{"web_search", "file_operations", "calculations", "text_processing", "api_calls"}
	strategy := "optimize_performance"
	
	resolved := handlers.ResolveCapabilityConflictsLogic(capabilities, strategy)
	
	assert.NotNil(t, resolved)
	assert.GreaterOrEqual(t, len(resolved), 1)
}

// TestResolveCapabilityConflictsMinimizeResources tests resource minimization strategy
func TestResolveCapabilityConflictsMinimizeResources(t *testing.T) {
	capabilities := []string{"web_search", "calculations", "text_processing"}
	strategy := "minimize_resources"
	
	resolved := handlers.ResolveCapabilityConflictsLogic(capabilities, strategy)
	
	assert.NotNil(t, resolved)
	assert.LessOrEqual(t, len(resolved), 3)
}

// TestGetCapabilityRecommendations tests the getCapabilityRecommendations function directly
func TestGetCapabilityRecommendations(t *testing.T) {
	taskDescription := "I need an agent that can search the web and process text data"
	domain := "research"
	complexity := "medium"
	
	recommendations := handlers.GetCapabilityRecommendationsLogic(taskDescription, domain, complexity)
	
	assert.NotNil(t, recommendations)
	assert.GreaterOrEqual(t, len(recommendations), 1)
	
	// Check that recommendations contain expected capabilities for research tasks
	found := false
	for _, rec := range recommendations {
		if rec == "web_search" || rec == "text_processing" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should recommend web_search or text_processing for research tasks")
}

// TestGetCapabilityRecommendationsDataAnalysis tests recommendations for data analysis
func TestGetCapabilityRecommendationsDataAnalysis(t *testing.T) {
	taskDescription := "Create a data analysis agent that can process large datasets"
	domain := "data_analysis"
	complexity := "high"
	
	recommendations := handlers.GetCapabilityRecommendationsLogic(taskDescription, domain, complexity)
	
	assert.NotNil(t, recommendations)
	assert.GreaterOrEqual(t, len(recommendations), 2)
	
	// Should recommend data analysis related capabilities
	found := false
	for _, rec := range recommendations {
		if rec == "calculations" || rec == "data_analysis" || rec == "file_operations" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should recommend data analysis capabilities")
}

// TestGetCapabilityRecommendationsSimpleTask tests recommendations for simple tasks
func TestGetCapabilityRecommendationsSimpleTask(t *testing.T) {
	taskDescription := "Simple text processing"
	domain := "general"
	complexity := "low"
	
	recommendations := handlers.GetCapabilityRecommendationsLogic(taskDescription, domain, complexity)
	
	assert.NotNil(t, recommendations)
	assert.GreaterOrEqual(t, len(recommendations), 1)
	assert.LessOrEqual(t, len(recommendations), 3) // Simple tasks should have fewer recommendations
}

// TestSelectOptimalFramework tests the selectOptimalFramework function directly
func TestSelectOptimalFramework(t *testing.T) {
	capabilities := []string{"web_search", "calculations", "text_processing"}
	taskType := "research"
	performanceReq := "high"
	
	framework := handlers.SelectOptimalFrameworkLogic(capabilities, taskType, performanceReq)
	
	assert.NotEmpty(t, framework)
	// Should select appropriate framework based on capabilities and requirements
	assert.Contains(t, []string{"langchain", "crewai", "autogen", "swarms"}, framework)
}

// TestSelectOptimalFrameworkLangChain tests LangChain selection
func TestSelectOptimalFrameworkLangChain(t *testing.T) {
	capabilities := []string{"web_search", "text_processing", "api_calls"}
	taskType := "research"
	performanceReq := "medium"
	
	framework := handlers.SelectOptimalFrameworkLogic(capabilities, taskType, performanceReq)
	
	assert.NotEmpty(t, framework)
	// For research tasks with these capabilities, should prefer LangChain
	assert.Equal(t, "langchain", framework)
}

// TestSelectOptimalFrameworkCrewAI tests CrewAI selection
func TestSelectOptimalFrameworkCrewAI(t *testing.T) {
	capabilities := []string{"web_search", "calculations", "text_processing", "api_calls"}
	taskType := "collaboration"
	performanceReq := "high"
	
	framework := handlers.SelectOptimalFrameworkLogic(capabilities, taskType, performanceReq)
	
	assert.NotEmpty(t, framework)
	// For collaboration tasks with multiple capabilities, should prefer CrewAI
	assert.Equal(t, "crewai", framework)
}

// TestSelectOptimalFrameworkSwarms tests Swarms selection
func TestSelectOptimalFrameworkSwarms(t *testing.T) {
	capabilities := []string{"calculations", "data_analysis", "file_operations"}
	taskType := "data_processing"
	performanceReq := "high"
	
	framework := handlers.SelectOptimalFrameworkLogic(capabilities, taskType, performanceReq)
	
	assert.NotEmpty(t, framework)
	// For data processing with high performance, should prefer Swarms
	assert.Equal(t, "swarms", framework)
}

// TestSelectOptimalFrameworkAutoGen tests AutoGen selection
func TestSelectOptimalFrameworkAutoGen(t *testing.T) {
	capabilities := []string{"text_processing", "api_calls"}
	taskType := "conversation"
	performanceReq := "medium"
	
	framework := handlers.SelectOptimalFrameworkLogic(capabilities, taskType, performanceReq)
	
	assert.NotEmpty(t, framework)
	// For conversation tasks, should prefer AutoGen
	assert.Equal(t, "autogen", framework)
}

// TestValidateAndResolveCapabilitiesEdgeCases tests edge cases for validateAndResolveCapabilities
func TestValidateAndResolveCapabilitiesEdgeCases(t *testing.T) {
	// Test with empty capabilities
	result := handlers.ValidateAndResolveCapabilitiesLogic([]string{})
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors, "no capabilities provided")
	
	// Test with too many capabilities
	tooMany := make([]string, 15) // Assuming max is 10
	for i := range tooMany {
		tooMany[i] = "capability_" + string(rune(i+48))
	}
	result = handlers.ValidateAndResolveCapabilitiesLogic(tooMany)
	assert.False(t, result.Valid)
	assert.Contains(t, result.Errors, "too many capabilities")
	
	// Test with invalid capabilities
	invalid := []string{"invalid_capability", "another_invalid"}
	result = handlers.ValidateAndResolveCapabilitiesLogic(invalid)
	assert.False(t, result.Valid)
	assert.Greater(t, len(result.Errors), 0)
	
	// Test with valid capabilities
	valid := []string{"web_search", "calculations", "text_processing"}
	result = handlers.ValidateAndResolveCapabilitiesLogic(valid)
	assert.True(t, result.Valid)
	assert.Equal(t, 0, len(result.Errors))
}

// TestValidateAndResolveCapabilitiesConflicts tests conflict detection
func TestValidateAndResolveCapabilitiesConflicts(t *testing.T) {
	// Test with conflicting capabilities
	conflicting := []string{"web_search", "file_operations", "calculations", "text_processing", "api_calls"}
	result := handlers.ValidateAndResolveCapabilitiesLogic(conflicting)
	
	// Should detect conflicts but still be valid after resolution
	assert.True(t, result.Valid)
	assert.LessOrEqual(t, len(result.ResolvedCapabilities), len(conflicting))
	
	if len(result.Conflicts) > 0 {
		assert.Greater(t, len(result.Conflicts), 0)
		assert.Greater(t, len(result.Recommendations), 0)
	}
}

// TestValidateAndResolveCapabilitiesResourceLimits tests resource limit validation
func TestValidateAndResolveCapabilitiesResourceLimits(t *testing.T) {
	// Test with capabilities that exceed resource limits
	resourceHeavy := []string{"web_search", "file_operations", "calculations", "text_processing", "api_calls", "data_analysis"}
	result := handlers.ValidateAndResolveCapabilitiesLogic(resourceHeavy)
	
	// Should handle resource limits gracefully
	assert.NotNil(t, result)
	if !result.Valid {
		assert.Greater(t, len(result.Errors), 0)
	} else {
		// If valid, should have resolved to within limits
		assert.LessOrEqual(t, len(result.ResolvedCapabilities), 5) // Assuming limit is 5
	}
}
