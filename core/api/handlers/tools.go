package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tuanle96/agentos-ecosystem/core/api/models"
)

// ListTools returns all available tools
func (h *Handler) ListTools(c *gin.Context) {
	rows, err := h.db.Query(`
		SELECT id, name, description, category, function_schema, is_active, version, created_at, updated_at
		FROM tools WHERE is_active = true ORDER BY category, name
	`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
		})
		return
	}
	defer rows.Close()

	var tools []models.Tool
	for rows.Next() {
		var tool models.Tool
		var schemaJSON []byte

		err := rows.Scan(
			&tool.ID, &tool.Name, &tool.Description, &tool.Category,
			&schemaJSON, &tool.IsActive, &tool.Version,
			&tool.CreatedAt, &tool.UpdatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scan tool",
			})
			return
		}

		// Parse JSON schema
		json.Unmarshal(schemaJSON, &tool.FunctionSchema)
		tools = append(tools, tool)
	}

	// Group tools by category
	toolsByCategory := make(map[string][]models.Tool)
	for _, tool := range tools {
		category := "general"
		if tool.Category != nil {
			category = *tool.Category
		}
		toolsByCategory[category] = append(toolsByCategory[category], tool)
	}

	c.JSON(http.StatusOK, gin.H{
		"tools":             tools,
		"tools_by_category": toolsByCategory,
		"count":             len(tools),
	})
}
