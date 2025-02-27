resource "contentful_role" "author" {
  space_id = var.space_id

  name = "Author"

  permissions = {}
  policies    = []
}
