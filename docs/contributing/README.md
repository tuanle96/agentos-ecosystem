# Contributing to AgentOS

> Guidelines and resources for contributing to the AgentOS ecosystem

## Overview

We welcome contributions to AgentOS! This section contains all the information you need to contribute effectively to the project.

## Contributing Guidelines

### Quick Links
- [Contributing Guide](CONTRIBUTING.md) - Main contribution guidelines
- [Code of Conduct](CODE_OF_CONDUCT.md) - Community standards
- [Code Style Guide](code-style.md) - Coding standards and conventions
- [Testing Guidelines](testing.md) - Testing requirements and best practices

### Development Process
- [Development Workflow](development-workflow.md) - Git workflow and branching
- [Pull Request Process](pull-request-process.md) - PR guidelines and review
- [Issue Guidelines](issue-guidelines.md) - Reporting bugs and requesting features
- [Release Process](release-process.md) - How releases are managed

## Getting Started

### Prerequisites
- **Go**: 1.21 or higher
- **Node.js**: 18.0 or higher
- **Git**: Latest version
- **Docker**: For testing and development
- **Make**: For build automation

### Development Setup
```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/agentos-ecosystem.git
cd agentos-ecosystem

# Set up development environment
make setup

# Configure environment
cp .env.example .env
# Edit .env with your configuration

# Start development services
make dev-services

# Run tests to ensure everything works
make test
```

### First Contribution
1. **Find an Issue**: Look for issues labeled `good first issue` or `help wanted`
2. **Fork Repository**: Create your own fork of the project
3. **Create Branch**: Create a feature branch for your changes
4. **Make Changes**: Implement your changes following our guidelines
5. **Test Changes**: Ensure all tests pass and add new tests if needed
6. **Submit PR**: Create a pull request with a clear description

## Types of Contributions

### Code Contributions
- **Bug Fixes**: Fix reported bugs and issues
- **New Features**: Implement new functionality
- **Performance Improvements**: Optimize existing code
- **Refactoring**: Improve code structure and maintainability

### Documentation Contributions
- **API Documentation**: Improve API reference docs
- **Tutorials**: Create step-by-step guides
- **Examples**: Add code examples and use cases
- **Translation**: Translate documentation to other languages

### Community Contributions
- **Issue Triage**: Help categorize and prioritize issues
- **Code Review**: Review pull requests from other contributors
- **Testing**: Test new features and report bugs
- **Support**: Help other users in discussions and forums

## Development Guidelines

### Code Quality Standards
- **Test Coverage**: Maintain >80% test coverage
- **Code Review**: All changes require peer review
- **Documentation**: Update docs for any API changes
- **Performance**: Consider performance impact of changes

