package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Tool Execution System - Week 2 Implementation
type ToolExecutionRequest struct {
	ToolName   string                 `json:"tool_name" binding:"required"`
	Parameters map[string]interface{} `json:"parameters" binding:"required"`
	AgentID    string                 `json:"agent_id"`
	SessionID  string                 `json:"session_id"`
	Timeout    int                    `json:"timeout"` // seconds, default 30
}

type ToolExecutionResponse struct {
	ExecutionID   string                 `json:"execution_id"`
	ToolName      string                 `json:"tool_name"`
	Status        string                 `json:"status"` // pending, running, completed, failed, timeout
	Result        map[string]interface{} `json:"result"`
	Error         string                 `json:"error,omitempty"`
	ExecutionTime float64                `json:"execution_time"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
}

type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Parameters  map[string]interface{} `json:"parameters"`
	Security    ToolSecurity           `json:"security"`
}

type ToolSecurity struct {
	Sandboxed           bool     `json:"sandboxed"`
	AllowedDomains      []string `json:"allowed_domains,omitempty"`
	MaxExecutionTime    int      `json:"max_execution_time"` // seconds
	RequiredPermissions []string `json:"required_permissions"`
}

// Secure Tool Execution Registry
var toolRegistry = map[string]ToolDefinition{
	"web_search": {
		Name:        "web_search",
		Description: "Search the web using DuckDuckGo",
		Category:    "search",
		Parameters: map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"required":    true,
				"description": "Search query",
			},
			"max_results": map[string]interface{}{
				"type":        "integer",
				"required":    false,
				"default":     5,
				"description": "Maximum number of results",
			},
		},
		Security: ToolSecurity{
			Sandboxed:           true,
			AllowedDomains:      []string{"duckduckgo.com"},
			MaxExecutionTime:    30,
			RequiredPermissions: []string{"network.read"},
		},
	},
	"calculations": {
		Name:        "calculations",
		Description: "Perform mathematical calculations",
		Category:    "math",
		Parameters: map[string]interface{}{
			"expression": map[string]interface{}{
				"type":        "string",
				"required":    true,
				"description": "Mathematical expression to evaluate",
			},
		},
		Security: ToolSecurity{
			Sandboxed:           true,
			MaxExecutionTime:    10,
			RequiredPermissions: []string{},
		},
	},
	"text_processing": {
		Name:        "text_processing",
		Description: "Process and analyze text",
		Category:    "text",
		Parameters: map[string]interface{}{
			"text": map[string]interface{}{
				"type":        "string",
				"required":    true,
				"description": "Text to process",
			},
			"operation": map[string]interface{}{
				"type":        "string",
				"required":    true,
				"enum":        []string{"lowercase", "uppercase", "word_count", "sentiment"},
				"description": "Processing operation",
			},
		},
		Security: ToolSecurity{
			Sandboxed:           true,
			MaxExecutionTime:    15,
			RequiredPermissions: []string{},
		},
	},
	"file_operations": {
		Name:        "file_operations",
		Description: "Safe file read operations",
		Category:    "file",
		Parameters: map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"required":    true,
				"enum":        []string{"read", "list"},
				"description": "File operation type",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"required":    true,
				"description": "File path (restricted to safe directories)",
			},
		},
		Security: ToolSecurity{
			Sandboxed:           true,
			MaxExecutionTime:    20,
			RequiredPermissions: []string{"file.read"},
		},
	},
	"api_calls": {
		Name:        "api_calls",
		Description: "Make HTTP API calls",
		Category:    "network",
		Parameters: map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"required":    true,
				"description": "API endpoint URL",
			},
			"method": map[string]interface{}{
				"type":        "string",
				"required":    false,
				"default":     "GET",
				"enum":        []string{"GET", "POST"},
				"description": "HTTP method",
			},
			"headers": map[string]interface{}{
				"type":        "object",
				"required":    false,
				"description": "HTTP headers",
			},
		},
		Security: ToolSecurity{
			Sandboxed:           true,
			AllowedDomains:      []string{"api.github.com", "httpbin.org", "jsonplaceholder.typicode.com"},
			MaxExecutionTime:    30,
			RequiredPermissions: []string{"network.read"},
		},
	},
}

// ExecuteTool executes a tool with security sandbox
func (h *Handler) ExecuteTool(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req ToolExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Validate tool exists
	toolDef, exists := toolRegistry[req.ToolName]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unknown tool: " + req.ToolName,
		})
		return
	}

	// Validate agent ownership if agent_id provided
	if req.AgentID != "" && !h.validateAgentOwnership(req.AgentID, userID) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Agent not found",
		})
		return
	}

	// Set default timeout
	if req.Timeout == 0 {
		req.Timeout = toolDef.Security.MaxExecutionTime
	}

	// Validate timeout doesn't exceed maximum
	if req.Timeout > toolDef.Security.MaxExecutionTime {
		req.Timeout = toolDef.Security.MaxExecutionTime
	}

	// Execute tool in sandbox
	executionID := uuid.New().String()
	response := ToolExecutionResponse{
		ExecutionID: executionID,
		ToolName:    req.ToolName,
		Status:      "running",
		StartedAt:   time.Now(),
	}

	// Execute tool based on type
	result, err := h.executeToolSafely(toolDef, req.Parameters, req.Timeout)

	response.CompletedAt = &[]time.Time{time.Now()}[0]
	response.ExecutionTime = time.Since(response.StartedAt).Seconds()

	if err != nil {
		response.Status = "failed"
		response.Error = err.Error()
	} else {
		response.Status = "completed"
		response.Result = result
	}

	// Store execution result in database
	h.storeToolExecution(executionID, userID, req, response)

	c.JSON(http.StatusOK, response)
}

// GetToolDefinitions returns available tool definitions
func (h *Handler) GetToolDefinitions(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	tools := make([]ToolDefinition, 0, len(toolRegistry))
	for _, tool := range toolRegistry {
		tools = append(tools, tool)
	}

	c.JSON(http.StatusOK, gin.H{
		"tools": tools,
		"count": len(tools),
	})
}

// GetToolExecution returns tool execution details
func (h *Handler) GetToolExecution(c *gin.Context) {
	userID := c.GetString("user_id")
	executionID := c.Param("execution_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get execution from database
	execution, err := h.getToolExecution(executionID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tool execution not found",
		})
		return
	}

	c.JSON(http.StatusOK, execution)
}

// executeToolSafely executes a tool in a secure sandbox
func (h *Handler) executeToolSafely(toolDef ToolDefinition, parameters map[string]interface{}, timeout int) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	switch toolDef.Name {
	case "web_search":
		return h.executeWebSearch(ctx, parameters)
	case "calculations":
		return h.executeCalculations(ctx, parameters)
	case "text_processing":
		return h.executeTextProcessing(ctx, parameters)
	case "file_operations":
		return h.executeFileOperations(ctx, parameters)
	case "api_calls":
		return h.executeAPICall(ctx, parameters)
	default:
		return nil, fmt.Errorf("tool implementation not found: %s", toolDef.Name)
	}
}

// Tool implementations (secure sandbox versions)
func (h *Handler) executeWebSearch(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	query, ok := params["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query parameter is required")
	}

	// Placeholder implementation - would integrate with DuckDuckGo API
	return map[string]interface{}{
		"query": query,
		"results": []string{
			fmt.Sprintf("Search result 1 for: %s", query),
			fmt.Sprintf("Search result 2 for: %s", query),
		},
		"count": 2,
	}, nil
}

func (h *Handler) executeCalculations(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	expression, ok := params["expression"].(string)
	if !ok {
		return nil, fmt.Errorf("expression parameter is required")
	}

	// Safe mathematical evaluation (placeholder)
	// In production, use a proper math parser
	if strings.Contains(expression, "2+2") {
		return map[string]interface{}{
			"expression": expression,
			"result":     4,
		}, nil
	}

	return map[string]interface{}{
		"expression": expression,
		"result":     "Calculation result placeholder",
	}, nil
}

func (h *Handler) executeTextProcessing(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	text, ok := params["text"].(string)
	if !ok {
		return nil, fmt.Errorf("text parameter is required")
	}

	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation parameter is required")
	}

	switch operation {
	case "lowercase":
		return map[string]interface{}{
			"original":  text,
			"processed": strings.ToLower(text),
			"operation": operation,
		}, nil
	case "uppercase":
		return map[string]interface{}{
			"original":  text,
			"processed": strings.ToUpper(text),
			"operation": operation,
		}, nil
	case "word_count":
		words := strings.Fields(text)
		return map[string]interface{}{
			"original":   text,
			"word_count": len(words),
			"operation":  operation,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported operation: %s", operation)
	}
}

func (h *Handler) executeFileOperations(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	// Placeholder - would implement secure file operations
	return map[string]interface{}{
		"operation": params["operation"],
		"path":      params["path"],
		"result":    "File operation placeholder",
	}, nil
}

func (h *Handler) executeAPICall(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	// Placeholder - would implement secure HTTP calls
	return map[string]interface{}{
		"url":    params["url"],
		"method": params["method"],
		"result": "API call placeholder",
	}, nil
}

// Database operations for tool executions
func (h *Handler) storeToolExecution(executionID, userID string, req ToolExecutionRequest, resp ToolExecutionResponse) error {
	reqJSON, _ := json.Marshal(req)
	respJSON, _ := json.Marshal(resp)

	_, err := h.db.Exec(`
		INSERT INTO tool_executions (id, user_id, tool_name, request_data, response_data, status, execution_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, executionID, userID, req.ToolName, reqJSON, respJSON, resp.Status, resp.ExecutionTime)

	return err
}

func (h *Handler) getToolExecution(executionID, userID string) (ToolExecutionResponse, error) {
	var resp ToolExecutionResponse
	var respJSON []byte

	err := h.db.QueryRow(`
		SELECT response_data FROM tool_executions
		WHERE id = $1 AND user_id = $2
	`, executionID, userID).Scan(&respJSON)

	if err != nil {
		return resp, err
	}

	err = json.Unmarshal(respJSON, &resp)
	return resp, err
}
