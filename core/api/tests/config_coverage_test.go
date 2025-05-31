package tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tuanle96/agentos-ecosystem/core/api/config"
)

// TestConfigCoverage contains tests specifically for config functions
// These tests target the config.go functions to improve coverage from 50% to 100%

// TestConfigLoad tests the Load function
func TestConfigLoad(t *testing.T) {
	// Save original environment variables
	originalDatabaseURL := os.Getenv("DATABASE_URL")
	originalRedisURL := os.Getenv("REDIS_URL")
	originalJWTSecret := os.Getenv("JWT_SECRET")
	originalPort := os.Getenv("PORT")
	originalGoEnv := os.Getenv("GO_ENV")

	// Clean up after test
	defer func() {
		os.Setenv("DATABASE_URL", originalDatabaseURL)
		os.Setenv("REDIS_URL", originalRedisURL)
		os.Setenv("JWT_SECRET", originalJWTSecret)
		os.Setenv("PORT", originalPort)
		os.Setenv("GO_ENV", originalGoEnv)
	}()

	// Test with custom environment variables
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test_db")
	os.Setenv("REDIS_URL", "localhost:6380")
	os.Setenv("JWT_SECRET", "test-jwt-secret-key")
	os.Setenv("CORE_API_PORT", "8081")
	os.Setenv("GO_ENV", "test")

	cfg := config.Load()

	assert.Equal(t, "postgres://test:test@localhost:5432/test_db", cfg.DatabaseURL)
	assert.Equal(t, "localhost:6380", cfg.RedisURL)
	assert.Equal(t, "test-jwt-secret-key", cfg.JWTSecret)
	assert.Equal(t, "8081", cfg.Port)
	assert.Equal(t, "test", cfg.Environment)
}

// TestConfigLoadDefaults tests the Load function with default values
func TestConfigLoadDefaults(t *testing.T) {
	// Clear environment variables to test defaults
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("CORE_API_PORT")
	os.Unsetenv("GO_ENV")

	cfg := config.Load()

	// Test default values
	assert.Equal(t, "postgres://agentos:agentos_dev_password@localhost:5432/agentos_dev?sslmode=disable", cfg.DatabaseURL)
	assert.Equal(t, "localhost:6379", cfg.RedisURL)
	assert.Equal(t, "dev-jwt-secret-change-in-production", cfg.JWTSecret)
	assert.Equal(t, "8000", cfg.Port)
	assert.Equal(t, "development", cfg.Environment)
}

// TestConfigGetEnv tests the getEnv function
func TestConfigGetEnv(t *testing.T) {
	// Test with existing environment variable
	os.Setenv("TEST_ENV_VAR", "test_value")
	defer os.Unsetenv("TEST_ENV_VAR")

	// This tests the internal getEnv function through Load()
	// Since getEnv is not exported, we test it indirectly
	os.Setenv("DATABASE_URL", "test_database_url")
	defer os.Unsetenv("DATABASE_URL")

	cfg := config.Load()
	assert.Equal(t, "test_database_url", cfg.DatabaseURL)

	// Test with non-existing environment variable (should use default)
	os.Unsetenv("DATABASE_URL")
	cfg = config.Load()
	assert.Equal(t, "postgres://agentos:agentos_dev_password@localhost:5432/agentos_dev?sslmode=disable", cfg.DatabaseURL)
}

// TestConfigGetEnvAsInt tests the getEnvAsInt function
func TestConfigGetEnvAsInt(t *testing.T) {
	// Test with valid integer environment variable
	os.Setenv("API_RATE_LIMIT", "5000")
	defer os.Unsetenv("API_RATE_LIMIT")

	// Test through the RateLimit configuration
	cfg := config.Load()
	assert.Equal(t, 5000, cfg.RateLimit)

	// Test with invalid integer environment variable (should use default)
	os.Setenv("API_RATE_LIMIT", "invalid_number")
	defer os.Unsetenv("API_RATE_LIMIT")

	cfg = config.Load()
	assert.Equal(t, 10000, cfg.RateLimit) // Should fall back to default

	// Test with empty environment variable (should use default)
	os.Setenv("API_RATE_LIMIT", "")
	cfg = config.Load()
	assert.Equal(t, 10000, cfg.RateLimit) // Should fall back to default
}

// TestConfigGetEnvAsIntEdgeCases tests edge cases for getEnvAsInt
func TestConfigGetEnvAsIntEdgeCases(t *testing.T) {
	// Test with zero value
	os.Setenv("API_RATE_LIMIT", "0")
	defer os.Unsetenv("API_RATE_LIMIT")

	cfg := config.Load()
	assert.Equal(t, 0, cfg.RateLimit)

	// Test with negative value
	os.Setenv("API_RATE_LIMIT", "-1")
	cfg = config.Load()
	assert.Equal(t, -1, cfg.RateLimit)

	// Test with very large number
	os.Setenv("API_RATE_LIMIT", "100000")
	cfg = config.Load()
	assert.Equal(t, 100000, cfg.RateLimit)

	// Test with floating point number (should fail and use default)
	os.Setenv("API_RATE_LIMIT", "5000.5")
	cfg = config.Load()
	assert.Equal(t, 10000, cfg.RateLimit) // Should fall back to default
}

