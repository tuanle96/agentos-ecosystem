# Auth Service

> Authentication and authorization service

## Overview

The Auth Service handles user authentication, authorization, and access control across the AgentOS ecosystem.

## Features

- JWT token management
- Role-based access control (RBAC)
- Multi-factor authentication
- OAuth2 integration
- Session management

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **JWT**: golang-jwt/jwt
- **Database**: PostgreSQL
- **Cache**: Redis

## Development

```bash
# Build service
cd services/auth-service && go build

# Run tests
cd services/auth-service && go test ./...
```