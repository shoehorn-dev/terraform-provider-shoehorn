# Manage a feature flag
resource "shoehorn_feature_flag" "dark_mode" {
  key             = "enable-dark-mode"
  name            = "Dark Mode"
  description     = "Enable dark mode for the platform UI"
  default_enabled = true
}
