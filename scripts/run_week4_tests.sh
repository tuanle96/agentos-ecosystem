#!/bin/bash

# Week 4 Advanced Memory System - Comprehensive Test Runner
# This script runs all Week 4 unit tests for both Go and Python components

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
VERBOSE=${VERBOSE:-false}
COVERAGE=${COVERAGE:-true}
PARALLEL=${PARALLEL:-true}
OUTPUT_DIR="test_results"

# Create output directory
mkdir -p "$OUTPUT_DIR"

echo -e "${BLUE}🧪 Week 4 Advanced Memory System - Test Suite${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Function to print section headers
print_section() {
    echo -e "${YELLOW}$1${NC}"
    echo -e "${YELLOW}$(printf '=%.0s' $(seq 1 ${#1}))${NC}"
}

# Function to print test results
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✅ $2 PASSED${NC}"
    else
        echo -e "${RED}❌ $2 FAILED${NC}"
        return 1
    fi
}

# Function to run Go tests
run_go_tests() {
    print_section "🔧 Go Memory System Tests"
    
    cd core/api
    
    # Check if Go modules are initialized
    if [ ! -f "go.mod" ]; then
        echo -e "${YELLOW}Initializing Go modules...${NC}"
        go mod init agentos-core-api
        go mod tidy
    fi
    
    # Install test dependencies
    echo -e "${BLUE}Installing Go test dependencies...${NC}"
    go get github.com/stretchr/testify/assert
    go get github.com/stretchr/testify/mock
    go get github.com/gin-gonic/gin
    
    # Run Go unit tests
    echo -e "${BLUE}Running Go unit tests...${NC}"
    
    if [ "$COVERAGE" = true ]; then
        if [ "$PARALLEL" = true ]; then
            go test -v -race -coverprofile="../../$OUTPUT_DIR/go_coverage.out" -covermode=atomic ./tests/memory_handlers_unit_test.go ./tests/setup_test.go -parallel 4
        else
            go test -v -race -coverprofile="../../$OUTPUT_DIR/go_coverage.out" -covermode=atomic ./tests/memory_handlers_unit_test.go ./tests/setup_test.go
        fi
        GO_UNIT_RESULT=$?
        
        # Generate coverage report
        if [ $GO_UNIT_RESULT -eq 0 ]; then
            echo -e "${BLUE}Generating Go coverage report...${NC}"
            go tool cover -html="../../$OUTPUT_DIR/go_coverage.out" -o "../../$OUTPUT_DIR/go_coverage.html"
            go tool cover -func="../../$OUTPUT_DIR/go_coverage.out" > "../../$OUTPUT_DIR/go_coverage.txt"
            
            # Extract coverage percentage
            COVERAGE_PERCENT=$(go tool cover -func="../../$OUTPUT_DIR/go_coverage.out" | grep "total:" | awk '{print $3}')
            echo -e "${GREEN}Go Coverage: $COVERAGE_PERCENT${NC}"
        fi
    else
        if [ "$PARALLEL" = true ]; then
            go test -v -race ./tests/memory_handlers_unit_test.go ./tests/setup_test.go -parallel 4
        else
            go test -v -race ./tests/memory_handlers_unit_test.go ./tests/setup_test.go
        fi
        GO_UNIT_RESULT=$?
    fi
    
    print_result $GO_UNIT_RESULT "Go Unit Tests"
    
    # Run Go integration tests
    echo -e "${BLUE}Running Go integration tests...${NC}"
    
    if [ "$PARALLEL" = true ]; then
        go test -v -race ./tests/memory_integration_test.go ./tests/setup_test.go -parallel 2
    else
        go test -v -race ./tests/memory_integration_test.go ./tests/setup_test.go
    fi
    GO_INTEGRATION_RESULT=$?
    
    print_result $GO_INTEGRATION_RESULT "Go Integration Tests"
    
    # Run Go benchmarks
    echo -e "${BLUE}Running Go benchmarks...${NC}"
    go test -bench=. -benchmem ./tests/memory_handlers_unit_test.go ./tests/setup_test.go > "../../$OUTPUT_DIR/go_benchmarks.txt" 2>&1
    GO_BENCHMARK_RESULT=$?
    
    print_result $GO_BENCHMARK_RESULT "Go Benchmarks"
    
    cd ../..
    
    return $((GO_UNIT_RESULT + GO_INTEGRATION_RESULT))
}

