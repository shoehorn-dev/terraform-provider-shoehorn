# Map an IdP group to a Cerbos role
resource "shoehorn_group_role_mapping" "platform_admin" {
  group_name  = "team-developer-platform"
  role_name   = "entity:editor"
  description = "Grant entity editor access to the developer platform group"
}

# Grant all users read-only entity access
resource "shoehorn_group_role_mapping" "everyone_viewer" {
  group_name  = "Everyone"
  role_name   = "entity:viewer"
  description = "All users can view catalog entities"
}
