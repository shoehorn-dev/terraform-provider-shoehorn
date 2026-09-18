# List the Forge molds your tenant can use
data "shoehorn_forge_molds" "all" {}

output "mold_slugs" {
  value = [for m in data.shoehorn_forge_molds.all.molds : m.slug]
}
