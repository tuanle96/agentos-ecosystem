# API Reference Documentation

> Comprehensive API documentation for AgentOS services

## Overview

This section contains detailed API documentation for all AgentOS services, including request/response schemas, authentication, and usage examples.

## API Services

### Core Services
- [Core API](core-api.md) - Main API gateway and orchestration
- [Agent Engine API](agent-engine.md) - Agent creation and execution
- [Memory Service API](memory-service.md) - Memory management and retrieval
- [Tool Registry API](tool-registry.md) - Tool discovery and execution

### Supporting Services
- [Auth Service API](auth-service.md) - Authentication and authorization
- [Billing Service API](billing-service.md) - Usage tracking and billing
- [Notification Service API](notification-service.md) - Real-time notifications

## API Standards

### Base URL
```
Production: https://api.agentos.ai/v1
Staging: https://staging-api.agentos.ai/v1
Development: http://localhost:8000/v1
```

### Authentication
All API requests require authentication using API keys or JWT tokens:

```bash
# API Key Authentication
curl -H "X-API-Key: your-api-key" https://api.agentos.ai/v1/agents

# JWT Token Authentication
curl -H "Authorization: Bearer your-jwt-token" https://api.agentos.ai/v1/agents
```

### Request/Response Format
- **Content Type**: `application/json`
- **Character Encoding**: UTF-8
- **Date Format**: ISO 8601 (e.g., `2024-01-15T10:30:00Z`)
- **ID Format**: UUID v4 (e.g., `550e8400-e29b-41d4-a716-446655440000`)

### HTTP Status Codes
- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request parameters
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

### Error Response Format
```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "The request is invalid",
    "details": {
      "field": "name",
      "reason": "Name is required"
    },
    "request_id": "req_123456789"
  }
}
```

### Rate Limiting
API requests are rate limited per API key:
- **Free Tier**: 100 requests/hour
- **Pro Tier**: 10,000 requests/hour
- **Enterprise**: Custom limits

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 10000
X-RateLimit-Remaining: 9999
X-RateLimit-Reset: 1640995200
```

## Common Data Types

### Agent
```json
{
  "id": "agent_123456789",
  "name": "Web Researcher",
  "description": "An agent that can research topics on the web",
  "capabilities": ["web-search", "data-analysis"],
  "status": "active",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  "metadata": {
    "version": "1.0.0",
    "tags": ["research", "web"]
  }
}
```

### Tool
```json
{
  "id": "tool_123456789",
  "name": "web-search",
  "description": "Search the web for information",
  "version": "1.2.0",
  "category": "search",
  "parameters": {
    "query": {
      "type": "string",
      "required": true,
      "description": "Search query"
    },
    "max_results": {
      "type": "integer",
      "default": 10,
      "description": "Maximum number of results"
    }
  },
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Execution
```json
{
  "id": "exec_123456789",
  "agent_id": "agent_123456789",
  "input": "Research the latest AI trends",
  "status": "running",
  "progress": 0.5,
  "result": null,
  "error": null,
  "started_at": "2024-01-15T10:30:00Z",
  "completed_at": null,
  "metadata": {
    "estimated_duration": 300,
    "priority": "normal"
  }
}
```

## Quick Start Examples

### Create an Agent
```bash
curl -X POST https://api.agentos.ai/v1/agents \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Web Researcher",
    "description": "An agent that researches topics on the web",
    "capabilities": ["web-search", "data-analysis"]
  }'
```

### Execute an Agent
```bash
curl -X POST https://api.agentos.ai/v1/executions \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "agent_id": "agent_123456789",
    "input": "Research the latest AI trends in 2024"
  }'
```

### Get Execution Status
```bash
curl https://api.agentos.ai/v1/executions/exec_123456789 \
  -H "X-API-Key: your-api-key"
```

## SDKs and Client Libraries

### Official SDKs
- [Go SDK](../sdk/go-sdk.md) - Native Go client
- [Python SDK](../sdk/python-sdk.md) - Python client library
- [JavaScript SDK](../sdk/javascript-sdk.md) - Browser and Node.js
- [Rust SDK](../sdk/rust-sdk.md) - High-performance Rust client

### Community SDKs
- Java SDK (community-maintained)
- C# SDK (community-maintained)
- PHP SDK (community-maintained)

## OpenAPI Specification

The complete API specification is available in OpenAPI 3.0 format:
- [OpenAPI Spec](openapi.yaml) - Complete API specification
- [Swagger UI](https://api.agentos.ai/docs) - Interactive API explorer
- [Redoc](https://api.agentos.ai/redoc) - Alternative API documentation

## Webhooks

AgentOS supports webhooks for real-time event notifications:

### Supported Events
- `agent.created` - Agent created
- `agent.updated` - Agent updated
- `execution.started` - Execution started
- `execution.completed` - Execution completed
- `execution.failed` - Execution failed

### Webhook Configuration
```bash
curl -X POST https://api.agentos.ai/v1/webhooks \
  -H "X-API-Key: your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-app.com/webhooks/agentos",
    "events": ["execution.completed", "execution.failed"],
    "secret": "your-webhook-secret"
  }'
```

## Testing and Development

### Sandbox Environment
Use the sandbox environment for testing:
```
Sandbox URL: https://sandbox-api.agentos.ai/v1
```

### API Testing Tools
- [Postman Collection](postman-collection.json) - Ready-to-use Postman collection
- [Insomnia Workspace](insomnia-workspace.json) - Insomnia API workspace
- [curl Examples](curl-examples.md) - Command-line examples

## Support and Resources

### Getting Help
- **API Issues**: [GitHub Issues](https://github.com/tuanle96/agentos-ecosystem/issues)
- **Developer Support**: developers@agentos.ai
- **Community Forum**: [community.agentos.ai](https://community.agentos.ai)
- **Discord**: [discord.gg/agentos](https://discord.gg/agentos)

### Additional Resources
- [API Changelog](changelog.md) - API version history
- [Migration Guides](migrations/) - Version migration guides
- [Best Practices](best-practices.md) - API usage best practices
- [Troubleshooting](troubleshooting.md) - Common issues and solutions