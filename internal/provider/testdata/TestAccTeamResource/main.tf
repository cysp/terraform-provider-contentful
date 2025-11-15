resource "contentful_team" "test" {
  organization_id = var.organization_id

  name = var.team_name
}