// TestConfigEnvironmentSpecific tests environment-specific configurations
func TestConfigEnvironmentSpecific(t *testing.T) {
	// Test production environment
	os.Setenv("GO_ENV", "production")
	defer os.Unsetenv("GO_ENV")

	cfg := config.Load()
	assert.Equal(t, "production", cfg.Environment)

	// Test development environment
	os.Setenv("GO_ENV", "development")
	cfg = config.Load()
	assert.Equal(t, "development", cfg.Environment)

	// Test test environment
	os.Setenv("GO_ENV", "test")
	cfg = config.Load()
	assert.Equal(t, "test", cfg.Environment)
}

// TestConfigValidation tests configuration validation
func TestConfigValidation(t *testing.T) {
	// Test with empty JWT secret (should still load but might be insecure)
	os.Setenv("JWT_SECRET", "")
	defer os.Unsetenv("JWT_SECRET")

	cfg := config.Load()
	assert.Equal(t, "dev-jwt-secret-change-in-production", cfg.JWTSecret) // Should use default

	// Test with very short JWT secret
	os.Setenv("JWT_SECRET", "abc")
	cfg = config.Load()
	assert.Equal(t, "abc", cfg.JWTSecret) // Should accept any value

	// Test with very long JWT secret
	longSecret := "this-is-a-very-long-jwt-secret-key-that-should-still-work-fine-for-testing-purposes"
	os.Setenv("JWT_SECRET", longSecret)
	cfg = config.Load()
	assert.Equal(t, longSecret, cfg.JWTSecret)
}

// TestConfigDatabaseURL tests database URL configurations
func TestConfigDatabaseURL(t *testing.T) {
	// Test with different database URL formats
	testCases := []string{
		"postgres://user:pass@localhost:5432/db",
		"postgresql://user:pass@localhost:5432/db?sslmode=require",
		"postgres://user@localhost/db",
		"postgres://localhost:5432/db",
	}

	for _, testURL := range testCases {
		os.Setenv("DATABASE_URL", testURL)
		cfg := config.Load()
		assert.Equal(t, testURL, cfg.DatabaseURL)
	}

	os.Unsetenv("DATABASE_URL")
}

// TestConfigRedisURL tests Redis URL configurations
func TestConfigRedisURL(t *testing.T) {
	// Test with different Redis URL formats
	testCases := []string{
		"localhost:6379",
		"redis://localhost:6379",
		"redis://user:pass@localhost:6379/0",
		"127.0.0.1:6380",
	}

	for _, testURL := range testCases {
		os.Setenv("REDIS_URL", testURL)
		cfg := config.Load()
		assert.Equal(t, testURL, cfg.RedisURL)
	}

	os.Unsetenv("REDIS_URL")
}

// TestConfigPortRange tests port range validation
func TestConfigPortRange(t *testing.T) {
	// Test with common port numbers
	testPorts := []struct {
		envValue string
		expected string
	}{
		{"80", "80"},
		{"443", "443"},
		{"3000", "3000"},
		{"8000", "8000"},
		{"8080", "8080"},
		{"9000", "9000"},
	}

	for _, testCase := range testPorts {
		os.Setenv("CORE_API_PORT", testCase.envValue)
		cfg := config.Load()
		assert.Equal(t, testCase.expected, cfg.Port)
	}

	os.Unsetenv("CORE_API_PORT")
}

// TestConfigConcurrency tests concurrent config loading
func TestConfigConcurrency(t *testing.T) {
	// Test that config loading is thread-safe
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			cfg := config.Load()
			assert.NotNil(t, cfg)
			assert.NotEmpty(t, cfg.DatabaseURL)
			assert.NotEmpty(t, cfg.RedisURL)
			assert.NotEmpty(t, cfg.JWTSecret)
			assert.NotEmpty(t, cfg.Port)
			assert.NotEmpty(t, cfg.Environment)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestConfigStructFields tests that all config struct fields are properly set
func TestConfigStructFields(t *testing.T) {
	cfg := config.Load()

	// Test that all fields are set to non-zero values
	assert.NotEmpty(t, cfg.DatabaseURL)
	assert.NotEmpty(t, cfg.RedisURL)
	assert.NotEmpty(t, cfg.JWTSecret)
	assert.NotEmpty(t, cfg.Port)
	assert.NotEmpty(t, cfg.Environment)

	// Test field types
	assert.IsType(t, "", cfg.DatabaseURL)
	assert.IsType(t, "", cfg.RedisURL)
	assert.IsType(t, "", cfg.JWTSecret)
	assert.IsType(t, "", cfg.Port)
	assert.IsType(t, "", cfg.Environment)
}