# Function to run Python tests
run_python_tests() {
    print_section "🐍 Python Memory System Tests"
    
    cd core/ai-worker
    
    # Check if virtual environment exists
    if [ ! -d "venv" ]; then
        echo -e "${YELLOW}Creating Python virtual environment...${NC}"
        python3 -m venv venv
    fi
    
    # Activate virtual environment
    source venv/bin/activate
    
    # Install test dependencies
    echo -e "${BLUE}Installing Python test dependencies...${NC}"
    pip install -q pytest pytest-asyncio pytest-cov pytest-mock pytest-benchmark
    pip install -q -r requirements.txt
    
    # Run Python unit tests
    echo -e "${BLUE}Running Python unit tests...${NC}"
    
    if [ "$COVERAGE" = true ]; then
        if [ "$PARALLEL" = true ]; then
            pytest tests/test_mem0_memory_engine.py tests/test_framework_adapters.py -v --tb=short --cov=memory --cov-report=html:"../../$OUTPUT_DIR/python_coverage_html" --cov-report=term --cov-report=xml:"../../$OUTPUT_DIR/python_coverage.xml" -n auto
        else
            pytest tests/test_mem0_memory_engine.py tests/test_framework_adapters.py -v --tb=short --cov=memory --cov-report=html:"../../$OUTPUT_DIR/python_coverage_html" --cov-report=term --cov-report=xml:"../../$OUTPUT_DIR/python_coverage.xml"
        fi
        PYTHON_UNIT_RESULT=$?
    else
        if [ "$PARALLEL" = true ]; then
            pytest tests/test_mem0_memory_engine.py tests/test_framework_adapters.py -v --tb=short -n auto
        else
            pytest tests/test_mem0_memory_engine.py tests/test_framework_adapters.py -v --tb=short
        fi
        PYTHON_UNIT_RESULT=$?
    fi
    
    print_result $PYTHON_UNIT_RESULT "Python Unit Tests"
    
    # Run Python performance tests
    echo -e "${BLUE}Running Python performance tests...${NC}"
    pytest tests/test_mem0_memory_engine.py::TestMem0MemoryEnginePerformance -v --tb=short --benchmark-only --benchmark-json="../../$OUTPUT_DIR/python_benchmarks.json" 2>/dev/null || true
    PYTHON_PERF_RESULT=$?
    
    print_result $PYTHON_PERF_RESULT "Python Performance Tests"
    
    # Deactivate virtual environment
    deactivate
    
    cd ../..
    
    return $PYTHON_UNIT_RESULT
}

# Function to run mem0 integration tests
run_mem0_integration_tests() {
    print_section "🧠 mem0 Integration Tests"
    
    cd core/api
    
    echo -e "${BLUE}Running mem0 integration tests...${NC}"
    
    if [ "$PARALLEL" = true ]; then
        go test -v -race ./tests/week4_mem0_integration_test.go ./tests/setup_test.go -parallel 2
    else
        go test -v -race ./tests/week4_mem0_integration_test.go ./tests/setup_test.go
    fi
    MEM0_INTEGRATION_RESULT=$?
    
    print_result $MEM0_INTEGRATION_RESULT "mem0 Integration Tests"
    
    cd ../..
    
    return $MEM0_INTEGRATION_RESULT
}

