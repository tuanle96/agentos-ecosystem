package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Memory System Implementation - Week 4: Advanced Memory System
type MemoryEntry struct {
	ID        string                 `json:"id"`
	AgentID   string                 `json:"agent_id"`
	Type      string                 `json:"type"` // working, episodic, semantic
	Content   map[string]interface{} `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	TTL       int                    `json:"ttl"` // seconds
}

// Week 4: Semantic Memory Structures
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

type MemoryConsolidation struct {
	ID                 string     `json:"id"`
	Framework          string     `json:"framework"`
	EpisodicCount      int        `json:"episodic_count"`
	SemanticCount      int        `json:"semantic_count"`
	ConsolidationScore float64    `json:"consolidation_score"`
	StartedAt          time.Time  `json:"started_at"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	Status             string     `json:"status"` // pending, running, completed, failed
}

type MemoryLink struct {
	ID             string    `json:"id"`
	SourceMemoryID string    `json:"source_memory_id"`
	TargetMemoryID string    `json:"target_memory_id"`
	LinkType       string    `json:"link_type"`
	Strength       float64   `json:"strength"`
	CreatedAt      time.Time `json:"created_at"`
}

type WorkingMemory struct {
	SessionID    string                 `json:"session_id"`
	AgentID      string                 `json:"agent_id"`
	Context      map[string]interface{} `json:"context"`
	Variables    map[string]interface{} `json:"variables"`
	LastActivity time.Time              `json:"last_activity"`
	ExpiresAt    time.Time              `json:"expires_at"`
}

// Redis keys for memory management
const (
	WorkingMemoryPrefix  = "agentos:memory:working:"
	EpisodicMemoryPrefix = "agentos:memory:episodic:"
	SemanticMemoryPrefix = "agentos:memory:semantic:"
	SessionPrefix        = "agentos:session:"
)

// GetAgentMemoryEnhanced returns agent's memory with Week 2 enhancements
func (h *Handler) GetAgentMemoryEnhanced(c *gin.Context) {
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

	// Get working memory from Redis
	workingMemory, err := h.getWorkingMemory(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve working memory",
		})
		return
	}

	// Get episodic memories from database
	episodicMemories, err := h.getEpisodicMemories(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve episodic memories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"agent_id":          agentID,
		"working_memory":    workingMemory,
		"episodic_memories": episodicMemories,
		"memory_stats": gin.H{
			"working_memory_size": len(workingMemory.Variables),
			"episodic_count":      len(episodicMemories),
			"last_activity":       workingMemory.LastActivity,
		},
	})
}

// ClearAgentMemoryEnhanced clears agent's memory with Week 2 enhancements
func (h *Handler) ClearAgentMemoryEnhanced(c *gin.Context) {
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

	// Clear working memory from Redis
	err := h.clearWorkingMemory(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear working memory",
		})
		return
	}

	// Clear episodic memories from database
	err = h.clearEpisodicMemories(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear episodic memories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Agent memory cleared successfully",
		"agent_id":   agentID,
		"cleared_at": time.Now(),
	})
}

// Week 2 Enhancement: Working Memory Management
func (h *Handler) CreateWorkingMemorySession(c *gin.Context) {
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

	// Create new working memory session
	sessionID := uuid.New().String()
	workingMemory := WorkingMemory{
		SessionID:    sessionID,
		AgentID:      agentID,
		Context:      make(map[string]interface{}),
		Variables:    make(map[string]interface{}),
		LastActivity: time.Now(),
		ExpiresAt:    time.Now().Add(24 * time.Hour), // 24 hour TTL
	}

	// Store in Redis
	err := h.storeWorkingMemory(agentID, workingMemory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create working memory session",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id": sessionID,
		"agent_id":   agentID,
		"expires_at": workingMemory.ExpiresAt,
		"message":    "Working memory session created",
	})
}

