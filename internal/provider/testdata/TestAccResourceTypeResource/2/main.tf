resource "contentful_resource_type" "test" {
  organization_id      = var.organization_id
  app_definition_id    = var.app_definition_id
  resource_provider_id = var.resource_provider_id
  resource_type_id     = "${var.resource_provider_id}:test"

  name = "Test"

  default_field_mapping = {
    title    = "{ /name }"
    subtitle = "{ /description }"
    image = {
      url = "{ /image }"
    }
    badge = {
      label   = "beta"
      variant = "beta"
    }
  }
}
