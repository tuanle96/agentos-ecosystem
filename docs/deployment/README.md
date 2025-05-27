# Deployment Documentation

> Comprehensive deployment guides for AgentOS ecosystem

## Overview

This section contains detailed deployment guides for various environments, from local development to production deployments across different platforms.

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
- **Go**: 1.21 or higher
- **Node.js**: 18.0 or higher
- **Docker**: 24.0 or higher
- **Kubernetes**: 1.28 or higher (for K8s deployments)

### Local Development (5 minutes)
```bash
# Clone repository
git clone https://github.com/tuanle96/agentos-ecosystem.git
cd agentos-ecosystem

# Setup environment
make setup
cp .env.example .env
# Edit .env with your configuration

# Start infrastructure
make dev-services

# Run migrations
make migrate-up

# Start development
make dev
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
```bash
# Core Configuration
GO_ENV=production
DATABASE_URL=postgres://user:pass@host:5432/agentos
REDIS_URL=redis://host:6379/0
NATS_URL=nats://host:4222

# Security
JWT_SECRET=your-jwt-secret-key
API_RATE_LIMIT=10000
CORS_ORIGINS=https://your-domain.com

# AI Services
OPENAI_API_KEY=your-openai-key
ANTHROPIC_API_KEY=your-anthropic-key
PINECONE_API_KEY=your-pinecone-key

# Monitoring
PROMETHEUS_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
LOG_LEVEL=info

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

### Minimum Requirements (Development)
- **CPU**: 2 cores
- **Memory**: 4 GB RAM
- **Storage**: 20 GB SSD
- **Network**: 10 Mbps

### Recommended Requirements (Production)
- **CPU**: 8 cores
- **Memory**: 16 GB RAM
- **Storage**: 100 GB SSD
- **Network**: 1 Gbps
- **Load Balancer**: Required for HA

### Scaling Guidelines
- **Small Deployment**: 1-100 users, 2-4 nodes
- **Medium Deployment**: 100-1000 users, 4-8 nodes
- **Large Deployment**: 1000+ users, 8+ nodes
- **Enterprise**: Custom sizing based on requirements

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
```bash
# Service health endpoints
curl http://localhost:8000/health  # Core API
curl http://localhost:8001/health  # Agent Engine
curl http://localhost:8002/health  # Memory Service
curl http://localhost:8003/health  # Tool Registry
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