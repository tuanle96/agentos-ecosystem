# AgentOS Cloud

> Hosted SaaS platform for AI agents

## Overview

AgentOS Cloud is a fully hosted SaaS platform that provides easy access to AI agent capabilities without infrastructure management.

## Features

- **Hosted Platform**: No infrastructure management required
- **Web Application**: Full-featured web interface
- **Mobile Apps**: iOS and Android applications
- **API Access**: RESTful APIs for integration
- **Scalable Infrastructure**: Auto-scaling based on usage
- **Pay-as-you-go**: Usage-based pricing

## License

**SaaS Subscription** - Monthly/annual subscriptions

## Tech Stack

- **Backend**: Go microservices on Kubernetes
- **Frontend**: React + TypeScript (PWA)
- **Mobile**: React Native
- **Infrastructure**: Multi-cloud (AWS, GCP, Azure)
- **CDN**: Global content delivery

## Directory Structure

```
products/cloud/
├── web-app/          # [PRIVATE] Cloud web application
├── mobile-app/       # [PRIVATE] React Native mobile app
├── landing-page/     # [PUBLIC] Marketing website
├── billing-portal/   # [PRIVATE] Customer billing
└── README.md        # This file
```

## Pricing Plans

### Free Tier
- 100 agent executions/month
- Basic tools and capabilities
- Community support
- **Price**: Free

### Pro Plan
- 10,000 agent executions/month
- Advanced tools and integrations
- Priority support
- **Price**: $29/month

### Business Plan
- 100,000 agent executions/month
- Custom integrations
- Dedicated support
- **Price**: $299/month

### Enterprise Plan
- Unlimited executions
- Custom deployment options
- 24/7 support
- **Price**: Custom pricing

## Features by Plan

| Feature | Free | Pro | Business | Enterprise |
|---------|------|-----|----------|------------|
| Agent Executions | 100/mo | 10K/mo | 100K/mo | Unlimited |
| Basic Tools | ✅ | ✅ | ✅ | ✅ |
| Advanced Tools | ❌ | ✅ | ✅ | ✅ |
| Custom Integrations | ❌ | ❌ | ✅ | ✅ |
| Priority Support | ❌ | ✅ | ✅ | ✅ |
| SLA | ❌ | 99.9% | 99.95% | 99.99% |

## Getting Started

1. **Sign Up**: [cloud.agentos.ai](https://cloud.agentos.ai)
2. **Create Agent**: Use the web interface or API
3. **Configure Tools**: Add capabilities to your agent
4. **Execute**: Run your agent with natural language

## API Access

```bash
# Get API key from dashboard
curl -H "Authorization: Bearer YOUR_API_KEY" \
     https://api.agentos.ai/v1/agents
```

## Support

- **Documentation**: [docs.agentos.ai](https://docs.agentos.ai)
- **Support**: support@agentos.ai
- **Status**: [status.agentos.ai](https://status.agentos.ai)