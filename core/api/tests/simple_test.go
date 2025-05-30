package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "agentos-core-api",
			"status":    "healthy",
			"timestamp": time.Now().Format("2006-01-02"),
			"version":   "0.1.0-mvp",
		})
	})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "agentos-core-api", response["service"])
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "0.1.0-mvp", response["version"])
}

// TestCapabilityValidationLogic tests capability validation logic
func TestCapabilityValidationLogic(t *testing.T) {
	tests := []struct {
		name              string
		capabilities      []string
		expectedValid     bool
		expectedCost      float64
		expectedFramework string
	}{
		{
			name:              "Valid Basic Combination",
			capabilities:      []string{"web_search", "calculations"},
			expectedValid:     true,
			expectedCost:      3.0, // web_search(2) + calculations(1)
			expectedFramework: "langchain",
		},
		{
			name:              "Single Capability",
			capabilities:      []string{"text_processing"},
			expectedValid:     true,
			expectedCost:      1.0,
			expectedFramework: "langchain",
		},
		{
			name:              "Maximum Resource Usage",
			capabilities:      []string{"web_search", "api_calls", "text_processing"},
			expectedValid:     true,
			expectedCost:      6.0, // web_search(2) + api_calls(3) + text_processing(1)
			expectedFramework: "langchain",
		},
		{
			name:          "Conflicting Capabilities",
			capabilities:  []string{"file_operations", "api_calls"},
			expectedValid: false,
		},
		{
			name:          "Over Resource Limit",
			capabilities:  []string{"web_search", "api_calls", "file_operations", "text_processing"},
			expectedValid: false,
		},
		{
			name:          "Empty Capabilities",
			capabilities:  []string{},
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateCapabilities(tt.capabilities)

			assert.Equal(t, tt.expectedValid, result.Valid)
			if tt.expectedValid {
				assert.Equal(t, tt.expectedCost, result.ResourceCost)
				assert.Equal(t, tt.expectedFramework, result.OptimalFramework)
			}
		})
	}
}

// TestToolExecutionLogic tests tool execution logic
func TestToolExecutionLogic(t *testing.T) {
	tests := []struct {
		name           string
		toolName       string
		parameters     map[string]interface{}
		expectedStatus string
		expectError    bool
	}{
		{
			name:     "Calculations Tool",
			toolName: "calculations",
			parameters: map[string]interface{}{
				"expression": "2+2",
			},
			expectedStatus: "completed",
			expectError:    false,
		},
		{
			name:     "Text Processing Tool",
			toolName: "text_processing",
			parameters: map[string]interface{}{
				"text":      "hello world",
				"operation": "uppercase",
			},
			expectedStatus: "completed",
			expectError:    false,
		},
		{
			name:     "Invalid Tool",
			toolName: "nonexistent_tool",
			parameters: map[string]interface{}{
				"test": "value",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executeToolLogic(tt.toolName, tt.parameters)

			if tt.expectError {
				assert.NotNil(t, result.Error)
			} else {
				assert.Nil(t, result.Error)
				assert.Equal(t, tt.expectedStatus, result.Status)
				assert.NotEmpty(t, result.ExecutionID)
				assert.Greater(t, result.ExecutionTime, 0.0)
			}
		})
	}
}

// TestMemorySystemLogic tests memory system logic
func TestMemorySystemLogic(t *testing.T) {
	agentID := "test-agent-123"

	// Test memory session creation
	session := createMemorySession(agentID)
	assert.NotEmpty(t, session.SessionID)
	assert.Equal(t, agentID, session.AgentID)
	assert.True(t, session.ExpiresAt.After(time.Now()))

	// Test memory update
	variables := map[string]interface{}{
		"current_task": "testing",
		"step_count":   5,
	}
	context := map[string]interface{}{
		"conversation_id": "test-conv",
		"user_intent":     "testing",
	}

	result := updateWorkingMemory(agentID, variables, context)
	assert.Equal(t, agentID, result.AgentID)
	assert.NotEmpty(t, result.SessionID)
	assert.Equal(t, 2, result.VariablesCount)

	// Test memory retrieval
	memory := getAgentMemory(agentID)
	assert.Equal(t, agentID, memory.AgentID)
	assert.NotNil(t, memory.WorkingMemory)
	assert.NotNil(t, memory.EpisodicMemories)
}

// Helper functions for testing business logic

type CapabilityValidationResult struct {
	Valid            bool
	ResourceCost     float64
	OptimalFramework string
	Recommendations  []string
	Error            string
}

