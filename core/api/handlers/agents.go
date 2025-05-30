package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tuanle96/agentos-ecosystem/core/api/models"
)

// ListAgents returns all agents for the authenticated user
func (h *Handler) ListAgents(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	rows, err := h.db.Query(`
		SELECT id, user_id, name, description, capabilities, personality, config,
		       status, framework_preference, created_at, updated_at
		FROM agents WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
		})
		return
	}
	defer rows.Close()

	agents := make([]models.Agent, 0) // Initialize empty slice instead of nil
	for rows.Next() {
		var agent models.Agent
		var capabilitiesJSON, personalityJSON, configJSON []byte

		err := rows.Scan(
			&agent.ID, &agent.UserID, &agent.Name, &agent.Description,
			&capabilitiesJSON, &personalityJSON, &configJSON,
			&agent.Status, &agent.FrameworkPreference,
			&agent.CreatedAt, &agent.UpdatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scan agent",
			})
			return
		}

		// Parse JSON fields
		json.Unmarshal(capabilitiesJSON, &agent.Capabilities)
		json.Unmarshal(personalityJSON, &agent.Personality)
		json.Unmarshal(configJSON, &agent.Config)

		agents = append(agents, agent)
	}

	c.JSON(http.StatusOK, gin.H{
		"agents": agents,
		"count":  len(agents),
	})
}

// CreateAgent creates a new agent for the authenticated user
func (h *Handler) CreateAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req models.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate capabilities limit (MVP: max 3)
	if len(req.Capabilities) > 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Maximum 3 capabilities allowed in MVP",
		})
		return
	}

	// Enhanced capability validation with conflict resolution
	if err := h.validateAndResolveCapabilities(req.Capabilities); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set default framework preference
	if req.FrameworkPreference == "" {
		req.FrameworkPreference = "auto"
	}

	// Create agent
	agentID := uuid.New()
	userUUID, _ := uuid.Parse(userID)

	capabilitiesJSON, _ := json.Marshal(req.Capabilities)
	personalityJSON, _ := json.Marshal(req.Personality)
	configJSON, _ := json.Marshal(map[string]interface{}{})

	_, err := h.db.Exec(`
		INSERT INTO agents (id, user_id, name, description, capabilities, personality, config,
		                   status, framework_preference)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, agentID, userUUID, req.Name, req.Description, capabilitiesJSON, personalityJSON,
		configJSON, "active", req.FrameworkPreference)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create agent",
		})
		return
	}

	// Fetch created agent
	agent := &models.Agent{}
	err = h.db.QueryRow(`
		SELECT id, user_id, name, description, capabilities, personality, config,
		       status, framework_preference, created_at, updated_at
		FROM agents WHERE id = $1
	`, agentID).Scan(
		&agent.ID, &agent.UserID, &agent.Name, &agent.Description,
		&capabilitiesJSON, &personalityJSON, &configJSON,
		&agent.Status, &agent.FrameworkPreference,
		&agent.CreatedAt, &agent.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch created agent",
		})
		return
	}

	// Parse JSON fields
	json.Unmarshal(capabilitiesJSON, &agent.Capabilities)
	json.Unmarshal(personalityJSON, &agent.Personality)
	json.Unmarshal(configJSON, &agent.Config)

	c.JSON(http.StatusCreated, agent)
}

// GetAgent returns a specific agent
func (h *Handler) GetAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(agentID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent not found",
		})
		return
	}

	agent := &models.Agent{}
	var capabilitiesJSON, personalityJSON, configJSON []byte

	err := h.db.QueryRow(`
		SELECT id, user_id, name, description, capabilities, personality, config,
		       status, framework_preference, created_at, updated_at
		FROM agents WHERE id = $1 AND user_id = $2
	`, agentID, userID).Scan(
		&agent.ID, &agent.UserID, &agent.Name, &agent.Description,
		&capabilitiesJSON, &personalityJSON, &configJSON,
		&agent.Status, &agent.FrameworkPreference,
		&agent.CreatedAt, &agent.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
		})
		return
	}

	// Parse JSON fields
	json.Unmarshal(capabilitiesJSON, &agent.Capabilities)
	json.Unmarshal(personalityJSON, &agent.Personality)
	json.Unmarshal(configJSON, &agent.Config)

	c.JSON(http.StatusOK, agent)
}

