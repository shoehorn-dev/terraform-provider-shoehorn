# Create an API key for service-to-service authentication
resource "shoehorn_api_key" "ci_pipeline" {
  name       = "CI Pipeline Key"
  expires_in = "90d"
  scopes     = ["entities:read", "catalog:read"]
}

# The raw key is only available after creation
output "api_key" {
  value     = shoehorn_api_key.ci_pipeline.raw_key
  sensitive = true
}
