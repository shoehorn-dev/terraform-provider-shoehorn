# Assign a role to a user
resource "shoehorn_user_role" "admin_user" {
  user_id = "user-abc-123"
  role    = "admin"
}
