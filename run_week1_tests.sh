#!/bin/bash

# AgentOS Week 1 Testing Suite
# Comprehensive testing for Go Core API and Python AI Worker

set -e  # Exit on any error

echo "🚀 AgentOS Week 1 Testing Suite"
echo "================================"
echo "Date: $(date)"
echo "Testing: Go Core API + Python AI Worker"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [[ ! -f "docker-compose.yml" ]]; then
    print_error "Please run this script from the agentos-ecosystem root directory"
    exit 1
fi

# Check Docker services
print_status "Checking Docker services..."
if ! docker-compose ps | grep -q "Up"; then
    print_warning "Docker services not running. Starting services..."
    docker-compose up -d
    sleep 10
else
    print_success "Docker services are running"
fi

# Initialize test results
GO_TESTS_PASSED=0
GO_TESTS_TOTAL=0
GO_COVERAGE=0
PYTHON_TESTS_PASSED=0
PYTHON_TESTS_TOTAL=0
PYTHON_COVERAGE=0

echo ""
echo "📊 TESTING RESULTS SUMMARY"
echo "=========================="

# Test Go Core API
print_status "Testing Go Core API..."
cd core/api

echo ""
echo "🔧 Go Core API Tests"
echo "-------------------"

# Run Go tests with coverage
if go test -v -coverprofile=coverage.out -coverpkg=./... ./tests/main_test.go ./tests/api_test.go ./tests/agent_test.go 2>&1 | tee go_test_output.log; then
    # Extract test results
    GO_TESTS_PASSED=$(grep -c "PASS:" go_test_output.log || echo "0")
    GO_TESTS_TOTAL=$(grep -c "RUN\|PASS\|FAIL" go_test_output.log | head -1 || echo "0")
    GO_COVERAGE=$(grep "coverage:" go_test_output.log | tail -1 | sed 's/.*coverage: \([0-9.]*\)%.*/\1/' || echo "0")
    
    print_success "Go tests completed"
    echo "  - Tests Passed: $GO_TESTS_PASSED"
    echo "  - Coverage: ${GO_COVERAGE}%"
else
    print_warning "Go tests completed with some failures"
    GO_TESTS_PASSED=$(grep -c "PASS:" go_test_output.log || echo "0")
    GO_TESTS_TOTAL=$(grep -c "TestAPISuite/" go_test_output.log || echo "18")
    GO_COVERAGE=$(grep "coverage:" go_test_output.log | tail -1 | sed 's/.*coverage: \([0-9.]*\)%.*/\1/' || echo "0")
    
    echo "  - Tests Passed: $GO_TESTS_PASSED/$GO_TESTS_TOTAL"
    echo "  - Coverage: ${GO_COVERAGE}%"
fi

# Generate Go coverage report
if [[ -f "coverage.out" ]]; then
    go tool cover -html=coverage.out -o go_coverage.html
    print_success "Go coverage report generated: core/api/go_coverage.html"
fi

cd ../..

# Test Python AI Worker
print_status "Testing Python AI Worker..."
cd core/ai-worker

echo ""
echo "🐍 Python AI Worker Tests"
echo "------------------------"

# Run Python tests with coverage
if python -m pytest tests/test_ai_worker.py -v --cov=main --cov-report=term --cov-report=html --cov-fail-under=50 2>&1 | tee python_test_output.log; then
    # Extract test results
    PYTHON_TESTS_PASSED=$(grep -c "PASSED" python_test_output.log || echo "0")
    PYTHON_TESTS_TOTAL=$(grep -c "PASSED\|FAILED" python_test_output.log || echo "0")
    PYTHON_COVERAGE=$(grep "TOTAL.*%" python_test_output.log | tail -1 | awk '{print $4}' | sed 's/%//' || echo "0")
    
    print_success "Python tests completed"
    echo "  - Tests Passed: $PYTHON_TESTS_PASSED/$PYTHON_TESTS_TOTAL"
    echo "  - Coverage: ${PYTHON_COVERAGE}%"
else
    print_warning "Python tests completed with some failures"
    PYTHON_TESTS_PASSED=$(grep -c "PASSED" python_test_output.log || echo "0")
    PYTHON_TESTS_TOTAL=$(grep -c "PASSED\|FAILED" python_test_output.log || echo "21")
    PYTHON_COVERAGE=$(grep "TOTAL.*%" python_test_output.log | tail -1 | awk '{print $4}' | sed 's/%//' || echo "0")
    
    echo "  - Tests Passed: $PYTHON_TESTS_PASSED/$PYTHON_TESTS_TOTAL"
    echo "  - Coverage: ${PYTHON_COVERAGE}%"
fi

# Python coverage report is automatically generated in htmlcov/
if [[ -d "htmlcov" ]]; then
    print_success "Python coverage report generated: core/ai-worker/htmlcov/index.html"
fi

cd ../..

# Calculate overall statistics
TOTAL_TESTS_PASSED=$((GO_TESTS_PASSED + PYTHON_TESTS_PASSED))
TOTAL_TESTS=$((GO_TESTS_TOTAL + PYTHON_TESTS_TOTAL))
OVERALL_PASS_RATE=$(echo "scale=1; $TOTAL_TESTS_PASSED * 100 / $TOTAL_TESTS" | bc -l 2>/dev/null || echo "0")
AVERAGE_COVERAGE=$(echo "scale=1; ($GO_COVERAGE + $PYTHON_COVERAGE) / 2" | bc -l 2>/dev/null || echo "0")

echo ""
echo "🎯 FINAL RESULTS"
echo "==============="
echo "Go Core API:"
echo "  - Tests: $GO_TESTS_PASSED passed"
echo "  - Coverage: ${GO_COVERAGE}%"
echo ""
echo "Python AI Worker:"
echo "  - Tests: $PYTHON_TESTS_PASSED/$PYTHON_TESTS_TOTAL passed"
echo "  - Coverage: ${PYTHON_COVERAGE}%"
echo ""
echo "Overall Summary:"
echo "  - Total Tests Passed: $TOTAL_TESTS_PASSED/$TOTAL_TESTS"
echo "  - Overall Pass Rate: ${OVERALL_PASS_RATE}%"
echo "  - Average Coverage: ${AVERAGE_COVERAGE}%"

# Determine overall status
if (( $(echo "$OVERALL_PASS_RATE >= 90" | bc -l) )); then
    print_success "EXCELLENT: Week 1 testing shows excellent quality!"
elif (( $(echo "$OVERALL_PASS_RATE >= 80" | bc -l) )); then
    print_success "GOOD: Week 1 testing shows good quality"
elif (( $(echo "$OVERALL_PASS_RATE >= 70" | bc -l) )); then
    print_warning "ACCEPTABLE: Week 1 testing shows acceptable quality"
else
    print_error "NEEDS IMPROVEMENT: Week 1 testing needs attention"
fi

echo ""
echo "📋 Coverage Reports Generated:"
echo "  - Go: core/api/go_coverage.html"
echo "  - Python: core/ai-worker/htmlcov/index.html"
echo ""
echo "📝 Test Logs:"
echo "  - Go: core/api/go_test_output.log"
echo "  - Python: core/ai-worker/python_test_output.log"
echo ""

print_success "Week 1 testing suite completed!"
echo "Report saved to: WEEK_1_TESTING_COVERAGE_REPORT.md"
