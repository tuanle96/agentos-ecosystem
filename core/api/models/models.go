package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JSONB represents a PostgreSQL JSONB field
type JSONB map[string]interface{}

// CapabilityArray represents an array of capabilities stored as JSON
type CapabilityArray []string

// Claims represents the JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONB)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// Value implements the driver.Valuer interface for CapabilityArray
func (ca CapabilityArray) Value() (driver.Value, error) {
	return json.Marshal(ca)
}

// Scan implements the sql.Scanner interface for CapabilityArray
func (ca *CapabilityArray) Scan(value interface{}) error {
	if value == nil {
		*ca = CapabilityArray{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into CapabilityArray", value)
	}

	return json.Unmarshal(bytes, ca)
}

// User represents a user in the system
type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FirstName    *string   `json:"first_name" db:"first_name"`
	LastName     *string   `json:"last_name" db:"last_name"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Agent represents an AI agent
type Agent struct {
	ID                  uuid.UUID `json:"id" db:"id"`
	UserID              uuid.UUID `json:"user_id" db:"user_id"`
	Name                string    `json:"name" db:"name"`
	Description         *string   `json:"description" db:"description"`
	Capabilities        JSONB     `json:"capabilities" db:"capabilities"`
	Personality         JSONB     `json:"personality" db:"personality"`
	Config              JSONB     `json:"config" db:"config"`
	Status              string    `json:"status" db:"status"`
	FrameworkPreference string    `json:"framework_preference" db:"framework_preference"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
}

// Tool represents a tool that agents can use
type Tool struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	Name               string    `json:"name" db:"name"`
	Description        *string   `json:"description" db:"description"`
	Category           *string   `json:"category" db:"category"`
	FunctionSchema     JSONB     `json:"function_schema" db:"function_schema"`
	ImplementationCode *string   `json:"implementation_code" db:"implementation_code"`
	IsActive           bool      `json:"is_active" db:"is_active"`
	Version            string    `json:"version" db:"version"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// Execution represents an agent execution
type Execution struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	AgentID         uuid.UUID  `json:"agent_id" db:"agent_id"`
	UserID          uuid.UUID  `json:"user_id" db:"user_id"`
	InputText       string     `json:"input_text" db:"input_text"`
	OutputText      *string    `json:"output_text" db:"output_text"`
	FrameworkUsed   *string    `json:"framework_used" db:"framework_used"`
	ToolsUsed       JSONB      `json:"tools_used" db:"tools_used"`
	ExecutionTimeMs *int       `json:"execution_time_ms" db:"execution_time_ms"`
	Status          string     `json:"status" db:"status"`
	ErrorMessage    *string    `json:"error_message" db:"error_message"`
	Metadata        JSONB      `json:"metadata" db:"metadata"`
	StartedAt       time.Time  `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time `json:"completed_at" db:"completed_at"`
}

// Memory represents agent memory
type Memory struct {
	ID              uuid.UUID `json:"id" db:"id"`
	AgentID         uuid.UUID `json:"agent_id" db:"agent_id"`
	MemoryType      string    `json:"memory_type" db:"memory_type"`
	Content         string    `json:"content" db:"content"`
	Metadata        JSONB     `json:"metadata" db:"metadata"`
	Embedding       []float32 `json:"embedding,omitempty" db:"embedding"`
	ImportanceScore float64   `json:"importance_score" db:"importance_score"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	AccessedAt      time.Time `json:"accessed_at" db:"accessed_at"`
}

// Session represents a user session
type Session struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	AgentID     uuid.UUID  `json:"agent_id" db:"agent_id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	SessionData JSONB      `json:"session_data" db:"session_data"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Request/Response DTOs

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// CreateAgentRequest represents an agent creation request
type CreateAgentRequest struct {
	Name                string                 `json:"name" binding:"required"`
	Description         string                 `json:"description"`
	Capabilities        []string               `json:"capabilities" binding:"required"`
	Personality         map[string]interface{} `json:"personality"`
	FrameworkPreference string                 `json:"framework_preference"`
}

// ExecuteAgentRequest represents an agent execution request
type ExecuteAgentRequest struct {
	InputText     string  `json:"input_text" binding:"required"`
	Framework     string  `json:"framework"`
	IncludeMemory bool    `json:"include_memory"`
	MaxTokens     int     `json:"max_tokens"`
	Temperature   float64 `json:"temperature"`
}

// ExecuteAgentResponse represents an agent execution response
type ExecuteAgentResponse struct {
	ExecutionID     string    `json:"execution_id"`
	OutputText      string    `json:"output_text"`
	ToolsUsed       []string  `json:"tools_used"`
	ExecutionTimeMs int       `json:"execution_time_ms"`
	FrameworkUsed   string    `json:"framework_used"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

// WorkingMemory represents working memory session data
type WorkingMemory struct {
	SessionID    string                 `json:"session_id"`
	AgentID      string                 `json:"agent_id"`
	Variables    map[string]interface{} `json:"variables"`
	Context      map[string]interface{} `json:"context"`
	LastActivity time.Time              `json:"last_activity"`
	ExpiresAt    time.Time              `json:"expires_at"`
}

// MemoryEntry represents a memory entry
type MemoryEntry struct {
	ID        string                 `json:"id"`
	AgentID   string                 `json:"agent_id"`
	Type      string                 `json:"type"`
	Content   map[string]interface{} `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
}

// SemanticMemoryEntry represents a semantic memory entry
type SemanticMemoryEntry struct {
	ID         string    `json:"id"`
	Content    string    `json:"content"`
	Embedding  []float32 `json:"embedding,omitempty"`
	Concepts   []string  `json:"concepts"`
	Importance float64   `json:"importance"`
	Framework  string    `json:"framework"`
	SourceType string    `json:"source_type"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
