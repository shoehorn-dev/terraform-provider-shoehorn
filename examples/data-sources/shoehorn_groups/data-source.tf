# List all IdP groups
data "shoehorn_groups" "all" {}

output "group_names" {
  value = [for g in data.shoehorn_groups.all.groups : g.name]
}

output "groups_with_roles" {
  value = [for g in data.shoehorn_groups.all.groups : {
    name  = g.name
    roles = [for r in g.roles : r.role_name]
  } if length(g.roles) > 0]
}
