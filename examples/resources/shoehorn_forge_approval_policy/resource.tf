# Two-step approval for production changes.
resource "shoehorn_forge_approval_policy" "production" {
  name        = "Production changes"
  description = "Platform lead signs off, then one person from security"
  enabled     = true

  steps = [
    {
      name           = "Platform lead"
      approvers      = ["alice@example.com", "bob@example.com"]
      required_count = 1
    },
    {
      name           = "Security"
      description    = "Any one security reviewer"
      approvers      = ["carol@example.com", "dave@example.com"]
      required_count = 1
    }
  ]
}
