package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// BenchmarkHealthEndpoint benchmarks the health check endpoint
func BenchmarkHealthEndpoint(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "agentos-core-api",
			"version":   "0.1.0-mvp-week6-day2",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkHealthEndpointWithMiddleware benchmarks health endpoint with middleware
func BenchmarkHealthEndpointWithMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	
	// Add performance middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		c.Header("X-Response-Time", duration.String())
	})
	
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "agentos-core-api",
			"version":   "0.1.0-mvp-week6-day2",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkJSONSerialization benchmarks JSON serialization performance
func BenchmarkJSONSerialization(b *testing.B) {
	data := map[string]interface{}{
		"id":          "test-id-123",
		"name":        "Test Agent",
		"description": "This is a test agent for benchmarking",
		"capabilities": []string{"web_search", "text_processing", "calculations"},
		"framework":   "langchain",
		"status":      "active",
		"created_at":  time.Now(),
		"metadata": map[string]interface{}{
			"version": "1.0.0",
			"author":  "AgentOS",
			"tags":    []string{"test", "benchmark", "performance"},
		},
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := json.Marshal(data)
			if err != nil {
				b.Errorf("JSON marshal error: %v", err)
			}
		}
	})
}

// BenchmarkJSONDeserialization benchmarks JSON deserialization performance
func BenchmarkJSONDeserialization(b *testing.B) {
	data := map[string]interface{}{
		"id":          "test-id-123",
		"name":        "Test Agent",
		"description": "This is a test agent for benchmarking",
		"capabilities": []string{"web_search", "text_processing", "calculations"},
		"framework":   "langchain",
		"status":      "active",
		"created_at":  time.Now(),
		"metadata": map[string]interface{}{
			"version": "1.0.0",
			"author":  "AgentOS",
			"tags":    []string{"test", "benchmark", "performance"},
		},
	}
	
	jsonData, _ := json.Marshal(data)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var result map[string]interface{}
			err := json.Unmarshal(jsonData, &result)
			if err != nil {
				b.Errorf("JSON unmarshal error: %v", err)
			}
		}
	})
}

// BenchmarkConcurrentRequests benchmarks concurrent request handling
func BenchmarkConcurrentRequests(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.POST("/api/test", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		// Simulate some processing
		time.Sleep(time.Microsecond * 100)
		
		c.JSON(http.StatusOK, gin.H{
			"status": "success",
			"data":   data,
		})
	})

	requestData := map[string]interface{}{
		"test": "data",
		"id":   123,
	}
	jsonData, _ := json.Marshal(requestData)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("POST", "/api/test", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				b.Errorf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// BenchmarkContextOperations benchmarks context operations
func BenchmarkContextOperations(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
			
			// Simulate context usage
			select {
			case <-ctx.Done():
				b.Error("Context cancelled unexpectedly")
			default:
				// Context is still valid
			}
			
			cancel()
		}
	})
}

// BenchmarkStringOperations benchmarks string operations
func BenchmarkStringOperations(b *testing.B) {
	testStrings := []string{
		"short",
		"medium length string for testing",
		"this is a much longer string that we will use for benchmarking string operations in our AgentOS application",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, s := range testStrings {
				// Common string operations
				_ = len(s)
				_ = s + "_suffix"
				_ = s[:min(len(s), 10)]
			}
		}
	})
}

// BenchmarkMapOperations benchmarks map operations
func BenchmarkMapOperations(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m := make(map[string]interface{})
			
			// Add items
			for i := 0; i < 100; i++ {
				key := "key_" + string(rune(i))
				m[key] = i
			}
			
			// Read items
			for i := 0; i < 100; i++ {
				key := "key_" + string(rune(i))
				_ = m[key]
			}
			
			// Delete items
			for i := 0; i < 50; i++ {
				key := "key_" + string(rune(i))
				delete(m, key)
			}
		}
	})
}

// BenchmarkSliceOperations benchmarks slice operations
func BenchmarkSliceOperations(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slice := make([]int, 0, 100)
			
			// Append items
			for i := 0; i < 100; i++ {
				slice = append(slice, i)
			}
			
			// Access items
			for i := 0; i < len(slice); i++ {
				_ = slice[i]
			}
			
			// Slice operations
			_ = slice[:50]
			_ = slice[25:75]
		}
	})
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Allocate various data structures
			_ = make([]byte, 1024)
			_ = make(map[string]int, 10)
			_ = make([]string, 0, 10)
			
			// Allocate struct
			type TestStruct struct {
				ID   int
				Name string
				Data []byte
			}
			_ = &TestStruct{
				ID:   1,
				Name: "test",
				Data: make([]byte, 256),
			}
		}
	})
}

// TestPerformanceBaseline provides a baseline performance test
func TestPerformanceBaseline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		start := time.Now()
		c.JSON(http.StatusOK, gin.H{
			"status":       "healthy",
			"service":      "agentos-core-api",
			"version":      "0.1.0-mvp-week6-day2",
			"timestamp":    time.Now().Format(time.RFC3339),
			"response_time": time.Since(start).String(),
		})
	})

	// Test response time
	start := time.Now()
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	duration := time.Since(start)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Less(t, duration, time.Millisecond*5, "Response time should be less than 5ms")
	
	t.Logf("Health endpoint response time: %v", duration)
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
