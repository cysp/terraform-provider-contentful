import {
  identity = {
    organization_id   = var.contentful_organization_id
    app_definition_id = var.app_definition_id
  }
  to = contentful_app_signing_secret.this
}
