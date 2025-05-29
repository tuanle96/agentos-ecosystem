# @agentos/ui-components

> Shared React UI components for AgentOS ecosystem

## Overview

A comprehensive library of reusable React components designed specifically for AgentOS applications.

## Features

- Modern React components with TypeScript
- Consistent design system
- Accessibility (a11y) compliant
- Storybook documentation
- Theme support

## Installation

```bash
npm install @agentos/ui-components
```

## Usage

```tsx
import { Button, Card, AgentCard } from '@agentos/ui-components';
import '@agentos/ui-components/dist/styles.css';

function App() {
  return (
    <Card>
      <AgentCard
        agent={{
          id: 'agent-123',
          name: 'Web Researcher',
          status: 'active'
        }}
        onExecute={() => console.log('Execute agent')}
      />
      <Button variant="primary" size="large">
        Create Agent
      </Button>
    </Card>
  );
}
```

## Components

### Core Components
- `Button` - Various button styles and states
- `Card` - Container component
- `Modal` - Modal dialogs
- `Input` - Form inputs
- `Select` - Dropdown selectors

### AgentOS Specific
- `AgentCard` - Agent display card
- `ToolCard` - Tool display card
- `ExecutionLog` - Execution history
- `CapabilityBadge` - Capability indicators

## Development

```bash
# Start Storybook
npm run storybook

# Build components
npm run build

# Run tests
npm run test
```