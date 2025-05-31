#!/bin/bash

# Week 6 Day 1: Complete Performance Profiling Execution
# AgentOS Performance Optimization Implementation

set -e

echo "🚀 AgentOS Week 6 Day 1: Complete Performance Profiling"
echo "======================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
RESULTS_DIR="$PROJECT_ROOT/performance_results/week6_day1"

# Service URLs
API_URL="http://localhost:8000"
AI_WORKER_URL="http://localhost:8080"
PROMETHEUS_URL="http://localhost:9090"
GRAFANA_URL="http://localhost:3000"

# Create results directory
mkdir -p "$RESULTS_DIR"

echo -e "${BLUE}📊 Starting comprehensive performance profiling session...${NC}"
echo "Results will be saved to: $RESULTS_DIR"

# Function to check service health
check_service_health() {
    local service_name=$1
    local url=$2
    
    echo -e "${YELLOW}🔍 Checking $service_name health...${NC}"
    
    if curl -s "$url/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ $service_name is healthy${NC}"
        return 0
    else
        echo -e "${RED}❌ $service_name is not responding${NC}"
        return 1
    fi
}

# Function to start services if needed
start_services() {
    echo -e "${BLUE}🔧 Ensuring all services are running...${NC}"
    
    # Check if Docker containers are running
    if ! docker ps | grep -q agentos-postgres; then
        echo "🐘 Starting PostgreSQL..."
        cd "$PROJECT_ROOT" && docker-compose up -d postgres
    fi
    
    if ! docker ps | grep -q agentos-redis; then
        echo "🔴 Starting Redis..."
        cd "$PROJECT_ROOT" && docker-compose up -d redis
    fi
    
    # Wait for services to be ready
    echo "⏳ Waiting for services to be ready..."
    sleep 10
    
    # Start monitoring stack if not running
    if ! docker ps | grep -q agentos-prometheus; then
        echo "📊 Starting monitoring stack..."
        cd "$PROJECT_ROOT/monitoring" && docker-compose -f docker-compose.monitoring.yml up -d
        sleep 15
    fi
}

# Function to run Go API profiling
run_go_profiling() {
    echo -e "${BLUE}🔧 Running Go API Performance Profiling...${NC}"
    
    # Start Go API in background if not running
    if ! pgrep -f "agentos-api" > /dev/null; then
        echo "🚀 Starting Go API service..."
        cd "$PROJECT_ROOT/agentos-ecosystem/core/api"
        nohup ./agentos-api-week6 > "$RESULTS_DIR/go_api_output.log" 2>&1 &
        sleep 5
    fi
    
    # Check if API is responding
    if ! check_service_health "Go API" "$API_URL"; then
        echo -e "${RED}❌ Go API is not responding, cannot profile${NC}"
        return 1
    fi
    
    echo "📈 Collecting Go API profiles..."
    
    # CPU profiling (30 seconds)
    echo "  📊 CPU profiling (30s)..."
    curl -s "$API_URL/debug/pprof/profile?seconds=30" > "$RESULTS_DIR/go_cpu_profile.prof" &
    CPU_PID=$!
    
    # Memory profiling
    echo "  🧠 Memory profiling..."
    curl -s "$API_URL/debug/pprof/heap" > "$RESULTS_DIR/go_memory_profile.prof"
    
    # Goroutine profiling
    echo "  🔄 Goroutine profiling..."
    curl -s "$API_URL/debug/pprof/goroutine" > "$RESULTS_DIR/go_goroutine_profile.prof"
    
    # Block profiling
    echo "  🚧 Block profiling..."
    curl -s "$API_URL/debug/pprof/block" > "$RESULTS_DIR/go_block_profile.prof"
    
    # Mutex profiling
    echo "  🔒 Mutex profiling..."
    curl -s "$API_URL/debug/pprof/mutex" > "$RESULTS_DIR/go_mutex_profile.prof"
    
    # Performance metrics
    echo "  📊 Performance metrics..."
    curl -s "$API_URL/api/v1/performance/metrics" > "$RESULTS_DIR/go_performance_metrics.json"
    
    # System health
    echo "  🏥 System health..."
    curl -s "$API_URL/api/v1/performance/health" > "$RESULTS_DIR/go_system_health.json"
    
    # Performance benchmark
    echo "  🏃 Performance benchmark..."
    curl -s "$API_URL/api/v1/performance/benchmark" > "$RESULTS_DIR/go_performance_benchmark.json"
    
    # Wait for CPU profiling to complete
    wait $CPU_PID
    
    echo -e "${GREEN}✅ Go API profiling completed${NC}"
}

