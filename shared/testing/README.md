# @agentos/testing

> Testing utilities and helpers for AgentOS ecosystem

## Overview

Shared testing utilities, mocks, and helpers for testing AgentOS applications and services.

## Features

- Mock API clients and services
- Test data factories
- Custom Jest matchers
- Testing utilities for React components
- Integration test helpers

## Installation

```bash
npm install --save-dev @agentos/testing
```

## Usage

### Mock API Client

```typescript
import { createMockClient } from '@agentos/testing';

const mockClient = createMockClient();
mockClient.agents.list.mockResolvedValue([
  { id: '1', name: 'Test Agent' }
]);
```

### Test Data Factories

```typescript
import { AgentFactory, ToolFactory } from '@agentos/testing';

const agent = AgentFactory.build({
  name: 'Custom Agent',
  capabilities: ['web-search']
});

const tool = ToolFactory.build();
```

### Custom Matchers

```typescript
import '@agentos/testing/matchers';

expect(agent).toBeValidAgent();
expect(execution).toHaveStatus('completed');
```

## Utilities

- `createMockClient()` - Mock API client
- `AgentFactory` - Agent test data factory
- `ToolFactory` - Tool test data factory
- `ExecutionFactory` - Execution test data factory
- `renderWithProviders()` - React testing utility

## Development

```bash
# Build package
npm run build

# Run tests
npm run test
```