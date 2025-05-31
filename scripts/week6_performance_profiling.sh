#!/bin/bash

# Week 6 Day 1: Performance Profiling & Analysis Script
# AgentOS Performance Optimization Implementation

set -e

echo "🚀 AgentOS Week 6 Day 1: Performance Profiling & Analysis"
echo "=========================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_URL="http://localhost:8000"
AI_WORKER_URL="http://localhost:8080"
POSTGRES_DB="agentos"
REDIS_URL="redis://localhost:6379"

# Create performance results directory
PERF_DIR="performance_results/week6_day1"
mkdir -p "$PERF_DIR"

echo -e "${BLUE}📊 Starting Performance Profiling Session...${NC}"
echo "Results will be saved to: $PERF_DIR"

# Function to check service health
check_service_health() {
    local service_name=$1
    local url=$2

    echo -e "${YELLOW}🔍 Checking $service_name health...${NC}"

    if curl -s "$url/health" > /dev/null; then
        echo -e "${GREEN}✅ $service_name is healthy${NC}"
        return 0
    else
        echo -e "${RED}❌ $service_name is not responding${NC}"
        return 1
    fi
}

# Function to profile Go API service
profile_go_api() {
    echo -e "${BLUE}🔧 Profiling Go API Service...${NC}"

    # CPU profiling
    echo "📈 Starting CPU profiling..."
    curl -s "$API_URL/debug/pprof/profile?seconds=30" > "$PERF_DIR/cpu_profile.prof" &
    CPU_PID=$!

    # Memory profiling
    echo "🧠 Capturing memory profile..."
    curl -s "$API_URL/debug/pprof/heap" > "$PERF_DIR/memory_profile.prof"

    # Goroutine profiling
    echo "🔄 Capturing goroutine profile..."
    curl -s "$API_URL/debug/pprof/goroutine" > "$PERF_DIR/goroutine_profile.prof"

    # Block profiling
    echo "🚧 Capturing block profile..."
    curl -s "$API_URL/debug/pprof/block" > "$PERF_DIR/block_profile.prof"

    # Mutex profiling
    echo "🔒 Capturing mutex profile..."
    curl -s "$API_URL/debug/pprof/mutex" > "$PERF_DIR/mutex_profile.prof"

    # Wait for CPU profiling to complete
    wait $CPU_PID
    echo -e "${GREEN}✅ Go API profiling completed${NC}"
}

# Function to profile Python AI Worker
profile_python_worker() {
    echo -e "${BLUE}🐍 Profiling Python AI Worker...${NC}"

    # Run the comprehensive Python profiler
    cd agentos-ecosystem/core/ai-worker

    echo "📊 Running comprehensive Python AI Worker profiling..."
    python performance_profiler.py

    cd - > /dev/null

    echo -e "${GREEN}✅ Python AI Worker profiling completed${NC}"
}

# Function to profile database performance
profile_database() {
    echo -e "${BLUE}🗄️ Profiling Database Performance...${NC}"

    # Use the comprehensive database analysis script
    echo "📊 Running comprehensive database performance analysis..."
    docker exec -i agentos-postgres psql -U postgres -d "$POSTGRES_DB" < "scripts/week6_database_analysis.sql" > "$PERF_DIR/db_performance_results.txt" 2>&1

    echo -e "${GREEN}✅ Database profiling completed${NC}"
}

# Function to profile Redis performance
profile_redis() {
    echo -e "${BLUE}🔴 Profiling Redis Performance...${NC}"

    # Redis performance analysis
    cat > "$PERF_DIR/redis_performance_analysis.sh" << 'EOF'
#!/bin/bash

echo "Redis Performance Analysis" > performance_results/week6_day1/redis_performance_results.txt
echo "=========================" >> performance_results/week6_day1/redis_performance_results.txt
echo "" >> performance_results/week6_day1/redis_performance_results.txt

# Redis info
echo "Redis Info:" >> performance_results/week6_day1/redis_performance_results.txt
docker exec agentos-redis redis-cli INFO >> performance_results/week6_day1/redis_performance_results.txt

echo "" >> performance_results/week6_day1/redis_performance_results.txt
echo "Memory Usage:" >> performance_results/week6_day1/redis_performance_results.txt
docker exec agentos-redis redis-cli INFO memory >> performance_results/week6_day1/redis_performance_results.txt

echo "" >> performance_results/week6_day1/redis_performance_results.txt
echo "Stats:" >> performance_results/week6_day1/redis_performance_results.txt
docker exec agentos-redis redis-cli INFO stats >> performance_results/week6_day1/redis_performance_results.txt

echo "" >> performance_results/week6_day1/redis_performance_results.txt
echo "Keyspace:" >> performance_results/week6_day1/redis_performance_results.txt
docker exec agentos-redis redis-cli INFO keyspace >> performance_results/week6_day1/redis_performance_results.txt

echo "" >> performance_results/week6_day1/redis_performance_results.txt
echo "Slow Log:" >> performance_results/week6_day1/redis_performance_results.txt
docker exec agentos-redis redis-cli SLOWLOG GET 10 >> performance_results/week6_day1/redis_performance_results.txt
EOF

    chmod +x "$PERF_DIR/redis_performance_analysis.sh"
    bash "$PERF_DIR/redis_performance_analysis.sh"

    echo -e "${GREEN}✅ Redis profiling completed${NC}"
}