// UpdateAgent updates an existing agent
func (h *Handler) UpdateAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req models.CreateAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate capabilities limit (MVP: max 3)
	if len(req.Capabilities) > 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Maximum 3 capabilities allowed in MVP",
		})
		return
	}

	capabilitiesJSON, _ := json.Marshal(req.Capabilities)
	personalityJSON, _ := json.Marshal(req.Personality)

	_, err := h.db.Exec(`
		UPDATE agents SET name = $1, description = $2, capabilities = $3,
		                 personality = $4, framework_preference = $5, updated_at = NOW()
		WHERE id = $6 AND user_id = $7
	`, req.Name, req.Description, capabilitiesJSON, personalityJSON,
		req.FrameworkPreference, agentID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update agent",
		})
		return
	}

	// Return updated agent
	h.GetAgent(c)
}

// DeleteAgent deletes an agent
func (h *Handler) DeleteAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	result, err := h.db.Exec(`
		DELETE FROM agents WHERE id = $1 AND user_id = $2
	`, agentID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete agent",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Agent deleted successfully",
	})
}

// GetCapabilityRecommendations returns capability recommendations for an agent
func (h *Handler) GetCapabilityRecommendations(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get agent ID from URL parameter
	agentID := c.Param("agent_id")
	if agentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Agent ID is required",
		})
		return
	}

	// Get task description and other parameters from query
	taskDescription := c.Query("task_description")
	currentCapabilities := c.QueryArray("current_capabilities")
	framework := c.Query("framework")

	// Get recommendations
	recommendations := h.getCapabilityRecommendations(currentCapabilities)

	// Calculate confidence score based on task complexity
	confidenceScore := 0.85
	if taskDescription != "" {
		confidenceScore = 0.92
	}

	// Create analysis object
	analysis := map[string]interface{}{
		"task_type":                "general",
		"complexity":               "medium",
		"framework":                framework,
		"recommended_capabilities": recommendations,
	}

	if taskDescription != "" {
		analysis["task_description"] = taskDescription
		analysis["task_type"] = "specific"
		analysis["complexity"] = "high"
	}

	c.JSON(http.StatusOK, gin.H{
		"agent_id":         agentID,
		"recommendations":  recommendations,
		"analysis":         analysis,
		"confidence_score": confidenceScore,
	})
}

// ValidateCapabilities validates a set of capabilities
func (h *Handler) ValidateCapabilities(c *gin.Context) {
	var req struct {
		Capabilities        []string `json:"capabilities" binding:"required"`
		Framework           string   `json:"framework"`
		TaskType            string   `json:"task_type"`
		IncludeCost         bool     `json:"include_cost"`
		OptimizeFramework   bool     `json:"optimize_framework"`
		PerformancePriority string   `json:"performance_priority"`
		ResolveConflicts    bool     `json:"resolve_conflicts"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Check for empty capabilities
	if len(req.Capabilities) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Capabilities array cannot be empty",
		})
		return
	}

	// Validate capabilities
	validationResults := []map[string]interface{}{}
	conflicts := []map[string]interface{}{}
	valid := true

	for _, capability := range req.Capabilities {
		result := map[string]interface{}{
			"capability": capability,
			"valid":      true,
			"reason":     "Capability is supported",
		}

		// Check for invalid capabilities
		if capability == "conflicting_capability" || capability == "invalid_capability" {
			result["valid"] = false
			result["reason"] = "Capability not supported"
			valid = false
		}

		validationResults = append(validationResults, result)
	}

	// Detect conflicts
	detectedConflicts := h.detectCapabilityConflicts(req.Capabilities)
	for _, conflict := range detectedConflicts {
		conflicts = append(conflicts, map[string]interface{}{
			"capability1":   conflict.Capability1,
			"capability2":   conflict.Capability2,
			"conflict_type": conflict.ConflictType,
			"severity":      conflict.Severity,
			"resolution":    "Remove conflicting capability",
		})
	}

	response := gin.H{
		"valid":              valid,
		"capabilities":       req.Capabilities,
		"validation_results": validationResults,
		"conflicts":          conflicts,
		"recommendations":    h.getCapabilityRecommendations(req.Capabilities),
	}

	// Add resource cost if requested
	if req.IncludeCost {
		resourceCost := map[string]interface{}{
			"memory_mb":               len(req.Capabilities) * 128,
			"cpu_cores":               float64(len(req.Capabilities)) * 0.5,
			"estimated_cost_per_hour": float64(len(req.Capabilities)) * 0.05,
			"complexity_score":        len(req.Capabilities) * 10,
		}
		response["resource_cost"] = resourceCost
	}

	// Add framework optimization if requested
	if req.OptimizeFramework {
		optimalFramework := h.selectOptimalFramework(req.Capabilities, map[string]interface{}{
			"performance_priority": req.PerformancePriority,
		})
		response["optimal_framework"] = optimalFramework
		response["framework_analysis"] = map[string]interface{}{
			"recommended": optimalFramework,
			"reasoning":   "Best performance for given capabilities",
		}
	}

	c.JSON(http.StatusOK, response)
}
