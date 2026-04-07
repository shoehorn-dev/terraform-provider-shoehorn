output "shoehorn_url" {
  description = "Shoehorn portal URL"
  value       = module.shoehorn.url
}

output "shoehorn_namespace" {
  description = "Kubernetes namespace"
  value       = module.shoehorn.namespace
}

output "cluster_id" {
  description = "UpCloud Kubernetes cluster ID"
  value       = upcloud_kubernetes_cluster.main.id
}

output "database_host" {
  description = "Managed PostgreSQL host"
  value       = upcloud_managed_database_postgresql.shoehorn.service_host
}

output "agent_deployed" {
  description = "Whether the K8s agent was deployed"
  value       = module.shoehorn.agent_deployed
}

output "agent_status" {
  description = "K8s agent status"
  value       = module.shoehorn.agent_status
}
