# Helm Charts

> Kubernetes application packaging and deployment

## Overview

Helm charts for deploying AgentOS ecosystem components to Kubernetes clusters.

## Directory Structure

```
helm-charts/
├── agentos-ecosystem/    # Main ecosystem chart
│   ├── Chart.yaml       # Chart metadata
│   ├── values.yaml      # Default values
│   ├── templates/       # Kubernetes templates
│   └── charts/         # Sub-charts
├── core-api/           # Core API service chart
├── agent-engine/       # Agent engine chart
├── memory-service/     # Memory service chart
├── tool-registry/      # Tool registry chart
└── README.md          # This file
```

## Installation

```bash
# Add Helm repository
helm repo add agentos https://charts.agentos.ai

# Install AgentOS ecosystem
helm install agentos agentos/agentos-ecosystem

# Install with custom values
helm install agentos agentos/agentos-ecosystem -f values-prod.yaml

# Upgrade deployment
helm upgrade agentos agentos/agentos-ecosystem
```

## Configuration

Key configuration options in `values.yaml`:

```yaml
# Global settings
global:
  imageRegistry: "registry.agentos.ai"
  storageClass: "gp2"
  
# Core API configuration
coreApi:
  replicaCount: 3
  image:
    tag: "1.0.0"
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi

# Database configuration
postgresql:
  enabled: true
  auth:
    database: "agentos"
    username: "agentos"
```

## Charts

- **agentos-ecosystem**: Main umbrella chart
- **core-api**: Core API service
- **agent-engine**: Agent execution engine
- **memory-service**: Memory management
- **tool-registry**: Tool registry
- **postgresql**: Database
- **redis**: Cache and queues
- **monitoring**: Observability stack