# =============================================================================
# Shoehorn Kubernetes Module
#
# Deploys Shoehorn Internal Developer Portal to any Kubernetes cluster.
#
# Phase 1 (always): Helm release + health gate
# Phase 2 (deploy_agent = true): K8s agent token + agent Helm release
# =============================================================================

locals {
  secret_name = "${var.release_name}-credentials"
  agent_name  = coalesce(var.cluster_name, var.cluster_id, "default")
  health_url  = "${var.health_check_protocol}://${var.domain}/healthz"
  org_slug    = coalesce(var.organization_slug, replace(lower(var.domain), "/[^a-z0-9]+/", "-"))
  org_name    = coalesce(var.organization_name, var.domain)

  # Build the complete Helm values as a map, rendered to YAML
  shoehorn_values = yamlencode({
    global = {
      domain       = var.domain
      storageClass = var.storage_class
      organization = {
        slug = local.org_slug
        name = local.org_name
      }
    }

    image = var.image_tag != null ? { tag = var.image_tag } : {}

    replicaCount = {
      api      = var.replica_count
      web      = var.replica_count
      eventbus = var.replica_count
      worker   = var.replica_count
      crawler  = var.replica_count
      forge    = var.replica_count
    }

    # External managed PostgreSQL
    postgresql = {
      enabled = false
      external = {
        enabled  = true
        host     = var.database_host
        port     = var.database_port
        database = var.database_name
        user     = var.database_user
      }
    }

    # Credentials secret
    secret = {
      existingSecret = local.secret_name
    }

    # Auth
    auth = merge(
      { provider = var.auth_provider },
      length(var.auth_config) > 0 ? { (var.auth_provider) = var.auth_config } : {}
    )

    # RBAC admin bootstrap
    rbac = var.admin_email != "" ? {
      roleAssignment = {
        tenantAdmin = { user = var.admin_email }
      }
    } : {}

    # Ingress
    ingressRoute = { enabled = var.ingress_type == "ingressRoute" }
    ingress = merge(
      { enabled = var.ingress_type == "ingress" },
      var.ingress_class != "" ? { className = var.ingress_class } : {}
    )
    httpRoute = { enabled = var.ingress_type == "httpRoute" }
  })
}

# =============================================================================
# 1. Namespace
# =============================================================================

resource "kubernetes_namespace_v1" "shoehorn" {
  metadata {
    name = var.namespace

    labels = {
      "app.kubernetes.io/managed-by" = "terraform"
      "app.kubernetes.io/part-of"    = "shoehorn"
    }
  }
}

# =============================================================================
# 2. Credentials Secret
# =============================================================================

resource "kubernetes_secret_v1" "credentials" {
  metadata {
    name      = local.secret_name
    namespace = kubernetes_namespace_v1.shoehorn.metadata[0].name

    labels = {
      "app.kubernetes.io/managed-by" = "terraform"
      "app.kubernetes.io/part-of"    = "shoehorn"
    }
  }

  data = var.credentials
}

# =============================================================================
# 3. Shoehorn Helm Release
# =============================================================================

resource "helm_release" "shoehorn" {
  name             = var.release_name
  repository       = var.chart_repository
  chart            = "shoehorn"
  version          = var.chart_version
  namespace        = kubernetes_namespace_v1.shoehorn.metadata[0].name
  create_namespace = false
  wait             = true
  wait_for_jobs    = true
  timeout          = var.helm_timeout

  # Module-generated values (lowest priority)
  # then user YAML overrides
  # then user set overrides (highest priority)
  values = concat(
    [local.shoehorn_values],
    var.helm_values,
  )

  set = [for k, v in var.helm_set : { name = k, value = v }]

  set_sensitive = [for k, v in var.helm_set_sensitive : { name = k, value = v }]
}

# =============================================================================
# 4. Health Gate
# =============================================================================

data "http" "health" {
  url = local.health_url

  retry {
    attempts     = var.health_check_attempts
    min_delay_ms = 3000
    max_delay_ms = 10000
  }

  depends_on = [helm_release.shoehorn]
}

# =============================================================================
# 5. K8s Agent Registration (Phase 2)
# =============================================================================

resource "shoehorn_k8s_agent" "cluster" {
  count = var.deploy_agent ? 1 : 0

  name       = local.agent_name
  cluster_id = var.cluster_id

  depends_on = [data.http.health]
}

# =============================================================================
# 6. K8s Agent Helm Release (Phase 2)
# =============================================================================

resource "helm_release" "k8s_agent" {
  count = var.deploy_agent ? 1 : 0

  name             = "${var.release_name}-k8s-agent"
  repository       = var.chart_repository
  chart            = "shoehorn-k8s-agent"
  version          = var.agent_chart_version
  namespace        = kubernetes_namespace_v1.shoehorn.metadata[0].name
  create_namespace = false
  wait             = true
  timeout          = 300

  values = concat(
    [yamlencode({
      shoehorn = {
        apiURL = "${var.health_check_protocol}://${var.domain}"
        cluster = {
          id   = var.cluster_id
          name = local.agent_name
        }
      }
      agent = var.agent_gitops_tool != "" ? {
        gitops = { tool = var.agent_gitops_tool }
      } : {}
    })],
  )

  set_sensitive = [
    { name = "shoehorn.apiToken", value = shoehorn_k8s_agent.cluster[0].token },
  ]

  set = [for k, v in var.agent_helm_set : { name = k, value = v }]
}
