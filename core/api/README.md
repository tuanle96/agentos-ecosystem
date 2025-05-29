# Core API Service

> Main API gateway for AgentOS ecosystem

## Overview

The Core API service is the main entry point for all AgentOS operations. Built with Go + Gin framework.

## Features

- RESTful API with OpenAPI documentation
- JWT authentication and authorization
- Rate limiting and request validation
- Health checks and metrics
- Database operations with GORM

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL with pgvector
- **Cache**: Redis
- **Documentation**: Swagger/OpenAPI

## Development

```bash
# Build service
make build-core-api

# Run tests
make test-core-api

# Start with hot reload
cd services/core-api && air
```

## API Endpoints

- `GET /health` - Health check
- `POST /auth/login` - User authentication
- `GET /agents` - List agents
- `POST /agents` - Create agent
- `GET /tools` - List tools
- `POST /executions` - Execute agent

## Environment Variables

```bash
DATABASE_URL=postgres://...
REDIS_URL=redis://...
JWT_SECRET=your-secret
OPENAI_API_KEY=your-key
```