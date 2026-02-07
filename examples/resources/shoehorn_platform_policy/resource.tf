# Configure a platform policy (policies are pre-seeded, cannot be created or destroyed)
# Terraform only manages the enabled/enforcement state.
resource "shoehorn_platform_policy" "require_description" {
  key         = "require-entity-description"
  enabled     = true
  enforcement = "warning"
}
