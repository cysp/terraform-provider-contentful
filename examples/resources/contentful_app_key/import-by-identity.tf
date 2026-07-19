import {
  identity = {
    organization_id   = var.contentful_organization_id
    app_definition_id = var.app_definition_id
    key_kid           = var.app_key_kid
  }
  to = contentful_app_key.this
}
