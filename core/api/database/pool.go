package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Database connection pool metrics
var (
	dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"operation", "table"},
	)

	dbQueryErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_query_errors_total",
			Help: "Total number of database query errors",
		},
		[]string{"operation", "table", "error_type"},
	)
)

// OptimizedDB represents an optimized database connection pool
type OptimizedDB struct {
	db           *sql.DB
	readReplicas []*sql.DB
	writeDB      *sql.DB
	config       *PoolConfig
	metrics      *PoolMetrics
	mutex        sync.RWMutex
}

// PoolConfig contains database pool configuration
type PoolConfig struct {
	MaxOpenConns        int
	MaxIdleConns        int
	ConnMaxLifetime     time.Duration
	ConnMaxIdleTime     time.Duration
	ReadReplicaCount    int
	HealthCheckInterval time.Duration
	QueryTimeout        time.Duration
}

// PoolMetrics tracks pool performance
type PoolMetrics struct {
	QueriesExecuted   int64
	QueriesSucceeded  int64
	QueriesFailed     int64
	AvgQueryTime      time.Duration
	ConnectionsOpened int64
	ConnectionsClosed int64
}

// DefaultPoolConfig returns optimized default configuration
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxOpenConns:        100,  // Increased for high concurrency
		MaxIdleConns:        25,   // Optimized idle connections
		ConnMaxLifetime:     time.Hour,
		ConnMaxIdleTime:     time.Minute * 15,
		ReadReplicaCount:    2,    // Read replicas for scaling
		HealthCheckInterval: time.Minute * 5,
		QueryTimeout:        time.Second * 30,
	}
}

// NewOptimizedDB creates a new optimized database connection pool
func NewOptimizedDB(dsn string, config *PoolConfig) (*OptimizedDB, error) {
	if config == nil {
		config = DefaultPoolConfig()
	}

	// Create main database connection
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)
	db.SetConnMaxLifetime(config.ConnMaxLifetime)
	db.SetConnMaxIdleTime(config.ConnMaxIdleTime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	optimizedDB := &OptimizedDB{
		db:      db,
		writeDB: db, // Use main connection for writes
		config:  config,
		metrics: &PoolMetrics{},
	}

	// Setup read replicas if configured
	if config.ReadReplicaCount > 0 {
		if err := optimizedDB.setupReadReplicas(dsn); err != nil {
			log.Printf("Warning: Failed to setup read replicas: %v", err)
		}
	}

	// Start health check routine
	go optimizedDB.healthCheckRoutine()

	// Start metrics collection
	go optimizedDB.metricsCollectionRoutine()

	log.Printf("Optimized database pool initialized with %d max connections", config.MaxOpenConns)
	return optimizedDB, nil
}

// setupReadReplicas configures read replica connections
func (odb *OptimizedDB) setupReadReplicas(dsn string) error {
	odb.readReplicas = make([]*sql.DB, odb.config.ReadReplicaCount)

	for i := 0; i < odb.config.ReadReplicaCount; i++ {
		replica, err := sql.Open("postgres", dsn)
		if err != nil {
			return fmt.Errorf("failed to open read replica %d: %w", i, err)
		}

		// Configure replica with read-optimized settings
		replica.SetMaxOpenConns(odb.config.MaxOpenConns / 2)
		replica.SetMaxIdleConns(odb.config.MaxIdleConns / 2)
		replica.SetConnMaxLifetime(odb.config.ConnMaxLifetime)
		replica.SetConnMaxIdleTime(odb.config.ConnMaxIdleTime)

		odb.readReplicas[i] = replica
	}

	log.Printf("Setup %d read replicas", odb.config.ReadReplicaCount)
	return nil
}

// GetReadDB returns an optimized read connection
func (odb *OptimizedDB) GetReadDB() *sql.DB {
	odb.mutex.RLock()
	defer odb.mutex.RUnlock()

	// Use read replicas if available
	if len(odb.readReplicas) > 0 {
		// Simple round-robin selection
		index := int(odb.metrics.QueriesExecuted) % len(odb.readReplicas)
		return odb.readReplicas[index]
	}

	// Fallback to main database
	return odb.db
}

// GetWriteDB returns the write database connection
func (odb *OptimizedDB) GetWriteDB() *sql.DB {
	return odb.writeDB
}

