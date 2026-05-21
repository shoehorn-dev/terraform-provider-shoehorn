# Route critical CVE events for a team to a Slack channel via an incoming webhook.
# The webhook URL lives in the customer's secret manager and is wired into the
# deployment as notifications.secrets. Terraform only references it.
resource "shoehorn_notification_subscription" "team_cve_slack" {
  scope    = "team"
  scope_id = shoehorn_team.platform.id

  name         = "Platform CVE alerts"
  event_types  = ["cve.detected", "cve.fix_available"]
  min_severity = "warning"
  cadence      = "realtime"

  channel_type = "slack"
  slack {
    mode       = "webhook"
    url_secret = "secret://slack-platform-webhook"
  }
}

# Send a daily digest of unhealthy workloads to an external webhook.
resource "shoehorn_notification_subscription" "team_digest_webhook" {
  scope    = "team"
  scope_id = shoehorn_team.platform.id

  name         = "Workload health digest"
  event_types  = ["workload.unhealthy"]
  min_severity = "info"
  cadence      = "daily"

  channel_type = "webhook"
  webhook {
    url            = "https://ops.example.com/hooks/shoehorn"
    signing_secret = "env://NOTIFICATIONS_WEBHOOK_SIGNING"
    content_type   = "application/json"
  }
}
