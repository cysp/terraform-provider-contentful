data "contentful_teams" "test" {
  organization_id = var.organization_id
}

locals {
  selected_team = one([
    for team in data.contentful_teams.test.teams : team
    if team.name == var.team_name
  ])
}

resource "contentful_team_space_membership" "test" {
  space_id = var.space_id
  team_id  = local.selected_team.team_id

  admin = false
  roles = []
}
