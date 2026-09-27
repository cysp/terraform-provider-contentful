data "contentful_app_definition" "this" {
  organization_id = var.contentful_organization_id

  app_definition_id = "app-definition-id"
}

locals {
  array_item_types = flatten([
    for location in coalesce(data.contentful_app_definition.this.locations, []) : [
      for field in coalesce(location.field_types, []) : field.items.type
      if field.items != null
    ]
  ])
}
