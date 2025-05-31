package handlers

import (
	"fmt"
	"strings"
)

// CapabilityConflict represents a conflict between capabilities
type CapabilityConflict struct {
	Capability1  string
	Capability2  string
	Reason       string
	ConflictType string
	Severity     string
}

// CapabilityMetadata contains metadata about each capability
type CapabilityMetadata struct {
	Name         string
	Category     string
	Dependencies []string
	Conflicts    []string
	Description  string
	ResourceCost int
}

// Agent Factory Enhancement - Week 2 Implementation
var capabilityRegistry = map[string]CapabilityMetadata{
	"web_search": {
		Name:         "web_search",
		Category:     "search",
		Dependencies: []string{},
		Conflicts:    []string{},
		Description:  "DuckDuckGo search integration",
		ResourceCost: 2,
	},
	"file_operations": {
		Name:         "file_operations",
		Category:     "file",
		Dependencies: []string{},
		Conflicts:    []string{"api_calls"}, // Potential security conflict
		Description:  "Safe file read/write operations",
		ResourceCost: 3,
	},
	"api_calls": {
		Name:         "api_calls",
		Category:     "network",
		Dependencies: []string{},
		Conflicts:    []string{"file_operations"}, // Security isolation
		Description:  "HTTP API call functionality",
		ResourceCost: 2,
	},
	"calculations": {
		Name:         "calculations",
		Category:     "math",
		Dependencies: []string{},
		Conflicts:    []string{},
		Description:  "Mathematical calculations",
		ResourceCost: 1,
	},
	"text_processing": {
		Name:         "text_processing",
		Category:     "text",
		Dependencies: []string{},
		Conflicts:    []string{},
		Description:  "Text analysis and processing",
		ResourceCost: 2,
	},
}

// validateAndResolveCapabilities validates capabilities and resolves conflicts
func (h *Handler) validateAndResolveCapabilities(capabilities []string) error {
	// Check if all capabilities exist
	for _, cap := range capabilities {
		if _, exists := capabilityRegistry[cap]; !exists {
			return fmt.Errorf("invalid capability: %s", cap)
		}
	}

	// Check for conflicts
	conflicts := h.detectCapabilityConflicts(capabilities)
	if len(conflicts) > 0 {
		return h.resolveCapabilityConflicts(conflicts, capabilities)
	}

	// Check resource limits (MVP: max total cost of 6)
	totalCost := h.calculateResourceCost(capabilities)
	if totalCost > 6 {
		return fmt.Errorf("capability resource cost too high: %d (max: 6)", totalCost)
	}

	return nil
}

// detectCapabilityConflicts detects conflicts between capabilities
func (h *Handler) detectCapabilityConflicts(capabilities []string) []CapabilityConflict {
	var conflicts []CapabilityConflict

	for i, cap1 := range capabilities {
		for j, cap2 := range capabilities {
			if i >= j {
				continue
			}

			meta1 := capabilityRegistry[cap1]
			meta2 := capabilityRegistry[cap2]

			// Check if cap1 conflicts with cap2
			for _, conflict := range meta1.Conflicts {
				if conflict == cap2 {
					conflicts = append(conflicts, CapabilityConflict{
						Capability1:  cap1,
						Capability2:  cap2,
						Reason:       fmt.Sprintf("%s conflicts with %s", meta1.Description, meta2.Description),
						ConflictType: "security",
						Severity:     "high",
					})
				}
			}
		}
	}

	return conflicts
}

// resolveCapabilityConflicts attempts to resolve conflicts
func (h *Handler) resolveCapabilityConflicts(conflicts []CapabilityConflict, capabilities []string) error {
	if len(conflicts) == 0 {
		return nil
	}

	// For MVP, we don't auto-resolve conflicts - just report them
	var conflictMessages []string
	for _, conflict := range conflicts {
		conflictMessages = append(conflictMessages,
			fmt.Sprintf("'%s' conflicts with '%s': %s",
				conflict.Capability1, conflict.Capability2, conflict.Reason))
	}

	return fmt.Errorf("capability conflicts detected: %s", strings.Join(conflictMessages, "; "))
}

// calculateResourceCost calculates total resource cost
func (h *Handler) calculateResourceCost(capabilities []string) int {
	totalCost := 0
	for _, cap := range capabilities {
		if meta, exists := capabilityRegistry[cap]; exists {
			totalCost += meta.ResourceCost
		}
	}
	return totalCost
}

// getCapabilityRecommendations suggests compatible capabilities
func (h *Handler) getCapabilityRecommendations(existingCapabilities []string) []string {
	var recommendations []string

	// Calculate current resource usage
	currentCost := h.calculateResourceCost(existingCapabilities)
	remainingBudget := 6 - currentCost

	// Find compatible capabilities within budget
	for name, meta := range capabilityRegistry {
		// Skip if already selected
		alreadySelected := false
		for _, existing := range existingCapabilities {
			if existing == name {
				alreadySelected = true
				break
			}
		}
		if alreadySelected {
			continue
		}

		// Check if within budget
		if meta.ResourceCost > remainingBudget {
			continue
		}

		// Check for conflicts
		hasConflict := false
		for _, existing := range existingCapabilities {
			existingMeta := capabilityRegistry[existing]
			for _, conflict := range existingMeta.Conflicts {
				if conflict == name {
					hasConflict = true
					break
				}
			}
			if hasConflict {
				break
			}
		}

		if !hasConflict {
			recommendations = append(recommendations, name)
		}
	}

	return recommendations
}

// Framework Selection Logic
type FrameworkSelector struct {
	capabilities []string
	preferences  map[string]interface{}
}

// selectOptimalFramework chooses the best framework for given capabilities
func (h *Handler) selectOptimalFramework(capabilities []string, preferences map[string]interface{}) string {
	// Framework scoring based on capabilities
	scores := map[string]int{
		"langchain": 0,
		"swarms":    0,
		"crewai":    0,
		"autogen":   0,
	}

	// Score based on capability compatibility
	for _, cap := range capabilities {
		switch cap {
		case "web_search", "api_calls":
			scores["langchain"] += 3 // LangChain excels at tool integration
			scores["swarms"] += 2
		case "text_processing":
			scores["langchain"] += 3
			scores["crewai"] += 2
		case "calculations":
			scores["autogen"] += 3 // AutoGen good for structured tasks
			scores["langchain"] += 2
		case "file_operations":
			scores["swarms"] += 3 // Swarms good for file handling
			scores["langchain"] += 2
		}
	}

	// Find framework with highest score
	bestFramework := "langchain" // Default
	bestScore := scores["langchain"]

	for framework, score := range scores {
		if score > bestScore {
			bestFramework = framework
			bestScore = score
		}
	}

	return bestFramework
}
