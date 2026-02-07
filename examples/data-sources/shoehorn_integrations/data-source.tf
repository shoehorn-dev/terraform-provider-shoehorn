# List all system integrations
data "shoehorn_integrations" "all" {}

output "integration_total" {
  value = data.shoehorn_integrations.all.total
}

output "integration_healthy" {
  value = data.shoehorn_integrations.all.healthy
}

output "connected_integrations" {
  value = [for i in data.shoehorn_integrations.all.integrations : {
    type     = i.type
    provider = i.provider
  } if i.status == "connected"]
}
