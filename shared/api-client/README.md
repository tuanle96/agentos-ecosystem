# @agentos/api-client

> TypeScript API client for AgentOS services

## Overview

A comprehensive TypeScript API client for interacting with AgentOS backend services.

## Features

- Full TypeScript support
- Automatic request/response validation
- Built-in error handling
- Authentication management
- Request/response interceptors
- Retry logic and rate limiting

## Installation

```bash
npm install @agentos/api-client
```

## Usage

```typescript
import { AgentOSClient } from '@agentos/api-client';

// Initialize client
const client = new AgentOSClient({
  baseURL: 'https://api.agentos.ai',
  apiKey: 'your-api-key'
});

// Use the client
async function example() {
  // List agents
  const agents = await client.agents.list();
  
  // Create agent
  const newAgent = await client.agents.create({
    name: 'Web Researcher',
    capabilities: ['web-search', 'data-analysis']
  });
  
  // Execute agent
  const execution = await client.executions.create({
    agentId: newAgent.id,
    input: 'Research the latest AI trends'
  });
  
  // Get execution status
  const status = await client.executions.get(execution.id);
}
```

## API Modules

- **agents**: Agent management
- **tools**: Tool registry operations
- **executions**: Agent execution
- **memory**: Memory operations
- **auth**: Authentication
- **billing**: Usage and billing

## Configuration

```typescript
const client = new AgentOSClient({
  baseURL: 'https://api.agentos.ai',
  apiKey: 'your-api-key',
  timeout: 30000,
  retries: 3,
  rateLimit: {
    requests: 100,
    window: 60000 // 1 minute
  }
});
```

## Development

```bash
# Build client
npm run build

# Run tests
npm run test

# Generate API docs
npm run docs
```