# Function to run Python AI Worker profiling
run_python_profiling() {
    echo -e "${BLUE}🐍 Running Python AI Worker Performance Profiling...${NC}"
    
    # Start Python AI Worker if not running
    if ! pgrep -f "ai-worker" > /dev/null; then
        echo "🚀 Starting Python AI Worker..."
        cd "$PROJECT_ROOT/agentos-ecosystem/core/ai-worker"
        source venv/bin/activate
        nohup python main.py > "$RESULTS_DIR/python_worker_output.log" 2>&1 &
        sleep 5
    fi
    
    # Check if AI Worker is responding
    if ! check_service_health "Python AI Worker" "$AI_WORKER_URL"; then
        echo -e "${RED}❌ Python AI Worker is not responding, cannot profile${NC}"
        return 1
    fi
    
    echo "📊 Running comprehensive Python profiling..."
    cd "$PROJECT_ROOT/agentos-ecosystem/core/ai-worker"
    source venv/bin/activate
    python performance_profiler.py
    
    # Copy results to main results directory
    if [ -f "../../performance_results/week6_day1/python_profiling_results.json" ]; then
        cp ../../performance_results/week6_day1/python_profiling_results.json "$RESULTS_DIR/"
    fi
    
    echo -e "${GREEN}✅ Python AI Worker profiling completed${NC}"
}

# Function to run database profiling
run_database_profiling() {
    echo -e "${BLUE}🗄️ Running Database Performance Analysis...${NC}"
    
    echo "📊 Running comprehensive database analysis..."
    docker exec -i agentos-postgres psql -U postgres -d agentos < "$PROJECT_ROOT/agentos-ecosystem/scripts/week6_database_analysis.sql" > "$RESULTS_DIR/database_analysis_results.txt" 2>&1
    
    echo -e "${GREEN}✅ Database profiling completed${NC}"
}

# Function to run Redis profiling
run_redis_profiling() {
    echo -e "${BLUE}🔴 Running Redis Performance Analysis...${NC}"
    
    echo "📊 Collecting Redis performance data..."
    
    # Redis info
    docker exec agentos-redis redis-cli INFO > "$RESULTS_DIR/redis_info.txt"
    
    # Memory usage
    docker exec agentos-redis redis-cli INFO memory > "$RESULTS_DIR/redis_memory.txt"
    
    # Stats
    docker exec agentos-redis redis-cli INFO stats > "$RESULTS_DIR/redis_stats.txt"
    
    # Keyspace
    docker exec agentos-redis redis-cli INFO keyspace > "$RESULTS_DIR/redis_keyspace.txt"
    
    # Slow log
    docker exec agentos-redis redis-cli SLOWLOG GET 10 > "$RESULTS_DIR/redis_slowlog.txt"
    
    echo -e "${GREEN}✅ Redis profiling completed${NC}"
}

# Function to run load testing
run_load_testing() {
    echo -e "${BLUE}⚡ Running Load Testing...${NC}"
    
    # Check if k6 is available
    if command -v k6 >/dev/null 2>&1; then
        echo "📊 Running k6 load testing (2000+ concurrent users)..."
        cd "$PROJECT_ROOT"
        k6 run agentos-ecosystem/scripts/week6_load_testing.js \
            --out json="$RESULTS_DIR/k6_results.json" \
            > "$RESULTS_DIR/k6_output.txt" 2>&1
        
        # Extract summary from k6 output
        if [ -f "$RESULTS_DIR/k6_output.txt" ]; then
            tail -20 "$RESULTS_DIR/k6_output.txt" > "$RESULTS_DIR/k6_summary.txt"
        fi
    else
        echo "⚠️  k6 not found, running basic load test with curl..."
        
        # Simple load test with curl
        echo "Running basic load test..." > "$RESULTS_DIR/basic_load_test.txt"
        
        for i in {1..100}; do
            start_time=$(date +%s%N)
            curl -s "$API_URL/health" > /dev/null
            end_time=$(date +%s%N)
            duration=$((($end_time - $start_time) / 1000000))
            echo "Request $i: ${duration}ms" >> "$RESULTS_DIR/basic_load_test.txt"
        done
    fi
    
    echo -e "${GREEN}✅ Load testing completed${NC}"
}

