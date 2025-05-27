# AgentOS Core Frontend

> React web application for AgentOS hybrid architecture (Go + Python)

## Overview

The frontend application provides a modern web interface for managing AI agents, tools, and executions. Built to work with the high-performance Go backend services and Python AI workers, delivering superior user experience with real-time updates and fast response times.

## Features

### **Core Features**
- **Agent Dashboard**: Visual agent management with Go backend performance
- **Tool Library**: Browse and configure tools from Python AI workers
- **Execution Monitor**: Real-time execution tracking via WebSocket
- **Memory Viewer**: Inspect agent memory with vector database integration
- **Settings**: Configuration management for hybrid architecture

### **Performance Features**
- **Real-time Updates**: WebSocket connections to Go services (<50ms latency)
- **Fast API Responses**: Go backend delivers <15ms response times
- **AI Framework Support**: Interface for LangChain, CrewAI, Swarms, AutoGen workers
- **Concurrent Operations**: Handle 10,000+ concurrent users
- **Responsive UI**: Optimized for high-performance backend

## Tech Stack

### **Frontend Technologies**
- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite (fast development and build)
- **UI Library**: Material-UI / Tailwind CSS
- **State Management**: Redux Toolkit
- **API Client**: @agentos/api-client (optimized for Go backend)
- **Testing**: Jest + React Testing Library

### **Backend Integration**
- **Go Backend**: High-performance API services (Gin framework)
- **Python Workers**: AI framework integration via HTTP APIs
- **WebSocket**: Real-time communication with Go services
- **Authentication**: JWT tokens from Go auth service
- **Performance**: Optimized for 10,000+ concurrent users

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

### **Go Backend Services**
```bash
# Core API (Go service)
VITE_API_URL=http://localhost:8000
VITE_WS_URL=ws://localhost:8000/ws

# Additional Go services
VITE_AGENT_ENGINE_URL=http://localhost:8001
VITE_MEMORY_SERVICE_URL=http://localhost:8002
VITE_TOOL_REGISTRY_URL=http://localhost:8003
VITE_AUTH_SERVICE_URL=http://localhost:8004
```

### **Python AI Workers**
```bash
# AI Worker endpoints
VITE_LANGCHAIN_WORKER_URL=http://localhost:8080
VITE_CREWAI_WORKER_URL=http://localhost:8081
VITE_SWARMS_WORKER_URL=http://localhost:8082
VITE_AUTOGEN_WORKER_URL=http://localhost:8083
```

### **Application Configuration**
```bash
VITE_APP_NAME=AgentOS Core
VITE_APP_VERSION=1.0.0
VITE_ENABLE_REAL_TIME=true
VITE_MAX_CONCURRENT_REQUESTS=1000
```

## API Integration

### **Go Backend API Client**
```typescript
// API client optimized for Go backend
import { AgentOSClient } from '@agentos/api-client';

const client = new AgentOSClient({
  baseURL: process.env.VITE_API_URL,
  timeout: 5000, // Fast timeout for Go services
  retries: 3,
  headers: {
    'Content-Type': 'application/json',
  }
});

// Fast agent creation with Go backend
const createAgent = async (config: AgentConfig) => {
  const response = await client.post('/agents', config);
  return response.data; // <15ms response time
};
```

### **WebSocket Integration**
```typescript
// Real-time WebSocket connection to Go services
import { useWebSocket } from '@agentos/websocket-client';

const AgentExecutionMonitor = () => {
  const { socket, isConnected } = useWebSocket(
    process.env.VITE_WS_URL + '/execution'
  );

  useEffect(() => {
    if (isConnected) {
      socket.on('execution_progress', (data) => {
        // Real-time updates from Go services (<50ms latency)
        updateExecutionProgress(data);
      });

      socket.on('ai_worker_response', (data) => {
        // Responses from Python AI workers
        updateAIWorkerStatus(data);
      });
    }
  }, [isConnected]);

  return (
    <div>
      <ExecutionProgress />
      <AIWorkerStatus />
    </div>
  );
};
```

### **AI Framework Integration**
```typescript
// Interface for Python AI workers
import { AIWorkerClient } from '@agentos/ai-worker-client';

const aiWorkers = {
  langchain: new AIWorkerClient(process.env.VITE_LANGCHAIN_WORKER_URL),
  crewai: new AIWorkerClient(process.env.VITE_CREWAI_WORKER_URL),
  swarms: new AIWorkerClient(process.env.VITE_SWARMS_WORKER_URL),
  autogen: new AIWorkerClient(process.env.VITE_AUTOGEN_WORKER_URL),
};

// Create agent with specific AI framework
const createAIAgent = async (framework: string, config: any) => {
  // Go backend orchestrates the request
  const agent = await client.post('/agents', {
    framework,
    config,
  });

  // Monitor AI worker progress
  const worker = aiWorkers[framework];
  const status = await worker.getStatus(agent.ai_agent_id);

  return { agent, status };
};
```

## Performance Optimization

### **Go Backend Benefits**
- **Fast API Responses**: <15ms response time from Go services
- **High Concurrency**: Support for 10,000+ concurrent users
- **Efficient Memory**: 5x less memory usage than Python backend
- **Real-time Updates**: WebSocket connections managed by Go services

### **Frontend Optimizations**
```typescript
// Optimized for high-performance Go backend
const optimizations = {
  // Fast API calls with minimal timeout
  apiTimeout: 5000, // Go services respond quickly

  // Efficient state management
  batchUpdates: true, // Batch multiple updates

  // Real-time optimizations
  websocketReconnect: true, // Auto-reconnect to Go services

  // Concurrent request handling
  maxConcurrentRequests: 100, // Go backend can handle high load
};
```

### **Monitoring Integration**
```typescript
// Performance monitoring for hybrid architecture
import { PerformanceMonitor } from '@agentos/monitoring';

const monitor = new PerformanceMonitor({
  // Track Go service performance
  trackAPILatency: true,
  trackWebSocketLatency: true,

  // Track AI worker performance
  trackAIWorkerLatency: true,
  trackFrameworkPerformance: true,

  // Performance targets
  apiLatencyTarget: 15, // ms
  websocketLatencyTarget: 50, // ms
  aiWorkerLatencyTarget: 2000, // ms
});
```