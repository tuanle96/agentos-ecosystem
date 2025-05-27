# Testing Tools

> Comprehensive testing utilities for AgentOS ecosystem

## Overview

Testing tools and utilities for ensuring quality and reliability across the AgentOS ecosystem.

## Test Types

### Unit Tests
- **Go Services**: Table-driven tests with testify
- **Frontend**: Jest + React Testing Library
- **API Clients**: Mock-based testing
- **Utilities**: Pure function testing

### Integration Tests
- **Service-to-Service**: Cross-service communication
- **Database**: Database integration testing
- **External APIs**: Third-party API integration
- **Message Queues**: NATS integration testing

### End-to-End Tests
- **User Workflows**: Complete user journeys
- **API Workflows**: Full API request/response cycles
- **Browser Testing**: Selenium/Playwright automation
- **Mobile Testing**: React Native app testing

### Performance Tests
- **Load Testing**: High-volume request testing
- **Stress Testing**: System breaking point analysis
- **Benchmark Testing**: Performance regression detection
- **Memory Testing**: Memory leak detection

## Directory Structure

```
testing/
├── unit/           # Unit test utilities
├── integration/    # Integration test suite
├── e2e/           # End-to-end tests
├── load/          # Load testing scripts
├── fixtures/      # Test data and fixtures
├── mocks/         # Mock implementations
└── README.md      # This file
```

## Test Utilities

### Test Data Factories

```go
// Go test factories
func CreateTestAgent() *models.Agent {
    return &models.Agent{
        ID:   uuid.New(),
        Name: "Test Agent",
        Capabilities: []string{"web-search"},
    }
}
```

```typescript
// TypeScript test factories
export const createTestAgent = (): Agent => ({
  id: 'test-agent-123',
  name: 'Test Agent',
  capabilities: ['web-search', 'data-analysis']
});
```

### Mock Services

```go
// Mock API client
type MockAPIClient struct {
    agents []models.Agent
}

func (m *MockAPIClient) CreateAgent(req *CreateAgentRequest) (*models.Agent, error) {
    agent := &models.Agent{
        ID:   uuid.New(),
        Name: req.Name,
    }
    m.agents = append(m.agents, *agent)
    return agent, nil
}
```

### Test Helpers

```go
// Database test helpers
func SetupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    
    err = db.AutoMigrate(&models.Agent{}, &models.Tool{})
    require.NoError(t, err)
    
    return db
}
```

## Running Tests

### All Tests

```bash
# Run all tests
./tools/testing/run-all-tests.sh

# Run with coverage
./tools/testing/run-tests-with-coverage.sh

# Run specific test suite
./tools/testing/run-unit-tests.sh
./tools/testing/run-integration-tests.sh
./tools/testing/run-e2e-tests.sh
```

### Service-Specific Tests

```bash
# Test Go services
cd services/core-api && go test ./...
cd services/agent-engine && go test ./...

# Test frontend packages
cd packages/core && npm test
cd packages/ui-components && npm test

# Test products
cd products/core/frontend && npm test
```

### Load Testing

```bash
# Run load tests
./tools/testing/load-tests/run-load-test.sh

# Specific scenarios
./tools/testing/load-tests/agent-creation-load.sh
./tools/testing/load-tests/execution-load.sh
```

## Test Configuration

### Test Environment

```yaml
# test-config.yml
test_environment:
  database_url: "postgres://test:test@localhost:5432/agentos_test"
  redis_url: "redis://localhost:6379/1"
  api_base_url: "http://localhost:8000"
  
coverage:
  threshold: 80
  exclude:
    - "**/mocks/**"
    - "**/testdata/**"
    
load_testing:
  duration: "5m"
  users: 100
  ramp_up: "30s"
```

### CI/CD Integration

```yaml
# GitHub Actions example
name: Test Suite
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Tests
        run: ./tools/testing/run-all-tests.sh
      - name: Upload Coverage
        uses: codecov/codecov-action@v3
```

## Test Reports

### Coverage Reports
- **Go**: `go test -coverprofile=coverage.out`
- **JavaScript**: Jest coverage reports
- **Combined**: Aggregated coverage across languages

### Performance Reports
- **Benchmark Results**: Performance regression tracking
- **Load Test Results**: Throughput and latency metrics
- **Memory Profiling**: Memory usage analysis

### Quality Reports
- **Test Results**: Pass/fail status and trends
- **Flaky Test Detection**: Unreliable test identification
- **Test Duration**: Test execution time tracking