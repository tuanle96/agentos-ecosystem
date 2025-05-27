# Agent Engine Service

> AI agent execution and orchestration engine

## Overview

The Agent Engine service handles agent creation, execution, and lifecycle management. Integrates with multiple AI frameworks.

## Features

- Agent lifecycle management
- Multi-framework support (Swarms, LangChain, CrewAI, AutoGen)
- Execution monitoring and logging
- Resource management and scaling
- Tool integration and execution

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **AI Integration**: HTTP APIs to Python services
- **Message Queue**: NATS
- **Monitoring**: Prometheus metrics

## Development

```bash
# Build service
make build-agent-engine

# Run tests
make test-agent-engine

# Start with hot reload
cd services/agent-engine && air
```

## Key Components

- **Agent Factory**: Create agents from specifications
- **Execution Engine**: Run agent tasks
- **Framework Adapters**: Interface with AI frameworks
- **Resource Manager**: Manage compute resources
- **Event Publisher**: Publish execution events

## Environment Variables

```bash
CORE_API_URL=http://core-api:8000
NATS_URL=nats://nats:4222
OPENAI_API_KEY=your-key
ANTHROPIC_API_KEY=your-key
```