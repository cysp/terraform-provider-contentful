data "contentful_role" "editor" {
  space_id = "your-space-id"
  role_id  = "existing-role-id"
}

# The lookup only reads the Role. This resource assigns its ID to an existing team.
resource "contentful_team_space_membership" "editor" {
  space_id = "your-space-id"
  team_id  = "existing-team-id"

  admin = false
  roles = [data.contentful_role.editor.role_id]
}
