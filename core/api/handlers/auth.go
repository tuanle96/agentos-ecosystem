package handlers

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/tuanle96/agentos-ecosystem/core/api/models"
)

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Check if user already exists
	var existingUserID string
	err := h.db.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&existingUserID)
	if err != sql.ErrNoRows {
		c.JSON(http.StatusConflict, gin.H{
			"error": "User with this email already exists",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	// Create user
	userID := uuid.New()
	var firstName, lastName *string
	if req.FirstName != "" {
		firstName = &req.FirstName
	}
	if req.LastName != "" {
		lastName = &req.LastName
	}

	_, err = h.db.Exec(`
		INSERT INTO users (id, email, password_hash, first_name, last_name, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, req.Email, string(hashedPassword), firstName, lastName, true)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user",
		})
		return
	}

	// Fetch created user
	user := &models.User{}
	err = h.db.QueryRow(`
		SELECT id, email, first_name, last_name, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch created user",
		})
		return
	}

	// Generate JWT token
	token, err := h.generateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusCreated, models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

// Login handles user authentication
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Fetch user by email
	user := &models.User{}
	var passwordHash string
	err := h.db.QueryRow(`
		SELECT id, email, password_hash, first_name, last_name, is_active, created_at, updated_at
		FROM users WHERE email = $1 AND is_active = true
	`, req.Email).Scan(
		&user.ID, &user.Email, &passwordHash, &user.FirstName, &user.LastName,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
		})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}

	// Generate JWT token
	token, err := h.generateJWT(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate token",
		})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		Token: token,
		User:  *user,
	})
}

// GetProfile returns the current user's profile
func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	user := &models.User{}
	err := h.db.QueryRow(`
		SELECT id, email, first_name, last_name, is_active, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.FirstName, &user.LastName,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile updates the current user's profile
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	var firstName, lastName *string
	if req.FirstName != "" {
		firstName = &req.FirstName
	}
	if req.LastName != "" {
		lastName = &req.LastName
	}

	_, err := h.db.Exec(`
		UPDATE users SET first_name = $1, last_name = $2, updated_at = NOW()
		WHERE id = $3
	`, firstName, lastName, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update profile",
		})
		return
	}

	// Return updated user
	h.GetProfile(c)
}
