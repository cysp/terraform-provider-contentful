resource "contentful_resource_type" "test" {
  organization_id   = var.organization_id
  app_definition_id = var.app_definition_id
  resource_type_id  = var.resource_type_id

  name = "Test"

  default_field_mapping = {
    title = "{ /name }"
  }
}
