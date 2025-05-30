package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Memory System Implementation - Week 2
type MemoryEntry struct {
	ID        string                 `json:"id"`
	AgentID   string                 `json:"agent_id"`
	Type      string                 `json:"type"` // working, episodic, semantic
	Content   map[string]interface{} `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	TTL       int                    `json:"ttl"` // seconds
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
