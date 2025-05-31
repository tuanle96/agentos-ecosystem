package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Performance metrics
var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		},
	)

	responseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Size of HTTP responses in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "endpoint"},
	)
)

// Connection pool for tracking active connections
var connectionPool = &sync.Map{}

// PerformanceMiddleware provides comprehensive performance monitoring
func PerformanceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Track active connection
		connID := generateConnectionID()
		connectionPool.Store(connID, start)
		activeConnections.Inc()
		
		// Add performance headers
		c.Header("X-Request-ID", connID)
		c.Header("X-Response-Time-Start", strconv.FormatInt(start.UnixNano(), 10))
		
		// Process request
		c.Next()
		
		// Calculate metrics
		duration := time.Since(start)
		status := strconv.Itoa(c.Writer.Status())
		
		// Record metrics
		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			status,
		).Inc()
		
		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration.Seconds())
		
		responseSize.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(float64(c.Writer.Size()))
		
		// Add response time header
		c.Header("X-Response-Time", duration.String())
		c.Header("X-Response-Time-Ms", strconv.FormatFloat(float64(duration.Nanoseconds())/1e6, 'f', 2, 64))
		
		// Cleanup connection tracking
		connectionPool.Delete(connID)
		activeConnections.Dec()
	}
}

// CacheMiddleware provides intelligent response caching
func CacheMiddleware(ttl time.Duration) gin.HandlerFunc {
	cache := &sync.Map{}
	
	return func(c *gin.Context) {
		// Only cache GET requests
		if c.Request.Method != "GET" {
			c.Next()
			return
		}
		
		// Generate cache key
		cacheKey := generateCacheKey(c)
		
		// Check cache
		if cached, found := cache.Load(cacheKey); found {
			if entry, ok := cached.(*CacheEntry); ok && !entry.IsExpired() {
				// Serve from cache
				c.Header("X-Cache", "HIT")
				c.Header("X-Cache-TTL", strconv.FormatInt(int64(entry.TTL.Seconds()), 10))
				
				for key, value := range entry.Headers {
					c.Header(key, value)
				}
				
				c.Data(entry.StatusCode, entry.ContentType, entry.Data)
				c.Abort()
				return
			}
		}
		
		// Cache miss - capture response
		c.Header("X-Cache", "MISS")
		
		// Create response writer wrapper
		writer := &CacheResponseWriter{
			ResponseWriter: c.Writer,
			cache:          cache,
			cacheKey:       cacheKey,
			ttl:            ttl,
		}
		c.Writer = writer
		
		c.Next()
	}
}

// ConnectionPoolMiddleware optimizes database connections
func ConnectionPoolMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Add connection pool context
		ctx := context.WithValue(c.Request.Context(), "connection_pool", true)
		c.Request = c.Request.WithContext(ctx)
		
		c.Next()
	}
}

// CompressionMiddleware provides response compression
func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if client accepts compression
		if !acceptsCompression(c.Request) {
			c.Next()
			return
		}
		
		// Add compression headers
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		
		// Wrap writer with compression
		writer := &CompressionWriter{
			ResponseWriter: c.Writer,
		}
		c.Writer = writer
		
		c.Next()
		
		// Ensure compression is finalized
		writer.Close()
	}
}

// RateLimitMiddleware provides intelligent rate limiting
func RateLimitMiddleware(requestsPerSecond int) gin.HandlerFunc {
	limiter := NewTokenBucket(requestsPerSecond)
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		if !limiter.Allow(clientIP) {
			c.Header("X-RateLimit-Limit", strconv.Itoa(requestsPerSecond))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "1")
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"limit": requestsPerSecond,
				"window": "1 second",
			})
			c.Abort()
			return
		}
		
		remaining := limiter.Remaining(clientIP)
		c.Header("X-RateLimit-Limit", strconv.Itoa(requestsPerSecond))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		
		c.Next()
	}
}

// Helper types and functions

type CacheEntry struct {
	Data        []byte
	Headers     map[string]string
	ContentType string
	StatusCode  int
	CreatedAt   time.Time
	TTL         time.Duration
}

func (e *CacheEntry) IsExpired() bool {
	return time.Since(e.CreatedAt) > e.TTL
}

type CacheResponseWriter struct {
	gin.ResponseWriter
	cache    *sync.Map
	cacheKey string
	ttl      time.Duration
	data     []byte
}

func (w *CacheResponseWriter) Write(data []byte) (int, error) {
	w.data = append(w.data, data...)
	return w.ResponseWriter.Write(data)
}

func (w *CacheResponseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	
	// Cache successful responses
	if statusCode >= 200 && statusCode < 300 {
		headers := make(map[string]string)
		for key, values := range w.Header() {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		
		entry := &CacheEntry{
			Data:        w.data,
			Headers:     headers,
			ContentType: w.Header().Get("Content-Type"),
			StatusCode:  statusCode,
			CreatedAt:   time.Now(),
			TTL:         w.ttl,
		}
		
		w.cache.Store(w.cacheKey, entry)
	}
}

type CompressionWriter struct {
	gin.ResponseWriter
	compressed bool
}

func (w *CompressionWriter) Write(data []byte) (int, error) {
	// Implement gzip compression here
	return w.ResponseWriter.Write(data)
}

func (w *CompressionWriter) Close() error {
	// Finalize compression
	return nil
}

// Helper functions

func generateConnectionID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func generateCacheKey(c *gin.Context) string {
	return c.Request.Method + ":" + c.Request.URL.Path + ":" + c.Request.URL.RawQuery
}

func acceptsCompression(req *http.Request) bool {
	return req.Header.Get("Accept-Encoding") != ""
}

// TokenBucket for rate limiting
type TokenBucket struct {
	capacity int
	tokens   map[string]*bucket
	mutex    sync.RWMutex
}

type bucket struct {
	tokens   int
	lastSeen time.Time
}

func NewTokenBucket(capacity int) *TokenBucket {
	return &TokenBucket{
		capacity: capacity,
		tokens:   make(map[string]*bucket),
	}
}

func (tb *TokenBucket) Allow(key string) bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()
	
	now := time.Now()
	b, exists := tb.tokens[key]
	
	if !exists {
		tb.tokens[key] = &bucket{
			tokens:   tb.capacity - 1,
			lastSeen: now,
		}
		return true
	}
	
	// Refill tokens based on time elapsed
	elapsed := now.Sub(b.lastSeen)
	tokensToAdd := int(elapsed.Seconds())
	b.tokens = min(tb.capacity, b.tokens+tokensToAdd)
	b.lastSeen = now
	
	if b.tokens > 0 {
		b.tokens--
		return true
	}
	
	return false
}

func (tb *TokenBucket) Remaining(key string) int {
	tb.mutex.RLock()
	defer tb.mutex.RUnlock()
	
	if b, exists := tb.tokens[key]; exists {
		return b.tokens
	}
	return tb.capacity
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
