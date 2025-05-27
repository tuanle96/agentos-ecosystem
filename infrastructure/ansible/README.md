# Ansible Playbooks

> Configuration management and application deployment

## Overview

Ansible playbooks for configuring servers, deploying applications, and managing infrastructure.

## Directory Structure

```
ansible/
├── playbooks/       # Ansible playbooks
│   ├── setup.yml   # Initial server setup
│   ├── deploy.yml  # Application deployment
│   └── update.yml  # System updates
├── roles/          # Ansible roles
│   ├── common/     # Common server configuration
│   ├── docker/     # Docker installation
│   ├── nginx/      # Nginx configuration
│   └── monitoring/ # Monitoring setup
├── inventories/    # Environment inventories
│   ├── dev/       # Development servers
│   ├── staging/   # Staging servers
│   └── prod/      # Production servers
└── README.md      # This file
```

## Usage

```bash
# Run playbook
ansible-playbook -i inventories/prod/hosts playbooks/deploy.yml

# Run specific role
ansible-playbook -i inventories/prod/hosts playbooks/setup.yml --tags docker

# Check syntax
ansible-playbook --syntax-check playbooks/deploy.yml
```

## Roles

- **common**: Basic server configuration
- **docker**: Docker and Docker Compose setup
- **nginx**: Reverse proxy configuration
- **monitoring**: Prometheus and Grafana setup
- **security**: Security hardening
- **backup**: Backup configuration

## Variables

Environment-specific variables in `group_vars/`:

```yaml
# group_vars/prod/main.yml
app_version: "1.0.0"
database_host: "prod-db.agentos.ai"
redis_host: "prod-redis.agentos.ai"
ssl_enabled: true
```