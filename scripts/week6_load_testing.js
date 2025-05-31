/**
 * Week 6 Day 1: Load Testing Script with k6
 * AgentOS Performance Optimization Implementation
 * 
 * This script tests the AgentOS API under various load conditions
 * to identify performance bottlenecks and validate scalability.
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const responseTime = new Trend('response_time');
const requestCount = new Counter('requests');

// Test configuration
export const options = {
  stages: [
    // Warm-up
    { duration: '30s', target: 10 },
    
    // Ramp up to 100 users
    { duration: '1m', target: 100 },
    
    // Stay at 100 users
    { duration: '2m', target: 100 },
    
    // Ramp up to 500 users
    { duration: '1m', target: 500 },
    
    // Stay at 500 users
    { duration: '2m', target: 500 },
    
    // Ramp up to 1000 users (stress test)
    { duration: '1m', target: 1000 },
    
    // Stay at 1000 users
    { duration: '2m', target: 1000 },
    
    // Ramp up to 2000 users (peak test)
    { duration: '1m', target: 2000 },
    
    // Stay at 2000 users
    { duration: '1m', target: 2000 },
    
    // Ramp down
    { duration: '1m', target: 0 },
  ],
  
  thresholds: {
    // 95% of requests should be below 15ms (Week 6 target: <5ms)
    'http_req_duration': ['p(95)<15'],
    
    // Error rate should be below 1%
    'errors': ['rate<0.01'],
    
    // 99% of requests should be below 50ms
    'http_req_duration': ['p(99)<50'],
  },
};

// Base URL configuration
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8000';

// Test data
const testUsers = [
  { email: 'test1@example.com', password: 'testpass123' },
  { email: 'test2@example.com', password: 'testpass123' },
  { email: 'test3@example.com', password: 'testpass123' },
  { email: 'test4@example.com', password: 'testpass123' },
  { email: 'test5@example.com', password: 'testpass123' },
];

let authTokens = [];

// Setup function - runs once before the test
export function setup() {
  console.log('🚀 Setting up load test...');
  
  // Register test users and get auth tokens
  const tokens = [];
  
  for (let i = 0; i < testUsers.length; i++) {
    const user = testUsers[i];
    
    // Register user
    const registerResponse = http.post(`${BASE_URL}/api/v1/auth/register`, JSON.stringify({
      email: user.email,
      password: user.password,
      name: `Test User ${i + 1}`
    }), {
      headers: { 'Content-Type': 'application/json' },
    });
    
    if (registerResponse.status === 201 || registerResponse.status === 409) {
      // Login to get token
      const loginResponse = http.post(`${BASE_URL}/api/v1/auth/login`, JSON.stringify({
        email: user.email,
        password: user.password
      }), {
        headers: { 'Content-Type': 'application/json' },
      });
      
      if (loginResponse.status === 200) {
        const loginData = JSON.parse(loginResponse.body);
        tokens.push(loginData.token);
        console.log(`✅ User ${i + 1} authenticated`);
      } else {
        console.log(`❌ Failed to authenticate user ${i + 1}`);
      }
    }
    
    sleep(0.1); // Small delay between registrations
  }
  
  console.log(`🔑 ${tokens.length} auth tokens obtained`);
  return { tokens: tokens };
}

// Main test function
export default function(data) {
  const token = data.tokens[Math.floor(Math.random() * data.tokens.length)];
  
  // Test scenario weights
  const scenario = Math.random();
  
  if (scenario < 0.3) {
    // 30% - Health check and basic endpoints
    testBasicEndpoints();
  } else if (scenario < 0.6) {
    // 30% - Agent operations
    testAgentOperations(token);
  } else if (scenario < 0.8) {
    // 20% - Tool operations
    testToolOperations(token);
  } else if (scenario < 0.95) {
    // 15% - Memory operations
    testMemoryOperations(token);
  } else {
    // 5% - Performance monitoring
    testPerformanceEndpoints(token);
  }
  
  // Random sleep between 1-3 seconds
  sleep(Math.random() * 2 + 1);
}

function testBasicEndpoints() {
  const responses = http.batch([
    ['GET', `${BASE_URL}/health`],
    ['GET', `${BASE_URL}/api/v1/tools`],
  ]);
  
  responses.forEach((response, index) => {
    const success = check(response, {
      'status is 200': (r) => r.status === 200,
      'response time < 15ms': (r) => r.timings.duration < 15,
    });
    
    errorRate.add(!success);
    responseTime.add(response.timings.duration);
    requestCount.add(1);
  });
}

function testAgentOperations(token) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };
  
  // List agents
  let response = http.get(`${BASE_URL}/api/v1/agents`, { headers });
  
  let success = check(response, {
    'list agents status is 200': (r) => r.status === 200,
    'list agents response time < 15ms': (r) => r.timings.duration < 15,
  });
  
  errorRate.add(!success);
  responseTime.add(response.timings.duration);
  requestCount.add(1);
  
  // Create agent (10% chance)
  if (Math.random() < 0.1) {
    const agentData = {
      name: `Load Test Agent ${Math.floor(Math.random() * 1000)}`,
      description: 'Agent created during load testing',
      capabilities: ['web_search', 'text_processing'],
      framework_preference: 'auto'
    };
    
    response = http.post(`${BASE_URL}/api/v1/agents`, JSON.stringify(agentData), { headers });
    
    success = check(response, {
      'create agent status is 201': (r) => r.status === 201,
      'create agent response time < 50ms': (r) => r.timings.duration < 50,
    });
    
    errorRate.add(!success);
    responseTime.add(response.timings.duration);
    requestCount.add(1);
    
    // If agent created successfully, test execution
    if (response.status === 201) {
      const agentData = JSON.parse(response.body);
      const agentId = agentData.data.id;
      
      // Execute agent
      const executionData = {
        task: 'Perform a simple test task for load testing',
        context: { test: 'load_testing' }
      };
      
      response = http.post(`${BASE_URL}/api/v1/agents/${agentId}/execute`, 
                          JSON.stringify(executionData), { headers });
      
      success = check(response, {
        'execute agent status is 200': (r) => r.status === 200,
        'execute agent response time < 100ms': (r) => r.timings.duration < 100,
      });
      
      errorRate.add(!success);
      responseTime.add(response.timings.duration);
      requestCount.add(1);
    }
  }
}

function testToolOperations(token) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };
  
  // Get tool definitions
  let response = http.get(`${BASE_URL}/api/v1/tools/definitions`, { headers });
  
  let success = check(response, {
    'tool definitions status is 200': (r) => r.status === 200,
    'tool definitions response time < 15ms': (r) => r.timings.duration < 15,
  });
  
  errorRate.add(!success);
  responseTime.add(response.timings.duration);
  requestCount.add(1);
  
  // Execute tool (5% chance)
  if (Math.random() < 0.05) {
    const toolData = {
      tool_name: 'web_search',
      parameters: {
        query: 'load testing performance',
        max_results: 5
      }
    };
    
    response = http.post(`${BASE_URL}/api/v1/tools/execute`, JSON.stringify(toolData), { headers });
    
    success = check(response, {
      'execute tool status is 200': (r) => r.status === 200,
      'execute tool response time < 200ms': (r) => r.timings.duration < 200,
    });
    
    errorRate.add(!success);
    responseTime.add(response.timings.duration);
    requestCount.add(1);
  }
}

function testMemoryOperations(token) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };
  
  // Test memory endpoints (if available)
  const responses = http.batch([
    ['GET', `${BASE_URL}/api/v1/memory/agents/test-agent`, { headers }],
  ]);
  
  responses.forEach((response) => {
    const success = check(response, {
      'memory operation status is 200 or 404': (r) => r.status === 200 || r.status === 404,
      'memory operation response time < 50ms': (r) => r.timings.duration < 50,
    });
    
    errorRate.add(!success);
    responseTime.add(response.timings.duration);
    requestCount.add(1);
  });
}

function testPerformanceEndpoints(token) {
  const headers = {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${token}`,
  };
  
  // Test performance monitoring endpoints (Week 6)
  const responses = http.batch([
    ['GET', `${BASE_URL}/api/v1/performance/metrics`, { headers }],
    ['GET', `${BASE_URL}/api/v1/performance/health`, { headers }],
    ['GET', `${BASE_URL}/api/v1/performance/benchmark`, { headers }],
  ]);
  
  responses.forEach((response, index) => {
    const endpointNames = ['metrics', 'health', 'benchmark'];
    const success = check(response, {
      [`${endpointNames[index]} status is 200`]: (r) => r.status === 200,
      [`${endpointNames[index]} response time < 100ms`]: (r) => r.timings.duration < 100,
    });
    
    errorRate.add(!success);
    responseTime.add(response.timings.duration);
    requestCount.add(1);
  });
}

// Teardown function - runs once after the test
export function teardown(data) {
  console.log('🧹 Cleaning up after load test...');
  
  // Could add cleanup logic here if needed
  // For now, just log completion
  console.log('✅ Load test completed successfully');
}

// Handle summary data
export function handleSummary(data) {
  const summary = {
    timestamp: new Date().toISOString(),
    test_duration: data.state.testRunDurationMs / 1000,
    total_requests: data.metrics.requests.values.count,
    error_rate: data.metrics.errors.values.rate,
    avg_response_time: data.metrics.http_req_duration.values.avg,
    p95_response_time: data.metrics.http_req_duration.values['p(95)'],
    p99_response_time: data.metrics.http_req_duration.values['p(99)'],
    max_response_time: data.metrics.http_req_duration.values.max,
    requests_per_second: data.metrics.http_reqs.values.rate,
    thresholds_passed: Object.keys(data.thresholds).every(
      threshold => data.thresholds[threshold].ok
    )
  };
  
  console.log('\n📊 LOAD TEST SUMMARY');
  console.log('='.repeat(50));
  console.log(`Duration: ${summary.test_duration}s`);
  console.log(`Total Requests: ${summary.total_requests}`);
  console.log(`Requests/sec: ${summary.requests_per_second.toFixed(2)}`);
  console.log(`Error Rate: ${(summary.error_rate * 100).toFixed(2)}%`);
  console.log(`Avg Response Time: ${summary.avg_response_time.toFixed(2)}ms`);
  console.log(`95th Percentile: ${summary.p95_response_time.toFixed(2)}ms`);
  console.log(`99th Percentile: ${summary.p99_response_time.toFixed(2)}ms`);
  console.log(`Max Response Time: ${summary.max_response_time.toFixed(2)}ms`);
  console.log(`Thresholds Passed: ${summary.thresholds_passed ? '✅' : '❌'}`);
  
  return {
    'performance_results/week6_day1/load_test_summary.json': JSON.stringify(summary, null, 2),
    stdout: JSON.stringify(summary, null, 2),
  };
}
