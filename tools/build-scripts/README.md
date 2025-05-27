# Build Scripts

> Automated build scripts for AgentOS ecosystem

## Overview

Collection of build automation scripts for compiling, testing, and packaging AgentOS components.

## Scripts

### Core Build Scripts

- **build-all.sh**: Build all services and packages
- **build-go-services.sh**: Build Go backend services
- **build-frontend.sh**: Build frontend packages
- **build-docker.sh**: Build Docker images
- **clean-all.sh**: Clean build artifacts

### Quality Scripts

- **lint-all.sh**: Run linters on all code
- **format-all.sh**: Format all code
- **test-all.sh**: Run all tests
- **security-scan.sh**: Security vulnerability scanning

### Utility Scripts

- **deps-update.sh**: Update dependencies
- **docs-generate.sh**: Generate documentation
- **version-bump.sh**: Bump version numbers
- **changelog-generate.sh**: Generate changelog

## Usage

```bash
# Make scripts executable
chmod +x tools/build-scripts/*.sh

# Build everything
./tools/build-scripts/build-all.sh

# Build specific components
./tools/build-scripts/build-go-services.sh
./tools/build-scripts/build-frontend.sh

# Quality checks
./tools/build-scripts/lint-all.sh
./tools/build-scripts/test-all.sh

# Clean up
./tools/build-scripts/clean-all.sh
```

## Configuration

Build configuration in `build-config.yml`:

```yaml
go:
  version: "1.21"
  services:
    - core-api
    - agent-engine
    - memory-service
    - tool-registry

node:
  version: "18"
  packages:
    - packages/core
    - packages/ui-components
    - packages/api-client

docker:
  registry: "registry.agentos.ai"
  tag_latest: true
  platforms:
    - linux/amd64
    - linux/arm64
```

## Build Targets

### Development
- Fast builds for development
- Debug symbols included
- Hot reload support

### Production
- Optimized builds
- Minified assets
- Security hardening

### Testing
- Test coverage reports
- Benchmark results
- Performance profiling