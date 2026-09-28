resource "terraform_data" "space" {
  input = "space"
}

resource "terraform_data" "role" {
  input = "item-b"
}

data "contentful_role" "selected" {
  space_id = terraform_data.space.output
  role_id  = terraform_data.role.output
}

data "contentful_roles" "all" {
  space_id = terraform_data.space.output
}

resource "contentful_team_space_membership" "assignment" {
  space_id = terraform_data.space.output
  team_id  = "team"
  admin    = false
  roles    = [data.contentful_role.selected.role_id]
}
