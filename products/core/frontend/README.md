# AgentOS Core Frontend

> React web application for AgentOS Core

## Overview

The frontend application provides a modern web interface for managing AI agents, tools, and executions.

## Features

- **Agent Dashboard**: Visual agent management
- **Tool Library**: Browse and configure tools
- **Execution Monitor**: Real-time execution tracking
- **Memory Viewer**: Inspect agent memory
- **Settings**: Configuration management

## Tech Stack

- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **UI Library**: Material-UI / Tailwind CSS
- **State Management**: Redux Toolkit
- **API Client**: @agentos/api-client
- **Testing**: Jest + React Testing Library

## Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Run tests
npm run test
```

## Environment Variables

```bash
VITE_API_URL=http://localhost:8000
VITE_WS_URL=ws://localhost:8000/ws
VITE_APP_NAME=AgentOS Core
```