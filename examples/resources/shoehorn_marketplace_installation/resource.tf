# Install a Marketplace item by its slug.
resource "shoehorn_marketplace_installation" "example" {
  slug    = "my-addon"
  enabled = true

  config_json = jsonencode({
    channel = "#platform-alerts"
  })
}
