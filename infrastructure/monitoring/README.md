# Monitoring Configurations

> Observability and monitoring setup

## Overview

Monitoring configurations for comprehensive observability of the AgentOS ecosystem.

## Directory Structure

```
monitoring/
├── prometheus/      # Prometheus configurations
│   ├── prometheus.yml # Main configuration
│   ├── rules/        # Alerting rules
│   └── targets/      # Service discovery
├── grafana/         # Grafana dashboards
│   ├── dashboards/  # Dashboard definitions
│   └── datasources/ # Data source configs
├── jaeger/          # Distributed tracing
├── elasticsearch/   # Log aggregation
└── README.md       # This file
```

## Components

### Metrics (Prometheus + Grafana)
- **Application Metrics**: Custom business metrics
- **Infrastructure Metrics**: CPU, memory, disk, network
- **Database Metrics**: PostgreSQL and Redis metrics
- **Kubernetes Metrics**: Cluster and pod metrics

### Logging (ELK Stack)
- **Elasticsearch**: Log storage and indexing
- **Logstash**: Log processing and transformation
- **Kibana**: Log visualization and analysis
- **Filebeat**: Log shipping from containers

### Tracing (Jaeger)
- **Distributed Tracing**: Request flow across services
- **Performance Analysis**: Latency and bottleneck identification
- **Error Tracking**: Error propagation analysis

### Alerting
- **Prometheus AlertManager**: Alert routing and grouping
- **PagerDuty**: Incident management
- **Slack**: Team notifications
- **Email**: Critical alerts

## Dashboards

### Application Dashboards
- **AgentOS Overview**: High-level system metrics
- **API Performance**: Request rates, latency, errors
- **Agent Execution**: Agent performance and success rates
- **Memory Usage**: Memory system performance
- **Tool Registry**: Tool usage and performance

### Infrastructure Dashboards
- **Kubernetes Cluster**: Cluster health and resource usage
- **Node Metrics**: Individual node performance
- **Database Performance**: PostgreSQL and Redis metrics
- **Network Traffic**: Network usage and latency

## Alerts

### Critical Alerts
- Service down (> 1 minute)
- High error rate (> 5%)
- Database connection failures
- Disk space low (< 10%)
- Memory usage high (> 90%)

### Warning Alerts
- High response time (> 1 second)
- Increased error rate (> 1%)
- High CPU usage (> 80%)
- Queue backlog growing

## Setup

```bash
# Deploy monitoring stack
kubectl apply -f monitoring/

# Access Grafana
kubectl port-forward svc/grafana 3000:3000

# Access Prometheus
kubectl port-forward svc/prometheus 9090:9090

# Access Jaeger
kubectl port-forward svc/jaeger 16686:16686
```