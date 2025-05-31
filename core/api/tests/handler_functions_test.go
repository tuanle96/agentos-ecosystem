package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tuanle96/agentos-ecosystem/core/api/handlers"
)

// TestHandlerLevelFunctions tests handler-level functions directly
func TestHandlerLevelFunctions(t *testing.T) {
	// Test capability validation at handler level
	capabilities := []string{"web_search", "calculations", "text_processing"}
	result := handlers.ValidateAndResolveCapabilitiesLogic(capabilities)
	assert.True(t, result.Valid)
	assert.Equal(t, capabilities, result.ResolvedCapabilities)

	// Test conflict resolution at handler level
	conflictingCaps := []string{"web_search", "file_operations", "calculations", "text_processing", "api_calls"}
	resolved := handlers.ResolveCapabilityConflictsLogic(conflictingCaps, "remove_conflicts")
	assert.NotNil(t, resolved)
	assert.LessOrEqual(t, len(resolved), len(conflictingCaps))

	// Test framework selection at handler level
	framework := handlers.SelectOptimalFrameworkLogic(capabilities, "research", "high")
	assert.NotEmpty(t, framework)
	assert.Contains(t, []string{"langchain", "crewai", "autogen", "swarms"}, framework)
}

// TestCapabilityRecommendationHandler tests capability recommendation logic
func TestCapabilityRecommendationHandler(t *testing.T) {
	// Test different domains
	domains := []string{"research", "data_analysis", "automation", "communication", "general"}
	complexities := []string{"low", "medium", "high"}

	for _, domain := range domains {
		for _, complexity := range complexities {
			recommendations := handlers.GetCapabilityRecommendationsLogic(
				"test task for "+domain, domain, complexity)
			assert.NotNil(t, recommendations)
			assert.GreaterOrEqual(t, len(recommendations), 1)

			// Verify recommendations are valid
			validCaps := []string{"web_search", "calculations", "text_processing",
				"file_operations", "api_calls", "data_analysis"}
			for _, rec := range recommendations {
				assert.Contains(t, validCaps, rec)
			}
		}
	}
}

// TestFrameworkSelectionLogic tests framework selection with various inputs
func TestFrameworkSelectionLogic(t *testing.T) {
	testCases := []struct {
		capabilities []string
		taskType     string
		performance  string
		expected     []string // possible frameworks
	}{
		{
			capabilities: []string{"web_search", "text_processing"},
			taskType:     "research",
			performance:  "high",
			expected:     []string{"langchain", "crewai"},
		},
		{
			capabilities: []string{"calculations", "data_analysis"},
			taskType:     "data_processing",
			performance:  "high",
			expected:     []string{"swarms", "langchain"},
		},
		{
			capabilities: []string{"text_processing", "api_calls"},
			taskType:     "conversation",
			performance:  "medium",
			expected:     []string{"autogen", "langchain"},
		},
	}

	for _, tc := range testCases {
		framework := handlers.SelectOptimalFrameworkLogic(tc.capabilities, tc.taskType, tc.performance)
		assert.NotEmpty(t, framework)
		// Framework should be one of the expected options
		found := false
		for _, expected := range tc.expected {
			if framework == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Framework %s not in expected list %v", framework, tc.expected)
	}
}

// TestConflictResolutionStrategies tests different conflict resolution strategies
func TestConflictResolutionStrategies(t *testing.T) {
	capabilities := []string{"web_search", "file_operations", "calculations", "text_processing", "api_calls"}

	strategies := []string{"remove_conflicts", "optimize_performance", "minimize_resources"}

	for _, strategy := range strategies {
		resolved := handlers.ResolveCapabilityConflictsLogic(capabilities, strategy)
		assert.NotNil(t, resolved)

		switch strategy {
		case "minimize_resources":
			assert.LessOrEqual(t, len(resolved), 3, "minimize_resources should reduce capabilities")
		case "optimize_performance":
			assert.GreaterOrEqual(t, len(resolved), 3, "optimize_performance should keep key capabilities")
		case "remove_conflicts":
			assert.LessOrEqual(t, len(resolved), len(capabilities), "remove_conflicts should not add capabilities")
		}
	}
}
