import {
  identity = {
    organization_id = var.contentful_organization_id
    team_id         = var.team_id
  }
  to = contentful_team.this
}
