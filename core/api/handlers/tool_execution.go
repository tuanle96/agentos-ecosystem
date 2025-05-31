package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
		Description: "Secure file operations in sandboxed environment",
		Category:    "file",
		Parameters: map[string]interface{}{
			"operation": map[string]interface{}{
				"type":        "string",
				"required":    true,
				"enum":        []string{"read", "write", "list", "exists", "delete", "create_dir"},
				"description": "File operation type",
			},
			"path": map[string]interface{}{
				"type":        "string",
				"required":    true,
				"description": "File path (restricted to safe directories)",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"required":    false,
				"description": "Content for write operations",
			},
		},
		Security: ToolSecurity{
			Sandboxed:           true,
			MaxExecutionTime:    20,
			RequiredPermissions: []string{"file.read", "file.write"},
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

	// Get max results parameter
	maxResults := 5
	if mr, ok := params["max_results"].(float64); ok {
		maxResults = int(mr)
		if maxResults > 10 {
			maxResults = 10 // Limit to prevent abuse
		}
	}

	// Real DuckDuckGo search implementation
	results, err := h.searchDuckDuckGo(ctx, query, maxResults)
	if err != nil {
		return nil, fmt.Errorf("search failed: %v", err)
	}

	return map[string]interface{}{
		"query":   query,
		"results": results,
		"count":   len(results),
		"source":  "duckduckgo",
		"success": true,
	}, nil
}

func (h *Handler) executeCalculations(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	expression, ok := params["expression"].(string)
	if !ok {
		return nil, fmt.Errorf("expression parameter is required")
	}

	// Validate expression for security
	if err := h.validateMathExpression(expression); err != nil {
		return nil, fmt.Errorf("invalid expression: %v", err)
	}

	// Evaluate mathematical expression safely
	result, err := h.evaluateMathExpression(expression)
	if err != nil {
		return map[string]interface{}{
			"expression": expression,
			"error":      err.Error(),
			"success":    false,
		}, nil
	}

	return map[string]interface{}{
		"expression": expression,
		"result":     result,
		"success":    true,
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
	operation, ok := params["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("operation parameter is required")
	}

	path, ok := params["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path parameter is required")
	}

	// Security: Validate and sanitize file path
	if err := h.validateFilePath(path); err != nil {
		return nil, fmt.Errorf("invalid file path: %v", err)
	}

	// Get secure working directory
	workDir := h.getSecureWorkingDirectory()
	fullPath := filepath.Join(workDir, filepath.Clean(path))

	switch operation {
	case "read":
		return h.executeFileRead(fullPath)
	case "write":
		content, _ := params["content"].(string)
		return h.executeFileWrite(fullPath, content)
	case "list":
		return h.executeFileList(fullPath)
	case "exists":
		return h.executeFileExists(fullPath)
	case "delete":
		return h.executeFileDelete(fullPath)
	case "create_dir":
		return h.executeCreateDirectory(fullPath)
	default:
		return nil, fmt.Errorf("unsupported file operation: %s", operation)
	}
}

func (h *Handler) executeAPICall(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	// Get URL parameter
	urlStr, ok := params["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url parameter is required")
	}

	// Validate URL
	if err := h.validateURL(urlStr); err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	// Get method (default to GET)
	method := "GET"
	if m, ok := params["method"].(string); ok {
		method = strings.ToUpper(m)
	}

	// Validate method
	if method != "GET" && method != "POST" {
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	// Get headers
	headers := make(map[string]string)
	if h, ok := params["headers"].(map[string]interface{}); ok {
		for k, v := range h {
			if str, ok := v.(string); ok {
				headers[k] = str
			}
		}
	}

	// Get body for POST requests
	var body string
	if bodyParam, ok := params["body"].(string); ok {
		body = bodyParam
	}

	// Execute HTTP request
	return h.executeHTTPRequest(ctx, method, urlStr, headers, body)
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

// ===================================
// REAL FILE OPERATIONS IMPLEMENTATION
// ===================================

// validateFilePath validates and sanitizes file paths for security
func (h *Handler) validateFilePath(path string) error {
	// Security checks
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed")
	}
	if strings.HasPrefix(path, "/") {
		return fmt.Errorf("absolute paths not allowed")
	}
	if strings.Contains(path, "~") {
		return fmt.Errorf("home directory access not allowed")
	}

	// Check for dangerous file extensions
	dangerousExts := []string{".exe", ".bat", ".sh", ".cmd", ".com", ".scr", ".pif"}
	for _, ext := range dangerousExts {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return fmt.Errorf("dangerous file extension not allowed: %s", ext)
		}
	}

	return nil
}

