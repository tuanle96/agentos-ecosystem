package tests

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tuanle96/agentos-ecosystem/core/api/models"
)

// TestCapabilityArrayModelFunctions tests CapabilityArray model functions
func TestCapabilityArrayModelFunctions(t *testing.T) {
	// Test Value method
	capabilities := models.CapabilityArray{"web_search", "calculations", "text_processing"}

	value, err := capabilities.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Test Scan method with valid JSON
	var scannedCapabilities models.CapabilityArray
	err = scannedCapabilities.Scan([]byte(`["web_search", "calculations"]`))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(scannedCapabilities))
	assert.Equal(t, "web_search", scannedCapabilities[0])
	assert.Equal(t, "calculations", scannedCapabilities[1])

	// Test Scan method with string
	var scannedCapabilities2 models.CapabilityArray
	err = scannedCapabilities2.Scan(`["text_processing", "file_operations"]`)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(scannedCapabilities2))

	// Test Scan method with nil
	var scannedCapabilities3 models.CapabilityArray
	err = scannedCapabilities3.Scan(nil)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(scannedCapabilities3))

	// Test Scan method with invalid JSON
	var scannedCapabilities4 models.CapabilityArray
	err = scannedCapabilities4.Scan(`{"invalid":"json"`)
	assert.Error(t, err)

	// Test Scan method with unsupported type
	var scannedCapabilities5 models.CapabilityArray
	err = scannedCapabilities5.Scan(123)
	assert.Error(t, err)
}

// TestJSONBModelFunctions tests JSONB model functions
func TestJSONBModelFunctions(t *testing.T) {
	// Test Value method with valid data
	jsonMap := models.JSONB{
		"key1": "value1",
		"key2": 123,
		"key3": map[string]interface{}{
			"nested": "value",
		},
	}

	value, err := jsonMap.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Test Value method with nil
	var nilMap models.JSONB
	value, err = nilMap.Value()
	assert.NoError(t, err)
	// Value returns []byte, not string
	assert.Equal(t, []byte("null"), value)

	// Test Scan method with valid JSON bytes
	var scannedMap models.JSONB
	err = scannedMap.Scan([]byte(`{"test": "value", "number": 42}`))
	assert.NoError(t, err)
	assert.Equal(t, "value", scannedMap["test"])
	assert.Equal(t, float64(42), scannedMap["number"]) // JSON numbers are float64

	// Test Scan method with valid JSON string
	var scannedMap2 models.JSONB
	err = scannedMap2.Scan(`{"string_test": "hello", "bool_test": true}`)
	assert.NoError(t, err)
	assert.Equal(t, "hello", scannedMap2["string_test"])
	assert.Equal(t, true, scannedMap2["bool_test"])

	// Test Scan method with nil
	var scannedMap3 models.JSONB
	err = scannedMap3.Scan(nil)
	assert.NoError(t, err)
	assert.NotNil(t, scannedMap3) // JSONB initializes empty map, not nil

	// Test Scan method with invalid JSON
	var scannedMap4 models.JSONB
	err = scannedMap4.Scan(`{"invalid":"json"`)
	assert.Error(t, err)

	// Test Scan method with unsupported type
	var scannedMap5 models.JSONB
	err = scannedMap5.Scan(123)
	assert.Error(t, err)
}

// TestModelEdgeCases tests edge cases for model functions
func TestModelEdgeCases(t *testing.T) {
	// Test CapabilityArray with empty array
	emptyCapabilities := models.CapabilityArray{}
	value, err := emptyCapabilities.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Test CapabilityArray scan with empty JSON array
	var scannedEmpty models.CapabilityArray
	err = scannedEmpty.Scan(`[]`)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(scannedEmpty))

	// Test JSONB with complex nested structure
	complexMap := models.JSONB{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": "deep_value",
				"array":  []interface{}{1, 2, 3},
			},
		},
		"simple": "value",
	}

	value, err = complexMap.Value()
	assert.NoError(t, err)
	assert.NotNil(t, value)

	// Test scanning complex JSON
	var scannedComplex models.JSONB
	complexJSON := `{
		"metadata": {
			"version": "1.0",
			"features": ["search", "calc"]
		},
		"config": {
			"timeout": 30,
			"enabled": true
		}
	}`

	err = scannedComplex.Scan(complexJSON)
	assert.NoError(t, err)
	assert.Contains(t, scannedComplex, "metadata")
	assert.Contains(t, scannedComplex, "config")

	// Verify nested structure
	metadata, ok := scannedComplex["metadata"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "1.0", metadata["version"])

	config, ok := scannedComplex["config"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(30), config["timeout"])
	assert.Equal(t, true, config["enabled"])
}

// TestDriverValueInterface tests driver.Value interface compliance
func TestDriverValueInterface(t *testing.T) {
	// Test CapabilityArray implements driver.Valuer
	capabilities := models.CapabilityArray{"test"}
	var _ driver.Valuer = capabilities

	// Test JSONB implements driver.Valuer
	jsonMap := models.JSONB{"test": "value"}
	var _ driver.Valuer = jsonMap

	// Test that Value() returns driver.Value compatible types
	capValue, err := capabilities.Value()
	assert.NoError(t, err)

	// driver.Value should be one of: int64, float64, bool, []byte, string, time.Time, or nil
	switch capValue.(type) {
	case int64, float64, bool, []byte, string, nil:
		// Valid driver.Value type
	default:
		t.Errorf("CapabilityArray.Value() returned invalid driver.Value type: %T", capValue)
	}

	mapValue, err := jsonMap.Value()
	assert.NoError(t, err)

	switch mapValue.(type) {
	case int64, float64, bool, []byte, string, nil:
		// Valid driver.Value type
	default:
		t.Errorf("JSONB.Value() returned invalid driver.Value type: %T", mapValue)
	}
}

// TestModelRoundTrip tests complete round-trip serialization/deserialization
func TestModelRoundTrip(t *testing.T) {
	// Test CapabilityArray round trip
	originalCaps := models.CapabilityArray{"web_search", "calculations", "text_processing", "api_calls"}

	// Serialize
	value, err := originalCaps.Value()
	assert.NoError(t, err)

	// Deserialize
	var roundTripCaps models.CapabilityArray
	err = roundTripCaps.Scan(value)
	assert.NoError(t, err)

	// Verify
	assert.Equal(t, len(originalCaps), len(roundTripCaps))
	for i, cap := range originalCaps {
		assert.Equal(t, cap, roundTripCaps[i])
	}

	// Test JSONB round trip
	originalMap := models.JSONB{
		"string_field":  "test_value",
		"number_field":  42.5,
		"boolean_field": true,
		"array_field":   []interface{}{"item1", "item2"},
		"object_field": map[string]interface{}{
			"nested": "value",
		},
	}

	// Serialize
	value, err = originalMap.Value()
	assert.NoError(t, err)

	// Deserialize
	var roundTripMap models.JSONB
	err = roundTripMap.Scan(value)
	assert.NoError(t, err)

	// Verify basic fields
	assert.Equal(t, "test_value", roundTripMap["string_field"])
	assert.Equal(t, 42.5, roundTripMap["number_field"])
	assert.Equal(t, true, roundTripMap["boolean_field"])

	// Verify array field
	arrayField, ok := roundTripMap["array_field"].([]interface{})
	assert.True(t, ok)
	assert.Equal(t, 2, len(arrayField))

	// Verify nested object
	objectField, ok := roundTripMap["object_field"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value", objectField["nested"])
}
