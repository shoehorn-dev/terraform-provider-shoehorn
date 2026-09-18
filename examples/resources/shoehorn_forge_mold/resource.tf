# A mold that creates an empty GitHub repository.
resource "shoehorn_forge_mold" "github_repo" {
  slug        = "create-github-repo"
  name        = "Create GitHub repository"
  description = "An empty repository with branch protection"
  version     = "1.0.0"
  visibility  = "tenant"
  category    = "repository"
  tags        = ["github", "repository"]

  actions = [
    {
      action  = "github.repo.create"
      label   = "Create repository"
      primary = true
    }
  ]
}
