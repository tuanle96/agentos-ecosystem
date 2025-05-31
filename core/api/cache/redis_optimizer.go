package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
)

// Redis cache metrics
var (
	cacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_type", "key_pattern"},
	)

	cacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_type", "key_pattern"},
	)

	cacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_operation_duration_seconds",
			Help:    "Duration of cache operations in seconds",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		},
		[]string{"operation", "cache_type"},
	)

	cacheSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_size_bytes",
			Help: "Current cache size in bytes",
		},
		[]string{"cache_type"},
	)
)

// OptimizedRedisCache provides high-performance Redis caching
type OptimizedRedisCache struct {
	client     redis.UniversalClient
	localCache *sync.Map
	config     *CacheConfig
	metrics    *CacheMetrics
	pipeline   redis.Pipeliner
	mutex      sync.RWMutex
}

// CacheConfig contains cache optimization settings
type CacheConfig struct {
	ClusterMode        bool
	PoolSize           int
	MinIdleConns       int
	MaxRetries         int
	RetryDelay         time.Duration
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	LocalCacheEnabled  bool
	LocalCacheTTL      time.Duration
	LocalCacheSize     int
	CompressionEnabled bool
	PipelineEnabled    bool
	PipelineSize       int
}

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	Hits             int64
	Misses           int64
	Sets             int64
	Deletes          int64
	LocalHits        int64
	LocalMisses      int64
	CompressionRatio float64
	AvgOperationTime time.Duration
}

// DefaultCacheConfig returns optimized default configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		ClusterMode:        false,
		PoolSize:           100,
		MinIdleConns:       10,
		MaxRetries:         3,
		RetryDelay:         time.Millisecond * 100,
		DialTimeout:        time.Second * 5,
		ReadTimeout:        time.Second * 3,
		WriteTimeout:       time.Second * 3,
		LocalCacheEnabled:  true,
		LocalCacheTTL:      time.Minute * 5,
		LocalCacheSize:     1000,
		CompressionEnabled: true,
		PipelineEnabled:    true,
		PipelineSize:       100,
	}
}

// NewOptimizedRedisCache creates a new optimized Redis cache
func NewOptimizedRedisCache(addr string, password string, config *CacheConfig) (*OptimizedRedisCache, error) {
	if config == nil {
		config = DefaultCacheConfig()
	}

	var client redis.UniversalClient

	if config.ClusterMode {
		// Redis Cluster configuration
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        []string{addr},
			Password:     password,
			PoolSize:     config.PoolSize,
			MinIdleConns: config.MinIdleConns,
			MaxRetries:   config.MaxRetries,
			DialTimeout:  config.DialTimeout,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
		})
	} else {
		// Single Redis instance configuration
		client = redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     password,
			DB:           0,
			PoolSize:     config.PoolSize,
			MinIdleConns: config.MinIdleConns,
			MaxRetries:   config.MaxRetries,
			DialTimeout:  config.DialTimeout,
			ReadTimeout:  config.ReadTimeout,
			WriteTimeout: config.WriteTimeout,
		})
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	cache := &OptimizedRedisCache{
		client:     client,
		localCache: &sync.Map{},
		config:     config,
		metrics:    &CacheMetrics{},
	}

	// Initialize pipeline if enabled
	if config.PipelineEnabled {
		cache.pipeline = client.Pipeline()
	}

	// Start metrics collection
	go cache.metricsCollectionRoutine()

	// Start local cache cleanup routine
	if config.LocalCacheEnabled {
		go cache.localCacheCleanupRoutine()
	}

	log.Printf("Optimized Redis cache initialized with pool size %d", config.PoolSize)
	return cache, nil
}

// Get retrieves a value from cache with multi-level optimization
func (orc *OptimizedRedisCache) Get(ctx context.Context, key string) (string, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		cacheOperationDuration.WithLabelValues("get", "redis").Observe(duration.Seconds())
		orc.updateAvgOperationTime(duration)
	}()

	// Check local cache first
	if orc.config.LocalCacheEnabled {
		if value, found := orc.getFromLocalCache(key); found {
			orc.metrics.LocalHits++
			cacheHits.WithLabelValues("local", orc.getKeyPattern(key)).Inc()
			return value, nil
		}
		orc.metrics.LocalMisses++
	}

	// Get from Redis
	value, err := orc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			orc.metrics.Misses++
			cacheMisses.WithLabelValues("redis", orc.getKeyPattern(key)).Inc()
			return "", ErrCacheMiss
		}
		return "", fmt.Errorf("Redis get error: %w", err)
	}

	orc.metrics.Hits++
	cacheHits.WithLabelValues("redis", orc.getKeyPattern(key)).Inc()

	// Store in local cache
	if orc.config.LocalCacheEnabled {
		orc.setInLocalCache(key, value)
	}

	// Decompress if needed
	if orc.config.CompressionEnabled {
		decompressed, err := orc.decompress(value)
		if err != nil {
			log.Printf("Decompression error for key %s: %v", key, err)
			return value, nil // Return original value
		}
		return decompressed, nil
	}

	return value, nil
}

