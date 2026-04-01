# =============================================================================
# UpCloud
# =============================================================================

variable "upcloud_username" {
  description = "UpCloud API username"
  type        = string
}

variable "upcloud_password" {
  description = "UpCloud API password"
  type        = string
  sensitive   = true
}

variable "upcloud_zone" {
  description = "UpCloud zone (e.g., fi-hel1, de-fra1, us-nyc1)"
  type        = string
  default     = "fi-hel1"
}

variable "cluster_plan" {
  description = "UpCloud Kubernetes cluster plan"
  type        = string
  default     = "production-small"
}

variable "node_plan" {
  description = "UpCloud node group plan"
  type        = string
  default     = "K8S-4xCPU-8GB"
}

variable "nodes_per_zone" {
  description = "Number of worker nodes per availability zone"
  type        = number
  default     = 2
}

variable "db_plan" {
  description = "UpCloud Managed Database plan"
  type        = string
  default     = "1x2xCPU-4GB-50GB"
}

# =============================================================================
# Shoehorn
# =============================================================================

variable "environment" {
  description = "Environment name (e.g., production, staging)"
  type        = string
  default     = "production"
}

variable "domain" {
  description = "Public domain for Shoehorn (e.g., portal.acme.com)"
  type        = string
}

variable "organization_name" {
  description = "Organization display name"
  type        = string
}

variable "organization_slug" {
  description = "URL-safe organization identifier"
  type        = string
}

variable "admin_email" {
  description = "Email of the initial tenant admin"
  type        = string
}

# =============================================================================
# Authentication
# =============================================================================

variable "auth_provider" {
  description = "Authentication provider: zitadel, okta, or entra-id"
  type        = string
  default     = "zitadel"
}

variable "auth_config" {
  description = "Auth provider-specific configuration (e.g., { externalUrl = \"https://auth.example.com\" })"
  type        = map(string)
  default     = {}
}

# =============================================================================
# Secrets
# =============================================================================

variable "app_user_password" {
  description = "Password for the PostgreSQL app_user (RLS-enforced runtime user)"
  type        = string
  sensitive   = true
}

variable "valkey_password" {
  description = "Valkey (Redis) password"
  type        = string
  sensitive   = true
}

variable "meilisearch_master_key" {
  description = "Meilisearch master key"
  type        = string
  sensitive   = true
}

variable "jwt_secret" {
  description = "JWT signing secret"
  type        = string
  sensitive   = true
}

variable "auth_encryption_key" {
  description = "Auth encryption key"
  type        = string
  sensitive   = true
}

variable "session_encryption_key" {
  description = "Session encryption key"
  type        = string
  sensitive   = true
}

# =============================================================================
# Phase 2 (set after initial deploy + API key creation)
# =============================================================================

variable "shoehorn_api_key" {
  description = "Shoehorn API key (create in UI after Phase 1 deploy, then set here for Phase 2)"
  type        = string
  sensitive   = true
  default     = ""
}
