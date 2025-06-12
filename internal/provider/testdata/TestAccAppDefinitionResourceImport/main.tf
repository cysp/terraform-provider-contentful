resource "contentful_app_definition" "test" {
  organization_id = var.organization_id

  name = "Test App"

  locations = []
}
