resource "contentful_resource_type" "this" {
  organization_id   = contentful_resource_provider.this.organization_id
  app_definition_id = contentful_resource_provider.this.app_definition_id
  resource_type_id  = "${contentful_resource_provider.this.resource_provider_id}:Product"

  name = "Product"

  default_field_mapping = {
    title    = "{ /title }"
    subtitle = "{ /subtitle }"
  }
}
