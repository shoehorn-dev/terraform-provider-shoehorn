# =============================================================================
# Core
# =============================================================================

variable "domain" {
  description = "Public domain for Shoehorn (e.g., portal.acme.com)"
  type        = string
}

variable "namespace" {
  description = "Kubernetes namespace for Shoehorn"
  type        = string
  default     = "shoehorn"
}

variable "release_name" {
  description = "Helm release name"
  type        = string
  default     = "shoehorn"
}

variable "organization_name" {
  description = "Organization display name shown in UI"
  type        = string
  default     = ""
}

variable "organization_slug" {
  description = "URL-safe organization identifier"
  type        = string
  default     = ""
}

# =============================================================================
# Helm Chart
# =============================================================================

variable "chart_repository" {
  description = "Helm chart repository URL"
  type        = string
  default     = "oci://ghcr.io/shoehorn-dev/helm-charts"
}

variable "chart_version" {
  description = "Shoehorn Helm chart version (null = latest)"
  type        = string
  default     = null
}

variable "helm_timeout" {
  description = "Timeout in seconds for Helm operations"
  type        = number
  default     = 600
}

variable "image_tag" {
  description = "Docker image tag for Shoehorn services (null = chart default)"
  type        = string
  default     = null
}

# =============================================================================
# Database (external managed PostgreSQL)
# =============================================================================

variable "database_host" {
  description = "PostgreSQL host (e.g., managed-db-1234.upcloud.com)"
  type        = string
}

variable "database_port" {
  description = "PostgreSQL port"
  type        = number
  default     = 5432
}

variable "database_name" {
  description = "PostgreSQL database name"
  type        = string
  default     = "shoehorn"
}

variable "database_user" {
  description = "PostgreSQL admin/migration user"
  type        = string
  default     = "shoehorn_user"
}

# =============================================================================
# Authentication
# =============================================================================

variable "auth_provider" {
  description = "Authentication provider: zitadel, okta, or entra-id"
  type        = string
  default     = "zitadel"

  validation {
    condition     = contains(["zitadel", "okta", "entra-id"], var.auth_provider)
    error_message = "auth_provider must be one of: zitadel, okta, entra-id"
  }
}

variable "auth_config" {
  description = "Auth provider configuration (passed as Helm set values under auth.<provider>.*)"
  type        = map(string)
  default     = {}
}

variable "admin_email" {
  description = "Email of the initial tenant admin user"
  type        = string
  default     = ""
}

# =============================================================================
# Credentials (becomes the K8s secret referenced by secret.existingSecret)
# =============================================================================

variable "credentials" {
  description = <<-EOT
    Map of secret keys for the shoehorn-credentials K8s secret.
    Required keys: postgres_password, db_password, jwt_secret, auth_encryption_key, session_encryption_key
    Optional keys: valkey_password, meilisearch_master_key, github_app_id, zitadel_project_id, etc.
    See Helm chart values.yaml secret.mappings for full key reference.
  EOT
  type        = map(string)
  sensitive   = true
}

# =============================================================================
# Infrastructure
# =============================================================================

variable "storage_class" {
  description = "Kubernetes storage class name"
  type        = string
  default     = ""
}

variable "ingress_type" {
  description = "Ingress type: ingressRoute (Traefik), ingress (standard), or httpRoute (Envoy Gateway)"
  type        = string
  default     = "ingress"

  validation {
    condition     = contains(["ingressRoute", "ingress", "httpRoute"], var.ingress_type)
    error_message = "ingress_type must be one of: ingressRoute, ingress, httpRoute"
  }
}

variable "ingress_class" {
  description = "Ingress class name (only used when ingress_type = ingress)"
  type        = string
  default     = ""
}

variable "replica_count" {
  description = "Replica count for application services"
  type        = number
  default     = 2
}

# =============================================================================
# K8s Agent (Phase 2 - requires shoehorn_api_key to be configured on provider)
# =============================================================================

variable "deploy_agent" {
  description = "Deploy the Shoehorn K8s discovery agent"
  type        = bool
  default     = false
}

variable "cluster_id" {
  description = "Unique identifier for this Kubernetes cluster"
  type        = string
  default     = ""
}

variable "cluster_name" {
  description = "Human-readable cluster name"
  type        = string
  default     = ""
}

variable "agent_chart_version" {
  description = "K8s agent Helm chart version (null = latest)"
  type        = string
  default     = null
}

variable "agent_gitops_tool" {
  description = "GitOps tool to monitor: argocd, fluxcd, or empty string"
  type        = string
  default     = ""
}

# =============================================================================
# Helm Value Overrides
# =============================================================================

variable "helm_values" {
  description = "List of raw YAML strings to pass as additional Helm values (applied in order)"
  type        = list(string)
  default     = []
}

variable "helm_set" {
  description = "Map of individual Helm value overrides (key = Helm value path, value = string value)"
  type        = map(string)
  default     = {}
}

variable "helm_set_sensitive" {
  description = "Map of sensitive Helm value overrides"
  type        = map(string)
  default     = {}
  sensitive   = true
}

variable "agent_helm_set" {
  description = "Map of Helm value overrides for the K8s agent chart"
  type        = map(string)
  default     = {}
}

# =============================================================================
# Health Check
# =============================================================================

variable "health_check_protocol" {
  description = "Protocol for health check (https or http)"
  type        = string
  default     = "https"
}

variable "health_check_attempts" {
  description = "Number of health check retry attempts"
  type        = number
  default     = 30
}
