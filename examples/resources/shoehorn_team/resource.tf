# Manage a team with member assignments
resource "shoehorn_team" "platform" {
  name         = "Platform Engineering"
  slug         = "platform-engineering"
  display_name = "Platform Engineering"
  description  = "Core platform engineering team"

  members = jsonencode([
    { user_id = "alice@example.com", role = "manager" },
    { user_id = "bob@example.com", role = "admin" },
    { user_id = "carol@example.com", role = "member" }
  ])
}
