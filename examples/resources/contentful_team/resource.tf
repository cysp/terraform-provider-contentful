resource "contentful_team" "this" {
  organization_id = var.contentful_organization_id

  name        = "Example Team"
  description = "An example team"
}
