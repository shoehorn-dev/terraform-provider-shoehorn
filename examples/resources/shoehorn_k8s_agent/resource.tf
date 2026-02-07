# Register a Kubernetes agent
resource "shoehorn_k8s_agent" "production" {
  name       = "production-cluster"
  cluster_id = "prod-us-east-1"
}

# The agent token is only available after creation
output "agent_token" {
  value     = shoehorn_k8s_agent.production.token
  sensitive = true
}
