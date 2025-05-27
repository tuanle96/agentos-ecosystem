# AgentOS Ecosystem

> A comprehensive AI agent operating system with 6 integrated products

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Node.js Version](https://img.shields.io/badge/node-%3E%3D18.0.0-brightgreen)](https://nodejs.org/)
[![Lerna](https://img.shields.io/badge/maintained%20with-lerna-cc00ff.svg)](https://lerna.js.org/)

## 🌟 Overview

AgentOS is a comprehensive AI agent ecosystem that provides a unified platform for creating, managing, and orchestrating intelligent agents. Built with a **monorepo architecture**, it supports multiple AI frameworks including LangChain, CrewAI, Swarms, and AutoGen through a universal orchestration layer.

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

## 📁 Repository Structure

```
agentos-ecosystem/
├── packages/                          # [PUBLIC] Shared libraries
│   ├── core/                         # Core types and utilities
│   ├── ui-components/                # Shared React components
│   ├── api-client/                   # API client library
│   └── testing/                      # Testing utilities
│
├── services/                         # [PRIVATE] Backend microservices
│   ├── core-api/                     # Core API service
│   ├── agent-engine/                 # Agent execution engine
│   ├── memory-service/               # Memory management
│   └── [additional services]
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

- Node.js >= 18.0.0
- npm >= 9.0.0
- Python >= 3.11 (for AI services)
- Docker (for development environment)

### Installation

```bash
# Clone the repository
git clone https://github.com/tuanle96/agentos-ecosystem.git
cd agentos-ecosystem

# Install dependencies
npm install

# Bootstrap packages
npm run bootstrap

# Start development environment
npm run dev
```

## 🛠️ Development

### Monorepo Management

This project uses [Lerna](https://lerna.js.org/) for monorepo management with npm workspaces.

```bash
# Build all packages
npm run build

# Run tests
npm run test

# Lint code
npm run lint

# Clean all packages
npm run clean
```

### Product-Specific Commands

```bash
# Build specific products
npm run build:core
npm run build:enterprise
npm run build:cloud
npm run build:store

# Deploy specific products
npm run deploy:core
npm run deploy:enterprise
```

## 📦 Products

### 🌐 AgentOS Core (Open Source)
**License**: MIT  
**Purpose**: Open source foundation for community adoption

- Basic agent creation and management
- Core tool execution capabilities
- Memory system foundation
- Community-driven development

### 🏢 AgentOS Enterprise (Commercial)
**License**: Commercial  
**Purpose**: Enterprise-grade features and compliance

- Advanced security and RBAC
- Multi-tenancy support
- Compliance tools and audit trails
- Enterprise integrations

### ☁️ AgentOS Cloud (SaaS)
**License**: SaaS Subscription  
**Purpose**: Hosted platform for easy adoption

- Web and mobile applications
- Hosted agent execution
- Scalable infrastructure
- Pay-as-you-go pricing

### 🛒 AgentOS Store (Marketplace)
**License**: Platform Fees  
**Purpose**: Agent and tool marketplace

- Agent marketplace
- Tool store and distribution
- Rating and review system
- Monetization platform

### 🔧 AgentOS SDK (Developer Tools)
**License**: MIT  
**Purpose**: Developer tools and integrations

- Multi-language SDKs (Python, JavaScript, Go, Rust)
- API clients and utilities
- Development tools and examples
- Integration guides

### 👥 AgentOS Community (Open Platform)
**License**: Open Community  
**Purpose**: Community building and support

- Developer forum and discussions
- Documentation and tutorials
- Community showcase
- Knowledge sharing platform

## 🤖 AI Framework Integration

AgentOS supports multiple AI frameworks through a universal orchestration layer:

- **Swarms**: 5.0.0+ (Primary orchestration)
- **LangChain**: 0.1.0+ (Tool ecosystem)
- **CrewAI**: 0.22.0+ (Multi-agent collaboration)
- **AutoGen**: 0.2.0+ (Conversational patterns)
- **mem0**: Latest (Memory management)

## 🔒 Public/Private Strategy

### Public Components 🌐
- AgentOS Core (complete open source)
- SDKs and developer tools
- Community platform and documentation
- Shared libraries and UI components

### Private Components 🔒
- Backend microservices
- Enterprise features
- Cloud SaaS application
- Infrastructure and deployment

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

- [ ] **Phase 0**: Repository setup and monorepo foundation
- [ ] **Phase 1**: AgentOS Core development (open source)
- [ ] **Phase 2**: Backend services and infrastructure
- [ ] **Phase 3**: Enterprise and Cloud products
- [ ] **Phase 4**: Store marketplace and community platform

---

**Built with ❤️ by the AgentOS Team**