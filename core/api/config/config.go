package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the API service
type Config struct {
	Environment   string
	Port          string
	DatabaseURL   string
	RedisURL      string
	NatsURL       string
	JWTSecret     string
	OpenAIAPIKey  string
	RateLimit     int
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Environment:   getEnv("GO_ENV", "development"),
		Port:          getEnv("CORE_API_PORT", "8000"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://agentos:agentos_dev_password@localhost:5432/agentos_dev?sslmode=disable"),
		RedisURL:      getEnv("REDIS_URL", "localhost:6379"),
		NatsURL:       getEnv("NATS_URL", "nats://localhost:4222"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-jwt-secret-change-in-production"),
		OpenAIAPIKey:  getEnv("OPENAI_API_KEY", ""),
		RateLimit:     getEnvAsInt("API_RATE_LIMIT", 10000),
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}
