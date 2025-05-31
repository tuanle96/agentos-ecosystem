package tests

import (
	"net/http/httptest"
	"strings"
	"testing"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tuanle96/agentos-ecosystem/core/api/handlers"
)

// TestToolExecutionHandlersDirect tests tool execution handlers directly to boost coverage
func TestToolExecutionHandlersDirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create handler
	h := &handlers.Handler{}
	
	// Test GetToolDefinitions
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "test-user-123")
	
	h.GetToolDefinitions(c)
	assert.Equal(t, 200, w.Code)
	
	// Test ExecuteTool with calculations
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Set("user_id", "test-user-123")
	c2.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{
		"tool_name": "calculations",
		"parameters": {
			"expression": "2+2"
		},
		"timeout": 10
	}`))
	c2.Request.Header.Set("Content-Type", "application/json")
	
	h.ExecuteTool(c2)
	assert.Equal(t, 200, w2.Code)
	
	// Test ExecuteTool with text_processing
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Set("user_id", "test-user-123")
	c3.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{
		"tool_name": "text_processing",
		"parameters": {
			"text": "Hello World",
			"operation": "lowercase"
		}
	}`))
	c3.Request.Header.Set("Content-Type", "application/json")
	
	h.ExecuteTool(c3)
	assert.Equal(t, 200, w3.Code)
	
	// Test ExecuteTool with web_search
	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Set("user_id", "test-user-123")
	c4.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{
		"tool_name": "web_search",
		"parameters": {
			"query": "AgentOS framework",
			"max_results": 5
		}
	}`))
	c4.Request.Header.Set("Content-Type", "application/json")
	
	h.ExecuteTool(c4)
	assert.Equal(t, 200, w4.Code)
	
	// Test ExecuteTool with file_operations
	w5 := httptest.NewRecorder()
	c5, _ := gin.CreateTestContext(w5)
	c5.Set("user_id", "test-user-123")
	c5.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{
		"tool_name": "file_operations",
		"parameters": {
			"operation": "read",
			"path": "/safe/path/file.txt"
		}
	}`))
	c5.Request.Header.Set("Content-Type", "application/json")
	
	h.ExecuteTool(c5)
	assert.Equal(t, 200, w5.Code)
	
	// Test ExecuteTool with api_calls
	w6 := httptest.NewRecorder()
	c6, _ := gin.CreateTestContext(w6)
	c6.Set("user_id", "test-user-123")
	c6.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{
		"tool_name": "api_calls",
		"parameters": {
			"url": "https://api.github.com/users/octocat",
			"method": "GET"
		}
	}`))
	c6.Request.Header.Set("Content-Type", "application/json")
	
	h.ExecuteTool(c6)
	assert.Equal(t, 200, w6.Code)
}

// TestToolExecutionErrorCases tests tool execution error scenarios
func TestToolExecutionErrorCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create handler
	h := &handlers.Handler{}
	
	// Test GetToolDefinitions without authentication
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// No user_id set
	
	h.GetToolDefinitions(c)
	assert.Equal(t, 401, w.Code)
	
	// Test ExecuteTool without authentication
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"tool_name":"calculations"}`))
	c2.Request.Header.Set("Content-Type", "application/json")
	
	h.ExecuteTool(c2)
	assert.Equal(t, 401, w2.Code)
	
	// Test ExecuteTool with invalid JSON
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Set("user_id", "test-user-123")
	c3.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"invalid":"json"`))
	c3.Request.Header.Set("Content-Type", "application/json")
	
	h.ExecuteTool(c3)
	assert.Equal(t, 400, w3.Code)
	
	// Test ExecuteTool with unknown tool
	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Set("user_id", "test-user-123")
	c4.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{
		"tool_name": "unknown_tool",
		"parameters": {}
	}`))
	c4.Request.Header.Set("Content-Type", "application/json")
	
	h.ExecuteTool(c4)
	assert.Equal(t, 400, w4.Code)
	
	// Test ExecuteTool with missing parameters
	w5 := httptest.NewRecorder()
	c5, _ := gin.CreateTestContext(w5)
	c5.Set("user_id", "test-user-123")
	c5.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{
		"tool_name": "calculations"
	}`))
	c5.Request.Header.Set("Content-Type", "application/json")
	
	h.ExecuteTool(c5)
	assert.Equal(t, 400, w5.Code)
	
	// Test GetToolExecution without authentication
	w6 := httptest.NewRecorder()
	c6, _ := gin.CreateTestContext(w6)
	c6.Params = gin.Params{gin.Param{Key: "execution_id", Value: "test-execution-id"}}
	
	h.GetToolExecution(c6)
	assert.Equal(t, 401, w6.Code)
}

// TestTextProcessingOperations tests different text processing operations
func TestTextProcessingOperations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create handler
	h := &handlers.Handler{}
	
	operations := []struct {
		operation string
		text      string
		expected  int
	}{
		{"lowercase", "HELLO WORLD", 200},
		{"uppercase", "hello world", 200},
		{"word_count", "hello world test", 200},
	}
	
	for _, op := range operations {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", "test-user-123")
		c.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{
			"tool_name": "text_processing",
			"parameters": {
				"text": "`+op.text+`",
				"operation": "`+op.operation+`"
			}
		}`))
		c.Request.Header.Set("Content-Type", "application/json")
		
		h.ExecuteTool(c)
		assert.Equal(t, op.expected, w.Code)
	}
	
	// Test unsupported operation
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", "test-user-123")
	c.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{
		"tool_name": "text_processing",
		"parameters": {
			"text": "test",
			"operation": "unsupported_operation"
		}
	}`))
	c.Request.Header.Set("Content-Type", "application/json")
	
	h.ExecuteTool(c)
	assert.Equal(t, 200, w.Code) // Should return 200 but with error in response
}

// TestCalculationsVariations tests different calculation scenarios
func TestCalculationsVariations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Create handler
	h := &handlers.Handler{}
	
	expressions := []string{"2+2", "5*3", "10/2", "unknown_expression"}
	
	for _, expr := range expressions {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("user_id", "test-user-123")
		c.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{
			"tool_name": "calculations",
			"parameters": {
				"expression": "`+expr+`"
			}
		}`))
		c.Request.Header.Set("Content-Type", "application/json")
		
		h.ExecuteTool(c)
		assert.Equal(t, 200, w.Code)
	}
}
