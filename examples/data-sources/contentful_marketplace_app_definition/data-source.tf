data "contentful_marketplace_app_definition" "this" {
  app_definition_id = "marketplace-app-definition-id"
}

locals {
  array_item_types = flatten([
    for location in coalesce(data.contentful_marketplace_app_definition.this.locations, []) : [
      for field in coalesce(location.field_types, []) : field.items.type
      if field.items != null
    ]
  ])
}
