# Configure a GitHub integration
resource "shoehorn_integration" "github" {
  name = "GitHub Production"
  type = "github"
  config_json = jsonencode({
    token        = var.github_token
    organization = "acme-corp"
  })
}

variable "github_token" {
  type      = string
  sensitive = true
}