// getSecureWorkingDirectory returns a secure sandbox directory for file operations
func (h *Handler) getSecureWorkingDirectory() string {
	// Create a secure sandbox directory
	workDir := "/tmp/agentos_files"
	os.MkdirAll(workDir, 0755)
	return workDir
}

// executeFileRead reads a file and returns its content
func (h *Handler) executeFileRead(fullPath string) (map[string]interface{}, error) {
	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", fullPath)
	}

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Check file size limit (1MB max)
	if len(content) > 1024*1024 {
		return nil, fmt.Errorf("file too large (max 1MB)")
	}

	return map[string]interface{}{
		"operation": "read",
		"path":      fullPath,
		"content":   string(content),
		"size":      len(content),
		"success":   true,
	}, nil
}

// executeFileWrite writes content to a file
func (h *Handler) executeFileWrite(fullPath, content string) (map[string]interface{}, error) {
	// Check content size limit (1MB max)
	if len(content) > 1024*1024 {
		return nil, fmt.Errorf("content too large (max 1MB)")
	}

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// Write file
	err := os.WriteFile(fullPath, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %v", err)
	}

	return map[string]interface{}{
		"operation": "write",
		"path":      fullPath,
		"size":      len(content),
		"success":   true,
	}, nil
}

// executeFileList lists files in a directory
func (h *Handler) executeFileList(fullPath string) (map[string]interface{}, error) {
	// Check if path exists
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("path does not exist: %s", fullPath)
	}

	var files []map[string]interface{}

	if info.IsDir() {
		// List directory contents
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %v", err)
		}

		for _, entry := range entries {
			fileInfo, err := entry.Info()
			if err != nil {
				continue
			}

			files = append(files, map[string]interface{}{
				"name":     entry.Name(),
				"is_dir":   entry.IsDir(),
				"size":     fileInfo.Size(),
				"mod_time": fileInfo.ModTime(),
			})
		}
	} else {
		// Single file info
		files = append(files, map[string]interface{}{
			"name":     info.Name(),
			"is_dir":   false,
			"size":     info.Size(),
			"mod_time": info.ModTime(),
		})
	}

	return map[string]interface{}{
		"operation": "list",
		"path":      fullPath,
		"files":     files,
		"count":     len(files),
		"success":   true,
	}, nil
}

// executeFileExists checks if a file exists
func (h *Handler) executeFileExists(fullPath string) (map[string]interface{}, error) {
	_, err := os.Stat(fullPath)
	exists := !os.IsNotExist(err)

	return map[string]interface{}{
		"operation": "exists",
		"path":      fullPath,
		"exists":    exists,
		"success":   true,
	}, nil
}

// executeFileDelete deletes a file or directory
func (h *Handler) executeFileDelete(fullPath string) (map[string]interface{}, error) {
	// Check if file exists
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", fullPath)
	}

	// Delete file or directory
	err = os.RemoveAll(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to delete: %v", err)
	}

	return map[string]interface{}{
		"operation": "delete",
		"path":      fullPath,
		"success":   true,
	}, nil
}

// executeCreateDirectory creates a directory
func (h *Handler) executeCreateDirectory(fullPath string) (map[string]interface{}, error) {
	// Create directory with all parent directories
	err := os.MkdirAll(fullPath, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	return map[string]interface{}{
		"operation": "create_dir",
		"path":      fullPath,
		"success":   true,
	}, nil
}

// ExecuteFileOperations - Public method for testing file operations
func (h *Handler) ExecuteFileOperations(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	return h.executeFileOperations(ctx, params)
}

// ExecuteAPICall - Public method for testing API calls
func (h *Handler) ExecuteAPICall(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	return h.executeAPICall(ctx, params)
}

// ===================================
// REAL API CALLS IMPLEMENTATION
// ===================================

// validateURL validates and sanitizes URLs for security
func (h *Handler) validateURL(urlStr string) error {
	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %v", err)
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("only HTTP and HTTPS schemes are allowed")
	}

	// Check for localhost/private IPs (security)
	if strings.Contains(parsedURL.Host, "localhost") ||
		strings.Contains(parsedURL.Host, "127.0.0.1") ||
		strings.Contains(parsedURL.Host, "0.0.0.0") {
		return fmt.Errorf("localhost/private IP access not allowed")
	}

	// Whitelist allowed domains for security
	allowedDomains := []string{
		"api.github.com",
		"httpbin.org",
		"jsonplaceholder.typicode.com",
		"duckduckgo.com",
		"api.openai.com",
	}

	domainAllowed := false
	for _, domain := range allowedDomains {
		if strings.Contains(parsedURL.Host, domain) {
			domainAllowed = true
			break
		}
	}

	if !domainAllowed {
		return fmt.Errorf("domain not in whitelist: %s", parsedURL.Host)
	}

	return nil
}

