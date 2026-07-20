data "contentful_teams" "example" {
  organization_id = "your-organization-id"
}

# Contentful does not document team names as unique. Filter before requiring
# exactly one result so duplicate names elsewhere do not affect this lookup.
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
