# Utility Scripts

> Collection of utility scripts for AgentOS ecosystem

## Overview

This directory contains utility scripts for various development, deployment, and maintenance tasks across the AgentOS ecosystem.

## Script Categories

### Public Scripts (Open Source)
- [setup.sh](setup.sh) - Environment setup and initialization
- [build-all.sh](build-all.sh) - Build all services and packages
- [test-all.sh](test-all.sh) - Run comprehensive test suite
- [clean.sh](clean.sh) - Clean build artifacts and temporary files

### Private Scripts (Internal)
- [deploy.sh](deploy.sh) - Production deployment script
- [backup.sh](backup.sh) - Database and application backup
- [restore.sh](restore.sh) - Restore from backup
- [health-check.sh](health-check.sh) - System health monitoring

### Database Scripts
- [init-db.sql](init-db.sql) - Database initialization
- [migrate.sh](migrate.sh) - Database migration runner
- [seed-data.sh](seed-data.sh) - Development data seeding
- [backup-db.sh](backup-db.sh) - Database backup utility

### Development Scripts
- [dev-setup.sh](dev-setup.sh) - Development environment setup
- [format-code.sh](format-code.sh) - Code formatting
- [lint-code.sh](lint-code.sh) - Code linting
- [generate-docs.sh](generate-docs.sh) - Documentation generation

## Usage

### Environment Setup
```bash
# Initial setup (run once)
./scripts/setup.sh

# Development environment setup
./scripts/dev-setup.sh

# Install development tools
./scripts/install-tools.sh
```

### Build and Test
```bash
# Build everything
./scripts/build-all.sh

# Run all tests
./scripts/test-all.sh

# Clean build artifacts
./scripts/clean.sh

# Format and lint code
./scripts/format-code.sh
./scripts/lint-code.sh
```

### Database Operations
```bash
# Initialize database
./scripts/init-db.sh

# Run migrations
./scripts/migrate.sh up

# Seed development data
./scripts/seed-data.sh

# Backup database
./scripts/backup-db.sh
```

### Deployment (Internal)
```bash
# Deploy to staging
./scripts/deploy.sh staging

# Deploy to production
./scripts/deploy.sh production

# Health check
./scripts/health-check.sh

# Backup before deployment
./scripts/backup.sh
```

## Script Standards

### Coding Standards
- Use `#!/bin/bash` shebang
- Set `set -euo pipefail` for error handling
- Use meaningful variable names
- Add comments for complex logic
- Include usage information

### Error Handling
```bash
#!/bin/bash
set -euo pipefail

# Function for error handling
error_exit() {
    echo "Error: $1" >&2
    exit 1
}

# Example usage
command_that_might_fail || error_exit "Command failed"
```

### Logging
```bash
# Logging functions
log_info() {
    echo "[INFO] $(date '+%Y-%m-%d %H:%M:%S') $1"
}

log_error() {
    echo "[ERROR] $(date '+%Y-%m-%d %H:%M:%S') $1" >&2
}

log_warn() {
    echo "[WARN] $(date '+%Y-%m-%d %H:%M:%S') $1"
}
```

### Configuration
```bash
# Load configuration from environment or config file
CONFIG_FILE="${CONFIG_FILE:-config/default.conf}"
if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
fi

# Set defaults
DATABASE_URL="${DATABASE_URL:-postgres://localhost:5432/agentos}"
REDIS_URL="${REDIS_URL:-redis://localhost:6379}"
```

## Script Documentation

### setup.sh
**Purpose**: Initial environment setup and dependency installation
**Usage**: `./scripts/setup.sh`
**Requirements**: Go 1.21+, Node.js 18+, Docker

### build-all.sh
**Purpose**: Build all Go services and Node.js packages
**Usage**: `./scripts/build-all.sh [--production]`
**Options**: 
- `--production`: Build for production with optimizations

### test-all.sh
**Purpose**: Run comprehensive test suite across all components
**Usage**: `./scripts/test-all.sh [--coverage] [--verbose]`
**Options**:
- `--coverage`: Generate coverage reports
- `--verbose`: Verbose test output

### deploy.sh (Private)
**Purpose**: Deploy application to specified environment
**Usage**: `./scripts/deploy.sh <environment> [--force]`
**Environments**: `dev`, `staging`, `production`
**Options**:
- `--force`: Force deployment without confirmation

## Environment Variables

### Required Variables
```bash
# Database configuration
DATABASE_URL=postgres://user:pass@host:5432/dbname
REDIS_URL=redis://host:6379/0

# API keys (for deployment)
OPENAI_API_KEY=your-openai-key
ANTHROPIC_API_KEY=your-anthropic-key

# Deployment configuration
DEPLOY_ENV=production
DEPLOY_REGION=us-west-2
```

### Optional Variables
```bash
# Build configuration
BUILD_ENV=production
SKIP_TESTS=false
PARALLEL_BUILDS=true

# Logging
LOG_LEVEL=info
LOG_FILE=/var/log/agentos/deploy.log

# Backup configuration
BACKUP_RETENTION_DAYS=30
BACKUP_STORAGE=s3://backups/agentos
```

## Security Considerations

### Sensitive Data
- Never commit API keys or passwords
- Use environment variables for secrets
- Encrypt sensitive configuration files
- Use secure file permissions (600/700)

### Script Permissions
```bash
# Set appropriate permissions
chmod 755 scripts/*.sh          # Executable scripts
chmod 600 scripts/config/*      # Configuration files
chmod 700 scripts/deploy.sh     # Deployment scripts (restricted)
```

### Audit Logging
```bash
# Log all script executions
SCRIPT_LOG="/var/log/agentos/scripts.log"
echo "$(date): $0 executed by $(whoami)" >> "$SCRIPT_LOG"
```

## Troubleshooting

### Common Issues
1. **Permission Denied**: Check script permissions with `ls -la`
2. **Command Not Found**: Ensure required tools are installed
3. **Environment Variables**: Verify all required variables are set
4. **Path Issues**: Use absolute paths or ensure correct working directory

### Debug Mode
```bash
# Enable debug mode
export DEBUG=true
./scripts/script-name.sh

# Or run with bash debug
bash -x ./scripts/script-name.sh
```

### Log Analysis
```bash
# View recent script logs
tail -f /var/log/agentos/scripts.log

# Search for errors
grep -i error /var/log/agentos/scripts.log

# View deployment logs
tail -f /var/log/agentos/deploy.log
```

## Contributing

### Adding New Scripts
1. Follow naming conventions (`kebab-case.sh`)
2. Include proper documentation header
3. Add usage information and examples
4. Test thoroughly before committing
5. Update this README with new script info

### Script Template
```bash
#!/bin/bash
# Script Name: example-script.sh
# Description: Brief description of what the script does
# Usage: ./scripts/example-script.sh [options]
# Author: Your Name
# Date: YYYY-MM-DD

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Functions
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Description of the script

OPTIONS:
    -h, --help      Show this help message
    -v, --verbose   Enable verbose output
    
EXAMPLES:
    $0 --verbose
    
EOF
}

main() {
    # Script logic here
    echo "Script executed successfully"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            usage
            exit 0
            ;;
        -v|--verbose)
            set -x
            shift
            ;;
        *)
            echo "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Execute main function
main "$@"
```

## Support

For issues with scripts or suggestions for new utilities:
- **GitHub Issues**: [Report script issues](https://github.com/tuanle96/agentos-ecosystem/issues)
- **Documentation**: [Script documentation](../docs/)
- **Community**: [Discord support](https://discord.gg/agentos)