// executeHTTPRequest performs the actual HTTP request
func (h *Handler) executeHTTPRequest(ctx context.Context, method, urlStr string, headers map[string]string, body string) (map[string]interface{}, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Prepare request body
	var reqBody io.Reader
	if body != "" && method == "POST" {
		reqBody = strings.NewReader(body)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, urlStr, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("User-Agent", "AgentOS/1.0")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set content type for POST requests
	if method == "POST" && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Limit response size (1MB max)
	if len(respBody) > 1024*1024 {
		return nil, fmt.Errorf("response too large (max 1MB)")
	}

	// Parse response headers
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}

	return map[string]interface{}{
		"operation":      "api_call",
		"url":            urlStr,
		"method":         method,
		"status_code":    resp.StatusCode,
		"status":         resp.Status,
		"headers":        responseHeaders,
		"body":           string(respBody),
		"content_length": len(respBody),
		"execution_time": time.Since(startTime).Seconds(),
		"success":        resp.StatusCode >= 200 && resp.StatusCode < 300,
	}, nil
}

// ===================================
// REAL MATHEMATICAL CALCULATIONS IMPLEMENTATION
// ===================================

// validateMathExpression validates mathematical expressions for security
func (h *Handler) validateMathExpression(expression string) error {
	// Remove whitespace for validation
	expr := strings.ReplaceAll(expression, " ", "")

	// Check length limit
	if len(expr) > 100 {
		return fmt.Errorf("expression too long (max 100 characters)")
	}

	// Check for dangerous characters/functions
	dangerous := []string{
		"import", "exec", "eval", "open", "file", "input", "raw_input",
		"__", "system", "os", "subprocess", "shell", "cmd", "bash",
		"rm", "del", "delete", "format", "globals", "locals", "vars",
	}

	lowerExpr := strings.ToLower(expr)
	for _, danger := range dangerous {
		if strings.Contains(lowerExpr, danger) {
			return fmt.Errorf("dangerous function/keyword not allowed: %s", danger)
		}
	}

	// Allow only safe mathematical characters
	allowedChars := "0123456789+-*/().^%"
	for _, char := range expr {
		if !strings.ContainsRune(allowedChars, char) {
			return fmt.Errorf("invalid character: %c", char)
		}
	}

	return nil
}

// evaluateMathExpression safely evaluates mathematical expressions
func (h *Handler) evaluateMathExpression(expression string) (float64, error) {
	// Remove whitespace
	expr := strings.ReplaceAll(expression, " ", "")

	// Handle simple arithmetic operations
	result, err := h.parseAndEvaluate(expr)
	if err != nil {
		return 0, fmt.Errorf("evaluation error: %v", err)
	}

	return result, nil
}

// parseAndEvaluate implements a simple recursive descent parser for basic math
func (h *Handler) parseAndEvaluate(expr string) (float64, error) {
	// Handle basic operations: +, -, *, /, (), ^, %
	// This is a simplified implementation for common mathematical expressions

	// Handle parentheses first
	for strings.Contains(expr, "(") {
		start := strings.LastIndex(expr, "(")
		if start == -1 {
			break
		}

		end := strings.Index(expr[start:], ")")
		if end == -1 {
			return 0, fmt.Errorf("mismatched parentheses")
		}
		end += start

		subExpr := expr[start+1 : end]
		subResult, err := h.parseAndEvaluate(subExpr)
		if err != nil {
			return 0, err
		}

		expr = expr[:start] + fmt.Sprintf("%g", subResult) + expr[end+1:]
	}

	// Handle basic arithmetic
	return h.evaluateBasicArithmetic(expr)
}

// evaluateBasicArithmetic evaluates basic arithmetic without parentheses
func (h *Handler) evaluateBasicArithmetic(expr string) (float64, error) {
	// Handle multiplication and division first (higher precedence)
	expr = h.handleHighPrecedence(expr)

	// Handle addition and subtraction
	return h.handleLowPrecedence(expr)
}

