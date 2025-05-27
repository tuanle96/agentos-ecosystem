# ADR-006: Python AI Workers Strategy

**Status**: Accepted  
**Date**: 2024-12-27  
**Deciders**: AgentOS Architecture Team  

## Context

While Go provides excellent performance for core backend services, AgentOS requires access to the rich Python AI/ML ecosystem including LangChain, CrewAI, Swarms, and AutoGen. These frameworks are Python-specific and cannot be easily ported to Go.

## Decision

We will implement **Python AI Workers** as specialized FastAPI services that handle AI-specific operations, communicating with Go backend services via HTTP APIs and message queues.

## Architecture

### AI Worker Services
- **LangChain Worker**: Tool chains and document processing
- **CrewAI Worker**: Multi-agent collaboration workflows
- **Swarms Worker**: Swarm intelligence coordination
- **AutoGen Worker**: Conversational AI patterns
- **Embedding Worker**: Vector embeddings and similarity
- **Model Worker**: Custom model inference and fine-tuning

### Communication Pattern
```
Go Backend Services ←→ HTTP APIs ←→ Python AI Workers
                   ←→ NATS Queues ←→ (async operations)
                   ←→ WebSockets ←→ (real-time streaming)
```

## Rationale

### AI Ecosystem Access
- **LangChain**: Comprehensive tool ecosystem and document processing
- **CrewAI**: Advanced multi-agent collaboration patterns
- **Swarms**: Cutting-edge swarm intelligence algorithms
- **AutoGen**: Microsoft's conversational AI framework
- **Rich Libraries**: NumPy, Pandas, Scikit-learn, Transformers, etc.

### Hybrid Architecture Benefits
1. **Best of Both Worlds**: Go performance + Python AI capabilities
2. **Specialization**: Each language optimized for its strengths
3. **Independent Scaling**: Scale Go and Python services separately
4. **Technology Evolution**: Easy to adopt new AI frameworks
5. **Risk Mitigation**: Isolated failures don't affect core services

### Performance Characteristics
```yaml
Go Services (High Frequency):
  - API routing and orchestration
  - Database operations
  - Authentication and authorization
  - Real-time communication
  - Performance: <15ms response time

Python Workers (AI Intensive):
  - AI model inference
  - Complex data processing
  - Framework-specific operations
  - Performance: <2s for most operations
```

## Implementation Strategy

### Worker Architecture
```python
# Standard AI Worker Template
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import asyncio

class AIWorker:
    def __init__(self, framework_name: str, port: int):
        self.app = FastAPI(title=f"{framework_name} AI Worker")
        self.framework = framework_name
        self.setup_routes()
    
    def setup_routes(self):
        @self.app.get("/health")
        async def health_check():
            return {"status": "healthy", "framework": self.framework}
        
        @self.app.post("/create-agent")
        async def create_agent(config: dict):
            return await self.create_agent_impl(config)
        
        @self.app.post("/execute")
        async def execute(request: dict):
            return await self.execute_impl(request)
```

### Go Integration
```go
// AI Worker Client in Go
type AIWorkerClient struct {
    baseURL    string
    httpClient *http.Client
    framework  string
}

func (c *AIWorkerClient) CreateAgent(config map[string]interface{}) (*AgentResponse, error) {
    payload, _ := json.Marshal(config)
    resp, err := c.httpClient.Post(
        c.baseURL+"/create-agent",
        "application/json",
        bytes.NewBuffer(payload),
    )
    // Handle response...
}
```

## Communication Patterns

### 1. Synchronous HTTP (Immediate Responses)
- Agent creation and configuration
- Quick AI operations (<5 seconds)
- Status queries and health checks
- Real-time user interactions

### 2. Asynchronous NATS (Long-running Tasks)
- Complex AI model training
- Large document processing
- Batch operations
- Background tasks

### 3. WebSocket Streaming (Real-time Updates)
- Agent execution progress
- Real-time AI responses
- Live collaboration sessions
- Streaming model outputs

