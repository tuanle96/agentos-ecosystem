# Billing Service

> Usage tracking and billing service

## Overview

The Billing Service tracks usage, manages subscriptions, and handles billing for AgentOS Cloud and Enterprise.

## Features

- Usage tracking and metering
- Subscription management
- Invoice generation
- Payment processing integration
- Usage analytics

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL
- **Payments**: Stripe integration
- **Analytics**: Time-series data

## Development

```bash
# Build service
cd services/billing-service && go build

# Run tests
cd services/billing-service && go test ./...
```