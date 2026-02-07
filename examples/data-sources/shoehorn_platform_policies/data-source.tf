# List all platform policies
data "shoehorn_platform_policies" "all" {}

output "enabled_policies" {
  value = [for p in data.shoehorn_platform_policies.all.policies : p.name if p.enabled]
}
