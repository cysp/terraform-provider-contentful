resource "contentful_resource_type" "this" {
  organization_id      = var.organization_id
  app_definition_id    = local.app_definition_id
  resource_provider_id = "ResourceProviderName"
  resource_type        = "ResourceProviderName:resourceType"

  name = "Resource"

  default_field_mapping = {
    title    = "{ /title }"
    subtitle = "{ /subtitle }"
  }
}
