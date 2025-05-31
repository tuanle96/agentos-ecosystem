package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tuanle96/agentos-ecosystem/core/api/models"
)

// CreateTool creates a new tool in the marketplace
func (h *Handler) CreateTool(c *gin.Context) {
	var req models.CreateToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse user ID to UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Create tool marketplace entry
	tool := models.ToolMarketplace{
		ID:             uuid.New(),
		DeveloperID:    userID,
		Name:           req.Name,
		DisplayName:    req.DisplayName,
		Description:    req.Description,
		Category:       req.Category,
		Version:        "1.0.0",
		LatestVersion:  "1.0.0",
		IsPublic:       req.IsPublic,
		IsVerified:     false,
		IsActive:       true,
		DownloadCount:  0,
		Rating:         0.0,
		RatingCount:    0,
		SecurityStatus: "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Convert tags to JSONB
	if req.Tags != nil {
		tagsJSON, _ := json.Marshal(req.Tags)
		tool.Tags = models.JSONB{}
		json.Unmarshal(tagsJSON, &tool.Tags)
	}

	// Convert function schema to JSONB
	schemaJSON, _ := json.Marshal(req.FunctionSchema)
	tool.FunctionSchema = models.JSONB{}
	json.Unmarshal(schemaJSON, &tool.FunctionSchema)

	// Convert examples to JSONB
	if req.Examples != nil {
		examplesJSON, _ := json.Marshal(req.Examples)
		tool.Examples = models.JSONB{}
		json.Unmarshal(examplesJSON, &tool.Examples)
	}

	// Convert dependencies to JSONB
	if req.Dependencies != nil {
		depsJSON, _ := json.Marshal(req.Dependencies)
		tool.Dependencies = models.JSONB{}
		json.Unmarshal(depsJSON, &tool.Dependencies)
	}

	// Convert requirements to JSONB
	if req.Requirements != nil {
		reqsJSON, _ := json.Marshal(req.Requirements)
		tool.Requirements = models.JSONB{}
		json.Unmarshal(reqsJSON, &tool.Requirements)
	}

	tool.SourceCode = req.SourceCode
	tool.Documentation = req.Documentation

	// Insert into database
	query := `
		INSERT INTO tool_marketplace (
			id, developer_id, name, display_name, description, category, tags,
			version, latest_version, function_schema, source_code, documentation,
			examples, dependencies, requirements, is_public, is_verified, is_active,
			download_count, rating, rating_count, security_status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24
		)`

	_, err = h.db.Exec(query,
		tool.ID, tool.DeveloperID, tool.Name, tool.DisplayName, tool.Description,
		tool.Category, tool.Tags, tool.Version, tool.LatestVersion, tool.FunctionSchema,
		tool.SourceCode, tool.Documentation, tool.Examples, tool.Dependencies,
		tool.Requirements, tool.IsPublic, tool.IsVerified, tool.IsActive,
		tool.DownloadCount, tool.Rating, tool.RatingCount, tool.SecurityStatus,
		tool.CreatedAt, tool.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tool"})
		return
	}

	// Create initial version
	version := models.ToolVersion{
		ID:             uuid.New(),
		ToolID:         tool.ID,
		Version:        "1.0.0",
		ChangeLog:      "Initial version",
		FunctionSchema: tool.FunctionSchema,
		SourceCode:     tool.SourceCode,
		Dependencies:   tool.Dependencies,
		IsStable:       true,
		SecurityStatus: "pending",
		CreatedAt:      time.Now(),
	}

	versionQuery := `
		INSERT INTO tool_versions (
			id, tool_id, version, changelog, function_schema, source_code,
			dependencies, is_stable, security_status, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = h.db.Exec(versionQuery,
		version.ID, version.ToolID, version.Version, version.ChangeLog,
		version.FunctionSchema, version.SourceCode, version.Dependencies,
		version.IsStable, version.SecurityStatus, version.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tool version"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tool created successfully",
		"tool":    tool,
	})
}

// GetTools retrieves tools from the marketplace with search and filtering
func (h *Handler) GetTools(c *gin.Context) {
	// Parse query parameters
	query := c.Query("query")
	category := c.Query("category")
	isVerified := c.Query("is_verified") == "true"
	sortBy := c.DefaultQuery("sort_by", "created_at")
	sortOrder := c.DefaultQuery("sort_order", "desc")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	// Build SQL query
	baseQuery := `
		SELECT id, developer_id, name, display_name, description, category, tags,
			   version, latest_version, function_schema, source_code, documentation,
			   examples, dependencies, requirements, is_public, is_verified, is_active,
			   download_count, rating, rating_count, security_status, created_at, updated_at, published_at
		FROM tool_marketplace
		WHERE is_active = true AND is_public = true`

	countQuery := `SELECT COUNT(*) FROM tool_marketplace WHERE is_active = true AND is_public = true`

	args := []interface{}{}
	argIndex := 1

	// Add filters
	if query != "" {
		baseQuery += fmt.Sprintf(" AND (name ILIKE $%d OR display_name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex, argIndex)
		countQuery += fmt.Sprintf(" AND (name ILIKE $%d OR display_name ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex, argIndex)
		args = append(args, "%"+query+"%")
		argIndex++
	}

	if category != "" {
		baseQuery += fmt.Sprintf(" AND category = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, category)
		argIndex++
	}

	if isVerified {
		baseQuery += " AND is_verified = true"
		countQuery += " AND is_verified = true"
	}

	// Add sorting
	validSortFields := map[string]bool{
		"created_at":     true,
		"updated_at":     true,
		"name":           true,
		"rating":         true,
		"download_count": true,
	}

	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	baseQuery += fmt.Sprintf(" ORDER BY %s %s LIMIT $%d OFFSET $%d", sortBy, sortOrder, argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute queries
	var totalCount int
	err := h.db.QueryRow(countQuery, args[:len(args)-2]...).Scan(&totalCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count tools"})
		return
	}

	rows, err := h.db.Query(baseQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tools"})
		return
	}
	defer rows.Close()

	var tools []models.ToolMarketplace
	for rows.Next() {
		var tool models.ToolMarketplace
		err := rows.Scan(
			&tool.ID, &tool.DeveloperID, &tool.Name, &tool.DisplayName, &tool.Description,
			&tool.Category, &tool.Tags, &tool.Version, &tool.LatestVersion, &tool.FunctionSchema,
			&tool.SourceCode, &tool.Documentation, &tool.Examples, &tool.Dependencies,
			&tool.Requirements, &tool.IsPublic, &tool.IsVerified, &tool.IsActive,
			&tool.DownloadCount, &tool.Rating, &tool.RatingCount, &tool.SecurityStatus,
			&tool.CreatedAt, &tool.UpdatedAt, &tool.PublishedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan tool"})
			return
		}
		tools = append(tools, tool)
	}

	response := models.ToolSearchResponse{
		Tools:      tools,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
		HasMore:    offset+len(tools) < totalCount,
	}

	c.JSON(http.StatusOK, response)
}

// GetTool retrieves a specific tool by ID with detailed information
func (h *Handler) GetTool(c *gin.Context) {
	toolID := c.Param("id")
	if toolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tool ID is required"})
		return
	}

	toolUUID, err := uuid.Parse(toolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tool ID format"})
		return
	}

	// Get tool details
	var tool models.ToolMarketplace
	query := `
		SELECT id, developer_id, name, display_name, description, category, tags,
			   version, latest_version, function_schema, source_code, documentation,
			   examples, dependencies, requirements, is_public, is_verified, is_active,
			   download_count, rating, rating_count, security_status, created_at, updated_at, published_at
		FROM tool_marketplace WHERE id = $1 AND is_active = true`

	err = h.db.QueryRow(query, toolUUID).Scan(
		&tool.ID, &tool.DeveloperID, &tool.Name, &tool.DisplayName, &tool.Description,
		&tool.Category, &tool.Tags, &tool.Version, &tool.LatestVersion, &tool.FunctionSchema,
		&tool.SourceCode, &tool.Documentation, &tool.Examples, &tool.Dependencies,
		&tool.Requirements, &tool.IsPublic, &tool.IsVerified, &tool.IsActive,
		&tool.DownloadCount, &tool.Rating, &tool.RatingCount, &tool.SecurityStatus,
		&tool.CreatedAt, &tool.UpdatedAt, &tool.PublishedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tool not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tool"})
		return
	}

	// Get tool versions
	versionsQuery := `
		SELECT id, tool_id, version, changelog, function_schema, source_code,
			   dependencies, is_stable, security_status, created_at
		FROM tool_versions WHERE tool_id = $1 ORDER BY created_at DESC`

	versionRows, err := h.db.Query(versionsQuery, toolUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tool versions"})
		return
	}
	defer versionRows.Close()

	var versions []models.ToolVersion
	for versionRows.Next() {
		var version models.ToolVersion
		err := versionRows.Scan(
			&version.ID, &version.ToolID, &version.Version, &version.ChangeLog,
			&version.FunctionSchema, &version.SourceCode, &version.Dependencies,
			&version.IsStable, &version.SecurityStatus, &version.CreatedAt,
		)
		if err != nil {
			continue
		}
		versions = append(versions, version)
	}

	response := models.ToolDetailsResponse{
		Tool:        tool,
		Versions:    versions,
		Reviews:     []models.ToolReview{},   // TODO: Implement reviews
		UsageStats:  models.ToolUsageStats{}, // TODO: Implement usage stats
		IsInstalled: false,                   // TODO: Check if user has installed this tool
	}

	c.JSON(http.StatusOK, response)
}

// UpdateTool updates an existing tool
func (h *Handler) UpdateTool(c *gin.Context) {
	toolID := c.Param("id")
	if toolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tool ID is required"})
		return
	}

	toolUUID, err := uuid.Parse(toolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tool ID format"})
		return
	}

	var req models.UpdateToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse user ID to UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if user owns the tool
	var developerID uuid.UUID
	checkQuery := `SELECT developer_id FROM tool_marketplace WHERE id = $1 AND is_active = true`
	err = h.db.QueryRow(checkQuery, toolUUID).Scan(&developerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tool not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check tool ownership"})
		return
	}

	if developerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update this tool"})
		return
	}

	// Build update query dynamically
	updateFields := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.DisplayName != "" {
		updateFields = append(updateFields, fmt.Sprintf("display_name = $%d", argIndex))
		args = append(args, req.DisplayName)
		argIndex++
	}

	if req.Description != "" {
		updateFields = append(updateFields, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, req.Description)
		argIndex++
	}

	if req.Category != "" {
		updateFields = append(updateFields, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, req.Category)
		argIndex++
	}

	if req.Tags != nil {
		tagsJSON, _ := json.Marshal(req.Tags)
		var tags models.JSONB
		json.Unmarshal(tagsJSON, &tags)
		updateFields = append(updateFields, fmt.Sprintf("tags = $%d", argIndex))
		args = append(args, tags)
		argIndex++
	}

	if req.Documentation != "" {
		updateFields = append(updateFields, fmt.Sprintf("documentation = $%d", argIndex))
		args = append(args, req.Documentation)
		argIndex++
	}

	if req.Examples != nil {
		examplesJSON, _ := json.Marshal(req.Examples)
		var examples models.JSONB
		json.Unmarshal(examplesJSON, &examples)
		updateFields = append(updateFields, fmt.Sprintf("examples = $%d", argIndex))
		args = append(args, examples)
		argIndex++
	}

	updateFields = append(updateFields, fmt.Sprintf("is_public = $%d", argIndex))
	args = append(args, req.IsPublic)
	argIndex++

	updateFields = append(updateFields, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	if len(updateFields) == 2 { // Only updated_at and is_public
		c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
		return
	}

	// Add tool ID as last parameter
	args = append(args, toolUUID)

	updateQuery := fmt.Sprintf("UPDATE tool_marketplace SET %s WHERE id = $%d",
		strings.Join(updateFields, ", "), argIndex)

	_, err = h.db.Exec(updateQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tool"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tool updated successfully"})
}

// DeleteTool soft deletes a tool
func (h *Handler) DeleteTool(c *gin.Context) {
	toolID := c.Param("id")
	if toolID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tool ID is required"})
		return
	}

	toolUUID, err := uuid.Parse(toolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tool ID format"})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse user ID to UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check if user owns the tool
	var developerID uuid.UUID
	checkQuery := `SELECT developer_id FROM tool_marketplace WHERE id = $1 AND is_active = true`
	err = h.db.QueryRow(checkQuery, toolUUID).Scan(&developerID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tool not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check tool ownership"})
		return
	}

	if developerID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this tool"})
		return
	}

	// Soft delete the tool
	deleteQuery := `UPDATE tool_marketplace SET is_active = false, updated_at = $1 WHERE id = $2`
	_, err = h.db.Exec(deleteQuery, time.Now(), toolUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tool"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tool deleted successfully"})
}

// InstallTool installs a tool for a user
func (h *Handler) InstallTool(c *gin.Context) {
	var req models.InstallToolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse user ID to UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	toolUUID, err := uuid.Parse(req.ToolID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tool ID format"})
		return
	}

	// Check if tool exists and is public
	var tool models.ToolMarketplace
	toolQuery := `SELECT id, name, latest_version FROM tool_marketplace WHERE id = $1 AND is_active = true AND is_public = true`
	err = h.db.QueryRow(toolQuery, toolUUID).Scan(&tool.ID, &tool.Name, &tool.LatestVersion)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tool not found or not available"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check tool availability"})
		return
	}

	// Use latest version if not specified
	version := req.Version
	if version == "" {
		version = tool.LatestVersion
	}

	// Get version details
	var versionID uuid.UUID
	versionQuery := `SELECT id FROM tool_versions WHERE tool_id = $1 AND version = $2`
	err = h.db.QueryRow(versionQuery, toolUUID, version).Scan(&versionID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tool version not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check tool version"})
		return
	}

	// Check if already installed
	var existingID uuid.UUID
	checkQuery := `SELECT id FROM tool_installations WHERE user_id = $1 AND tool_id = $2`
	err = h.db.QueryRow(checkQuery, userID, toolUUID).Scan(&existingID)
	if err == nil {
		// Update existing installation
		updateQuery := `UPDATE tool_installations SET version_id = $1, status = 'installed', updated_at = $2 WHERE id = $3`
		_, err = h.db.Exec(updateQuery, versionID, time.Now(), existingID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tool installation"})
			return
		}
	} else if err == sql.ErrNoRows {
		// Create new installation
		installation := models.ToolInstallation{
			ID:          uuid.New(),
			UserID:      userID,
			ToolID:      toolUUID,
			VersionID:   versionID,
			Status:      "installed",
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Convert configuration to JSONB
		if req.Configuration != nil {
			configJSON, _ := json.Marshal(req.Configuration)
			installation.Configuration = models.JSONB{}
			json.Unmarshal(configJSON, &installation.Configuration)
		}

		installQuery := `
			INSERT INTO tool_installations (id, user_id, tool_id, version_id, status, configuration, installed_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

		_, err = h.db.Exec(installQuery,
			installation.ID, installation.UserID, installation.ToolID, installation.VersionID,
			installation.Status, installation.Configuration, installation.InstalledAt, installation.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to install tool"})
			return
		}
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing installation"})
		return
	}

	// Update download count
	updateCountQuery := `UPDATE tool_marketplace SET download_count = download_count + 1 WHERE id = $1`
	h.db.Exec(updateCountQuery, toolUUID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Tool installed successfully",
		"tool_id": toolUUID,
		"version": version,
	})
}
