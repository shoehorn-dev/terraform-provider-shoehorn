# Manage a catalog entity with full metadata
resource "shoehorn_entity" "payments_service" {
  name             = "payments-service"
  type             = "service"
  description      = "Handles payment processing and billing"
  entity_lifecycle = "production"
  tier             = "critical"
  owner            = "platform-engineering"
  tags             = ["payments", "go", "grpc"]

  links = jsonencode([
    { name = "Repository", url = "https://github.com/org/payments-service", icon = "github" },
    { name = "Dashboard", url = "https://grafana.internal/d/payments", icon = "grafana" },
    { name = "Runbook", url = "https://wiki.internal/runbooks/payments", icon = "docs" }
  ])

  relations = jsonencode([
    { type = "depends_on", target = "resource:postgres-primary" },
    { type = "calls", target = "service:notification-service" }
  ])

  licenses = jsonencode([
    { title = "Stripe Enterprise", vendor = "Stripe", expires = "2026-12-31", seats = 50, cost = "$25000/year" }
  ])

  interfaces = jsonencode({
    http = {
      openapi = "https://api.example.com/payments/openapi.json"
    }
  })

  changelog_path = "CHANGELOG.md"
}