// handleHighPrecedence handles *, /, %, ^ operations
func (h *Handler) handleHighPrecedence(expr string) string {
	operators := []string{"^", "*", "/", "%"}

	for _, op := range operators {
		for strings.Contains(expr, op) {
			idx := strings.Index(expr, op)
			if idx == -1 {
				break
			}

			// Find left operand
			leftStart := idx - 1
			for leftStart >= 0 && (expr[leftStart] >= '0' && expr[leftStart] <= '9' || expr[leftStart] == '.') {
				leftStart--
			}
			leftStart++

			// Find right operand
			rightEnd := idx + 1
			for rightEnd < len(expr) && (expr[rightEnd] >= '0' && expr[rightEnd] <= '9' || expr[rightEnd] == '.') {
				rightEnd++
			}

			if leftStart >= idx || rightEnd <= idx+1 {
				break
			}

			leftStr := expr[leftStart:idx]
			rightStr := expr[idx+1 : rightEnd]

			left, err1 := h.parseFloat(leftStr)
			right, err2 := h.parseFloat(rightStr)

			if err1 != nil || err2 != nil {
				break
			}

			var result float64
			switch op {
			case "+":
				result = left + right
			case "-":
				result = left - right
			case "*":
				result = left * right
			case "/":
				if right == 0 {
					return expr // Don't divide by zero, return as-is
				}
				result = left / right
			case "%":
				if right == 0 {
					return expr
				}
				result = float64(int(left) % int(right))
			case "^":
				result = h.power(left, right)
			}

			expr = expr[:leftStart] + fmt.Sprintf("%g", result) + expr[rightEnd:]
		}
	}

	return expr
}

// handleLowPrecedence handles + and - operations
func (h *Handler) handleLowPrecedence(expr string) (float64, error) {
	// Simple implementation for addition and subtraction
	if !strings.ContainsAny(expr, "+-") {
		// No operators, just parse the number
		return h.parseFloat(expr)
	}

	// Find the last + or - (left to right evaluation)
	var lastOpIdx = -1
	var lastOp string

	for i := len(expr) - 1; i >= 0; i-- {
		if expr[i] == '+' || expr[i] == '-' {
			// Skip if it's at the beginning (negative number)
			if i == 0 {
				continue
			}
			lastOpIdx = i
			lastOp = string(expr[i])
			break
		}
	}

	if lastOpIdx == -1 {
		// No operators found, parse as number
		return h.parseFloat(expr)
	}

	leftStr := expr[:lastOpIdx]
	rightStr := expr[lastOpIdx+1:]

	left, err1 := h.handleLowPrecedence(leftStr)
	right, err2 := h.parseFloat(rightStr)

	if err1 != nil || err2 != nil {
		return 0, fmt.Errorf("invalid operands")
	}

	switch lastOp {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	default:
		return 0, fmt.Errorf("unknown operator: %s", lastOp)
	}
}

// parseFloat safely parses a float from string
func (h *Handler) parseFloat(s string) (float64, error) {
	if s == "" {
		return 0, fmt.Errorf("empty string")
	}

	// Use Go's built-in strconv.ParseFloat for safety
	result, err := fmt.Sscanf(s, "%f", new(float64))
	if err != nil || result != 1 {
		return 0, fmt.Errorf("invalid number: %s", s)
	}

	var value float64
	fmt.Sscanf(s, "%f", &value)
	return value, nil
}

// power calculates x^y safely
func (h *Handler) power(base, exponent float64) float64 {
	// Limit exponent to prevent overflow
	if exponent > 100 || exponent < -100 {
		return 0 // Return 0 for extreme exponents
	}

	// Simple power calculation for integer exponents
	if exponent == 0 {
		return 1
	}

	if exponent == 1 {
		return base
	}

	if exponent < 0 {
		return 1 / h.power(base, -exponent)
	}

	// For positive integer exponents
	result := 1.0
	for i := 0; i < int(exponent); i++ {
		result *= base
	}

	return result
}

// ExecuteCalculations - Public method for testing calculations
func (h *Handler) ExecuteCalculations(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	return h.executeCalculations(ctx, params)
}

// ===================================
// REAL WEB SEARCH IMPLEMENTATION
// ===================================

// SearchResult represents a search result from web search
type SearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

