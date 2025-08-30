resource "contentful_resource_type" "this" {
  organization_id   = var.organization_id
  app_definition_id = local.app_definition_id
  resource_type_id  = "ResourceProvider:resourceType"

  name = "Resource"

  default_field_mapping = {
    title    = "{ /title }"
    subtitle = "{ /subtitle }"
  }
}
