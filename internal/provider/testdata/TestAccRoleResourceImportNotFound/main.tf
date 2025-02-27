resource "contentful_role" "admin" {
  space_id = var.space_id

  name = "Admin"

  permissions = {}
  policies    = []
}