// UpdateWorkingMemory updates working memory variables
func (h *Handler) UpdateWorkingMemory(c *gin.Context) {
	userID := c.GetString("user_id")
	agentID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req struct {
		Variables map[string]interface{} `json:"variables"`
		Context   map[string]interface{} `json:"context"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// Get existing working memory
	workingMemory, err := h.getWorkingMemory(agentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve working memory",
		})
		return
	}

	// Update variables and context
	if req.Variables != nil {
		for key, value := range req.Variables {
			workingMemory.Variables[key] = value
		}
	}

	if req.Context != nil {
		for key, value := range req.Context {
			workingMemory.Context[key] = value
		}
	}

	workingMemory.LastActivity = time.Now()

	// Store updated memory
	err = h.storeWorkingMemory(agentID, workingMemory)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update working memory",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Working memory updated",
		"agent_id":        agentID,
		"session_id":      workingMemory.SessionID,
		"variables_count": len(workingMemory.Variables),
		"context_count":   len(workingMemory.Context),
	})
}

// Helper functions for memory management
func (h *Handler) validateAgentOwnership(agentID, userID string) bool {
	var count int
	err := h.db.QueryRow(`
		SELECT COUNT(*) FROM agents WHERE id = $1 AND user_id = $2
	`, agentID, userID).Scan(&count)

	return err == nil && count > 0
}

func (h *Handler) getWorkingMemory(agentID string) (WorkingMemory, error) {
	ctx := context.Background()
	key := WorkingMemoryPrefix + agentID

	val, err := h.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		// Return empty working memory if not found
		return WorkingMemory{
			SessionID:    uuid.New().String(),
			AgentID:      agentID,
			Context:      make(map[string]interface{}),
			Variables:    make(map[string]interface{}),
			LastActivity: time.Now(),
			ExpiresAt:    time.Now().Add(24 * time.Hour),
		}, nil
	}
	if err != nil {
		return WorkingMemory{}, err
	}

	var workingMemory WorkingMemory
	err = json.Unmarshal([]byte(val), &workingMemory)
	return workingMemory, err
}

func (h *Handler) storeWorkingMemory(agentID string, memory WorkingMemory) error {
	ctx := context.Background()
	key := WorkingMemoryPrefix + agentID

	data, err := json.Marshal(memory)
	if err != nil {
		return err
	}

	// Store with TTL
	ttl := time.Until(memory.ExpiresAt)
	return h.redis.Set(ctx, key, data, ttl).Err()
}

func (h *Handler) clearWorkingMemory(agentID string) error {
	ctx := context.Background()
	key := WorkingMemoryPrefix + agentID
	return h.redis.Del(ctx, key).Err()
}

func (h *Handler) getEpisodicMemories(agentID string) ([]MemoryEntry, error) {
	rows, err := h.db.Query(`
		SELECT id, content, created_at FROM memories
		WHERE agent_id = $1 AND type = 'episodic'
		ORDER BY created_at DESC LIMIT 50
	`, agentID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []MemoryEntry
	for rows.Next() {
		var memory MemoryEntry
		var contentJSON []byte

		err := rows.Scan(&memory.ID, &contentJSON, &memory.Timestamp)
		if err != nil {
			continue
		}

		json.Unmarshal(contentJSON, &memory.Content)
		memory.AgentID = agentID
		memory.Type = "episodic"

		memories = append(memories, memory)
	}

	return memories, nil
}

func (h *Handler) clearEpisodicMemories(agentID string) error {
	_, err := h.db.Exec(`
		DELETE FROM memories WHERE agent_id = $1 AND type = 'episodic'
	`, agentID)
	return err
}

// ===================================
// WEEK 4: MEM0-POWERED MEMORY API ENDPOINTS
// ===================================

// SemanticMemorySearch performs semantic search using mem0 engine
func (h *Handler) SemanticMemorySearch(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Query     string  `json:"query" binding:"required"`
		Framework string  `json:"framework,omitempty"`
		Limit     int     `json:"limit,omitempty"`
		Threshold float64 `json:"threshold,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Set defaults
	if req.Limit == 0 {
		req.Limit = 10
	}
	if req.Threshold == 0 {
		req.Threshold = 0.7
	}

	// Call Python mem0 engine for semantic search
	memories, err := h.callMem0Search(userID, req.Query, req.Framework, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to perform mem0 search"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":     req.Query,
		"memories":  memories,
		"count":     len(memories),
		"threshold": req.Threshold,
		"engine":    "mem0",
	})
}

