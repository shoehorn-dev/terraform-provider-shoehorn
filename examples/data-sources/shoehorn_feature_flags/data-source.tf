# List all feature flags
data "shoehorn_feature_flags" "all" {}

output "enabled_flags" {
  value = [for f in data.shoehorn_feature_flags.all.feature_flags : f.key if f.enabled]
}
