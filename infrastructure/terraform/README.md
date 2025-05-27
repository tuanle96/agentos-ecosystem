# Terraform Configurations

> Infrastructure as Code using Terraform

## Overview

Terraform configurations for provisioning cloud infrastructure across AWS, GCP, and Azure.

## Directory Structure

```
terraform/
├── modules/          # Reusable Terraform modules
│   ├── vpc/         # VPC and networking
│   ├── eks/         # Kubernetes clusters
│   ├── rds/         # Database instances
│   └── redis/       # Redis clusters
├── environments/     # Environment-specific configs
│   ├── dev/         # Development environment
│   ├── staging/     # Staging environment
│   └── prod/        # Production environment
├── aws/             # AWS-specific configurations
├── gcp/             # GCP-specific configurations
├── azure/           # Azure-specific configurations
└── README.md       # This file
```

## Usage

```bash
# Initialize Terraform
terraform init

# Select workspace
terraform workspace select prod

# Plan changes
terraform plan -var-file="environments/prod/terraform.tfvars"

# Apply changes
terraform apply -var-file="environments/prod/terraform.tfvars"
```

## Modules

- **VPC**: Virtual Private Cloud setup
- **EKS**: Elastic Kubernetes Service
- **RDS**: Relational Database Service
- **Redis**: ElastiCache Redis clusters
- **S3**: Object storage buckets
- **IAM**: Identity and Access Management

## Variables

Key variables defined in `terraform.tfvars`:

```hcl
region = "us-west-2"
environment = "prod"
cluster_name = "agentos-prod"
node_count = 3
instance_type = "t3.large"
db_instance_class = "db.t3.medium"
```