# Kubernetes Manifests

> Raw Kubernetes YAML manifests

## Overview

Kubernetes manifests for deploying AgentOS components without Helm.

## Directory Structure

```
kubernetes/
├── namespaces/      # Namespace definitions
├── services/        # Service manifests
│   ├── core-api/   # Core API service
│   ├── agent-engine/ # Agent engine
│   └── memory-service/ # Memory service
├── databases/       # Database deployments
├── monitoring/      # Monitoring stack
├── ingress/        # Ingress configurations
└── README.md       # This file
```

## Deployment

```bash
# Create namespace
kubectl apply -f namespaces/

# Deploy services
kubectl apply -f services/

# Deploy databases
kubectl apply -f databases/

# Deploy monitoring
kubectl apply -f monitoring/

# Configure ingress
kubectl apply -f ingress/
```

## Components

### Services
- Core API (3 replicas)
- Agent Engine (2 replicas)
- Memory Service (2 replicas)
- Tool Registry (2 replicas)

### Databases
- PostgreSQL (primary + replica)
- Redis (cluster mode)
- Elasticsearch (3 nodes)

### Monitoring
- Prometheus
- Grafana
- Jaeger
- AlertManager

## Configuration

Environment-specific configurations using ConfigMaps and Secrets:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: agentos-config
data:
  DATABASE_URL: "postgres://..."
  REDIS_URL: "redis://..."
  LOG_LEVEL: "info"
```