resource "contentful_team_space_membership" "this" {
  space_id = var.space_id
  team_id  = var.team_id

  admin = false
  roles = [var.role_id]
}
