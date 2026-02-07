# List all catalog entities
data "shoehorn_entities" "all" {}

output "entity_names" {
  value = [for e in data.shoehorn_entities.all.entities : e.name]
}