# Function to run load testing
run_load_testing() {
    echo -e "${BLUE}⚡ Running Load Testing (2000+ concurrent users with k6)...${NC}"

    # Check if k6 is installed
    if command -v k6 >/dev/null 2>&1; then
        echo "📊 Running k6 load testing..."
        k6 run scripts/week6_load_testing.js --out json="$PERF_DIR/k6_results.json" > "$PERF_DIR/k6_output.txt" 2>&1
    else
        echo "⚠️  k6 not found, falling back to Apache Bench..."

        # Fallback to Apache Bench
        echo "Load Testing Results (Apache Bench)" > "$PERF_DIR/load_test_results.txt"
        echo "====================================" >> "$PERF_DIR/load_test_results.txt"

        # Test 1: Health check endpoint
        echo "Test 1: Health Check (100 concurrent, 1000 requests)" >> "$PERF_DIR/load_test_results.txt"
        ab -n 1000 -c 100 http://localhost:8000/health >> "$PERF_DIR/load_test_results.txt" 2>&1

        echo "" >> "$PERF_DIR/load_test_results.txt"

        # Test 2: Performance endpoints
        echo "Test 2: Performance Metrics (50 concurrent, 500 requests)" >> "$PERF_DIR/load_test_results.txt"
        ab -n 500 -c 50 http://localhost:8000/api/v1/performance/metrics >> "$PERF_DIR/load_test_results.txt" 2>&1
    fi

    echo -e "${GREEN}✅ Load testing completed${NC}"
}

# Main execution
main() {
    echo -e "${BLUE}🚀 Starting Week 6 Day 1 Performance Profiling...${NC}"

    # Check service health
    check_service_health "Go API" "$API_URL" || exit 1
    check_service_health "Python AI Worker" "$AI_WORKER_URL" || exit 1

    # Run profiling
    profile_go_api
    profile_python_worker
    profile_database
    profile_redis
    run_load_testing

    # Generate summary report
    cat > "$PERF_DIR/performance_summary.md" << EOF
# Week 6 Day 1: Performance Profiling Summary

**Date**: $(date)
**Status**: ✅ Completed Successfully

## 📊 Profiling Results

### Go API Service
- CPU Profile: \`cpu_profile.prof\`
- Memory Profile: \`memory_profile.prof\`
- Goroutine Profile: \`goroutine_profile.prof\`
- Block Profile: \`block_profile.prof\`
- Mutex Profile: \`mutex_profile.prof\`

### Python AI Worker
- CPU Profile: \`python_cpu_profile.prof\`
- Memory Stats: \`python_memory_stats.txt\`

### Database Performance
- Analysis Results: \`db_performance_results.txt\`

### Redis Performance
- Analysis Results: \`redis_performance_results.txt\`

### Load Testing
- Results: \`load_test_results.txt\`

## 🎯 Next Steps
1. Analyze profiling results
2. Identify performance bottlenecks
3. Implement optimizations
4. Validate improvements

## 📋 Files Generated
$(ls -la $PERF_DIR)
EOF

    echo -e "${GREEN}🎉 Week 6 Day 1 Performance Profiling Completed!${NC}"
    echo -e "${BLUE}📊 Results saved to: $PERF_DIR${NC}"
    echo -e "${YELLOW}📋 Summary report: $PERF_DIR/performance_summary.md${NC}"
}

# Execute main function
main "$@"
