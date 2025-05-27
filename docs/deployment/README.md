# Deployment Documentation

> Comprehensive deployment guides for AgentOS hybrid ecosystem (Go + Python)

## Overview

This section contains detailed deployment guides for the AgentOS hybrid architecture, covering Go backend services and Python AI workers across various environments, from local development to production deployments.

## Deployment Options

### Local Development
- [Local Development Setup](local-development.md) - Development environment setup
- [Docker Compose](docker-compose.md) - Containerized local development
- [Hot Reload Setup](hot-reload.md) - Development with hot reload

### Container Deployments
- [Docker Deployment](docker.md) - Single-node Docker deployment
- [Docker Swarm](docker-swarm.md) - Multi-node Docker Swarm
- [Kubernetes](kubernetes.md) - Kubernetes cluster deployment

### Cloud Platforms
- [AWS Deployment](aws.md) - Amazon Web Services deployment
- [GCP Deployment](gcp.md) - Google Cloud Platform deployment
- [Azure Deployment](azure.md) - Microsoft Azure deployment
- [DigitalOcean](digitalocean.md) - DigitalOcean deployment

### Specialized Deployments
- [Production Setup](production.md) - Production-ready deployment
- [High Availability](high-availability.md) - HA configuration
- [Disaster Recovery](disaster-recovery.md) - Backup and recovery
- [Multi-Region](multi-region.md) - Global deployment

## Quick Start

### Prerequisites

#### **Hybrid Architecture Requirements**
- **Go**: 1.21 or higher (for backend services)
- **Python**: 3.11 or higher (for AI workers)
- **Node.js**: 18.0 or higher (for frontend)
- **Docker**: 24.0 or higher (for containerization)
- **Kubernetes**: 1.28 or higher (for K8s deployments)

#### **Infrastructure Services**
- **PostgreSQL**: 15+ with pgvector extension
- **Redis**: 7+ for caching and sessions
- **NATS**: For Go ↔ Python communication
- **Prometheus**: For monitoring (optional)

### Local Development (5 minutes)
```bash
# Clone repository
git clone https://github.com/tuanle96/agentos-ecosystem.git
cd agentos-ecosystem

# Setup hybrid environment
make setup
cp .env.example .env
# Edit .env with your configuration (API keys, database URLs)

# Start infrastructure services (PostgreSQL, Redis, NATS)
make dev-services

# Run database migrations
make migrate-up

# Start Go backend services
make dev-go

# Start Python AI workers (in separate terminal)
make dev-python

# Start frontend (in separate terminal)
cd products/core/frontend && npm run dev
```

### Docker Deployment (10 minutes)
```bash
# Clone repository
git clone https://github.com/tuanle96/agentos-ecosystem.git
cd agentos-ecosystem

# Configure environment
cp .env.example .env
# Edit .env with production values

# Start with Docker Compose
docker-compose -f docker-compose.prod.yml up -d

# Check health
curl http://localhost:8000/health
```

### Kubernetes Deployment (15 minutes)
```bash
# Add Helm repository
helm repo add agentos https://charts.agentos.ai
helm repo update

# Install AgentOS
helm install agentos agentos/agentos-ecosystem \
  --set global.domain=your-domain.com \
  --set postgresql.auth.password=your-db-password

# Check status
kubectl get pods -l app.kubernetes.io/name=agentos-ecosystem
```

## Environment Configuration

### Environment Variables

#### **Go Backend Services Configuration**
```bash
# Core Configuration
GO_ENV=production
NODE_ENV=production
LOG_LEVEL=info

# Service Ports
CORE_API_PORT=8000
AGENT_ENGINE_PORT=8001
MEMORY_SERVICE_PORT=8002
TOOL_REGISTRY_PORT=8003
AUTH_SERVICE_PORT=8004

# Database & Infrastructure
DATABASE_URL=postgres://user:pass@host:5432/agentos?sslmode=require
REDIS_URL=redis://host:6379/0
NATS_URL=nats://host:4222

# Security
JWT_SECRET=your-jwt-secret-key
API_RATE_LIMIT=10000
CORS_ORIGINS=https://your-domain.com
```

#### **Python AI Workers Configuration**
```bash
# AI Worker URLs
LANGCHAIN_WORKER_URL=http://langchain-worker:8080
CREWAI_WORKER_URL=http://crewai-worker:8081
SWARMS_WORKER_URL=http://swarms-worker:8082
AUTOGEN_WORKER_URL=http://autogen-worker:8083

# AI Services
OPENAI_API_KEY=your-openai-key
ANTHROPIC_API_KEY=your-anthropic-key
PINECONE_API_KEY=your-pinecone-key
PINECONE_ENVIRONMENT=your-pinecone-env

# Python Worker Ports
LANGCHAIN_WORKER_PORT=8080
CREWAI_WORKER_PORT=8081
SWARMS_WORKER_PORT=8082
AUTOGEN_WORKER_PORT=8083
```

