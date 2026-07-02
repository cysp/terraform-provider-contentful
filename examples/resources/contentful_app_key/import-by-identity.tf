import {
  identity = {
    organization_id   = var.organization_id
    app_definition_id = var.app_definition_id
    key_kid           = var.key_kid
  }

  to = contentful_app_key.this
}
