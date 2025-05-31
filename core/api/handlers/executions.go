package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
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
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Get agent information including framework preference
	var agent struct {
		ID                  string
		Name                string
		FrameworkPreference string
		Status              string
	}

	err := h.db.QueryRow(`
		SELECT id, name, framework_preference, status
		FROM agents
		WHERE id = $1 AND user_id = $2 AND status = 'active'
	`, agentID, userID).Scan(&agent.ID, &agent.Name, &agent.FrameworkPreference, &agent.Status)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent not found or not accessible",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve agent information",
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

	// Execute agent using real Python AI Worker
	outputText, frameworkUsed, err := h.executeAgentReal(agentID, req.InputText, agent.FrameworkPreference)
	if err != nil {
		// Update execution record with error
		_, updateErr := h.db.Exec(`
			UPDATE executions
			SET status = $1, error_message = $2, completed_at = $3
			WHERE id = $4
		`, "failed", err.Error(), time.Now(), executionID)

		if updateErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update execution record: " + updateErr.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to execute agent: " + err.Error(),
		})
		return
	}
	executionTime := int(time.Since(startTime).Milliseconds())

	// Update execution record
	toolsUsedJSON, _ := json.Marshal([]string{})
	metadataJSON, _ := json.Marshal(map[string]interface{}{
		"real_execution": true,
		"ai_worker_used": true,
		"agent_name":     agent.Name,
	})

	completedAt := time.Now()
	_, err = h.db.Exec(`
		UPDATE executions
		SET output_text = $1, framework_used = $2, tools_used = $3,
		    execution_time_ms = $4, status = $5, metadata = $6, completed_at = $7
		WHERE id = $8
	`, outputText, frameworkUsed, toolsUsedJSON, executionTime, "completed", metadataJSON, completedAt, executionID)

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
		FrameworkUsed:   frameworkUsed,
		Status:          "completed",
		CreatedAt:       startTime,
	}

	c.JSON(http.StatusOK, response)
}

// executeAgentReal executes agent using real Python AI Worker
func (h *Handler) executeAgentReal(agentID, input, framework string) (string, string, error) {
	// Prepare request to Python AI Worker
	requestBody := map[string]interface{}{
		"input":     input,
		"framework": framework,
		"agent_id":  agentID,
		"timeout":   30,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", framework, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 35 * time.Second}

	// Call Python AI Worker (assuming it runs on localhost:8080 in Docker)
	aiWorkerURL := os.Getenv("AI_WORKER_URL")
	if aiWorkerURL == "" {
		aiWorkerURL = "http://localhost:8080" // Default for development
	}

	resp, err := client.Post(aiWorkerURL+"/api/execute",
		"application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", framework, fmt.Errorf("failed to call AI worker: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", framework, fmt.Errorf("AI worker returned status %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		Output        string  `json:"output"`
		FrameworkUsed string  `json:"framework_used"`
		Status        string  `json:"status"`
		Error         string  `json:"error,omitempty"`
		ExecutionTime float64 `json:"execution_time"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", framework, fmt.Errorf("failed to decode AI worker response: %v", err)
	}

	if result.Status == "failed" {
		return "", framework, fmt.Errorf("AI worker execution failed: %s", result.Error)
	}

	return result.Output, result.FrameworkUsed, nil
}

// simulateAgentExecution provides a simple simulation for MVP (DEPRECATED - Use executeAgentReal)
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

// GetAgentMemory returns memory for an agent with real database integration
func (h *Handler) GetAgentMemory(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Validate agent ownership
	if !h.validateAgentOwnership(agentID, userID) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent not found",
		})
		return
	}

	// Get real memory data from database
	memory, err := h.getAgentMemoryFromDB(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve agent memory",
		})
		return
	}

	c.JSON(http.StatusOK, memory)
}

// ClearAgentMemory clears memory for an agent with real database operations
func (h *Handler) ClearAgentMemory(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Validate agent ownership
	if !h.validateAgentOwnership(agentID, userID) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent not found",
		})
		return
	}

	// Clear real memory data from database
	err := h.clearAgentMemoryFromDB(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear agent memory",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Agent memory cleared successfully",
		"agent_id":   agentID,
		"cleared_at": time.Now(),
	})
}

// getAgentMemoryFromDB retrieves agent memory from database
func (h *Handler) getAgentMemoryFromDB(agentID string) (map[string]interface{}, error) {
	// Query working memory (recent conversations)
	workingMemory, err := h.getWorkingMemory(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get working memory: %v", err)
	}

	// Query episodic memory (stored memories)
	episodicMemory, err := h.getEpisodicMemory(agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodic memory: %v", err)
	}

	// Calculate memory statistics
	stats := map[string]interface{}{
		"total_memories": len(workingMemory) + len(episodicMemory),
		"working_size":   len(workingMemory),
		"episodic_size":  len(episodicMemory),
		"last_updated":   time.Now(),
	}

	return map[string]interface{}{
		"agent_id":        agentID,
		"working_memory":  workingMemory,
		"episodic_memory": episodicMemory,
		"memory_stats":    stats,
	}, nil
}

// getWorkingMemory retrieves recent conversations for working memory
func (h *Handler) getWorkingMemory(agentID string) ([]map[string]interface{}, error) {
	rows, err := h.db.Query(`
		SELECT id, input_text, output_text, started_at, completed_at
		FROM executions
		WHERE agent_id = $1 AND status = 'completed'
		ORDER BY started_at DESC
		LIMIT 10
	`, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []map[string]interface{}
	for rows.Next() {
		var id, input, output string
		var startedAt, completedAt time.Time

		err := rows.Scan(&id, &input, &output, &startedAt, &completedAt)
		if err != nil {
			continue
		}

		memories = append(memories, map[string]interface{}{
			"id":           id,
			"type":         "conversation",
			"input":        input,
			"output":       output,
			"started_at":   startedAt,
			"completed_at": completedAt,
		})
	}

	return memories, nil
}

// getEpisodicMemory retrieves stored episodic memories
func (h *Handler) getEpisodicMemory(agentID string) ([]map[string]interface{}, error) {
	// For now, return empty array as episodic memory system is not fully implemented
	// This would integrate with the memory consolidation system in the future
	return []map[string]interface{}{}, nil
}

// clearAgentMemoryFromDB clears agent memory from database
func (h *Handler) clearAgentMemoryFromDB(agentID string) error {
	// Clear execution history (working memory)
	_, err := h.db.Exec(`
		DELETE FROM executions
		WHERE agent_id = $1
	`, agentID)
	if err != nil {
		return fmt.Errorf("failed to clear execution history: %v", err)
	}

	// Clear tool executions
	_, err = h.db.Exec(`
		DELETE FROM tool_executions
		WHERE user_id IN (
			SELECT user_id FROM agents WHERE id = $1
		)
	`, agentID)
	if err != nil {
		return fmt.Errorf("failed to clear tool executions: %v", err)
	}

	// In the future, this would also clear episodic memory from memory system
	// For now, we've cleared the main conversation history

	return nil
}
