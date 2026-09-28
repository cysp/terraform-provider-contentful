data "contentful_roles" "existing" {
  space_id = "your-space-id"
}

# Select exactly one Role by name before using its ID. Names are not lookup IDs.
locals {
  selected_role = one([
    for role in data.contentful_roles.existing.roles : role
    if role.name == "Editors"
  ])
}
