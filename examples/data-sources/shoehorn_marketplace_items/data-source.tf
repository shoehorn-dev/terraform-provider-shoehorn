# List verified marketplace items
data "shoehorn_marketplace_items" "all" {}

output "verified_items" {
  value = [
    for i in data.shoehorn_marketplace_items.all.items : i.slug if i.verified
  ]
}
