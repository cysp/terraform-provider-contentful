import {
  id = "${var.contentful_organization_id}/${var.app_definition_id}"
  to = contentful_app_signing_secret.this
}