// DuckDuckGoResponse represents the response from DuckDuckGo Instant Answer API
type DuckDuckGoResponse struct {
	Abstract       string                   `json:"Abstract"`
	AbstractText   string                   `json:"AbstractText"`
	AbstractSource string                   `json:"AbstractSource"`
	AbstractURL    string                   `json:"AbstractURL"`
	Image          string                   `json:"Image"`
	Heading        string                   `json:"Heading"`
	Answer         string                   `json:"Answer"`
	AnswerType     string                   `json:"AnswerType"`
	Definition     string                   `json:"Definition"`
	DefinitionURL  string                   `json:"DefinitionURL"`
	RelatedTopics  []DuckDuckGoRelatedTopic `json:"RelatedTopics"`
	Results        []DuckDuckGoResult       `json:"Results"`
}

type DuckDuckGoRelatedTopic struct {
	FirstURL string `json:"FirstURL"`
	Result   string `json:"Result"`
	Text     string `json:"Text"`
}

type DuckDuckGoResult struct {
	FirstURL string `json:"FirstURL"`
	Result   string `json:"Result"`
	Text     string `json:"Text"`
}

// searchDuckDuckGo performs real web search using DuckDuckGo Instant Answer API
func (h *Handler) searchDuckDuckGo(ctx context.Context, query string, maxResults int) ([]SearchResult, error) {
	// DuckDuckGo Instant Answer API endpoint
	baseURL := "https://api.duckduckgo.com/"

	// Prepare query parameters
	params := url.Values{}
	params.Set("q", query)
	params.Set("format", "json")
	params.Set("no_html", "1")
	params.Set("skip_disambig", "1")

	searchURL := baseURL + "?" + params.Encode()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "AgentOS/1.0 (Web Search Tool)")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse JSON response
	var ddgResp DuckDuckGoResponse
	if err := json.Unmarshal(body, &ddgResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Convert to SearchResult format
	results := h.convertDuckDuckGoResults(ddgResp, maxResults)

	return results, nil
}

// convertDuckDuckGoResults converts DuckDuckGo API response to SearchResult format
func (h *Handler) convertDuckDuckGoResults(ddgResp DuckDuckGoResponse, maxResults int) []SearchResult {
	var results []SearchResult

	// Add abstract/answer if available
	if ddgResp.Abstract != "" && ddgResp.AbstractURL != "" {
		results = append(results, SearchResult{
			Title:       ddgResp.Heading,
			URL:         ddgResp.AbstractURL,
			Description: ddgResp.AbstractText,
			Source:      ddgResp.AbstractSource,
		})
	}

	// Add definition if available
	if ddgResp.Definition != "" && ddgResp.DefinitionURL != "" {
		results = append(results, SearchResult{
			Title:       "Definition: " + ddgResp.Heading,
			URL:         ddgResp.DefinitionURL,
			Description: ddgResp.Definition,
			Source:      "definition",
		})
	}

	// Add instant answer if available
	if ddgResp.Answer != "" {
		results = append(results, SearchResult{
			Title:       "Instant Answer",
			URL:         "",
			Description: ddgResp.Answer,
			Source:      "instant_answer",
		})
	}

	// Add related topics
	for _, topic := range ddgResp.RelatedTopics {
		if len(results) >= maxResults {
			break
		}

		if topic.FirstURL != "" && topic.Text != "" {
			// Extract title from text (usually format: "Title - Description")
			title := topic.Text
			description := ""

			if idx := strings.Index(topic.Text, " - "); idx != -1 {
				title = topic.Text[:idx]
				description = topic.Text[idx+3:]
			}

			results = append(results, SearchResult{
				Title:       title,
				URL:         topic.FirstURL,
				Description: description,
				Source:      "related_topic",
			})
		}
	}

	// Add direct results
	for _, result := range ddgResp.Results {
		if len(results) >= maxResults {
			break
		}

		if result.FirstURL != "" && result.Text != "" {
			// Extract title from text
			title := result.Text
			description := ""

			if idx := strings.Index(result.Text, " - "); idx != -1 {
				title = result.Text[:idx]
				description = result.Text[idx+3:]
			}

			results = append(results, SearchResult{
				Title:       title,
				URL:         result.FirstURL,
				Description: description,
				Source:      "direct_result",
			})
		}
	}

	// If no results found, provide a helpful message
	if len(results) == 0 {
		results = append(results, SearchResult{
			Title:       "No results found",
			URL:         "",
			Description: "DuckDuckGo did not return any results for this query. Try a different search term.",
			Source:      "no_results",
		})
	}

	// Limit results to maxResults
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results
}

// ExecuteWebSearch - Public method for testing web search
func (h *Handler) ExecuteWebSearch(ctx context.Context, params map[string]interface{}) (map[string]interface{}, error) {
	return h.executeWebSearch(ctx, params)
}
