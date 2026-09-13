data "contentful_teams" "example" {
  organization_id = "your-organization-id"
}

# Choose a name that identifies exactly one team in this organization.
# one(...) returns null for no matches and errors for multiple matches;
# the membership below requires a matching team.
locals {
  selected_team = one([
    for team in data.contentful_teams.example.teams : team
    if team.name == "Your team name"
  ])
}

resource "contentful_team_space_membership" "example" {
  space_id = "your-space-id"
  team_id  = local.selected_team.team_id

  admin = false
  roles = []
}
