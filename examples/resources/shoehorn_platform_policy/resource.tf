# Shoehorn defines its own policies. Two can be changed; the rest are always on.

resource "shoehorn_platform_policy" "governance_actions" {
  key     = "governance-auto-actions"
  enabled = true
}

resource "shoehorn_platform_policy" "api_key_expiry" {
  key     = "api-key-expiration"
  enabled = true
}