# Function to generate test summary
generate_test_summary() {
    print_section "📊 Test Summary Report"
    
    SUMMARY_FILE="$OUTPUT_DIR/test_summary.txt"
    
    echo "Week 4 Advanced Memory System - Test Summary" > "$SUMMARY_FILE"
    echo "=============================================" >> "$SUMMARY_FILE"
    echo "Date: $(date)" >> "$SUMMARY_FILE"
    echo "" >> "$SUMMARY_FILE"
    
    echo "Test Results:" >> "$SUMMARY_FILE"
    echo "- Go Unit Tests: $([ $GO_TESTS_RESULT -eq 0 ] && echo "PASSED" || echo "FAILED")" >> "$SUMMARY_FILE"
    echo "- Python Unit Tests: $([ $PYTHON_TESTS_RESULT -eq 0 ] && echo "PASSED" || echo "FAILED")" >> "$SUMMARY_FILE"
    echo "- mem0 Integration Tests: $([ $MEM0_TESTS_RESULT -eq 0 ] && echo "PASSED" || echo "FAILED")" >> "$SUMMARY_FILE"
    echo "" >> "$SUMMARY_FILE"
    
    if [ "$COVERAGE" = true ]; then
        echo "Coverage Reports:" >> "$SUMMARY_FILE"
        echo "- Go Coverage: $OUTPUT_DIR/go_coverage.html" >> "$SUMMARY_FILE"
        echo "- Python Coverage: $OUTPUT_DIR/python_coverage_html/index.html" >> "$SUMMARY_FILE"
        echo "" >> "$SUMMARY_FILE"
    fi
    
    echo "Benchmark Reports:" >> "$SUMMARY_FILE"
    echo "- Go Benchmarks: $OUTPUT_DIR/go_benchmarks.txt" >> "$SUMMARY_FILE"
    echo "- Python Benchmarks: $OUTPUT_DIR/python_benchmarks.json" >> "$SUMMARY_FILE"
    echo "" >> "$SUMMARY_FILE"
    
    # Calculate overall result
    OVERALL_RESULT=$((GO_TESTS_RESULT + PYTHON_TESTS_RESULT + MEM0_TESTS_RESULT))
    
    if [ $OVERALL_RESULT -eq 0 ]; then
        echo -e "${GREEN}🎉 ALL TESTS PASSED!${NC}"
        echo "Overall Result: PASSED" >> "$SUMMARY_FILE"
    else
        echo -e "${RED}❌ SOME TESTS FAILED${NC}"
        echo "Overall Result: FAILED" >> "$SUMMARY_FILE"
    fi
    
    echo ""
    echo -e "${BLUE}📁 Test results saved to: $OUTPUT_DIR/${NC}"
    echo -e "${BLUE}📄 Summary report: $SUMMARY_FILE${NC}"
    
    return $OVERALL_RESULT
}

# Main execution
main() {
    echo -e "${BLUE}Starting Week 4 Memory System Tests...${NC}"
    echo ""
    
    # Check prerequisites
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go is not installed${NC}"
        exit 1
    fi
    
    if ! command -v python3 &> /dev/null; then
        echo -e "${RED}❌ Python3 is not installed${NC}"
        exit 1
    fi
    
    # Run test suites
    run_go_tests
    GO_TESTS_RESULT=$?
    
    echo ""
    
    run_python_tests
    PYTHON_TESTS_RESULT=$?
    
    echo ""
    
    run_mem0_integration_tests
    MEM0_TESTS_RESULT=$?
    
    echo ""
    
    # Generate summary
    generate_test_summary
    SUMMARY_RESULT=$?
    
    exit $SUMMARY_RESULT
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        --no-coverage)
            COVERAGE=false
            shift
            ;;
        --no-parallel)
            PARALLEL=false
            shift
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -h|--help)
            echo "Week 4 Memory System Test Runner"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -v, --verbose      Enable verbose output"
            echo "  --no-coverage      Disable coverage reporting"
            echo "  --no-parallel      Disable parallel test execution"
            echo "  -o, --output DIR   Set output directory (default: test_results)"
            echo "  -h, --help         Show this help message"
            echo ""
            echo "Environment Variables:"
            echo "  VERBOSE=true       Enable verbose output"
            echo "  COVERAGE=false     Disable coverage reporting"
            echo "  PARALLEL=false     Disable parallel execution"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Run main function
main