## Deployment Strategy

### Development Environment
```yaml
# docker-compose.yml
services:
  # Go Services
  core-api:
    build: ./services/core-api
    ports: ["8000:8000"]
  
  # Python AI Workers
  langchain-worker:
    build: ./ai-workers/langchain-worker
    ports: ["8080:8080"]
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
  
  crewai-worker:
    build: ./ai-workers/crewai-worker
    ports: ["8081:8081"]
```

### Production Kubernetes
```yaml
# Go services: Small resource requirements
apiVersion: apps/v1
kind: Deployment
metadata:
  name: core-api
spec:
  template:
    spec:
      containers:
      - name: core-api
        resources:
          requests:
            memory: "32Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "200m"

# Python workers: Higher resource requirements
apiVersion: apps/v1
kind: Deployment
metadata:
  name: langchain-worker
spec:
  template:
    spec:
      containers:
      - name: langchain-worker
        resources:
          requests:
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
```

## Consequences

### Positive
- **AI Ecosystem Access**: Full access to Python AI/ML libraries
- **Performance Optimization**: Go handles high-frequency operations
- **Independent Scaling**: Scale AI workers based on AI workload
- **Technology Flexibility**: Easy to adopt new AI frameworks
- **Fault Isolation**: AI worker failures don't affect core services

### Negative
- **Complexity**: Additional operational complexity
- **Network Latency**: HTTP calls between Go and Python
- **Resource Usage**: Python workers require more memory
- **Deployment**: More services to deploy and monitor

### Mitigation Strategies
- **Connection Pooling**: Efficient HTTP client management
- **Circuit Breakers**: Graceful handling of AI worker failures
- **Caching**: Cache AI results to reduce worker load
- **Monitoring**: Comprehensive monitoring of all services
- **Auto-scaling**: Automatic scaling based on workload

## Performance Targets

### AI Worker Performance
- **Agent Creation**: <2 seconds for standard agents
- **Tool Execution**: <5 seconds for most operations
- **Model Inference**: <1 second for standard models
- **Concurrent Operations**: 100+ per worker instance

### Communication Performance
- **HTTP Latency**: <10ms between Go and Python services
- **Throughput**: 10,000+ HTTP calls per minute
- **Message Queue**: <100ms for async operations
- **WebSocket**: Real-time streaming with <50ms latency

## Alternatives Considered

### Go-only Architecture
- **Pros**: Single language, simpler deployment
- **Cons**: Limited AI ecosystem, significant development effort
- **Verdict**: Would require reimplementing Python AI frameworks

### Python-only Architecture
- **Pros**: Single language, full AI ecosystem access
- **Cons**: Performance limitations, higher resource usage
- **Verdict**: Cannot meet performance requirements

### Microservices with gRPC
- **Pros**: Better performance than HTTP
- **Cons**: More complex, additional protocol overhead
- **Verdict**: HTTP is simpler and sufficient for current needs

## Success Metrics

### Integration Metrics
- **API Response Time**: <10ms for Go ↔ Python communication
- **AI Operation Success Rate**: >99%
- **Worker Availability**: >99.9%
- **Error Rate**: <0.1%

### Business Metrics
- **AI Feature Adoption**: Track usage of AI capabilities
- **User Satisfaction**: Measure AI response quality
- **Development Velocity**: Time to implement new AI features
- **Operational Efficiency**: Monitor resource utilization

## Related ADRs
- [ADR-001: Go Backend Services](001-go-backend-services.md)
- [ADR-007: Hybrid Communication Patterns](007-hybrid-communication.md)
- [ADR-008: Technology Migration Strategy](008-technology-migration.md)

## References
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [LangChain Documentation](https://python.langchain.com/)
- [CrewAI Documentation](https://docs.crewai.com/)
- [Swarms Documentation](https://docs.swarms.world/)
- [AutoGen Documentation](https://microsoft.github.io/autogen/)

**Decision**: Python AI Workers provide the optimal balance of AI capability and system performance in our hybrid architecture.
