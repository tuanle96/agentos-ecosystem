# ADR-007: Hybrid Communication Patterns

**Status**: Accepted  
**Date**: 2024-12-27  
**Deciders**: AgentOS Architecture Team  

## Context

The AgentOS hybrid architecture requires efficient communication between Go backend services and Python AI workers. Different types of operations require different communication patterns to optimize for performance, reliability, and user experience.

## Decision

We will implement **multiple communication patterns** optimized for different use cases:
1. **Synchronous HTTP** for immediate responses
2. **Asynchronous NATS** for long-running tasks
3. **WebSocket streaming** for real-time updates

## Communication Patterns

### 1. Synchronous HTTP (Immediate Responses)

**Use Cases**:
- Agent creation and configuration
- Quick AI operations (<5 seconds)
- Status queries and health checks
- Real-time user interactions

**Implementation**:
```go
// Go Service → Python Worker
type AIWorkerClient struct {
    baseURL    string
    httpClient *http.Client
    timeout    time.Duration
}

func (c *AIWorkerClient) CreateAgent(ctx context.Context, config AgentConfig) (*Agent, error) {
    payload, _ := json.Marshal(config)
    
    req, _ := http.NewRequestWithContext(ctx, "POST", 
        c.baseURL+"/create-agent", 
        bytes.NewBuffer(payload))
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close()
    
    var agent Agent
    if err := json.NewDecoder(resp.Body).Decode(&agent); err != nil {
        return nil, fmt.Errorf("response decode failed: %w", err)
    }
    
    return &agent, nil
}
```

**Performance Targets**:
- Response Time: <10ms for Go ↔ Python communication
- Timeout: 30 seconds for most operations
- Concurrent Requests: 1,000+ per worker
- Success Rate: >99.9%

### 2. Asynchronous NATS (Long-running Tasks)

**Use Cases**:
- Complex AI model training
- Large document processing
- Batch operations
- Background tasks

**Implementation**:
```go
// Go Service publishes task
type TaskMessage struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    Payload   map[string]interface{} `json:"payload"`
    Callback  string                 `json:"callback"`
    Timeout   time.Duration          `json:"timeout"`
}

func (s *AgentService) ExecuteAgentAsync(agentID string, input string) (*ExecutionID, error) {
    taskID := uuid.New().String()
    
    msg := &TaskMessage{
        ID:       taskID,
        Type:     "agent_execution",
        Payload: map[string]interface{}{
            "agent_id": agentID,
            "input":    input,
        },
        Callback: fmt.Sprintf("http://agent-engine/execution-complete/%s", taskID),
        Timeout:  5 * time.Minute,
    }
    
    data, _ := json.Marshal(msg)
    if err := s.nats.Publish("agent.execute", data); err != nil {
        return nil, fmt.Errorf("failed to publish task: %w", err)
    }
    
    return &ExecutionID{ID: taskID}, nil
}
```

```python
# Python Worker subscribes to tasks
import asyncio
import nats
import json

class LangChainWorker:
    async def setup_nats_subscriber(self):
        self.nc = await nats.connect("nats://localhost:4222")
        
        async def message_handler(msg):
            try:
                task = json.loads(msg.data.decode())
                result = await self.process_task(task)
                
                # Send result back via callback
                await self.send_callback(task['callback'], result)
                
            except Exception as e:
                await self.send_error_callback(task['callback'], str(e))
        
        await self.nc.subscribe("agent.execute", cb=message_handler)
    
    async def process_task(self, task):
        # Process the AI task
        agent_id = task['payload']['agent_id']
        input_text = task['payload']['input']
        
        agent = self.agents.get(agent_id)
        result = await agent.arun(input_text)
        
        return {
            "task_id": task['id'],
            "status": "completed",
            "result": result,
            "metadata": {"execution_time": "2.5s"}
        }
```

**Performance Targets**:
- Message Latency: <100ms
- Throughput: 10,000+ messages per minute
- Reliability: At-least-once delivery
- Retention: 24 hours for unprocessed messages

### 3. WebSocket Streaming (Real-time Updates)

**Use Cases**:
- Agent execution progress
- Real-time AI responses
- Live collaboration sessions
- Streaming model outputs

**Implementation**:
```go
// Go Service manages WebSocket connections
func (s *AgentService) StreamExecution(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Printf("WebSocket upgrade failed: %v", err)
        return
    }
    defer conn.Close()
    
    executionID := r.URL.Query().Get("execution_id")
    
    // Subscribe to execution updates
    updates := s.subscribeToExecutionUpdates(executionID)
    
    for update := range updates {
        if err := conn.WriteJSON(update); err != nil {
            log.Printf("WebSocket write failed: %v", err)
            break
        }
    }
}
```

```python
# Python Worker sends streaming updates
import websockets
import json

class LangChainWorker:
    async def execute_with_streaming(self, agent_id: str, input_text: str, websocket_url: str):
        # Connect to WebSocket for streaming updates
        async with websockets.connect(websocket_url) as websocket:
            
            # Send initial status
            await websocket.send(json.dumps({
                "status": "started",
                "progress": 0,
                "message": "Initializing agent execution"
            }))
            
            # Execute agent with progress updates
            agent = self.agents[agent_id]
            
            async for chunk in agent.astream(input_text):
                await websocket.send(json.dumps({
                    "status": "processing",
                    "progress": chunk.progress,
                    "partial_result": chunk.content,
                    "message": f"Processing step {chunk.step}"
                }))
            
            # Send final result
            await websocket.send(json.dumps({
                "status": "completed",
                "progress": 100,
                "result": final_result,
                "message": "Execution completed"
            }))
```

