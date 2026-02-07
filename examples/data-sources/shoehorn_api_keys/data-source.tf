# List all API keys (secrets are not returned)
data "shoehorn_api_keys" "all" {}

output "active_keys" {
  value = [for k in data.shoehorn_api_keys.all.api_keys : k.name if k.status == "active"]
}
