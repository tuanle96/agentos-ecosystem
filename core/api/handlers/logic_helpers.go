package handlers

import (
	"strings"
)

// ValidationResult represents the result of capability validation
type ValidationResult struct {
	Valid                 bool     `json:"valid"`
	ResolvedCapabilities []string `json:"resolved_capabilities"`
	Conflicts            []string `json:"conflicts"`
	Errors               []string `json:"errors"`
	Recommendations      []string `json:"recommendations"`
}

// ResolveCapabilityConflictsLogic resolves conflicts between capabilities
func ResolveCapabilityConflictsLogic(capabilities []string, strategy string) []string {
	if len(capabilities) == 0 {
		return []string{}
	}

	// Define conflicting capability groups
	conflictGroups := map[string][]string{
		"io_operations": {"web_search", "file_operations", "api_calls"},
		"processing":    {"text_processing", "data_analysis", "calculations"},
	}

	resolved := make([]string, 0, len(capabilities))
	used := make(map[string]bool)

	switch strategy {
	case "remove_conflicts":
		// Remove conflicting capabilities, keep first in each group
		for _, cap := range capabilities {
			conflict := false
			for _, group := range conflictGroups {
				if contains(group, cap) {
					// Check if we already have a capability from this group
					for _, existing := range resolved {
						if contains(group, existing) {
							conflict = true
							break
						}
					}
					break
				}
			}
			if !conflict && !used[cap] {
				resolved = append(resolved, cap)
				used[cap] = true
			}
		}

	case "optimize_performance":
		// Prioritize high-performance capabilities
		priority := []string{"calculations", "web_search", "text_processing", "api_calls", "file_operations"}
		for _, cap := range priority {
			if contains(capabilities, cap) && !used[cap] {
				resolved = append(resolved, cap)
				used[cap] = true
			}
		}
		// Add remaining capabilities
		for _, cap := range capabilities {
			if !used[cap] {
				resolved = append(resolved, cap)
				used[cap] = true
			}
		}

	case "minimize_resources":
		// Keep only essential capabilities
		essential := []string{"calculations", "text_processing"}
		for _, cap := range essential {
			if contains(capabilities, cap) && !used[cap] {
				resolved = append(resolved, cap)
				used[cap] = true
			}
		}

	default:
		// Default: return original capabilities
		return capabilities
	}

	return resolved
}

// GetCapabilityRecommendationsLogic generates capability recommendations based on task description
func GetCapabilityRecommendationsLogic(taskDescription, domain, complexity string) []string {
	recommendations := []string{}

	// Convert to lowercase for easier matching
	desc := strings.ToLower(taskDescription)
	domain = strings.ToLower(domain)

	// Base recommendations by domain
	switch domain {
	case "research":
		recommendations = append(recommendations, "web_search", "text_processing")
	case "data_analysis":
		recommendations = append(recommendations, "calculations", "data_analysis", "file_operations")
	case "automation":
		recommendations = append(recommendations, "api_calls", "file_operations")
	case "communication":
		recommendations = append(recommendations, "text_processing", "api_calls")
	default:
		recommendations = append(recommendations, "text_processing")
	}

	// Add recommendations based on task description keywords
	if strings.Contains(desc, "search") || strings.Contains(desc, "web") {
		if !contains(recommendations, "web_search") {
			recommendations = append(recommendations, "web_search")
		}
	}

	if strings.Contains(desc, "calculate") || strings.Contains(desc, "math") || strings.Contains(desc, "number") {
		if !contains(recommendations, "calculations") {
			recommendations = append(recommendations, "calculations")
		}
	}

	if strings.Contains(desc, "file") || strings.Contains(desc, "document") || strings.Contains(desc, "data") {
		if !contains(recommendations, "file_operations") {
			recommendations = append(recommendations, "file_operations")
		}
	}

	if strings.Contains(desc, "api") || strings.Contains(desc, "service") || strings.Contains(desc, "call") {
		if !contains(recommendations, "api_calls") {
			recommendations = append(recommendations, "api_calls")
		}
	}

	// Adjust based on complexity
	switch complexity {
	case "low":
		// Keep only 1-2 capabilities for simple tasks
		if len(recommendations) > 2 {
			recommendations = recommendations[:2]
		}
	case "high":
		// Add more capabilities for complex tasks
		if !contains(recommendations, "data_analysis") {
			recommendations = append(recommendations, "data_analysis")
		}
	}

	return recommendations
}

