# Changelog

All notable changes to the Shoehorn Terraform Provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-02-06

### Added

- **Provider**: Initial release with API key authentication and environment variable support (`SHOEHORN_HOST`, `SHOEHORN_API_KEY`)
- **Resources**:
  - `shoehorn_entity` - Manage catalog entities with links, relations, licenses, interfaces, tags, and changelog path
  - `shoehorn_team` - Manage teams with member assignments and roles
  - `shoehorn_tenant_settings` - Configure portal branding and appearance
  - `shoehorn_platform_policy` - Configure governance policies (pre-seeded, config-only)
  - `shoehorn_feature_flag` - Manage feature flags
  - `shoehorn_api_key` - Provision API keys with scoped permissions
  - `shoehorn_user_role` - Assign RBAC roles to users
  - `shoehorn_integration` - Configure third-party integrations
  - `shoehorn_k8s_agent` - Register Kubernetes cluster agents
- **Data Sources**:
  - `shoehorn_entities` - List all catalog entities
  - `shoehorn_teams` - List all teams
  - `shoehorn_feature_flags` - List all feature flags
  - `shoehorn_integrations` - List all integrations
  - `shoehorn_api_keys` - List all API keys
  - `shoehorn_k8s_agents` - List all K8s agents
  - `shoehorn_platform_policies` - List all platform policies
- **Import support** for all resources
- **Retry logic** with 3 attempts for EOF, connection reset, and 5xx errors
- 197+ unit tests with full CRUD lifecycle coverage

[Unreleased]: https://github.com/shoehorn-dev/terraform-provider-shoehorn/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/shoehorn-dev/terraform-provider-shoehorn/releases/tag/v0.1.0
