# @agentos/core

> Core types, utilities, and shared functionality

## Overview

The core package provides shared types, utilities, and common functionality used across the AgentOS ecosystem.

## Features

- TypeScript type definitions
- Common utilities and helpers
- Shared constants and enums
- API client interfaces
- Error handling utilities

## Installation

```bash
npm install @agentos/core
```

## Usage

```typescript
import { Agent, Tool, Capability } from '@agentos/core';
import { createApiClient } from '@agentos/core/client';

// Create API client
const client = createApiClient({
  baseURL: 'https://api.agentos.ai',
  apiKey: 'your-api-key'
});

// Use types
const agent: Agent = {
  id: 'agent-123',
  name: 'My Agent',
  capabilities: ['web-search', 'data-analysis']
};
```

## Exports

- **Types**: Agent, Tool, Capability, Execution
- **Utilities**: Logger, Validator, Formatter
- **Constants**: API endpoints, error codes
- **Client**: API client factory

## Development

```bash
# Build package
npm run build

# Run tests
npm run test

# Type checking
npm run type-check
```