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

	// Get existing capabilities from query params
	existingCaps := c.QueryArray("capabilities")

	// Get recommendations
	recommendations := h.getCapabilityRecommendations(existingCaps)

	c.JSON(http.StatusOK, gin.H{
		"existing_capabilities": existingCaps,
		"recommendations":       recommendations,
		"resource_usage":        h.calculateResourceCost(existingCaps),
		"resource_limit":        6,
	})
}

// ValidateCapabilities validates a set of capabilities
func (h *Handler) ValidateCapabilities(c *gin.Context) {
	var req struct {
		Capabilities []string `json:"capabilities" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate capabilities
	if err := h.validateAndResolveCapabilities(req.Capabilities); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"valid":        false,
			"error":        err.Error(),
			"capabilities": req.Capabilities,
		})
		return
	}

	// Get optimal framework
	optimalFramework := h.selectOptimalFramework(req.Capabilities, map[string]interface{}{})

	c.JSON(http.StatusOK, gin.H{
		"valid":             true,
		"capabilities":      req.Capabilities,
		"resource_cost":     h.calculateResourceCost(req.Capabilities),
		"optimal_framework": optimalFramework,
		"recommendations":   h.getCapabilityRecommendations(req.Capabilities),
	})
}