// StoreSemanticMemory stores a new semantic memory using mem0
func (h *Handler) StoreSemanticMemory(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Content    string   `json:"content" binding:"required"`
		Concepts   []string `json:"concepts"`
		Framework  string   `json:"framework"`
		SourceType string   `json:"source_type"`
		Importance float64  `json:"importance"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Set defaults
	if req.Importance == 0 {
		req.Importance = 0.5
	}
	if req.SourceType == "" {
		req.SourceType = "user_input"
	}
	if req.Framework == "" {
		req.Framework = "universal"
	}

	// Prepare metadata for mem0
	metadata := map[string]interface{}{
		"concepts":    req.Concepts,
		"source_type": req.SourceType,
		"importance":  req.Importance,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	// Store using mem0 engine
	memoryID, err := h.callMem0Store(userID, req.Content, req.Framework, metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store memory via mem0"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"memory_id":  memoryID,
		"content":    req.Content,
		"concepts":   req.Concepts,
		"framework":  req.Framework,
		"importance": req.Importance,
		"engine":     "mem0",
		"created_at": time.Now(),
	})
}

// ===================================
// WEEK 4: SEMANTIC MEMORY HELPER FUNCTIONS
// ===================================

// generateEmbedding generates vector embedding for text (placeholder implementation)
func (h *Handler) generateEmbedding(text string) []float32 {
	// Placeholder implementation - in production would use OpenAI embeddings or similar
	// For now, generate a simple hash-based embedding
	embedding := make([]float32, 1536) // OpenAI embedding dimension

	// Simple hash-based embedding for testing
	hash := 0
	for _, char := range text {
		hash = hash*31 + int(char)
	}

	for i := range embedding {
		embedding[i] = float32((hash+i)%1000) / 1000.0
	}

	return embedding
}

// performSemanticSearch performs vector similarity search
func (h *Handler) performSemanticSearch(userID string, queryEmbedding []float32, framework string, limit int, threshold float64) ([]SemanticMemoryEntry, error) {
	query := `
		SELECT id, content, embedding, concepts, importance, framework, source_type, created_at, updated_at
		FROM semantic_memories sm
		JOIN agents a ON sm.framework = a.framework_preference OR sm.framework = 'universal'
		WHERE a.user_id = $1
	`

	args := []interface{}{userID}

	if framework != "" {
		query += " AND sm.framework = $2"
		args = append(args, framework)
	}

	query += " ORDER BY sm.importance DESC, sm.created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []SemanticMemoryEntry
	for rows.Next() {
		var memory SemanticMemoryEntry
		var embeddingBytes []byte
		var conceptsJSON []byte

		err := rows.Scan(
			&memory.ID,
			&memory.Content,
			&embeddingBytes,
			&conceptsJSON,
			&memory.Importance,
			&memory.Framework,
			&memory.SourceType,
			&memory.CreatedAt,
			&memory.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// Parse concepts JSON
		if len(conceptsJSON) > 0 {
			json.Unmarshal(conceptsJSON, &memory.Concepts)
		}

		// Calculate similarity (placeholder - would use actual vector similarity)
		similarity := h.calculateSimilarity(queryEmbedding, memory.Embedding)
		if similarity >= threshold {
			memories = append(memories, memory)
		}
	}

	return memories, nil
}

// storeSemanticMemory stores a semantic memory entry
func (h *Handler) storeSemanticMemory(userID, content string, embedding []float32, concepts []string, framework, sourceType string, importance float64) (string, error) {
	memoryID := uuid.New().String()

	conceptsJSON, _ := json.Marshal(concepts)

	_, err := h.db.Exec(`
		INSERT INTO semantic_memories (id, content, embedding, concepts, importance, framework, source_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`, memoryID, content, embedding, conceptsJSON, importance, framework, sourceType)

	return memoryID, err
}

// calculateSimilarity calculates cosine similarity between embeddings (placeholder)
func (h *Handler) calculateSimilarity(embedding1, embedding2 []float32) float64 {
	if len(embedding1) != len(embedding2) {
		return 0.0
	}

	var dotProduct, norm1, norm2 float64

	for i := range embedding1 {
		dotProduct += float64(embedding1[i] * embedding2[i])
		norm1 += float64(embedding1[i] * embedding1[i])
		norm2 += float64(embedding2[i] * embedding2[i])
	}

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// ===================================
// WEEK 4: MEMORY CONSOLIDATION API ENDPOINTS
// ===================================

// GetConsolidationStatus returns the status of memory consolidation
func (h *Handler) GetConsolidationStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	framework := c.Query("framework")
	if framework == "" {
		framework = "universal"
	}

	// Get recent consolidations
	consolidations, err := h.getRecentConsolidations(userID, framework)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get consolidation status"})
		return
	}

	// Calculate consolidation metrics
	metrics := h.calculateConsolidationMetrics(consolidations)

	c.JSON(http.StatusOK, gin.H{
		"framework":      framework,
		"consolidations": consolidations,
		"metrics":        metrics,
		"status":         "operational",
	})
}

// TriggerMemoryConsolidation triggers memory consolidation for a framework
func (h *Handler) TriggerMemoryConsolidation(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		Framework       string  `json:"framework" binding:"required"`
		TimeWindowHours float64 `json:"time_window_hours"`
		ForceRun        bool    `json:"force_run"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Set default time window
	if req.TimeWindowHours == 0 {
		req.TimeWindowHours = 24.0
	}

	// Check if consolidation is already running
	if !req.ForceRun {
		running, err := h.isConsolidationRunning(userID, req.Framework)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check consolidation status"})
			return
		}
		if running {
			c.JSON(http.StatusConflict, gin.H{"error": "Consolidation already running for this framework"})
			return
		}
	}

	// Create consolidation record
	consolidationID := uuid.New().String()
	err := h.createConsolidationRecord(consolidationID, userID, req.Framework, req.TimeWindowHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create consolidation record"})
		return
	}

	// Trigger mem0 consolidation
	consolidationResult, err := h.callMem0Consolidate(userID, req.Framework)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to trigger mem0 consolidation"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"consolidation_id":  consolidationID,
		"framework":         req.Framework,
		"time_window_hours": req.TimeWindowHours,
		"status":            "completed",
		"engine":            "mem0",
		"result":            consolidationResult,
		"started_at":        time.Now(),
	})
}

