package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tuanle96/agentos-ecosystem/core/api/models"
)

// ExecuteAgent executes an agent with given input
func (h *Handler) ExecuteAgent(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req models.ExecuteAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Verify agent exists and belongs to user
	var agentExists bool
	err := h.db.QueryRow(`
		SELECT EXISTS(SELECT 1 FROM agents WHERE id = $1 AND user_id = $2 AND status = 'active')
	`, agentID, userID).Scan(&agentExists)

	if err != nil || !agentExists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent not found or not accessible",
		})
		return
	}

	// Create execution record
	executionID := uuid.New()
	userUUID, _ := uuid.Parse(userID)
	agentUUID, _ := uuid.Parse(agentID)

	startTime := time.Now()

	_, err = h.db.Exec(`
		INSERT INTO executions (id, agent_id, user_id, input_text, status, started_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, executionID, agentUUID, userUUID, req.InputText, "pending", startTime)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create execution record",
		})
		return
	}

	// For MVP, we'll simulate execution with a simple response
	// In Week 3-4, this will be replaced with actual AI framework integration
	outputText := h.simulateAgentExecution(req.InputText)
	executionTime := int(time.Since(startTime).Milliseconds())

	// Update execution record
	toolsUsedJSON, _ := json.Marshal([]string{})
	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"simulated": true,
		"mvp_mode": true,
	})

	completedAt := time.Now()
	_, err = h.db.Exec(`
		UPDATE executions 
		SET output_text = $1, framework_used = $2, tools_used = $3, 
		    execution_time_ms = $4, status = $5, metadata = $6, completed_at = $7
		WHERE id = $8
	`, outputText, "simulated", toolsUsedJSON, executionTime, "completed", metadataJSON, completedAt, executionID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update execution record",
		})
		return
	}

	response := models.ExecuteAgentResponse{
		ExecutionID:     executionID.String(),
		OutputText:      outputText,
		ToolsUsed:       []string{},
		ExecutionTimeMs: executionTime,
		FrameworkUsed:   "simulated",
		Status:          "completed",
		CreatedAt:       startTime,
	}

	c.JSON(http.StatusOK, response)
}

// simulateAgentExecution provides a simple simulation for MVP
func (h *Handler) simulateAgentExecution(input string) string {
	return "Hello! I'm an AgentOS MVP agent. You said: \"" + input + "\". " +
		"I'm currently in simulation mode. In Week 3-4, I'll be integrated with " +
		"LangChain and Swarms frameworks to provide real AI capabilities!"
}

// GetAgentExecutions returns execution history for an agent
func (h *Handler) GetAgentExecutions(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	rows, err := h.db.Query(`
		SELECT id, agent_id, user_id, input_text, output_text, framework_used,
		       tools_used, execution_time_ms, status, error_message, metadata,
		       started_at, completed_at
		FROM executions 
		WHERE agent_id = $1 AND user_id = $2 
		ORDER BY started_at DESC
		LIMIT 50
	`, agentID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
		})
		return
	}
	defer rows.Close()

	var executions []models.Execution
	for rows.Next() {
		var execution models.Execution
		var toolsUsedJSON, metadataJSON []byte

		err := rows.Scan(
			&execution.ID, &execution.AgentID, &execution.UserID,
			&execution.InputText, &execution.OutputText, &execution.FrameworkUsed,
			&toolsUsedJSON, &execution.ExecutionTimeMs, &execution.Status,
			&execution.ErrorMessage, &metadataJSON,
			&execution.StartedAt, &execution.CompletedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scan execution",
			})
			return
		}

		// Parse JSON fields
		json.Unmarshal(toolsUsedJSON, &execution.ToolsUsed)
		json.Unmarshal(metadataJSON, &execution.Metadata)

		executions = append(executions, execution)
	}

	c.JSON(http.StatusOK, gin.H{
		"executions": executions,
		"count":      len(executions),
	})
}

// GetExecution returns a specific execution
func (h *Handler) GetExecution(c *gin.Context) {
	userID := c.GetString("user_id")
	executionID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	execution := &models.Execution{}
	var toolsUsedJSON, metadataJSON []byte

	err := h.db.QueryRow(`
		SELECT id, agent_id, user_id, input_text, output_text, framework_used,
		       tools_used, execution_time_ms, status, error_message, metadata,
		       started_at, completed_at
		FROM executions 
		WHERE id = $1 AND user_id = $2
	`, executionID, userID).Scan(
		&execution.ID, &execution.AgentID, &execution.UserID,
		&execution.InputText, &execution.OutputText, &execution.FrameworkUsed,
		&toolsUsedJSON, &execution.ExecutionTimeMs, &execution.Status,
		&execution.ErrorMessage, &metadataJSON,
		&execution.StartedAt, &execution.CompletedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Execution not found",
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
	json.Unmarshal(toolsUsedJSON, &execution.ToolsUsed)
	json.Unmarshal(metadataJSON, &execution.Metadata)

	c.JSON(http.StatusOK, execution)
}

// GetExecutionLogs returns logs for a specific execution
func (h *Handler) GetExecutionLogs(c *gin.Context) {
	userID := c.GetString("user_id")
	executionID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// For MVP, return simple logs
	// In later weeks, this will integrate with actual logging system
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Add(-time.Second * 5),
			"level":     "info",
			"message":   "Execution started",
		},
		{
			"timestamp": time.Now().Add(-time.Second * 3),
			"level":     "info",
			"message":   "Processing input with simulated agent",
		},
		{
			"timestamp": time.Now().Add(-time.Second * 1),
			"level":     "info",
			"message":   "Execution completed successfully",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"execution_id": executionID,
		"logs":         logs,
		"count":        len(logs),
	})
}

// GetAgentMemory returns memory for an agent (placeholder for MVP)
func (h *Handler) GetAgentMemory(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// For MVP, return empty memory
	// In Week 5-6, this will integrate with actual memory system
	c.JSON(http.StatusOK, gin.H{
		"agent_id":       agentID,
		"working_memory": []interface{}{},
		"episodic_memory": []interface{}{},
		"memory_stats": map[string]interface{}{
			"total_memories": 0,
			"working_size":   0,
			"episodic_size":  0,
		},
	})
}

// ClearAgentMemory clears memory for an agent (placeholder for MVP)
func (h *Handler) ClearAgentMemory(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// For MVP, just return success
	// In Week 5-6, this will actually clear memory
	c.JSON(http.StatusOK, gin.H{
		"message":  "Agent memory cleared successfully",
		"agent_id": agentID,
	})
}
