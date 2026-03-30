terraform {
  required_providers {
    upcloud = {
      source  = "UpCloudLtd/upcloud"
      version = ">= 5.0.0"
    }
    helm = {
      source  = "hashicorp/helm"
      version = ">= 2.12.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">= 2.25.0"
    }
    shoehorn = {
      source  = "shoehorn-dev/shoehorn"
      version = ">= 0.2.0"
    }
  }
}

# =============================================================================
# Providers
# =============================================================================

provider "upcloud" {
  username = var.upcloud_username
  password = var.upcloud_password
}

provider "kubernetes" {
  host                   = upcloud_kubernetes_cluster.main.network[0].dns_name
  client_certificate     = base64decode(upcloud_kubernetes_cluster.main.client_certificate)
  client_key             = base64decode(upcloud_kubernetes_cluster.main.client_key)
  cluster_ca_certificate = base64decode(upcloud_kubernetes_cluster.main.cluster_ca_certificate)
}

provider "helm" {
  kubernetes {
    host                   = upcloud_kubernetes_cluster.main.network[0].dns_name
    client_certificate     = base64decode(upcloud_kubernetes_cluster.main.client_certificate)
    client_key             = base64decode(upcloud_kubernetes_cluster.main.client_key)
    cluster_ca_certificate = base64decode(upcloud_kubernetes_cluster.main.cluster_ca_certificate)
  }
}

# Shoehorn provider - configure after Phase 1 deploy, used by Phase 2 (agent)
provider "shoehorn" {
  host    = "https://${var.domain}"
  api_key = var.shoehorn_api_key
}

# =============================================================================
# UpCloud Infrastructure
# =============================================================================

resource "upcloud_kubernetes_cluster" "main" {
  name                    = "${var.environment}-shoehorn"
  zone                    = var.upcloud_zone
  network_cidr            = "172.16.0.0/24"
  control_plane_ip_filter = ["0.0.0.0/0"]

  plan = var.cluster_plan
}

resource "upcloud_kubernetes_node_group" "workers" {
  cluster = upcloud_kubernetes_cluster.main.id
  name    = "workers"

  plan           = var.node_plan
  count_per_zone = var.nodes_per_zone
  storage_size   = 50

  labels = {
    "app" = "shoehorn"
  }
}

resource "upcloud_managed_database_postgresql" "shoehorn" {
  name  = "${var.environment}-shoehorn-db"
  title = "Shoehorn PostgreSQL"
  zone  = var.upcloud_zone
  plan  = var.db_plan

  properties {
    version = "17"
  }
}

# =============================================================================
# Shoehorn Platform (Phase 1)
# =============================================================================

module "shoehorn" {
  source = "../../"

  domain            = var.domain
  organization_name = var.organization_name
  organization_slug = var.organization_slug
  storage_class     = "upcloud-block-storage-maxiops"

  # UpCloud Managed PostgreSQL
  database_host = upcloud_managed_database_postgresql.shoehorn.service_host
  database_port = upcloud_managed_database_postgresql.shoehorn.service_port
  database_name = "shoehorn"
  database_user = upcloud_managed_database_postgresql.shoehorn.service_username

  # Auth
  auth_provider = var.auth_provider
  auth_config   = var.auth_config
  admin_email   = var.admin_email

  # Credentials
  credentials = {
    postgres_password      = upcloud_managed_database_postgresql.shoehorn.service_password
    db_password            = var.app_user_password
    valkey_password        = var.valkey_password
    meilisearch_master_key = var.meilisearch_master_key
    jwt_secret             = var.jwt_secret
    auth_encryption_key    = var.auth_encryption_key
    session_encryption_key = var.session_encryption_key
  }

  # Ingress
  ingress_type  = "ingress"
  ingress_class = "nginx"

  # K8s Agent (Phase 2 - set deploy_agent = true after creating API key in UI)
  deploy_agent = var.shoehorn_api_key != ""
  cluster_id   = "${var.environment}-${var.upcloud_zone}"
  cluster_name = "${var.organization_name} (${var.environment})"

  # UpCloud-specific overrides
  helm_set = {
    "cloudProviders.upcloud.enabled" = "true"
  }

  depends_on = [
    upcloud_kubernetes_node_group.workers,
  ]
}