#### **Monitoring & Storage**
```bash
# Monitoring
PROMETHEUS_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
GRAFANA_ADMIN_PASSWORD=your-grafana-password

# Storage
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=your-access-key
MINIO_SECRET_KEY=your-secret-key
```

### Configuration Files
- [Environment Templates](configs/env-templates/) - Environment file templates
- [Docker Configs](configs/docker/) - Docker-specific configurations
- [Kubernetes Configs](configs/kubernetes/) - K8s configuration files
- [Production Configs](configs/production/) - Production settings

## Infrastructure Requirements

### Hybrid Architecture Resource Planning

#### **Minimum Requirements (Development)**
```yaml
Go Backend Services:
  CPU: 1 core per service (4 cores total)
  Memory: 512MB per service (2GB total)

Python AI Workers:
  CPU: 1 core per worker (4 cores total)
  Memory: 1GB per worker (4GB total)

Infrastructure:
  PostgreSQL: 1 core, 1GB RAM
  Redis: 0.5 core, 512MB RAM
  NATS: 0.5 core, 256MB RAM

Total Development:
  CPU: 10 cores
  Memory: 8 GB RAM
  Storage: 20 GB SSD
  Network: 10 Mbps
```

#### **Recommended Requirements (Production)**
```yaml
Go Backend Services (High Performance):
  CPU: 2 cores per service (8 cores total)
  Memory: 1GB per service (4GB total)

Python AI Workers (AI Intensive):
  CPU: 2 cores per worker (8 cores total)
  Memory: 2GB per worker (8GB total)

Infrastructure:
  PostgreSQL: 4 cores, 8GB RAM
  Redis: 2 cores, 4GB RAM
  NATS: 1 core, 1GB RAM

Total Production:
  CPU: 23 cores
  Memory: 25 GB RAM
  Storage: 200 GB SSD
  Network: 1 Gbps
  Load Balancer: Required for HA
```

#### **Scaling Guidelines (Hybrid Architecture)**
```yaml
Small Deployment (1-1,000 users):
  Go Services: 2-3 replicas each
  Python Workers: 1-2 replicas each
  Nodes: 3-4 nodes
  Performance: 1,000+ concurrent users

Medium Deployment (1,000-10,000 users):
  Go Services: 3-5 replicas each
  Python Workers: 2-4 replicas each
  Nodes: 6-8 nodes
  Performance: 10,000+ concurrent users

Large Deployment (10,000+ users):
  Go Services: 5+ replicas each
  Python Workers: 4+ replicas each
  Nodes: 10+ nodes
  Performance: 100,000+ concurrent users

Enterprise: Custom sizing with dedicated AI worker clusters
```

## Database Setup

### PostgreSQL Configuration
```sql
-- Create database and user
CREATE DATABASE agentos;
CREATE USER agentos WITH PASSWORD 'secure_password';
GRANT ALL PRIVILEGES ON DATABASE agentos TO agentos;

-- Enable required extensions
\c agentos
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgvector";
```

### Redis Configuration
```redis
# redis.conf
maxmemory 2gb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

### Database Migrations
```bash
# Run migrations
make migrate-up

# Check migration status
make migrate-status

# Rollback if needed
make migrate-down
```

## Security Configuration

### SSL/TLS Setup
- [SSL Certificate Setup](security/ssl-setup.md)
- [Let's Encrypt Integration](security/letsencrypt.md)
- [Certificate Management](security/cert-management.md)

### Authentication & Authorization
- [JWT Configuration](security/jwt-config.md)
- [API Key Management](security/api-keys.md)
- [Role-Based Access Control](security/rbac.md)

### Network Security
- [Firewall Configuration](security/firewall.md)
- [VPC Setup](security/vpc-setup.md)
- [Security Groups](security/security-groups.md)

## Monitoring & Observability

### Metrics Collection
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'agentos-core-api'
    static_configs:
      - targets: ['core-api:8000']

  - job_name: 'agentos-agent-engine'
    static_configs:
      - targets: ['agent-engine:8001']
```

### Log Aggregation
```yaml
# filebeat.yml
filebeat.inputs:
- type: container
  paths:
    - '/var/lib/docker/containers/*/*.log'
  processors:
    - add_docker_metadata: ~

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
```

