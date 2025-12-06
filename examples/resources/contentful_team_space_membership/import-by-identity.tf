import {
  identity = {
    space_id                 = var.space_id
    team_space_membership_id = var.team_space_membership_id
  }
  to = contentful_team_space_membership.this
}
