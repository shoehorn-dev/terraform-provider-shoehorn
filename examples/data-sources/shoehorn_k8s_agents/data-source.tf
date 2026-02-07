# List all Kubernetes agents
data "shoehorn_k8s_agents" "all" {}

output "connected_agents" {
  value = [for a in data.shoehorn_k8s_agents.all.agents : a.name if a.status == "connected"]
}
