package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tuanle96/agentos-ecosystem/core/api/config"
	"github.com/tuanle96/agentos-ecosystem/core/api/models"
)

// TestDatabaseConnection tests database connection and basic operations
func TestDatabaseConnection(t *testing.T) {
	// Load config
	cfg := config.Load()
	assert.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.DatabaseURL)

	// Test database connection would be here
	// For now, we test the config loading which is part of database setup
	assert.Contains(t, cfg.DatabaseURL, "postgres")
}

// TestConfigurationLoading tests configuration loading for database
func TestConfigurationLoading(t *testing.T) {
	// Test config loading
	cfg := config.Load()

	// Verify required database fields
	assert.NotEmpty(t, cfg.DatabaseURL)
	assert.NotEmpty(t, cfg.JWTSecret)
	assert.Greater(t, cfg.Port, 0)

	// Test environment variable handling
	assert.NotNil(t, cfg)
}

// TestCapabilityArrayOperations tests CapabilityArray operations
func TestCapabilityArrayOperations(t *testing.T) {
	// Test empty array
	emptyArray := models.CapabilityArray{}
	assert.Equal(t, 0, len(emptyArray))

	// Test array with capabilities
	capabilities := models.CapabilityArray{"web_search", "calculations", "text_processing"}
	assert.Equal(t, 3, len(capabilities))

	// Test contains functionality (if implemented)
	assert.Equal(t, "web_search", capabilities[0])
	assert.Equal(t, "calculations", capabilities[1])
	assert.Equal(t, "text_processing", capabilities[2])

	// Test serialization
	value, err := capabilities.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Test deserialization
	var newCapabilities models.CapabilityArray
	err = newCapabilities.Scan(value)
	assert.NoError(t, err)
	assert.Equal(t, len(capabilities), len(newCapabilities))

	for i, cap := range capabilities {
		assert.Equal(t, cap, newCapabilities[i])
	}
}

// TestJSONBBasicOperations tests basic JSONB operations
func TestJSONBBasicOperations(t *testing.T) {
	// Test empty JSONB
	emptyJSON := models.JSONB{}
	value, err := emptyJSON.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Test JSONB with simple data
	jsonData := models.JSONB{
		"string_field":  "test_value",
		"number_field":  42,
		"boolean_field": true,
	}

	// Test serialization
	value, err = jsonData.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Test deserialization
	var newJSONData models.JSONB
	err = newJSONData.Scan(value)
	assert.NoError(t, err)

	// Verify data integrity
	assert.Equal(t, "test_value", newJSONData["string_field"])
	assert.Equal(t, float64(42), newJSONData["number_field"]) // JSON numbers are float64
	assert.Equal(t, true, newJSONData["boolean_field"])
}

// TestJSONBOperations tests JSONB operations
func TestJSONBOperations(t *testing.T) {
	// Test empty JSONB
	emptyJSON := models.JSONB{}
	value, err := emptyJSON.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Test JSONB with data
	jsonData := models.JSONB{
		"string_field":  "test_value",
		"number_field":  42,
		"boolean_field": true,
		"array_field":   []interface{}{"item1", "item2"},
		"object_field": map[string]interface{}{
			"nested_key": "nested_value",
		},
	}

	// Test serialization
	value, err = jsonData.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Test deserialization
	var newJSONData models.JSONB
	err = newJSONData.Scan(value)
	assert.NoError(t, err)

	// Verify data integrity
	assert.Equal(t, "test_value", newJSONData["string_field"])
	assert.Equal(t, float64(42), newJSONData["number_field"]) // JSON numbers are float64
	assert.Equal(t, true, newJSONData["boolean_field"])

	// Verify array field
	arrayField, ok := newJSONData["array_field"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(arrayField))

	// Verify nested object
	objectField, ok := newJSONData["object_field"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "nested_value", objectField["nested_key"])
}
