# Manage tenant appearance settings (singleton per tenant)
resource "shoehorn_tenant_settings" "main" {
  platform_name        = "Acme Developer Portal"
  platform_description = "Internal developer portal for Acme Corp"
  company_name         = "Acme Corp"
  primary_color        = "#3b82f6"
  secondary_color      = "#64748b"
  default_theme        = "system"
}
