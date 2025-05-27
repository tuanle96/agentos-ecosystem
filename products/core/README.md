# AgentOS Core

> Open source foundation for AI agent management

## Overview

AgentOS Core is the open source foundation of the AgentOS ecosystem. It provides basic agent creation, management, and execution capabilities.

## Features

- **Agent Management**: Create, configure, and manage AI agents
- **Tool Integration**: Connect agents with various tools and APIs
- **Memory System**: Basic working memory for agents
- **Execution Engine**: Run agent tasks and workflows
- **Web Interface**: React-based dashboard
- **CLI Tool**: Command-line interface for developers

## License

**MIT License** - Open source and free to use

## Tech Stack

- **Backend**: Go + Gin + GORM
- **Frontend**: React + TypeScript + Vite
- **Database**: PostgreSQL
- **Cache**: Redis
- **AI**: OpenAI, Anthropic integration

## Quick Start

```bash
# Clone repository
git clone https://github.com/tuanle96/agentos-ecosystem.git
cd agentos-ecosystem/products/core

# Start development
make dev

# Access dashboard
open http://localhost:3001
```

## Directory Structure

```
products/core/
├── frontend/          # React web application
├── cli/              # Command-line interface
├── docs/             # Documentation
├── examples/         # Usage examples
└── README.md         # This file
```

## API Endpoints

- `GET /api/agents` - List agents
- `POST /api/agents` - Create agent
- `GET /api/agents/:id` - Get agent details
- `POST /api/agents/:id/execute` - Execute agent
- `GET /api/tools` - List available tools

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](../../docs/contributing/CONTRIBUTING.md) for guidelines.

## Support

- **Documentation**: [docs.agentos.ai](https://docs.agentos.ai)
- **Community**: [GitHub Discussions](https://github.com/tuanle96/agentos-ecosystem/discussions)
- **Issues**: [GitHub Issues](https://github.com/tuanle96/agentos-ecosystem/issues)