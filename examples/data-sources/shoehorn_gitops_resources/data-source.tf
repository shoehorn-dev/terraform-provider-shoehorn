# Argo CD applications and their sync status
data "shoehorn_gitops_resources" "argocd" {
  tool = "argocd"
}

output "argocd_apps" {
  value = [
    for r in data.shoehorn_gitops_resources.argocd.resources :
    "${r.namespace}/${r.name}: ${r.sync_status}"
  ]
}
