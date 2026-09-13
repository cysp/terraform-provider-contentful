import {
  identity = {
    organization_id   = var.contentful_organization_id
    app_definition_id = var.app_definition_id
    resource_type_id  = "Catalog:Product"
  }
  to = contentful_resource_type.this
}
