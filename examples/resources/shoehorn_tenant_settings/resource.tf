# Manage tenant appearance settings and announcement bar (singleton per tenant)
resource "shoehorn_tenant_settings" "main" {
  # Platform Branding
  platform_name        = "Acme Developer Portal"
  platform_description = "Internal developer portal for Acme Corp"
  company_name         = "Acme Corporation"

  # Color Palette
  primary_color   = "#3b82f6" # Blue - active states, primary buttons
  secondary_color = "#64748b" # Slate - hover states, secondary UI
  accent_color    = "#8b5cf6" # Purple - highlights, badges

  # Assets
  logo_url     = "https://cdn.example.com/logo.png"
  favicon_url  = "https://cdn.example.com/favicon.ico"
  default_theme = "dark" # Options: light, dark, system

  # Announcement Bar (optional)
  announcement = {
    enabled   = true
    message   = "Scheduled maintenance: Saturday 2AM-4AM UTC"
    type      = "warning" # Options: info, warning, error, success
    pinned    = false     # If true, users cannot dismiss
    link_url  = "https://status.acme.com"
    link_text = "View Status Page"
  }
}

# Minimal configuration (all fields optional)
resource "shoehorn_tenant_settings" "minimal" {
  platform_name = "My Portal"
}

# Configuration without announcement
resource "shoehorn_tenant_settings" "no_announcement" {
  platform_name = "Acme Portal"
  primary_color = "#1E88E5"
  default_theme = "system"
}