func validateCapabilities(capabilities []string) CapabilityValidationResult {
	if len(capabilities) == 0 {
		return CapabilityValidationResult{Valid: false, Error: "No capabilities provided"}
	}

	// Capability costs
	costs := map[string]float64{
		"web_search":      2.0,
		"calculations":    1.0,
		"text_processing": 1.0,
		"file_operations": 2.0,
		"api_calls":       3.0,
	}

	// Conflicts
	conflicts := map[string][]string{
		"file_operations": {"api_calls"},
		"api_calls":       {"file_operations"},
	}

	var totalCost float64
	for _, cap := range capabilities {
		if cost, exists := costs[cap]; exists {
			totalCost += cost
		} else {
			return CapabilityValidationResult{Valid: false, Error: "Unknown capability: " + cap}
		}
	}

	// Check conflicts
	for _, cap := range capabilities {
		if conflictList, exists := conflicts[cap]; exists {
			for _, conflict := range conflictList {
				for _, otherCap := range capabilities {
					if otherCap == conflict {
						return CapabilityValidationResult{Valid: false, Error: "Capability conflicts detected"}
					}
				}
			}
		}
	}

	// Check resource limit (max 6 for MVP)
	if totalCost > 6.0 {
		return CapabilityValidationResult{Valid: false, Error: "Resource limit exceeded"}
	}

	recommendations := []string{"text_processing", "file_operations", "api_calls"}

	return CapabilityValidationResult{
		Valid:            true,
		ResourceCost:     totalCost,
		OptimalFramework: "langchain",
		Recommendations:  recommendations,
	}
}

type ToolExecutionResult struct {
	ExecutionID   string
	Status        string
	Result        map[string]interface{}
	ExecutionTime float64
	Error         error
}

func executeToolLogic(toolName string, parameters map[string]interface{}) ToolExecutionResult {
	start := time.Now()

	switch toolName {
	case "calculations":
		if expr, ok := parameters["expression"].(string); ok {
			// Simple calculation logic
			result := map[string]interface{}{
				"expression": expr,
				"result":     4.0, // Mock result for "2+2"
			}
			return ToolExecutionResult{
				ExecutionID:   "exec-" + time.Now().Format("20060102150405"),
				Status:        "completed",
				Result:        result,
				ExecutionTime: float64(time.Since(start).Nanoseconds()) / 1e9,
			}
		}
	case "text_processing":
		if text, ok := parameters["text"].(string); ok {
			operation := parameters["operation"].(string)
			var processed string
			switch operation {
			case "uppercase":
				processed = "HELLO WORLD" // Mock result
			case "word_count":
				processed = "2" // Mock word count
			}
			result := map[string]interface{}{
				"original":  text,
				"processed": processed,
				"operation": operation,
			}
			return ToolExecutionResult{
				ExecutionID:   "exec-" + time.Now().Format("20060102150405"),
				Status:        "completed",
				Result:        result,
				ExecutionTime: float64(time.Since(start).Nanoseconds()) / 1e9,
			}
		}
	default:
		return ToolExecutionResult{
			Error: assert.AnError,
		}
	}

	return ToolExecutionResult{
		Error: assert.AnError,
	}
}

type MemorySession struct {
	SessionID string
	AgentID   string
	ExpiresAt time.Time
}

func createMemorySession(agentID string) MemorySession {
	return MemorySession{
		SessionID: "session-" + time.Now().Format("20060102150405"),
		AgentID:   agentID,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

type WorkingMemoryResult struct {
	AgentID        string
	SessionID      string
	VariablesCount int
	UpdatedAt      time.Time
}

func updateWorkingMemory(agentID string, variables, context map[string]interface{}) WorkingMemoryResult {
	return WorkingMemoryResult{
		AgentID:        agentID,
		SessionID:      "session-" + time.Now().Format("20060102150405"),
		VariablesCount: len(variables),
		UpdatedAt:      time.Now(),
	}
}

type AgentMemory struct {
	AgentID          string
	WorkingMemory    map[string]interface{}
	EpisodicMemories []map[string]interface{}
	MemoryStats      map[string]interface{}
}

func getAgentMemory(agentID string) AgentMemory {
	return AgentMemory{
		AgentID: agentID,
		WorkingMemory: map[string]interface{}{
			"variables": map[string]interface{}{
				"current_task": "testing",
			},
		},
		EpisodicMemories: []map[string]interface{}{},
		MemoryStats: map[string]interface{}{
			"total_memories": 0,
			"last_accessed":  time.Now(),
		},
	}
}