### Health Checks

#### **Go Backend Services**
```bash
# Go service health endpoints
curl http://localhost:8000/health  # Core API
curl http://localhost:8001/health  # Agent Engine
curl http://localhost:8002/health  # Memory Service
curl http://localhost:8003/health  # Tool Registry
curl http://localhost:8004/health  # Auth Service
```

#### **Python AI Workers**
```bash
# Python AI worker health endpoints
curl http://localhost:8080/health  # LangChain Worker
curl http://localhost:8081/health  # CrewAI Worker
curl http://localhost:8082/health  # Swarms Worker
curl http://localhost:8083/health  # AutoGen Worker
curl http://localhost:8084/health  # Embedding Worker
curl http://localhost:8085/health  # Model Worker
```

#### **Infrastructure Services**
```bash
# Infrastructure health checks
curl http://localhost:8000/health/db     # Database connectivity
curl http://localhost:8000/health/redis  # Redis connectivity
curl http://localhost:8000/health/nats   # NATS connectivity
```

#### **Comprehensive Health Check Script**
```bash
#!/bin/bash
# health-check.sh - Comprehensive health check for hybrid architecture

echo "=== AgentOS Hybrid Health Check ==="

# Go Backend Services
echo "Checking Go Backend Services..."
services=("core-api:8000" "agent-engine:8001" "memory-service:8002" "tool-registry:8003")
for service in "${services[@]}"; do
    if curl -f -s "http://${service}/health" > /dev/null; then
        echo "✅ ${service} - Healthy"
    else
        echo "❌ ${service} - Unhealthy"
    fi
done

# Python AI Workers
echo "Checking Python AI Workers..."
workers=("langchain-worker:8080" "crewai-worker:8081" "swarms-worker:8082" "autogen-worker:8083")
for worker in "${workers[@]}"; do
    if curl -f -s "http://${worker}/health" > /dev/null; then
        echo "✅ ${worker} - Healthy"
    else
        echo "❌ ${worker} - Unhealthy"
    fi
done

echo "Health check complete."
```

## Backup & Recovery

### Database Backup
```bash
# Automated backup script
#!/bin/bash
BACKUP_DIR="/backups/postgresql"
DATE=$(date +%Y%m%d_%H%M%S)

pg_dump -h localhost -U agentos agentos > \
  "$BACKUP_DIR/agentos_backup_$DATE.sql"

# Compress backup
gzip "$BACKUP_DIR/agentos_backup_$DATE.sql"

# Clean old backups (keep 30 days)
find $BACKUP_DIR -name "*.gz" -mtime +30 -delete
```

### Application Data Backup
```bash
# Backup application data
kubectl create backup agentos-backup \
  --include-namespaces=agentos \
  --storage-location=s3-backup
```

## Troubleshooting

### Common Issues
- [Service Won't Start](troubleshooting/service-startup.md)
- [Database Connection Issues](troubleshooting/database.md)
- [Performance Problems](troubleshooting/performance.md)
- [Memory Issues](troubleshooting/memory.md)

### Debugging Tools
- [Log Analysis](troubleshooting/log-analysis.md)
- [Performance Profiling](troubleshooting/profiling.md)
- [Network Debugging](troubleshooting/network.md)
- [Container Debugging](troubleshooting/containers.md)

### Support Resources
- [Deployment Checklist](checklists/deployment.md)
- [Production Readiness](checklists/production.md)
- [Security Checklist](checklists/security.md)
- [Performance Checklist](checklists/performance.md)

## Deployment Automation

### CI/CD Pipelines
- [GitHub Actions](ci-cd/github-actions.md)
- [GitLab CI](ci-cd/gitlab-ci.md)
- [Jenkins](ci-cd/jenkins.md)
- [Azure DevOps](ci-cd/azure-devops.md)

### Infrastructure as Code
- [Terraform](iac/terraform.md)
- [Ansible](iac/ansible.md)
- [Helm Charts](iac/helm.md)
- [Pulumi](iac/pulumi.md)

## Support

### Getting Help
- **Deployment Issues**: [GitHub Issues](https://github.com/tuanle96/agentos-ecosystem/issues)
- **Enterprise Support**: enterprise@agentos.ai
- **Community Forum**: [community.agentos.ai](https://community.agentos.ai)
- **Documentation**: [docs.agentos.ai](https://docs.agentos.ai)

### Professional Services
- **Deployment Consulting**: Available for enterprise customers
- **Custom Deployment**: Tailored deployment solutions
- **Training**: Deployment and operations training
- **Support Contracts**: 24/7 support available