# Open critical governance actions
data "shoehorn_governance_actions" "critical" {
  status   = "open"
  priority = "critical"
}

output "critical_action_count" {
  value = data.shoehorn_governance_actions.critical.total
}

output "critical_action_entities" {
  value = [
    for a in data.shoehorn_governance_actions.critical.actions : a.entity_name
  ]
}