// SelectOptimalFrameworkLogic selects the best framework based on capabilities and requirements
func SelectOptimalFrameworkLogic(capabilities []string, taskType, performanceReq string) string {
	// Framework scoring based on capabilities and requirements
	scores := map[string]int{
		"langchain": 0,
		"crewai":    0,
		"autogen":   0,
		"swarms":    0,
	}

	// Score based on capabilities
	for _, cap := range capabilities {
		switch cap {
		case "web_search":
			scores["langchain"] += 3
			scores["crewai"] += 2
		case "text_processing":
			scores["langchain"] += 3
			scores["autogen"] += 3
			scores["crewai"] += 2
		case "calculations":
			scores["swarms"] += 3
			scores["langchain"] += 2
		case "api_calls":
			scores["langchain"] += 2
			scores["autogen"] += 2
			scores["crewai"] += 3
		case "file_operations":
			scores["swarms"] += 3
			scores["langchain"] += 2
		case "data_analysis":
			scores["swarms"] += 4
			scores["langchain"] += 2
		}
	}

	// Score based on task type
	switch strings.ToLower(taskType) {
	case "research":
		scores["langchain"] += 4
		scores["crewai"] += 2
	case "collaboration":
		scores["crewai"] += 4
		scores["autogen"] += 3
	case "conversation":
		scores["autogen"] += 4
		scores["crewai"] += 2
	case "data_processing":
		scores["swarms"] += 4
		scores["langchain"] += 1
	}

	// Score based on performance requirements
	switch strings.ToLower(performanceReq) {
	case "high":
		scores["swarms"] += 2
		scores["crewai"] += 1
	case "medium":
		scores["langchain"] += 2
		scores["autogen"] += 1
	}

	// Find framework with highest score
	maxScore := 0
	selectedFramework := "langchain" // default

	for framework, score := range scores {
		if score > maxScore {
			maxScore = score
			selectedFramework = framework
		}
	}

	return selectedFramework
}

// ValidateAndResolveCapabilitiesLogic validates and resolves capability conflicts
func ValidateAndResolveCapabilitiesLogic(capabilities []string) ValidationResult {
	result := ValidationResult{
		Valid:                 true,
		ResolvedCapabilities: make([]string, 0),
		Conflicts:            make([]string, 0),
		Errors:               make([]string, 0),
		Recommendations:      make([]string, 0),
	}

	// Check if capabilities are provided
	if len(capabilities) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "no capabilities provided")
		return result
	}

	// Check maximum capabilities limit
	maxCapabilities := 10
	if len(capabilities) > maxCapabilities {
		result.Valid = false
		result.Errors = append(result.Errors, "too many capabilities")
		return result
	}

	// Valid capabilities list
	validCapabilities := []string{
		"web_search", "calculations", "text_processing", "file_operations",
		"api_calls", "data_analysis", "image_processing", "audio_processing",
	}

	// Validate each capability
	validCaps := make([]string, 0)
	for _, cap := range capabilities {
		if contains(validCapabilities, cap) {
			validCaps = append(validCaps, cap)
		} else {
			result.Errors = append(result.Errors, "invalid capability: "+cap)
		}
	}

	// If we have invalid capabilities, mark as invalid
	if len(result.Errors) > 0 {
		result.Valid = false
		return result
	}

	// Check for conflicts
	conflicts := detectConflicts(validCaps)
	if len(conflicts) > 0 {
		result.Conflicts = conflicts
		// Resolve conflicts
		result.ResolvedCapabilities = ResolveCapabilityConflictsLogic(validCaps, "remove_conflicts")
		// Add recommendations
		result.Recommendations = []string{"Consider reducing capability overlap", "Review resource requirements"}
	} else {
		result.ResolvedCapabilities = validCaps
	}

	return result
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func detectConflicts(capabilities []string) []string {
	conflicts := []string{}

	// Define conflicting pairs
	conflictPairs := map[string][]string{
		"web_search":     {"file_operations"},
		"file_operations": {"api_calls"},
	}

	for _, cap := range capabilities {
		if conflictingCaps, exists := conflictPairs[cap]; exists {
			for _, conflictCap := range conflictingCaps {
				if contains(capabilities, conflictCap) {
					conflict := cap + " conflicts with " + conflictCap
					if !contains(conflicts, conflict) {
						conflicts = append(conflicts, conflict)
					}
				}
			}
		}
	}

	return conflicts
}
