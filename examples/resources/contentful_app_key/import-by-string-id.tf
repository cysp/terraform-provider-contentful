import {
  id = "${var.contentful_organization_id}/${var.app_definition_id}/${var.app_key_kid}"
  to = contentful_app_key.this
}