// Set stores a value in cache with optimization
func (orc *OptimizedRedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		cacheOperationDuration.WithLabelValues("set", "redis").Observe(duration.Seconds())
		orc.updateAvgOperationTime(duration)
	}()

	// Compress if enabled
	if orc.config.CompressionEnabled {
		compressed, err := orc.compress(value)
		if err != nil {
			log.Printf("Compression error for key %s: %v", key, err)
		} else {
			value = compressed
		}
	}

	// Set in Redis
	err := orc.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("Redis set error: %w", err)
	}

	orc.metrics.Sets++

	// Store in local cache
	if orc.config.LocalCacheEnabled {
		orc.setInLocalCache(key, value)
	}

	return nil
}

// GetJSON retrieves and unmarshals JSON data
func (orc *OptimizedRedisCache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	value, err := orc.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(value), dest)
}

// SetJSON marshals and stores JSON data
func (orc *OptimizedRedisCache) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("JSON marshal error: %w", err)
	}

	return orc.Set(ctx, key, string(data), ttl)
}

// Delete removes a key from cache
func (orc *OptimizedRedisCache) Delete(ctx context.Context, key string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		cacheOperationDuration.WithLabelValues("delete", "redis").Observe(duration.Seconds())
	}()

	// Delete from Redis
	err := orc.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("Redis delete error: %w", err)
	}

	orc.metrics.Deletes++

	// Delete from local cache
	if orc.config.LocalCacheEnabled {
		orc.localCache.Delete(key)
	}

	return nil
}

// Pipeline operations for batch processing
func (orc *OptimizedRedisCache) Pipeline() redis.Pipeliner {
	return orc.client.Pipeline()
}

// GetMulti retrieves multiple keys efficiently
func (orc *OptimizedRedisCache) GetMulti(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		cacheOperationDuration.WithLabelValues("mget", "redis").Observe(duration.Seconds())
	}()

	// Use pipeline for efficiency
	pipe := orc.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("Pipeline exec error: %w", err)
	}

	result := make(map[string]string)
	for i, cmd := range cmds {
		value, err := cmd.Result()
		if err == nil {
			result[keys[i]] = value
			orc.metrics.Hits++
		} else if err == redis.Nil {
			orc.metrics.Misses++
		}
	}

	return result, nil
}

// Local cache operations
type localCacheEntry struct {
	value     string
	createdAt time.Time
}

func (orc *OptimizedRedisCache) getFromLocalCache(key string) (string, bool) {
	if value, found := orc.localCache.Load(key); found {
		if entry, ok := value.(*localCacheEntry); ok {
			if time.Since(entry.createdAt) < orc.config.LocalCacheTTL {
				return entry.value, true
			}
			// Expired, remove it
			orc.localCache.Delete(key)
		}
	}
	return "", false
}

func (orc *OptimizedRedisCache) setInLocalCache(key, value string) {
	entry := &localCacheEntry{
		value:     value,
		createdAt: time.Now(),
	}
	orc.localCache.Store(key, entry)
}

// Helper functions
func (orc *OptimizedRedisCache) getKeyPattern(key string) string {
	// Extract pattern from key (e.g., "user:123" -> "user:*")
	if len(key) > 0 {
		parts := []rune(key)
		for i, char := range parts {
			if char >= '0' && char <= '9' {
				return string(parts[:i]) + "*"
			}
		}
	}
	return "unknown"
}

func (orc *OptimizedRedisCache) compress(data string) (string, error) {
	// Implement compression logic here (e.g., gzip)
	// For now, return original data
	return data, nil
}

func (orc *OptimizedRedisCache) decompress(data string) (string, error) {
	// Implement decompression logic here
	// For now, return original data
	return data, nil
}

func (orc *OptimizedRedisCache) updateAvgOperationTime(duration time.Duration) {
	orc.mutex.Lock()
	defer orc.mutex.Unlock()

	totalOps := orc.metrics.Hits + orc.metrics.Misses + orc.metrics.Sets + orc.metrics.Deletes
	if totalOps > 0 {
		totalTime := time.Duration(totalOps-1) * orc.metrics.AvgOperationTime
		orc.metrics.AvgOperationTime = (totalTime + duration) / time.Duration(totalOps)
	}
}

// Cleanup routines
func (orc *OptimizedRedisCache) localCacheCleanupRoutine() {
	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for range ticker.C {
		count := 0
		orc.localCache.Range(func(key, value interface{}) bool {
			if entry, ok := value.(*localCacheEntry); ok {
				if time.Since(entry.createdAt) > orc.config.LocalCacheTTL {
					orc.localCache.Delete(key)
				}
			}
			count++
			return count < orc.config.LocalCacheSize
		})
	}
}

func (orc *OptimizedRedisCache) metricsCollectionRoutine() {
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for range ticker.C {
		// Update cache size metrics
		info := orc.client.Info(context.Background(), "memory")
		if info.Err() == nil {
			// Parse memory usage from info result
			// This is a simplified version
			cacheSize.WithLabelValues("redis").Set(1024 * 1024) // Placeholder
		}
	}
}

// GetMetrics returns cache performance metrics
func (orc *OptimizedRedisCache) GetMetrics() *CacheMetrics {
	orc.mutex.RLock()
	defer orc.mutex.RUnlock()

	metrics := *orc.metrics
	return &metrics
}

// Close closes the cache connection
func (orc *OptimizedRedisCache) Close() error {
	return orc.client.Close()
}

// Custom errors
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)
