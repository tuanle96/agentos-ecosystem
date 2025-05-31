package tests

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tuanle96/agentos-ecosystem/core/api/handlers"
)

// TestMemoryHandlersDirect tests memory handlers directly to boost coverage
func TestMemoryHandlersDirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler
	h := &handlers.Handler{}

	// Test GetAgentMemoryEnhanced
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "agent_id", Value: "test-agent-123"}}

	h.GetAgentMemoryEnhanced(c)
	assert.Equal(t, 200, w.Code)

	// Test ClearAgentMemoryEnhanced
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Params = gin.Params{gin.Param{Key: "agent_id", Value: "test-agent-123"}}

	h.ClearAgentMemoryEnhanced(c2)
	assert.Equal(t, 200, w2.Code)

	// Test CreateWorkingMemorySession with valid JSON
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"agent_id":"test","session_name":"test"}`))
	c3.Request.Header.Set("Content-Type", "application/json")

	h.CreateWorkingMemorySession(c3)
	assert.Equal(t, 201, w3.Code)

	// Test UpdateWorkingMemory with valid JSON
	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Params = gin.Params{gin.Param{Key: "session_id", Value: "test-session-123"}}
	c4.Request = httptest.NewRequest("PUT", "/", strings.NewReader(`{"context":{"test":"data"},"status":"active"}`))
	c4.Request.Header.Set("Content-Type", "application/json")

	h.UpdateWorkingMemory(c4)
	assert.Equal(t, 200, w4.Code)
}

// TestMemoryHandlersErrorCases tests memory handlers error cases
func TestMemoryHandlersErrorCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create handler
	h := &handlers.Handler{}

	// Test GetAgentMemoryEnhanced with empty agent ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "agent_id", Value: ""}}

	h.GetAgentMemoryEnhanced(c)
	assert.Equal(t, 400, w.Code)

	// Test ClearAgentMemoryEnhanced with empty agent ID
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Params = gin.Params{gin.Param{Key: "agent_id", Value: ""}}

	h.ClearAgentMemoryEnhanced(c2)
	assert.Equal(t, 400, w2.Code)

	// Test CreateWorkingMemorySession with invalid JSON
	w3 := httptest.NewRecorder()
	c3, _ := gin.CreateTestContext(w3)
	c3.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"invalid":"json"`))
	c3.Request.Header.Set("Content-Type", "application/json")

	h.CreateWorkingMemorySession(c3)
	assert.Equal(t, 400, w3.Code)

	// Test CreateWorkingMemorySession with missing required fields
	w4 := httptest.NewRecorder()
	c4, _ := gin.CreateTestContext(w4)
	c4.Request = httptest.NewRequest("POST", "/", strings.NewReader(`{"session_name":""}`))
	c4.Request.Header.Set("Content-Type", "application/json")

	h.CreateWorkingMemorySession(c4)
	assert.Equal(t, 400, w4.Code)

	// Test UpdateWorkingMemory with empty session ID
	w5 := httptest.NewRecorder()
	c5, _ := gin.CreateTestContext(w5)
	c5.Params = gin.Params{gin.Param{Key: "session_id", Value: ""}}
	c5.Request = httptest.NewRequest("PUT", "/", strings.NewReader(`{"context":{}}`))
	c5.Request.Header.Set("Content-Type", "application/json")

	h.UpdateWorkingMemory(c5)
	assert.Equal(t, 400, w5.Code)

	// Test UpdateWorkingMemory with invalid JSON
	w6 := httptest.NewRecorder()
	c6, _ := gin.CreateTestContext(w6)
	c6.Params = gin.Params{gin.Param{Key: "session_id", Value: "test-session"}}
	c6.Request = httptest.NewRequest("PUT", "/", strings.NewReader(`{"invalid":"json"`))
	c6.Request.Header.Set("Content-Type", "application/json")

	h.UpdateWorkingMemory(c6)
	assert.Equal(t, 400, w6.Code)
}