**Performance Targets**:
- Latency: <50ms for real-time updates
- Concurrent Connections: 1,000+ per Go service
- Message Rate: 100+ messages per second per connection
- Connection Stability: >99% uptime

## Error Handling & Resilience

### Circuit Breaker Pattern
```go
type CircuitBreaker struct {
    maxFailures int
    timeout     time.Duration
    failures    int
    lastFailure time.Time
    state       string // "closed", "open", "half-open"
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == "open" {
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = "half-open"
        } else {
            return errors.New("circuit breaker is open")
        }
    }
    
    err := fn()
    if err != nil {
        cb.failures++
        cb.lastFailure = time.Now()
        
        if cb.failures >= cb.maxFailures {
            cb.state = "open"
        }
        return err
    }
    
    cb.failures = 0
    cb.state = "closed"
    return nil
}
```

### Retry Logic
```go
func (c *AIWorkerClient) CallWithRetry(ctx context.Context, fn func() error) error {
    backoff := time.Second
    maxRetries := 3
    
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        if i == maxRetries-1 {
            return fmt.Errorf("max retries exceeded: %w", err)
        }
        
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(backoff):
            backoff *= 2 // Exponential backoff
        }
    }
    
    return nil
}
```

### Graceful Degradation
```go
func (s *AgentService) CreateAgentWithFallback(ctx context.Context, config AgentConfig) (*Agent, error) {
    // Try primary AI worker
    if agent, err := s.primaryWorker.CreateAgent(ctx, config); err == nil {
        return agent, nil
    }
    
    // Fallback to secondary worker
    if agent, err := s.secondaryWorker.CreateAgent(ctx, config); err == nil {
        return agent, nil
    }
    
    // Fallback to basic agent without AI capabilities
    return s.createBasicAgent(config), nil
}
```

## Monitoring & Observability

### Metrics Collection
```go
// Prometheus metrics for communication patterns
var (
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "agentos_http_request_duration_seconds",
            Help: "HTTP request duration for Go ↔ Python communication",
        },
        []string{"service", "endpoint", "status"},
    )
    
    natsMessageCount = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "agentos_nats_messages_total",
            Help: "Total number of NATS messages",
        },
        []string{"subject", "status"},
    )
    
    websocketConnections = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "agentos_websocket_connections",
            Help: "Current number of WebSocket connections",
        },
        []string{"service"},
    )
)
```

### Distributed Tracing
```go
import "go.opentelemetry.io/otel/trace"

func (c *AIWorkerClient) CreateAgentWithTracing(ctx context.Context, config AgentConfig) (*Agent, error) {
    tracer := otel.Tracer("agentos/ai-worker-client")
    ctx, span := tracer.Start(ctx, "ai_worker.create_agent")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("ai.framework", c.framework),
        attribute.String("agent.type", config.Type),
    )
    
    agent, err := c.CreateAgent(ctx, config)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }
    
    span.SetAttributes(attribute.String("agent.id", agent.ID))
    return agent, nil
}
```

## Consequences

### Positive
- **Optimized Performance**: Each pattern optimized for specific use cases
- **Scalability**: Independent scaling of communication channels
- **Reliability**: Multiple fallback mechanisms and error handling
- **Real-time Capability**: WebSocket streaming for immediate feedback
- **Flexibility**: Easy to add new communication patterns

### Negative
- **Complexity**: Multiple protocols to manage and monitor
- **Network Dependencies**: Increased network traffic between services
- **Debugging Complexity**: Distributed tracing required
- **Operational Overhead**: More monitoring and alerting needed

### Mitigation Strategies
- **Comprehensive Monitoring**: Detailed metrics and tracing
- **Standardized Patterns**: Consistent implementation across services
- **Documentation**: Clear guidelines for when to use each pattern
- **Testing**: Extensive integration testing of communication patterns

## Success Metrics

### Performance Metrics
- **HTTP Response Time**: <10ms for Go ↔ Python calls
- **NATS Message Latency**: <100ms end-to-end
- **WebSocket Latency**: <50ms for real-time updates
- **Overall System Throughput**: 100,000+ operations per minute

### Reliability Metrics
- **HTTP Success Rate**: >99.9%
- **NATS Message Delivery**: >99.99%
- **WebSocket Connection Stability**: >99%
- **Circuit Breaker Effectiveness**: <1% false positives

## Related ADRs
- [ADR-001: Go Backend Services](001-go-backend-services.md)
- [ADR-006: Python AI Workers](006-python-ai-workers.md)
- [ADR-008: Technology Migration Strategy](008-technology-migration.md)

## References
- [NATS Documentation](https://docs.nats.io/)
- [WebSocket RFC 6455](https://tools.ietf.org/html/rfc6455)
- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)

**Decision**: Multiple communication patterns provide the optimal balance of performance, reliability, and real-time capability for our hybrid architecture.