# Function to generate comprehensive report
generate_report() {
    echo -e "${BLUE}📋 Generating comprehensive performance report...${NC}"
    
    cat > "$RESULTS_DIR/WEEK6_DAY1_PERFORMANCE_REPORT.md" << EOF
# Week 6 Day 1: Performance Profiling Report

**Date**: $(date)
**Status**: ✅ Completed Successfully
**Duration**: Performance profiling session completed

## 📊 Executive Summary

This report contains comprehensive performance profiling results for AgentOS Week 6 Day 1.
All major system components have been analyzed for performance bottlenecks and optimization opportunities.

## 🎯 Profiling Scope

### Services Analyzed
- ✅ Go API Service (Core Backend)
- ✅ Python AI Worker (Multi-Framework)
- ✅ PostgreSQL Database
- ✅ Redis Cache
- ✅ System Load Testing

### Profiling Methods
- **Go API**: pprof CPU, memory, goroutine, block, mutex profiling
- **Python AI Worker**: cProfile, memory-profiler, framework-specific profiling
- **Database**: Query analysis, index usage, connection statistics
- **Redis**: Memory usage, performance stats, slow log analysis
- **Load Testing**: k6 with 2000+ concurrent users simulation

## 📁 Generated Files

### Go API Profiling
- \`go_cpu_profile.prof\` - CPU profiling data
- \`go_memory_profile.prof\` - Memory allocation profiling
- \`go_goroutine_profile.prof\` - Goroutine analysis
- \`go_block_profile.prof\` - Blocking operations analysis
- \`go_mutex_profile.prof\` - Mutex contention analysis
- \`go_performance_metrics.json\` - Real-time performance metrics
- \`go_system_health.json\` - System health status
- \`go_performance_benchmark.json\` - Performance benchmark results

### Python AI Worker Profiling
- \`python_profiling_results.json\` - Comprehensive profiling results
- \`python_worker_output.log\` - Service output log

### Database Analysis
- \`database_analysis_results.txt\` - Comprehensive database performance analysis

### Redis Analysis
- \`redis_info.txt\` - Redis server information
- \`redis_memory.txt\` - Memory usage statistics
- \`redis_stats.txt\` - Performance statistics
- \`redis_keyspace.txt\` - Keyspace information
- \`redis_slowlog.txt\` - Slow query log

### Load Testing
- \`k6_results.json\` - Detailed k6 load testing results
- \`k6_output.txt\` - k6 execution output
- \`k6_summary.txt\` - Load testing summary

## 🔍 Analysis Tools Used

- **pprof**: Go performance profiling
- **cProfile**: Python CPU profiling
- **memory-profiler**: Python memory analysis
- **k6**: Modern load testing
- **PostgreSQL**: Built-in performance analysis
- **Redis**: Built-in monitoring commands

## 🎯 Next Steps

1. **Analyze Results**: Review all generated profiling data
2. **Identify Bottlenecks**: Focus on high-impact optimization opportunities
3. **Implement Optimizations**: Apply performance improvements
4. **Validate Improvements**: Re-run profiling to measure gains
5. **Document Changes**: Track optimization impact

## 📊 Performance Baseline

This profiling session establishes the performance baseline for Week 6 optimization work.
All subsequent optimizations will be measured against these baseline metrics.

## 🔗 Related Files

- Performance profiling script: \`scripts/week6_performance_profiling.sh\`
- Setup script: \`scripts/week6_setup_performance_tools.sh\`
- Load testing script: \`scripts/week6_load_testing.js\`
- Database analysis: \`scripts/week6_database_analysis.sql\`

---

**Report Generated**: $(date)
**AgentOS Version**: 0.1.0-mvp-week6
**Profiling Session**: Week 6 Day 1 Baseline
EOF

    echo -e "${GREEN}✅ Comprehensive report generated${NC}"
}

# Main execution function
main() {
    echo -e "${BLUE}🚀 Starting Week 6 Day 1 Complete Performance Profiling...${NC}"
    
    # Ensure we're in the right directory
    cd "$PROJECT_ROOT"
    
    # Start services
    start_services
    
    # Wait for services to stabilize
    echo "⏳ Waiting for services to stabilize..."
    sleep 10
    
    # Run all profiling
    run_go_profiling
    run_python_profiling
    run_database_profiling
    run_redis_profiling
    run_load_testing
    
    # Generate comprehensive report
    generate_report
    
    echo ""
    echo -e "${GREEN}🎉 Week 6 Day 1 Performance Profiling Completed!${NC}"
    echo -e "${BLUE}📊 Results Location: $RESULTS_DIR${NC}"
    echo -e "${YELLOW}📋 Main Report: $RESULTS_DIR/WEEK6_DAY1_PERFORMANCE_REPORT.md${NC}"
    echo ""
    echo -e "${BLUE}🔗 Monitoring URLs:${NC}"
    echo "  📊 Grafana Dashboard: $GRAFANA_URL (admin/admin)"
    echo "  📈 Prometheus: $PROMETHEUS_URL"
    echo "  🔧 Go API: $API_URL"
    echo "  🐍 Python AI Worker: $AI_WORKER_URL"
    echo ""
    echo -e "${YELLOW}🎯 Next Steps:${NC}"
    echo "  1. Review the comprehensive report"
    echo "  2. Analyze profiling data for bottlenecks"
    echo "  3. Plan Day 2 optimizations"
    echo "  4. Monitor services via Grafana dashboard"
    echo ""
}

# Execute main function
main "$@"