### Coding Standards
- **Go**: Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- **JavaScript/TypeScript**: Follow [Airbnb Style Guide](https://github.com/airbnb/javascript)
- **Commit Messages**: Use [Conventional Commits](https://conventionalcommits.org/)
- **Branch Naming**: Use descriptive branch names (e.g., `feature/add-memory-search`)

### Testing Requirements
- **Unit Tests**: Required for all new code
- **Integration Tests**: Required for API changes
- **End-to-End Tests**: Required for user-facing features
- **Performance Tests**: Required for performance-critical changes

## Project Structure

### Repository Organization
```
agentos-ecosystem/
├── services/           # Backend Go services (PRIVATE)
├── packages/          # Shared libraries (PUBLIC)
├── products/          # Product implementations (MIXED)
├── infrastructure/    # Infrastructure code (PRIVATE)
├── tools/            # Development tools (MIXED)
├── docs/             # Documentation (PUBLIC)
└── scripts/          # Utility scripts (MIXED)
```

### Public vs Private Components
- **PUBLIC**: Open source components (Core, SDK, Community, Docs)
- **PRIVATE**: Commercial components (Services, Enterprise, Infrastructure)
- **MIXED**: Partially open components (Store, Cloud, Tools)

### Component Ownership
- **Core Team**: Maintains core services and architecture
- **Community**: Contributes to public components
- **Product Teams**: Maintains specific product areas
- **Documentation Team**: Maintains all documentation

## Communication

### Channels
- **GitHub Issues**: Bug reports and feature requests
- **GitHub Discussions**: General discussions and Q&A
- **Discord**: Real-time chat and community support
- **Email**: Direct contact for sensitive issues

### Community Guidelines
- **Be Respectful**: Treat all community members with respect
- **Be Constructive**: Provide helpful and actionable feedback
- **Be Patient**: Remember that everyone is learning
- **Be Inclusive**: Welcome contributors from all backgrounds

### Getting Help
- **Documentation**: Check existing docs first
- **Search Issues**: Look for existing discussions
- **Ask Questions**: Use GitHub Discussions for questions
- **Join Discord**: Get real-time help from the community

## Recognition

### Contributor Recognition
- **Contributors List**: All contributors are listed in CONTRIBUTORS.md
- **Release Notes**: Significant contributions are highlighted
- **Community Spotlight**: Featured contributors in blog posts
- **Swag**: Contributors receive AgentOS merchandise

### Maintainer Program
- **Requirements**: Consistent high-quality contributions
- **Benefits**: Commit access, decision-making input, early access
- **Responsibilities**: Code review, issue triage, community support
- **Application**: Contact the core team to express interest

## Resources

### Development Resources
- [Architecture Documentation](../architecture/) - System design and patterns
- [API Reference](../api-reference/) - Complete API documentation
- [Deployment Guides](../deployment/) - Setup and deployment instructions
- [Examples Repository](https://github.com/agentos/examples) - Code examples

### Learning Resources
- [Go Documentation](https://golang.org/doc/) - Go programming language
- [React Documentation](https://reactjs.org/docs/) - React framework
- [Kubernetes Documentation](https://kubernetes.io/docs/) - Container orchestration
- [PostgreSQL Documentation](https://www.postgresql.org/docs/) - Database

### Tools and Utilities
- [Development Tools](../tools/) - Build and development tools
- [Testing Utilities](../tools/testing/) - Testing frameworks and helpers
- [Code Generators](../tools/generators/) - Code generation tools
- [Linting Tools](../tools/linting/) - Code quality tools

## Frequently Asked Questions

### General Questions
**Q: How do I get started contributing?**
A: Start by reading this guide, setting up your development environment, and looking for issues labeled "good first issue".

**Q: What if I'm new to Go/React/Kubernetes?**
A: We welcome contributors of all skill levels! Start with documentation or simple bug fixes to get familiar with the codebase.

**Q: How long does it take to review pull requests?**
A: We aim to provide initial feedback within 48 hours and complete reviews within a week.

### Technical Questions
**Q: Can I contribute to private components?**
A: Private components are only accessible to core team members and enterprise customers. Focus on public components for open source contributions.

**Q: How do I test my changes?**
A: Run `make test` to run all tests, or see our [Testing Guidelines](testing.md) for more detailed instructions.

**Q: What's the branching strategy?**
A: We use Git Flow with `main` for production, `develop` for integration, and feature branches for development.

### Process Questions
**Q: Do I need to sign a CLA?**
A: Yes, all contributors must sign our Contributor License Agreement (CLA) before their first contribution.

**Q: How are releases managed?**
A: Releases follow semantic versioning and are managed by the core team. See our [Release Process](release-process.md) for details.

**Q: Can I propose new features?**
A: Absolutely! Create an issue with the "feature request" label and provide a detailed description of your proposal.

## Contact

### Core Team
- **Project Lead**: Lê Anh Tuấn (tuanle96@example.com)
- **Technical Lead**: TBD
- **Community Manager**: TBD
- **Documentation Lead**: TBD

### Community
- **GitHub**: [github.com/tuanle96/agentos-ecosystem](https://github.com/tuanle96/agentos-ecosystem)
- **Discord**: [discord.gg/agentos](https://discord.gg/agentos)
- **Email**: community@agentos.ai
- **Twitter**: [@AgentOS](https://twitter.com/AgentOS)

Thank you for your interest in contributing to AgentOS! 🚀