// GetFrameworkMemory returns framework-specific memory information
func (h *Handler) GetFrameworkMemory(c *gin.Context) {
	userID := c.GetString("user_id")
	framework := c.Param("framework")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Validate framework
	validFrameworks := []string{"langchain", "swarms", "crewai", "autogen", "universal"}
	isValid := false
	for _, valid := range validFrameworks {
		if framework == valid {
			isValid = true
			break
		}
	}

	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid framework"})
		return
	}

	// Get framework memory statistics
	stats, err := h.getFrameworkMemoryStats(userID, framework)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get framework memory"})
		return
	}

	// Get recent memories
	recentMemories, err := h.getRecentFrameworkMemories(userID, framework, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent memories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"framework":       framework,
		"statistics":      stats,
		"recent_memories": recentMemories,
		"last_updated":    time.Now(),
	})
}

// ===================================
// WEEK 4: MEMORY CONSOLIDATION HELPER FUNCTIONS
// ===================================

// getRecentConsolidations retrieves recent consolidation records
func (h *Handler) getRecentConsolidations(userID, framework string) ([]MemoryConsolidation, error) {
	query := `
		SELECT mc.id, mc.framework, mc.episodic_count, mc.semantic_count,
		       mc.consolidation_score, mc.started_at, mc.completed_at, mc.status
		FROM memory_consolidations mc
		JOIN agents a ON mc.framework = a.framework_preference OR mc.framework = 'universal'
		WHERE a.user_id = $1
	`

	args := []interface{}{userID}

	if framework != "universal" {
		query += " AND mc.framework = $2"
		args = append(args, framework)
	}

	query += " ORDER BY mc.started_at DESC LIMIT 10"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var consolidations []MemoryConsolidation
	for rows.Next() {
		var consolidation MemoryConsolidation
		var completedAt sql.NullTime

		err := rows.Scan(
			&consolidation.ID,
			&consolidation.Framework,
			&consolidation.EpisodicCount,
			&consolidation.SemanticCount,
			&consolidation.ConsolidationScore,
			&consolidation.StartedAt,
			&completedAt,
			&consolidation.Status,
		)
		if err != nil {
			continue
		}

		if completedAt.Valid {
			consolidation.CompletedAt = &completedAt.Time
		}

		consolidations = append(consolidations, consolidation)
	}

	return consolidations, nil
}

// calculateConsolidationMetrics calculates metrics from consolidation history
func (h *Handler) calculateConsolidationMetrics(consolidations []MemoryConsolidation) map[string]interface{} {
	if len(consolidations) == 0 {
		return map[string]interface{}{
			"total_consolidations": 0,
			"avg_score":            0.0,
			"success_rate":         0.0,
			"last_consolidation":   nil,
		}
	}

	totalScore := 0.0
	successCount := 0
	var lastConsolidation *MemoryConsolidation

	for i, consolidation := range consolidations {
		totalScore += consolidation.ConsolidationScore
		if consolidation.Status == "completed" {
			successCount++
		}
		if i == 0 {
			lastConsolidation = &consolidation
		}
	}

	return map[string]interface{}{
		"total_consolidations": len(consolidations),
		"avg_score":            totalScore / float64(len(consolidations)),
		"success_rate":         float64(successCount) / float64(len(consolidations)),
		"last_consolidation":   lastConsolidation,
	}
}

