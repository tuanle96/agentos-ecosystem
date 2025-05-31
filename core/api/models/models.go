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

// Tool Marketplace Models

// ToolMarketplace represents a tool in the marketplace
type ToolMarketplace struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	DeveloperID    uuid.UUID  `json:"developer_id" db:"developer_id"`
	Name           string     `json:"name" db:"name"`
	DisplayName    string     `json:"display_name" db:"display_name"`
	Description    string     `json:"description" db:"description"`
	Category       string     `json:"category" db:"category"`
	Tags           JSONB      `json:"tags" db:"tags"`
	Version        string     `json:"version" db:"version"`
	LatestVersion  string     `json:"latest_version" db:"latest_version"`
	FunctionSchema JSONB      `json:"function_schema" db:"function_schema"`
	SourceCode     string     `json:"source_code" db:"source_code"`
	Documentation  string     `json:"documentation" db:"documentation"`
	Examples       JSONB      `json:"examples" db:"examples"`
	Dependencies   JSONB      `json:"dependencies" db:"dependencies"`
	Requirements   JSONB      `json:"requirements" db:"requirements"`
	IsPublic       bool       `json:"is_public" db:"is_public"`
	IsVerified     bool       `json:"is_verified" db:"is_verified"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	DownloadCount  int        `json:"download_count" db:"download_count"`
	Rating         float64    `json:"rating" db:"rating"`
	RatingCount    int        `json:"rating_count" db:"rating_count"`
	SecurityStatus string     `json:"security_status" db:"security_status"`
	ValidationHash string     `json:"validation_hash" db:"validation_hash"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	PublishedAt    *time.Time `json:"published_at" db:"published_at"`
}

// ToolVersion represents a specific version of a tool
type ToolVersion struct {
	ID             uuid.UUID `json:"id" db:"id"`
	ToolID         uuid.UUID `json:"tool_id" db:"tool_id"`
	Version        string    `json:"version" db:"version"`
	ChangeLog      string    `json:"changelog" db:"changelog"`
	FunctionSchema JSONB     `json:"function_schema" db:"function_schema"`
	SourceCode     string    `json:"source_code" db:"source_code"`
	Dependencies   JSONB     `json:"dependencies" db:"dependencies"`
	IsStable       bool      `json:"is_stable" db:"is_stable"`
	SecurityStatus string    `json:"security_status" db:"security_status"`
	ValidationHash string    `json:"validation_hash" db:"validation_hash"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// ToolInstallation represents a user's tool installation
type ToolInstallation struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	ToolID        uuid.UUID  `json:"tool_id" db:"tool_id"`
	VersionID     uuid.UUID  `json:"version_id" db:"version_id"`
	Status        string     `json:"status" db:"status"`
	Configuration JSONB      `json:"configuration" db:"configuration"`
	InstalledAt   time.Time  `json:"installed_at" db:"installed_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	LastUsedAt    *time.Time `json:"last_used_at" db:"last_used_at"`
}

// ToolReview represents a user review of a tool
type ToolReview struct {
	ID        uuid.UUID `json:"id" db:"id"`
	ToolID    uuid.UUID `json:"tool_id" db:"tool_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Rating    int       `json:"rating" db:"rating"`
	Title     string    `json:"title" db:"title"`
	Comment   string    `json:"comment" db:"comment"`
	IsPublic  bool      `json:"is_public" db:"is_public"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ToolUsageStats represents tool usage statistics
type ToolUsageStats struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	ToolID         uuid.UUID  `json:"tool_id" db:"tool_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	ExecutionCount int        `json:"execution_count" db:"execution_count"`
	SuccessCount   int        `json:"success_count" db:"success_count"`
	ErrorCount     int        `json:"error_count" db:"error_count"`
	TotalTimeMs    int64      `json:"total_time_ms" db:"total_time_ms"`
	AverageTimeMs  float64    `json:"average_time_ms" db:"average_time_ms"`
	LastExecutedAt *time.Time `json:"last_executed_at" db:"last_executed_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// Tool Marketplace DTOs

// CreateToolRequest represents a tool creation request
type CreateToolRequest struct {
	Name           string                   `json:"name" binding:"required"`
	DisplayName    string                   `json:"display_name" binding:"required"`
	Description    string                   `json:"description" binding:"required"`
	Category       string                   `json:"category" binding:"required"`
	Tags           []string                 `json:"tags"`
	FunctionSchema map[string]interface{}   `json:"function_schema" binding:"required"`
	SourceCode     string                   `json:"source_code" binding:"required"`
	Documentation  string                   `json:"documentation"`
	Examples       []map[string]interface{} `json:"examples"`
	Dependencies   []string                 `json:"dependencies"`
	Requirements   map[string]interface{}   `json:"requirements"`
	IsPublic       bool                     `json:"is_public"`
}

// UpdateToolRequest represents a tool update request
type UpdateToolRequest struct {
	DisplayName   string                   `json:"display_name"`
	Description   string                   `json:"description"`
	Category      string                   `json:"category"`
	Tags          []string                 `json:"tags"`
	Documentation string                   `json:"documentation"`
	Examples      []map[string]interface{} `json:"examples"`
	IsPublic      bool                     `json:"is_public"`
}

// CreateToolVersionRequest represents a tool version creation request
type CreateToolVersionRequest struct {
	Version        string                 `json:"version" binding:"required"`
	ChangeLog      string                 `json:"changelog" binding:"required"`
	FunctionSchema map[string]interface{} `json:"function_schema" binding:"required"`
	SourceCode     string                 `json:"source_code" binding:"required"`
	Dependencies   []string               `json:"dependencies"`
	IsStable       bool                   `json:"is_stable"`
}

// InstallToolRequest represents a tool installation request
type InstallToolRequest struct {
	ToolID        string                 `json:"tool_id" binding:"required"`
	Version       string                 `json:"version"`
	Configuration map[string]interface{} `json:"configuration"`
}

// ToolSearchRequest represents a tool search request
type ToolSearchRequest struct {
	Query      string   `json:"query"`
	Category   string   `json:"category"`
	Tags       []string `json:"tags"`
	IsVerified bool     `json:"is_verified"`
	SortBy     string   `json:"sort_by"`
	SortOrder  string   `json:"sort_order"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
}

// CreateReviewRequest represents a tool review creation request
type CreateReviewRequest struct {
	Rating  int    `json:"rating" binding:"required,min=1,max=5"`
	Title   string `json:"title" binding:"required"`
	Comment string `json:"comment" binding:"required"`
}

// ToolSearchResponse represents a tool search response
type ToolSearchResponse struct {
	Tools      []ToolMarketplace `json:"tools"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	HasMore    bool              `json:"has_more"`
}

// ToolDetailsResponse represents a detailed tool response
type ToolDetailsResponse struct {
	Tool         ToolMarketplace   `json:"tool"`
	Versions     []ToolVersion     `json:"versions"`
	Reviews      []ToolReview      `json:"reviews"`
	UsageStats   ToolUsageStats    `json:"usage_stats"`
	IsInstalled  bool              `json:"is_installed"`
	Installation *ToolInstallation `json:"installation,omitempty"`
}

// ToolValidationResponse represents a tool validation response
type ToolValidationResponse struct {
	IsValid        bool     `json:"is_valid"`
	SecurityStatus string   `json:"security_status"`
	ValidationHash string   `json:"validation_hash"`
	Issues         []string `json:"issues"`
	Warnings       []string `json:"warnings"`
}
