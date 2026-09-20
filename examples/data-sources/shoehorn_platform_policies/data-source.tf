data "shoehorn_platform_policies" "all" {}

# The policies you can turn on or off.
output "configurable_policies" {
  value = [for p in data.shoehorn_platform_policies.all.policies : p.key if p.configurable]
}

# The protections Shoehorn keeps on for every tenant.
output "always_on_policies" {
  value = [for p in data.shoehorn_platform_policies.all.policies : p.key if !p.configurable]
}
