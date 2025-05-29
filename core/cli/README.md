# AgentOS Core CLI

> Command-line interface for AgentOS Core

## Overview

The CLI provides a powerful command-line interface for managing agents, tools, and executions from the terminal.

## Installation

```bash
# Install globally
npm install -g @agentos/cli

# Or use npx
npx @agentos/cli --help
```

## Usage

### Authentication

```bash
# Login to AgentOS
agentos auth login

# Set API endpoint
agentos config set api-url https://api.agentos.ai
```

### Agent Management

```bash
# List agents
agentos agents list

# Create agent
agentos agents create --name "Web Researcher" --capabilities web-search,data-analysis

# Get agent details
agentos agents get agent-123

# Execute agent
agentos agents execute agent-123 "Research AI trends in 2024"
```

### Tool Management

```bash
# List tools
agentos tools list

# Install tool
agentos tools install web-search

# Tool details
agentos tools info web-search
```

### Execution Management

```bash
# List executions
agentos executions list

# Get execution status
agentos executions get exec-456

# Stream execution logs
agentos executions logs exec-456 --follow
```

## Configuration

Configuration is stored in `~/.agentos/config.json`:

```json
{
  "api-url": "https://api.agentos.ai",
  "api-key": "your-api-key",
  "default-agent": "agent-123"
}
```

## Development

```bash
# Build CLI
npm run build

# Test locally
npm link
agentos --help
```