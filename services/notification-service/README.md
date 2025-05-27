# Notification Service

> Real-time notifications and messaging service

## Overview

The Notification Service handles real-time notifications, webhooks, and messaging across the AgentOS ecosystem.

## Features

- Real-time notifications
- Webhook management
- Email and SMS notifications
- WebSocket connections
- Event streaming

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **WebSockets**: Gorilla WebSocket
- **Message Queue**: NATS
- **Email**: SendGrid integration

## Development

```bash
# Build service
cd services/notification-service && go build

# Run tests
cd services/notification-service && go test ./...
```