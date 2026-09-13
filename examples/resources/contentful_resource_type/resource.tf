# The app must already have a Resource Provider with ID Catalog.
# The resource_provider example creates that provider; deploy its function separately.
resource "contentful_resource_type" "this" {
  organization_id   = var.contentful_organization_id
  app_definition_id = var.app_definition_id
  resource_type_id  = "Catalog:Product"

  name = "Product"

  default_field_mapping = {
    title    = "{ /title }"
    subtitle = "{ /subtitle }"
  }
}
