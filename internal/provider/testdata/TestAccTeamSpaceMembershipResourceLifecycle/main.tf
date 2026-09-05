resource "contentful_team_space_membership" "test" {
  space_id = var.space_id
  team_id  = var.team_id

  admin = var.admin
  roles = []
}
