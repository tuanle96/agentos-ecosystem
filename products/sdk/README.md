# AgentOS SDK

> Multi-language SDKs for AgentOS integration

## Overview

AgentOS SDK provides official software development kits for multiple programming languages, enabling easy integration with AgentOS services.

## Supported Languages

- **Go**: Native Go SDK with full feature support
- **Python**: Comprehensive Python SDK
- **JavaScript/TypeScript**: Browser and Node.js support
- **Rust**: High-performance Rust SDK
- **Java**: Enterprise Java SDK (planned)
- **C#**: .NET SDK (planned)

## License

**MIT License** - Open source and free to use

## Directory Structure

```
products/sdk/
├── go-sdk/           # Go SDK
├── python-sdk/       # Python SDK
├── javascript-sdk/   # JavaScript/TypeScript SDK
├── rust-sdk/         # Rust SDK
├── docs/            # SDK documentation
└── README.md        # This file
```

## Quick Start

### Go SDK

```go
package main

import (
    "github.com/agentos/go-sdk"
)

func main() {
    client := agentos.NewClient("your-api-key")
    
    agent, err := client.Agents.Create(&agentos.AgentRequest{
        Name: "Web Researcher",
        Capabilities: []string{"web-search", "data-analysis"},
    })
    
    execution, err := client.Executions.Create(&agentos.ExecutionRequest{
        AgentID: agent.ID,
        Input: "Research AI trends in 2024",
    })
}
```

### Python SDK

```python
from agentos import AgentOSClient

client = AgentOSClient(api_key="your-api-key")

# Create agent
agent = client.agents.create(
    name="Web Researcher",
    capabilities=["web-search", "data-analysis"]
)

# Execute agent
execution = client.executions.create(
    agent_id=agent.id,
    input="Research AI trends in 2024"
)
```

### JavaScript SDK

```javascript
import { AgentOSClient } from '@agentos/sdk';

const client = new AgentOSClient({
  apiKey: 'your-api-key'
});

// Create agent
const agent = await client.agents.create({
  name: 'Web Researcher',
  capabilities: ['web-search', 'data-analysis']
});

// Execute agent
const execution = await client.executions.create({
  agentId: agent.id,
  input: 'Research AI trends in 2024'
});
```

### Rust SDK

```rust
use agentos_sdk::{AgentOSClient, AgentRequest, ExecutionRequest};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = AgentOSClient::new("your-api-key");
    
    let agent = client.agents().create(AgentRequest {
        name: "Web Researcher".to_string(),
        capabilities: vec!["web-search".to_string(), "data-analysis".to_string()],
    }).await?;
    
    let execution = client.executions().create(ExecutionRequest {
        agent_id: agent.id,
        input: "Research AI trends in 2024".to_string(),
    }).await?;
    
    Ok(())
}
```

## Features

### Core Features
- **Agent Management**: Create, list, update, delete agents
- **Tool Integration**: Manage and execute tools
- **Execution Control**: Start, monitor, and control executions
- **Memory Operations**: Access and manage agent memory
- **Authentication**: API key and OAuth2 support

### Advanced Features
- **Streaming**: Real-time execution streaming
- **Webhooks**: Event-driven integrations
- **Batch Operations**: Bulk operations for efficiency
- **Error Handling**: Comprehensive error handling
- **Retry Logic**: Automatic retry with exponential backoff

## Installation

### Go
```bash
go get github.com/agentos/go-sdk
```

### Python
```bash
pip install agentos-sdk
```

### JavaScript
```bash
npm install @agentos/sdk
```

### Rust
```toml
[dependencies]
agentos-sdk = "0.1.0"
```

## Documentation

- **API Reference**: [docs.agentos.ai/sdk](https://docs.agentos.ai/sdk)
- **Examples**: [github.com/agentos/examples](https://github.com/agentos/examples)
- **Tutorials**: [tutorials.agentos.ai](https://tutorials.agentos.ai)

## Support

- **SDK Issues**: [GitHub Issues](https://github.com/tuanle96/agentos-ecosystem/issues)
- **Developer Support**: developers@agentos.ai
- **Community**: [Discord](https://discord.gg/agentos)