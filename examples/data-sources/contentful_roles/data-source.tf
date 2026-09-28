data "contentful_roles" "existing" {
  space_id = "your-space-id"

  lifecycle {
    postcondition {
      condition     = length([for role in self.roles : role if role.name == "Editors"]) == 1
      error_message = "Exactly one Role named Editors must exist in this space."
    }
  }
}

locals {
  selected_role = one([
    for role in data.contentful_roles.existing.roles : role
    if role.name == "Editors"
  ])
}
