#!/bin/bash

# Test script for File Operations Mock Elimination
# Tests real file operations implementation in Go API

echo "🚀 Testing File Operations Mock Elimination"
echo "=========================================="

# Base URL for API
API_URL="http://localhost:8080/api/tools/execute"

# Test JWT token (you may need to get a real token)
JWT_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdF91c2VyIiwiZXhwIjo5OTk5OTk5OTk5fQ.test"

echo "📁 Test 1: Create Directory"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "tool_name": "file_operations",
    "parameters": {
      "operation": "create_dir",
      "path": "test_dir"
    }
  }' | jq .

echo ""
echo "📝 Test 2: Write File"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "tool_name": "file_operations",
    "parameters": {
      "operation": "write",
      "path": "test_dir/hello.txt",
      "content": "Hello AgentOS! This is a real file operation, not a mock!"
    }
  }' | jq .

echo ""
echo "📖 Test 3: Read File"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "tool_name": "file_operations",
    "parameters": {
      "operation": "read",
      "path": "test_dir/hello.txt"
    }
  }' | jq .

echo ""
echo "📋 Test 4: List Directory"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "tool_name": "file_operations",
    "parameters": {
      "operation": "list",
      "path": "test_dir"
    }
  }' | jq .

echo ""
echo "❓ Test 5: Check File Exists"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "tool_name": "file_operations",
    "parameters": {
      "operation": "exists",
      "path": "test_dir/hello.txt"
    }
  }' | jq .

echo ""
echo "🛡️ Test 6: Security Test - Path Traversal (Should Fail)"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "tool_name": "file_operations",
    "parameters": {
      "operation": "read",
      "path": "../../../etc/passwd"
    }
  }' | jq .

echo ""
echo "🗑️ Test 7: Delete File"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "tool_name": "file_operations",
    "parameters": {
      "operation": "delete",
      "path": "test_dir/hello.txt"
    }
  }' | jq .

echo ""
echo "✅ File Operations Mock Elimination Test Complete!"
echo "Check /tmp/agentos_files for actual file operations"
