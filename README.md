# AgentOS Ecosystem

> A comprehensive AI agent operating system with 6 integrated products

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)](https://golang.org/)
[![Node.js Version](https://img.shields.io/badge/node-%3E%3D18.0.0-brightgreen)](https://nodejs.org/)
[![Lerna](https://img.shields.io/badge/maintained%20with-lerna-cc00ff.svg)](https://lerna.js.org/)

## 🌟 Overview

AgentOS is a comprehensive AI agent ecosystem that provides a unified platform for creating, managing, and orchestrating intelligent agents. Built with a **monorepo architecture**, it features **Go-based backend microservices** and supports multiple AI frameworks including LangChain, CrewAI, Swarms, and AutoGen through a universal orchestration layer.

## 🏗️ Ecosystem Architecture

```
AgentOS (Master Brand)
├── AgentOS Core (Open Source Foundation)
├── AgentOS Enterprise (Business Platform)
├── AgentOS Cloud (SaaS Offering)
├── AgentOS Store (Capability Marketplace)
├── AgentOS SDK (Developer Tools)
└── AgentOS Community (Developer Ecosystem)
```

## 🔧 Technology Stack

### **Backend (Go)**
- **Language**: Go 1.21+
- **Framework**: Gin (HTTP), GORM (ORM)
- **Database**: PostgreSQL 15+ with pgvector
- **Cache**: Redis 7+ with clustering
- **Message Queue**: NATS with JetStream
- **Monitoring**: Prometheus + Grafana + Jaeger

### **Frontend (JavaScript/TypeScript)**
- **Framework**: React 18+ with TypeScript
- **State Management**: Redux Toolkit
- **UI Library**: Material-UI or Tailwind CSS
- **Build Tool**: Vite
- **Package Management**: Lerna + npm workspaces

### **AI Integration**
- **Swarms**: 5.0.0+ (Primary orchestration)
- **LangChain**: 0.1.0+ (Tool ecosystem)
- **CrewAI**: 0.22.0+ (Multi-agent collaboration)
- **AutoGen**: 0.2.0+ (Conversational patterns)
- **Vector DBs**: Pinecone, Weaviate, Qdrant

## 📁 Repository Structure

```
agentos-ecosystem/
├── services/                         # [PRIVATE] Go backend microservices
│   ├── core-api/                     # Core API service (Gin + GORM)
│   ├── agent-engine/                 # Agent execution engine
│   ├── memory-service/               # Memory management with vector DBs
│   └── tool-registry/                # Tool registry and execution
│
├── packages/                         # [PUBLIC] Shared libraries
│   ├── core/                         # Core types and utilities
│   ├── ui-components/                # Shared React components
│   ├── api-client/                   # API client library
│   └── testing/                      # Testing utilities
│
├── products/                         # AgentOS Products
│   ├── core/                        # [PUBLIC] AgentOS Core
│   ├── enterprise/                  # [PRIVATE] AgentOS Enterprise
│   ├── cloud/                       # [MIXED] AgentOS Cloud
│   ├── store/                       # [MIXED] AgentOS Store
│   ├── sdk/                         # [PUBLIC] AgentOS SDK
│   └── community/                   # [PUBLIC] AgentOS Community
│
├── infrastructure/                   # [PRIVATE] Infrastructure as Code
├── tools/                           # [MIXED] Development tools
├── docs/                            # [PUBLIC] Documentation
└── scripts/                         # [MIXED] Utility scripts
```

## 🚀 Quick Start

### Prerequisites

- **Go**: 1.21 or higher
- **Node.js**: 18.0.0 or higher
- **Docker**: For development environment
- **Make**: For build automation

### Installation

```bash
# Clone the repository
git clone https://github.com/tuanle96/agentos-ecosystem.git
cd agentos-ecosystem

# Setup development environment
make setup

# Copy environment configuration
cp .env.example .env
# Edit .env with your API keys

# Start infrastructure services
make dev-services

# Run database migrations
make migrate-up

# Start development with hot reload
make dev
```

## 🛠️ Development

### Go Backend Development

```bash
# Build all services
make build

# Run tests
make test

# Run with hot reload
make dev

# Lint and format code
make lint
make format

# Generate API documentation
make swagger
```

### Service-Specific Commands

```bash
# Build specific services
make build-core-api
make build-agent-engine
make build-memory-service
make build-tool-registry

# Test specific services
make test-core-api
make test-agent-engine
make test-memory-service
make test-tool-registry
```

### Frontend Development

```bash
# Install frontend dependencies
npm install

# Build all frontend packages
npm run build

# Run frontend in development mode
npm run dev

# Test frontend packages
npm run test
```

## 📦 Products

### 🌐 AgentOS Core (Open Source)
**License**: MIT  
**Tech Stack**: Go + React  
**Purpose**: Open source foundation for community adoption

- Basic agent creation and management
- Core tool execution capabilities
- Memory system foundation
- Community-driven development

### 🏢 AgentOS Enterprise (Commercial)
**License**: Commercial  
**Tech Stack**: Go + React  
**Purpose**: Enterprise-grade features and compliance

- Advanced security and RBAC
- Multi-tenancy support
- Compliance tools and audit trails
- Enterprise integrations

### ☁️ AgentOS Cloud (SaaS)
**License**: SaaS Subscription  
**Tech Stack**: Go + React + React Native  
**Purpose**: Hosted platform for easy adoption

- Web and mobile applications
- Hosted agent execution
- Scalable infrastructure
- Pay-as-you-go pricing

### 🛒 AgentOS Store (Marketplace)
**License**: Platform Fees  
**Tech Stack**: Go + React  
**Purpose**: Agent and tool marketplace

- Agent marketplace
- Tool store and distribution
- Rating and review system
- Monetization platform

### 🔧 AgentOS SDK (Developer Tools)
**License**: MIT  
**Tech Stack**: Go + Multiple Languages  
**Purpose**: Developer tools and integrations

- Multi-language SDKs (Go, Python, JavaScript, Rust)
- API clients and utilities
- Development tools and examples
- Integration guides

### 👥 AgentOS Community (Open Platform)
**License**: Open Community  
**Tech Stack**: Go + React  
**Purpose**: Community building and support

- Developer forum and discussions
- Documentation and tutorials
- Community showcase
- Knowledge sharing platform

## 🤖 AI Framework Integration

AgentOS supports multiple AI frameworks through Go-based orchestration:

- **Swarms**: 5.0.0+ (Primary orchestration)
- **LangChain**: 0.1.0+ (Tool ecosystem via HTTP APIs)
- **CrewAI**: 0.22.0+ (Multi-agent collaboration)
- **AutoGen**: 0.2.0+ (Conversational patterns)
- **Custom Integration**: Go-based AI service wrappers

## 🔒 Public/Private Strategy

### Public Components 🌐
- AgentOS Core (complete open source)
- SDKs and developer tools
- Community platform and documentation
- Shared libraries and UI components

### Private Components 🔒
- Go backend microservices
- Enterprise features
- Cloud SaaS application
- Infrastructure and deployment

## 🐳 Docker Development

```bash
# Start all services
make docker-up

# View logs
make docker-logs

# Stop services
make docker-down

# Clean everything
make docker-clean
```

## 📊 Monitoring & Observability

- **Metrics**: Prometheus (http://localhost:9090)
- **Dashboards**: Grafana (http://localhost:3000)
- **Tracing**: Jaeger (http://localhost:16686)
- **Logs**: Structured logging with logrus

## 🗄️ Database

```bash
# Run migrations
make migrate-up

# Rollback migrations
make migrate-down

# Create new migration
make migrate-create name=add_users_table
```

## 📚 Documentation

- **Architecture**: [docs/architecture/](docs/architecture/)
- **API Reference**: [docs/api-reference/](docs/api-reference/)
- **Deployment**: [docs/deployment/](docs/deployment/)
- **Contributing**: [docs/contributing/](docs/contributing/)

## 🤝 Contributing

We welcome contributions to the public components of AgentOS! Please read our [Contributing Guide](docs/contributing/CONTRIBUTING.md) for details.

### Development Setup

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests and documentation
5. Submit a pull request

## 📄 License

- **Public Components**: MIT License
- **Private Components**: Proprietary License
- **Documentation**: CC BY 4.0

See [LICENSE](LICENSE) for more details.

## 🔗 Links

- **Website**: https://agentos.ai (coming soon)
- **Documentation**: https://docs.agentos.ai (coming soon)
- **Community**: https://community.agentos.ai (coming soon)
- **Blog**: https://blog.agentos.ai (coming soon)

## 📞 Support

- **Community Support**: [GitHub Discussions](https://github.com/tuanle96/agentos-ecosystem/discussions)
- **Enterprise Support**: enterprise@agentos.ai
- **Security Issues**: security@agentos.ai

## 🗺️ Roadmap

- [ ] **Phase 0**: Repository setup and Go backend foundation
- [ ] **Phase 1**: AgentOS Core development (open source)
- [ ] **Phase 2**: Backend microservices and AI integration
- [ ] **Phase 3**: Enterprise and Cloud products
- [ ] **Phase 4**: Store marketplace and community platform

## ⚡ Performance

- **API Response**: <100ms (95th percentile)
- **Agent Creation**: <500ms
- **Memory Operations**: <10ms
- **Concurrent Users**: 10,000+
- **Throughput**: 100,000+ requests/min

---

**Built with ❤️ and ⚡ Go by the AgentOS Team**