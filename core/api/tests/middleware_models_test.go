package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tuanle96/agentos-ecosystem/core/api/middleware"
	"github.com/tuanle96/agentos-ecosystem/core/api/models"
)

// TestAuthMiddlewareValidToken tests auth middleware with valid token
func (suite *TestSuite) TestAuthMiddlewareValidToken() {
	// Create a test route with auth middleware
	router := gin.New()
	router.Use(middleware.AuthMiddleware("test-jwt-secret"))
	router.GET("/test", func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		userEmail := middleware.GetUserEmail(c)

		c.JSON(http.StatusOK, gin.H{
			"user_id":    userID,
			"user_email": userEmail,
			"message":    "authenticated",
		})
	})

	// Test with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser.Token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "authenticated", response["message"])
	assert.NotEmpty(suite.T(), response["user_id"])
	assert.NotEmpty(suite.T(), response["user_email"])
}

// TestAuthMiddlewareInvalidToken tests auth middleware with invalid token
func (suite *TestSuite) TestAuthMiddlewareInvalidToken() {
	router := gin.New()
	router.Use(middleware.AuthMiddleware("test-jwt-secret"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	// Test with invalid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "error")
}

// TestAuthMiddlewareMissingToken tests auth middleware without token
func (suite *TestSuite) TestAuthMiddlewareMissingToken() {
	router := gin.New()
	router.Use(middleware.AuthMiddleware("test-jwt-secret"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	// Test without token
	req := httptest.NewRequest("GET", "/test", nil)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "error")
	assert.Contains(suite.T(), response["error"].(string), "Authorization header required")
}

// TestAuthMiddlewareInvalidFormat tests auth middleware with invalid header format
func (suite *TestSuite) TestAuthMiddlewareInvalidFormat() {
	router := gin.New()
	router.Use(middleware.AuthMiddleware("test-jwt-secret"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	// Test with invalid header format (missing Bearer)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat "+suite.testUser.Token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "error")
	assert.Contains(suite.T(), response["error"].(string), "Invalid authorization header format")
}

// TestAuthMiddlewareExpiredToken tests auth middleware with expired token
func (suite *TestSuite) TestAuthMiddlewareExpiredToken() {
	router := gin.New()
	router.Use(middleware.AuthMiddleware("test-jwt-secret"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	// Use a known expired token (this would be a token with past expiration)
	expiredToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzkwMjJ9.invalid"

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// TestGetUserIDFromContext tests getting user ID from context
func (suite *TestSuite) TestGetUserIDFromContext() {
	router := gin.New()
	router.Use(middleware.AuthMiddleware("test-jwt-secret"))
	router.GET("/test", func(c *gin.Context) {
		userID := middleware.GetUserID(c)

		// Test that we can get a valid user ID
		assert.NotEmpty(suite.T(), userID)
		assert.IsType(suite.T(), "", userID) // Should be string, not uint

		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser.Token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestGetUserEmailFromContext tests getting user email from context
func (suite *TestSuite) TestGetUserEmailFromContext() {
	router := gin.New()
	router.Use(middleware.AuthMiddleware("test-jwt-secret"))
	router.GET("/test", func(c *gin.Context) {
		userEmail := middleware.GetUserEmail(c)

		// Test that we can get a valid user email
		assert.NotEmpty(suite.T(), userEmail)
		assert.Contains(suite.T(), userEmail, "@")

		c.JSON(http.StatusOK, gin.H{
			"user_email": userEmail,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser.Token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// TestCapabilityArrayValue tests CapabilityArray Value method
func (suite *TestSuite) TestCapabilityArrayValue() {
	capabilities := models.CapabilityArray{"web_search", "calculations", "text_processing"}

	value, err := capabilities.Value()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), value)

	// Should be JSON bytes
	jsonBytes, ok := value.([]byte)
	assert.True(suite.T(), ok)
	jsonStr := string(jsonBytes)
	assert.Contains(suite.T(), jsonStr, "web_search")
	assert.Contains(suite.T(), jsonStr, "calculations")
	assert.Contains(suite.T(), jsonStr, "text_processing")
}

// TestCapabilityArrayScan tests CapabilityArray Scan method
func (suite *TestSuite) TestCapabilityArrayScan() {
	var capabilities models.CapabilityArray

	// Test scanning from JSON string
	jsonData := `["web_search", "calculations", "text_processing"]`
	err := capabilities.Scan(jsonData)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), capabilities, 3)
	assert.Contains(suite.T(), capabilities, "web_search")
	assert.Contains(suite.T(), capabilities, "calculations")
	assert.Contains(suite.T(), capabilities, "text_processing")
}

// TestCapabilityArrayScanBytes tests CapabilityArray Scan method with bytes
func (suite *TestSuite) TestCapabilityArrayScanBytes() {
	var capabilities models.CapabilityArray

	// Test scanning from byte slice
	jsonData := []byte(`["api_calls", "file_operations"]`)
	err := capabilities.Scan(jsonData)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), capabilities, 2)
	assert.Contains(suite.T(), capabilities, "api_calls")
	assert.Contains(suite.T(), capabilities, "file_operations")
}

// TestCapabilityArrayScanInvalidJSON tests CapabilityArray Scan with invalid JSON
func (suite *TestSuite) TestCapabilityArrayScanInvalidJSON() {
	var capabilities models.CapabilityArray

	// Test scanning from invalid JSON
	invalidJSON := `["web_search", "calculations"` // Missing closing bracket
	err := capabilities.Scan(invalidJSON)
	assert.Error(suite.T(), err)
}

// TestCapabilityArrayScanNil tests CapabilityArray Scan with nil value
func (suite *TestSuite) TestCapabilityArrayScanNil() {
	var capabilities models.CapabilityArray

	// Test scanning from nil
	err := capabilities.Scan(nil)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), capabilities, 0)
}

// TestCapabilityArrayScanUnsupportedType tests CapabilityArray Scan with unsupported type
func (suite *TestSuite) TestCapabilityArrayScanUnsupportedType() {
	var capabilities models.CapabilityArray

	// Test scanning from unsupported type (int)
	err := capabilities.Scan(123)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "cannot scan")
}

// TestCapabilityArrayEmpty tests CapabilityArray with empty array
func (suite *TestSuite) TestCapabilityArrayEmpty() {
	capabilities := models.CapabilityArray{}

	value, err := capabilities.Value()
	assert.NoError(suite.T(), err)

	jsonBytes, ok := value.([]byte)
	assert.True(suite.T(), ok)
	jsonStr := string(jsonBytes)
	assert.Equal(suite.T(), "[]", jsonStr)
}

// TestCapabilityArrayRoundTrip tests Value and Scan round trip
func (suite *TestSuite) TestCapabilityArrayRoundTrip() {
	original := models.CapabilityArray{"web_search", "calculations", "text_processing", "api_calls"}

	// Convert to value
	value, err := original.Value()
	assert.NoError(suite.T(), err)

	// Scan back
	var scanned models.CapabilityArray
	err = scanned.Scan(value)
	assert.NoError(suite.T(), err)

	// Should be equal
	assert.Equal(suite.T(), original, scanned)
	assert.Len(suite.T(), scanned, 4)

	for _, cap := range original {
		assert.Contains(suite.T(), scanned, cap)
	}
}

// TestAuthMiddlewareWithCORS tests auth middleware with CORS headers
func (suite *TestSuite) TestAuthMiddlewareWithCORS() {
	router := gin.New()

	// Add CORS middleware first
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	})

	router.Use(middleware.AuthMiddleware("test-jwt-secret"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test OPTIONS request (preflight)
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "*", w.Header().Get("Access-Control-Allow-Origin"))

	// Test actual request with auth
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+suite.testUser.Token)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}
