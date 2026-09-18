# Track a scorecard gap on a service, due in two weeks.
resource "shoehorn_governance_action" "checkout_runbook" {
  entity_id   = "checkout-api"
  title       = "Add an on-call runbook"
  description = "The scorecard flags checkout-api for a missing runbook."
  priority    = "high"
  source_type = "scorecard"
  assigned_to = "checkout"
  sla_days    = 14
}