// isConsolidationRunning checks if consolidation is currently running
func (h *Handler) isConsolidationRunning(userID, framework string) (bool, error) {
	var count int
	err := h.db.QueryRow(`
		SELECT COUNT(*) FROM memory_consolidations mc
		JOIN agents a ON mc.framework = a.framework_preference OR mc.framework = 'universal'
		WHERE a.user_id = $1 AND mc.framework = $2 AND mc.status IN ('pending', 'running')
	`, userID, framework).Scan(&count)

	return count > 0, err
}

// createConsolidationRecord creates a new consolidation record
func (h *Handler) createConsolidationRecord(consolidationID, userID, framework string, timeWindowHours float64) error {
	_, err := h.db.Exec(`
		INSERT INTO memory_consolidations (id, framework, started_at, status, metadata)
		VALUES ($1, $2, NOW(), 'pending', $3)
	`, consolidationID, framework, fmt.Sprintf(`{"user_id": "%s", "time_window_hours": %f}`, userID, timeWindowHours))

	return err
}

// performMemoryConsolidation performs async memory consolidation
func (h *Handler) performMemoryConsolidation(consolidationID, userID, framework string, timeWindowHours float64) {
	// Update status to running
	h.db.Exec(`
		UPDATE memory_consolidations
		SET status = 'running'
		WHERE id = $1
	`, consolidationID)

	// Simulate consolidation process (in production, would call Python consolidation engine)
	time.Sleep(2 * time.Second) // Simulate processing time

	// Mock consolidation results
	episodicCount := 15
	semanticCount := 3
	consolidationScore := 0.75

	// Update consolidation record with results
	_, err := h.db.Exec(`
		UPDATE memory_consolidations
		SET episodic_count = $1, semantic_count = $2, consolidation_score = $3,
		    completed_at = NOW(), status = 'completed'
		WHERE id = $4
	`, episodicCount, semanticCount, consolidationScore, consolidationID)

	if err != nil {
		// Mark as failed
		h.db.Exec(`
			UPDATE memory_consolidations
			SET status = 'failed', error_message = $1
			WHERE id = $2
		`, err.Error(), consolidationID)
	}
}