// QueryContext executes a read query with optimization
func (odb *OptimizedDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	
	// Add query timeout
	ctx, cancel := context.WithTimeout(ctx, odb.config.QueryTimeout)
	defer cancel()

	// Use read database
	db := odb.GetReadDB()
	
	// Execute query
	rows, err := db.QueryContext(ctx, query, args...)
	
	// Record metrics
	duration := time.Since(start)
	odb.recordQueryMetrics("SELECT", "", duration, err)
	
	return rows, err
}

// QueryRowContext executes a single row query with optimization
func (odb *OptimizedDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	
	// Add query timeout
	ctx, cancel := context.WithTimeout(ctx, odb.config.QueryTimeout)
	defer cancel()

	// Use read database
	db := odb.GetReadDB()
	
	// Execute query
	row := db.QueryRowContext(ctx, query, args...)
	
	// Record metrics
	duration := time.Since(start)
	odb.recordQueryMetrics("SELECT", "", duration, nil)
	
	return row
}

// ExecContext executes a write query with optimization
func (odb *OptimizedDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	
	// Add query timeout
	ctx, cancel := context.WithTimeout(ctx, odb.config.QueryTimeout)
	defer cancel()

	// Use write database
	db := odb.GetWriteDB()
	
	// Execute query
	result, err := db.ExecContext(ctx, query, args...)
	
	// Record metrics
	duration := time.Since(start)
	operation := "INSERT"
	if len(query) > 6 {
		switch query[:6] {
		case "UPDATE":
			operation = "UPDATE"
		case "DELETE":
			operation = "DELETE"
		}
	}
	odb.recordQueryMetrics(operation, "", duration, err)
	
	return result, err
}

// BeginTx starts an optimized transaction
func (odb *OptimizedDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	// Add transaction timeout
	ctx, cancel := context.WithTimeout(ctx, odb.config.QueryTimeout*2)
	defer cancel()

	// Use write database for transactions
	return odb.writeDB.BeginTx(ctx, opts)
}

// Ping checks database connectivity
func (odb *OptimizedDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	
	return odb.db.PingContext(ctx)
}

// Stats returns database statistics
func (odb *OptimizedDB) Stats() sql.DBStats {
	return odb.db.Stats()
}

// GetMetrics returns pool performance metrics
func (odb *OptimizedDB) GetMetrics() *PoolMetrics {
	odb.mutex.RLock()
	defer odb.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	metrics := *odb.metrics
	return &metrics
}

// Close closes all database connections
func (odb *OptimizedDB) Close() error {
	var errors []error

	// Close main database
	if err := odb.db.Close(); err != nil {
		errors = append(errors, err)
	}

	// Close read replicas
	for i, replica := range odb.readReplicas {
		if err := replica.Close(); err != nil {
			errors = append(errors, fmt.Errorf("failed to close replica %d: %w", i, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing database connections: %v", errors)
	}

	return nil
}

// recordQueryMetrics records query performance metrics
func (odb *OptimizedDB) recordQueryMetrics(operation, table string, duration time.Duration, err error) {
	odb.mutex.Lock()
	defer odb.mutex.Unlock()

	odb.metrics.QueriesExecuted++
	
	if err != nil {
		odb.metrics.QueriesFailed++
		dbQueryErrors.WithLabelValues(operation, table, "query_error").Inc()
	} else {
		odb.metrics.QueriesSucceeded++
	}

	// Update average query time
	if odb.metrics.QueriesExecuted > 0 {
		totalTime := time.Duration(odb.metrics.QueriesExecuted-1) * odb.metrics.AvgQueryTime
		odb.metrics.AvgQueryTime = (totalTime + duration) / time.Duration(odb.metrics.QueriesExecuted)
	}

	// Record Prometheus metrics
	dbQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())
}

// healthCheckRoutine performs periodic health checks
func (odb *OptimizedDB) healthCheckRoutine() {
	ticker := time.NewTicker(odb.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		// Check main database
		if err := odb.Ping(); err != nil {
			log.Printf("Main database health check failed: %v", err)
		}

		// Check read replicas
		for i, replica := range odb.readReplicas {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			if err := replica.PingContext(ctx); err != nil {
				log.Printf("Read replica %d health check failed: %v", i, err)
			}
			cancel()
		}
	}
}

// metricsCollectionRoutine collects and updates metrics
func (odb *OptimizedDB) metricsCollectionRoutine() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for range ticker.C {
		stats := odb.Stats()
		
		// Update Prometheus metrics
		dbConnectionsActive.Set(float64(stats.OpenConnections))
		dbConnectionsIdle.Set(float64(stats.Idle))
	}
}
