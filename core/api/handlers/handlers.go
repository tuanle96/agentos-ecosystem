package handlers

import (
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/tuanle96/agentos-ecosystem/core/api/config"
	"github.com/tuanle96/agentos-ecosystem/core/api/models"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	db     *sql.DB
	redis  *redis.Client
	config *config.Config
}

// New creates a new handler instance
func New(db *sql.DB, redis *redis.Client, cfg *config.Config) *Handler {
	return &Handler{
		db:     db,
		redis:  redis,
		config: cfg,
	}
}

// generateJWT generates a JWT token for a user
func (h *Handler) generateJWT(user *models.User) (string, error) {
	claims := &models.Claims{
		UserID: user.ID.String(),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.config.JWTSecret))
}
