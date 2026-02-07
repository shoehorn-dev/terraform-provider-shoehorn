# Shoehorn Terraform Provider

Terraform provider for managing [Shoehorn](https://shoehorn.dev) Internal Developer Portal resources as code.

## Features

- **Catalog Entities** - Define services, libraries, APIs, and infrastructure as code with full metadata (links, relations, licenses, interfaces)
- **Teams** - Manage teams with members and role assignments
- **Feature Flags** - Toggle feature flags across environments
- **Tenant Settings** - Configure portal branding and appearance
- **Platform Policies** - Enforce organizational standards and governance
- **API Keys** - Provision API keys for service-to-service authentication
- **User Roles** - Assign RBAC roles to users
- **Integrations** - Configure third-party integrations (GitHub, PagerDuty, etc.)
- **Kubernetes Agents** - Register K8s cluster agents for workload discovery

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (to build the provider plugin)
- A running Shoehorn instance with an API key

## Installation

### From the Terraform Registry

```hcl
terraform {
  required_providers {
    shoehorn = {
      source  = "shoehorn-dev/shoehorn"
      version = "~> 0.1"
    }
  }
}
```

### Local Development

```bash
# Clone the repository
git clone https://github.com/shoehorn-dev/terraform-provider-shoehorn.git
cd terraform-provider-shoehorn

# Build the provider
go build -o terraform-provider-shoehorn

# Install to local plugin directory
# Linux/macOS:
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/shoehorn-dev/shoehorn/0.1.0/$(go env GOOS)_$(go env GOARCH)
cp terraform-provider-shoehorn ~/.terraform.d/plugins/registry.terraform.io/shoehorn-dev/shoehorn/0.1.0/$(go env GOOS)_$(go env GOARCH)/

# Windows: Use a dev_overrides block in your .terraformrc (see Development section below)
```

## Authentication

The provider requires a Shoehorn API host and API key. These can be configured in three ways:

### 1. Provider Block

```hcl
provider "shoehorn" {
  host    = "https://shoehorn.example.com"
  api_key = var.shoehorn_api_key
}
```

### 2. Environment Variables (Recommended for CI/CD)

```bash
export SHOEHORN_HOST="https://shoehorn.example.com"
export SHOEHORN_API_KEY="shp_your_api_key_here"
```

```hcl
provider "shoehorn" {}
```

### 3. Mixed

```hcl
provider "shoehorn" {
  host = "https://shoehorn.example.com"
  # api_key read from SHOEHORN_API_KEY env var
}
```

## Quick Start

```hcl
terraform {
  required_providers {
    shoehorn = {
      source  = "shoehorn-dev/shoehorn"
      version = "~> 0.1"
    }
  }
}

provider "shoehorn" {
  host    = "https://shoehorn.example.com"
  api_key = var.shoehorn_api_key
}

variable "shoehorn_api_key" {
  type      = string
  sensitive = true
}

# Create a team
resource "shoehorn_team" "platform" {
  name         = "Platform Engineering"
  slug         = "platform-engineering"
  display_name = "Platform Engineering"
  description  = "Core platform team"

  members = jsonencode([
    { user_id = "alice@example.com", role = "manager" },
    { user_id = "bob@example.com", role = "member" }
  ])
}

# Create a service entity
resource "shoehorn_entity" "api_gateway" {
  name             = "api-gateway"
  type             = "service"
  description      = "Central API gateway for all backend services"
  entity_lifecycle = "production"
  tier             = "critical"
  owner            = shoehorn_team.platform.slug
  tags             = ["gateway", "go", "infrastructure"]

  links = jsonencode([
    { name = "Repository", url = "https://github.com/org/api-gateway", icon = "github" },
    { name = "Dashboard",  url = "https://grafana.internal/d/api-gw",  icon = "grafana" },
    { name = "Runbook",    url = "https://wiki.internal/api-gateway",  icon = "docs" }
  ])

  relations = jsonencode([
    { type = "depends_on", target = "service:auth-service" },
    { type = "calls",      target = "service:user-service" }
  ])
}
```

## Resources

### shoehorn_entity

Manages a catalog entity (service, library, API, resource, etc.).

```hcl
resource "shoehorn_entity" "payments" {
  name             = "payments-service"
  type             = "service"
  description      = "Payment processing microservice"
  entity_lifecycle = "production"
  tier             = "critical"
  owner            = "platform-engineering"
  tags             = ["payments", "go", "grpc"]

  links = jsonencode([
    { name = "Repository", url = "https://github.com/org/payments", icon = "github" },
    { name = "API Docs",   url = "https://docs.internal/payments",  icon = "api" }
  ])

  relations = jsonencode([
    { type = "depends_on",  target = "resource:postgres-primary" },
    { type = "calls",       target = "service:notification-service" },
    { type = "reads_from",  target = "resource:kafka-cluster" }
  ])

  licenses = jsonencode([
    {
      title   = "Stripe Enterprise"
      vendor  = "Stripe"
      expires = "2026-12-31"
      seats   = 50
      cost    = "$25000/year"
    }
  ])

  interfaces = jsonencode({
    http = {
      openapi = "https://api.example.com/openapi.json"
      baseUrl = "https://api.example.com/v1"
    }
    grpc = {
      package = "payments.v1"
      proto   = "https://github.com/org/payments/proto/payments.proto"
    }
  })

  changelog_path = "CHANGELOG.md"
}
```

#### Attributes

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | Yes | Entity name (used as service ID). Forces replacement if changed. |
| `type` | String | Yes | Entity type: `service`, `library`, `api`, `website`, `resource`, `system`, `component` |
| `description` | String | No | Description of the entity |
| `entity_lifecycle` | String | No | Lifecycle stage: `experimental`, `production`, `deprecated` |
| `tier` | String | No | Service tier: `critical`, `gold`, `silver`, `bronze`, `tier1`-`tier4` |
| `owner` | String | No | Owner team slug |
| `tags` | Set of String | No | Tags for categorization |
| `links` | JSON String | No | Array of `{name, url, icon}` objects. Icons: `github`, `grafana`, `docs`, `api`, `slack`, `jira`, etc. |
| `relations` | JSON String | No | Array of `{type, target}` objects. Types: `depends_on`, `calls`, `reads_from`, `writes_to`, `owned_by` |
| `licenses` | JSON String | No | Array of `{title, vendor, expires, seats, cost}` objects |
| `interfaces` | JSON String | No | Object with `http` and/or `grpc` interface definitions |
| `changelog_path` | String | No | Path to changelog file in repository |

**Computed**: `id`, `repository_path`, `created_at`, `updated_at`

### shoehorn_team

Manages a team with optional member assignments.

```hcl
resource "shoehorn_team" "backend" {
  name         = "Backend Engineering"
  slug         = "backend-engineering"
  display_name = "Backend Engineering"
  description  = "Backend services team"

  members = jsonencode([
    { user_id = "alice@example.com", role = "manager" },
    { user_id = "bob@example.com",   role = "admin" },
    { user_id = "carol@example.com", role = "member" }
  ])
}
```

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | String | Yes | Team name |
| `slug` | String | Yes | Unique slug (forces replacement if changed) |
| `display_name` | String | No | Display name |
| `description` | String | No | Team description |
| `members` | JSON String | No | Array of `{user_id, role}` objects. Roles: `manager`, `admin`, `member` |
| `metadata` | JSON String | No | Arbitrary JSON metadata |

**Computed**: `id`, `is_active`, `member_count`, `created_at`, `updated_at`

### shoehorn_tenant_settings

Manages portal branding and appearance (singleton per tenant).

```hcl
resource "shoehorn_tenant_settings" "main" {
  platform_name        = "Acme Developer Portal"
  platform_description = "Internal developer portal for Acme Corp"
  company_name         = "Acme Corp"
  primary_color        = "#3b82f6"
  default_theme        = "system"
}
```

### shoehorn_platform_policy

Configures platform governance policies. Policies are pre-seeded and cannot be created or destroyed - Terraform only manages their `enabled` and `enforcement` state.

```hcl
resource "shoehorn_platform_policy" "require_docs" {
  key         = "required-entity-docs"
  enabled     = true
  enforcement = "warning"
}
```

### shoehorn_feature_flag

Manages feature flags.

```hcl
resource "shoehorn_feature_flag" "dark_mode" {
  key             = "enable-dark-mode"
  name            = "Dark Mode"
  description     = "Enable dark mode for the platform UI"
  default_enabled = true
}
```

### shoehorn_api_key

Provisions API keys for service-to-service authentication.

```hcl
resource "shoehorn_api_key" "ci" {
  name       = "CI Pipeline Key"
  expires_in = "90d"
  scopes     = ["entities:read", "catalog:read"]
}

output "api_key" {
  value     = shoehorn_api_key.ci.raw_key
  sensitive = true
}
```

### shoehorn_integration

Configures third-party integrations.

```hcl
resource "shoehorn_integration" "github" {
  name = "GitHub Production"
  type = "github"
  config_json = jsonencode({
    token        = var.github_token
    organization = "acme-corp"
  })
}
```

### shoehorn_k8s_agent

Registers Kubernetes cluster agents.

```hcl
resource "shoehorn_k8s_agent" "production" {
  name       = "production-cluster"
  cluster_id = "prod-us-east-1"
}

output "agent_token" {
  value     = shoehorn_k8s_agent.production.token
  sensitive = true
}
```

### shoehorn_user_role

Assigns RBAC roles to users.

```hcl
resource "shoehorn_user_role" "admin" {
  user_id = "user-abc-123"
  role    = "admin"
}
```

## Data Sources

All resources have corresponding data sources for reading existing state:

```hcl
# List all entities
data "shoehorn_entities" "all" {}

# List all teams
data "shoehorn_teams" "all" {}

# List all feature flags
data "shoehorn_feature_flags" "all" {}

# List all integrations
data "shoehorn_integrations" "all" {}

# List all API keys
data "shoehorn_api_keys" "all" {}

# List all K8s agents
data "shoehorn_k8s_agents" "all" {}

# List all platform policies
data "shoehorn_platform_policies" "all" {}
```

## Importing Existing Resources

Resources can be imported into Terraform state:

```bash
# Import an entity by name
terraform import shoehorn_entity.api_gateway api-gateway

# Import a team by slug
terraform import shoehorn_team.platform platform-engineering

# Import tenant settings (singleton, use any ID)
terraform import shoehorn_tenant_settings.main singleton

# Import a platform policy by key
terraform import shoehorn_platform_policy.require_docs required-entity-docs
```

## Development

### Building

```bash
go build -o terraform-provider-shoehorn
```

### Testing

```bash
# Run all unit tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with race detector
go test -race ./...
```

### Local Testing with dev_overrides

Create or update your Terraform CLI configuration file:

**Linux/macOS** (`~/.terraformrc`):
```hcl
provider_installation {
  dev_overrides {
    "shoehorn-dev/shoehorn" = "/path/to/terraform-provider-shoehorn"
  }
  direct {}
}
```

**Windows** (`%APPDATA%\terraform.rc`):
```hcl
provider_installation {
  dev_overrides {
    "shoehorn-dev/shoehorn" = "C:\\path\\to\\terraform-provider-shoehorn"
  }
  direct {}
}
```

Then build and test:

```bash
go build -o terraform-provider-shoehorn
cd test-local
terraform plan
terraform apply
```

### Generating Documentation

Documentation is generated with [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs):

```bash
go generate ./...
```

## Support

- **Bug reports and feature requests**: [GitHub Issues](https://github.com/shoehorn-dev/terraform-provider-shoehorn/issues)
- **Security vulnerabilities**: See [SECURITY.md](SECURITY.md)
- **Documentation**: [Terraform Registry](https://registry.terraform.io/providers/shoehorn-dev/shoehorn/latest/docs)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
