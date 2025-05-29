# Tool Registry Service

> Tool discovery, registration, and execution service

## Overview

The Tool Registry service manages the ecosystem of tools and capabilities available to agents. Provides discovery, registration, and execution of tools.

## Features

- Tool registration and discovery
- Tool execution and sandboxing
- Capability composition
- Tool marketplace integration
- Security and validation

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **Storage**: MinIO (S3-compatible)
- **Sandboxing**: Docker containers
- **Registry**: Tool metadata storage

## Development

```bash
# Build service
make build-tool-registry

# Run tests
make test-tool-registry

# Start with hot reload
cd services/tool-registry && air
```

## Key Components

- **Tool Registry**: Tool metadata management
- **Execution Engine**: Sandboxed tool execution
- **Capability Composer**: Combine tools into capabilities
- **Security Scanner**: Tool security validation
- **Marketplace API**: Integration with AgentOS Store

## Environment Variables

```bash
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=agentos
MINIO_SECRET_KEY=password
DOCKER_HOST=unix:///var/run/docker.sock
```