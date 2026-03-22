# Changelog

All notable changes to the Shoehorn Terraform Provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-03-22

### Added

- **`shoehorn_group_role_mapping`** resource: Maps IdP groups to Cerbos roles so all members of a group inherit that role
  - Attributes: `group_name`, `role_name`, `auth_provider`, `description`
  - Composite ID format: `group_name:role_name`
  - Import support via `terraform import`
  - All fields force replacement (immutable resource)
- **`shoehorn_users`** data source: Lists all users from the IdP directory
  - Returns: `id`, `username`, `first_name`, `last_name`, `email`, `enabled`, `git_provider`
  - Nested `bundles` with role bundle assignments per user
- **`shoehorn_groups`** data source: Lists all IdP groups with their role mappings
  - Returns: `id`, `name`, `path`, `member_count`
  - Nested `roles` with `role_name`, `bundle_display_name`, `provider`
- **Client APIs**: `ListGroups`, `GetGroupRoles`, `AssignGroupRole`, `RemoveGroupRole`, `ListDirectoryUsers`, `GetDirectoryUser`
- **Error helpers**: `IsNotFound` utility and `ErrNotFound` sentinel error (refactored from client)
- **CI pipeline**: GitHub Actions workflow for build, vet, and test on push/PR
- **Dependabot**: Automated dependency updates for Go modules and GitHub Actions

### Fixed

- **`shoehorn_group_role_mapping`**: Renamed `provider` attribute to `auth_provider` — `provider` is a reserved Terraform root attribute name which prevented the provider schema from loading
- **`shoehorn_team`**: Fixed members state drift in Create and Update — when the API response omits members, the provider now preserves the planned value instead of setting it to null, preventing "inconsistent result after apply" errors

### Changed

- **Dependencies**: Upgraded core Terraform SDK packages
  - `terraform-plugin-framework` v1.17.0 → v1.19.0
  - `terraform-plugin-go` v0.29.0 → v0.31.0
  - `google.golang.org/grpc` v1.75.1 → v1.79.2
  - `google.golang.org/protobuf` v1.36.9 → v1.36.11
- **Client**: Added trace logging for all API requests/responses, warn logging for retries, error logging for final failures

## [0.1.1] - 2026-02-15

### Added

- **`shoehorn_tenant_settings`**: Announcement bar configuration support
  - New nested `announcement` block with fields: `enabled`, `message`, `type`, `pinned`, `link_url`, `link_text`, `updated_at`
  - Announcement types: `info`, `warning`, `error`, `success`
  - Optional dismissible or pinned announcements
- **Validators**: Input validation for tenant settings
  - Hex color validation for `primary_color`, `secondary_color`, `accent_color` (format: `#RRGGBB`)
  - Enum validation for `default_theme` (values: `light`, `dark`, `system`)
  - Enum validation for `announcement.type` (values: `info`, `warning`, `error`, `success`)

### Changed

- **`shoehorn_tenant_settings`**: Enhanced documentation with complete examples
  - Full configuration example with all appearance and announcement fields
  - Minimal configuration example
  - Example without announcement
- **Tests**: Increased test coverage from 194 to 229 tests
  - Added 5 announcement-specific tests
  - Added 8 client tests for new functionality
  - Added 27 resource tests for validation and announcement handling

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

[Unreleased]: https://github.com/shoehorn-dev/terraform-provider-shoehorn/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/shoehorn-dev/terraform-provider-shoehorn/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/shoehorn-dev/terraform-provider-shoehorn/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/shoehorn-dev/terraform-provider-shoehorn/releases/tag/v0.1.0