// getFrameworkMemoryStats gets memory statistics for a framework
func (h *Handler) getFrameworkMemoryStats(userID, framework string) (map[string]interface{}, error) {
	// Get semantic memory count
	var semanticCount int
	err := h.db.QueryRow(`
		SELECT COUNT(*) FROM semantic_memories sm
		JOIN agents a ON sm.framework = a.framework_preference OR sm.framework = 'universal'
		WHERE a.user_id = $1 AND (sm.framework = $2 OR $2 = 'universal')
	`, userID, framework).Scan(&semanticCount)
	if err != nil {
		semanticCount = 0
	}

	// Get episodic memory count
	var episodicCount int
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM memories m
		JOIN agents a ON m.agent_id = a.id
		WHERE a.user_id = $1 AND m.memory_type = 'episodic' AND (a.framework_preference = $2 OR $2 = 'universal')
	`, userID, framework).Scan(&episodicCount)
	if err != nil {
		episodicCount = 0
	}

	// Get average importance
	var avgImportance float64
	err = h.db.QueryRow(`
		SELECT COALESCE(AVG(importance), 0.0) FROM semantic_memories sm
		JOIN agents a ON sm.framework = a.framework_preference OR sm.framework = 'universal'
		WHERE a.user_id = $1 AND (sm.framework = $2 OR $2 = 'universal')
	`, userID, framework).Scan(&avgImportance)
	if err != nil {
		avgImportance = 0.0
	}

	// Get recent consolidation count
	var recentConsolidations int
	err = h.db.QueryRow(`
		SELECT COUNT(*) FROM memory_consolidations mc
		JOIN agents a ON mc.framework = a.framework_preference OR mc.framework = 'universal'
		WHERE a.user_id = $1 AND mc.framework = $2 AND mc.started_at > NOW() - INTERVAL '7 days'
	`, userID, framework).Scan(&recentConsolidations)
	if err != nil {
		recentConsolidations = 0
	}

	return map[string]interface{}{
		"semantic_memories":     semanticCount,
		"episodic_memories":     episodicCount,
		"total_memories":        semanticCount + episodicCount,
		"avg_importance":        avgImportance,
		"recent_consolidations": recentConsolidations,
		"framework":             framework,
	}, nil
}

// getRecentFrameworkMemories gets recent memories for a framework
func (h *Handler) getRecentFrameworkMemories(userID, framework string, limit int) ([]SemanticMemoryEntry, error) {
	query := `
		SELECT sm.id, sm.content, sm.concepts, sm.importance, sm.framework,
		       sm.source_type, sm.created_at, sm.updated_at
		FROM semantic_memories sm
		JOIN agents a ON sm.framework = a.framework_preference OR sm.framework = 'universal'
		WHERE a.user_id = $1 AND (sm.framework = $2 OR $2 = 'universal')
		ORDER BY sm.created_at DESC LIMIT $3
	`

	rows, err := h.db.Query(query, userID, framework, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []SemanticMemoryEntry
	for rows.Next() {
		var memory SemanticMemoryEntry
		var conceptsJSON []byte

		err := rows.Scan(
			&memory.ID,
			&memory.Content,
			&conceptsJSON,
			&memory.Importance,
			&memory.Framework,
			&memory.SourceType,
			&memory.CreatedAt,
			&memory.UpdatedAt,
		)
		if err != nil {
			continue
		}

		// Parse concepts JSON
		if len(conceptsJSON) > 0 {
			json.Unmarshal(conceptsJSON, &memory.Concepts)
		}

		memories = append(memories, memory)
	}

	return memories, nil
}

// ===================================
// WEEK 4: MEM0 INTEGRATION HELPER FUNCTIONS
// ===================================

// callMem0Search calls Python mem0 engine for semantic search
func (h *Handler) callMem0Search(userID, query, framework string, limit int) ([]interface{}, error) {
	// Real implementation: Call Python mem0 service via HTTP
	client := &http.Client{Timeout: 10 * time.Second}

	requestBody := map[string]interface{}{
		"query":     query,
		"user_id":   userID,
		"framework": framework,
		"limit":     limit,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Call Python mem0 service (assuming it runs on port 8001)
	resp, err := client.Post("http://localhost:8001/api/memory/search",
		"application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to call mem0 service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mem0 service returned status %d", resp.StatusCode)
	}

	var result struct {
		Memories []interface{} `json:"memories"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode mem0 response: %v", err)
	}

	return result.Memories, nil
}

// callMem0Store calls Python mem0 engine to store memory
func (h *Handler) callMem0Store(userID, content, framework string, metadata map[string]interface{}) (string, error) {
	// Real implementation: Call Python mem0 service via HTTP
	client := &http.Client{Timeout: 10 * time.Second}

	requestBody := map[string]interface{}{
		"content":   content,
		"user_id":   userID,
		"framework": framework,
		"metadata":  metadata,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	// Call Python mem0 service
	resp, err := client.Post("http://localhost:8001/api/memory/store",
		"application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to call mem0 service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("mem0 service returned status %d", resp.StatusCode)
	}

	var result struct {
		MemoryID string `json:"memory_id"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode mem0 response: %v", err)
	}

	// Also store in local database for backup
	h.storeMemoryBackup(userID, content, framework, result.MemoryID, metadata)

	return result.MemoryID, nil
}

// storeMemoryBackup stores memory in local database as backup
func (h *Handler) storeMemoryBackup(userID, content, framework, memoryID string, metadata map[string]interface{}) error {
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		metadataJSON = []byte("{}")
	}

	_, err = h.db.Exec(`
		INSERT INTO memories (id, user_id, content, framework, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO NOTHING
	`, memoryID, userID, content, framework, string(metadataJSON), time.Now())

	return err
}

// callMem0Consolidate calls Python mem0 engine for memory consolidation
func (h *Handler) callMem0Consolidate(userID, framework string) (map[string]interface{}, error) {
	// Real implementation: Call Python mem0 service via HTTP
	client := &http.Client{Timeout: 30 * time.Second} // Longer timeout for consolidation

	requestBody := map[string]interface{}{
		"user_id":   userID,
		"framework": framework,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	// Call Python mem0 service
	resp, err := client.Post("http://localhost:8001/api/memory/consolidate",
		"application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to call mem0 service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("mem0 service returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode mem0 response: %v", err)
	}

	return result, nil
}
