resource "contentful_role" "test" {
  space_id    = var.space_id
  name        = "Test"
  permissions = {}
  policies    = []
}
