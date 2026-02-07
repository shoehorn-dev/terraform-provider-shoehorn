terraform {
  required_providers {
    shoehorn = {
      source = "shoehorn-dev/shoehorn"
    }
  }
}

# Configure the Shoehorn provider.
# Credentials can also be set via SHOEHORN_HOST and SHOEHORN_API_KEY environment variables.
provider "shoehorn" {
  host    = "https://shoehorn.example.com"
  api_key = var.shoehorn_api_key
  timeout = 30
}

variable "shoehorn_api_key" {
  type      = string
  sensitive = true
}
