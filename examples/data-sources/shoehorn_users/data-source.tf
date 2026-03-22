# List all IdP directory users
data "shoehorn_users" "all" {}

output "user_emails" {
  value = [for u in data.shoehorn_users.all.users : u.email]
}

output "enabled_users" {
  value = [for u in data.shoehorn_users.all.users : {
    username = u.username
    email    = u.email
  } if u.enabled]
}
