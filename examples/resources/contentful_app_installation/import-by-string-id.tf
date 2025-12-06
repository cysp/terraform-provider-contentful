import {
  id = "${var.contentful_space_id}/${var.contentful_environment_id}/${var.app_definition_id}"
  to = contentful_app_installation.this
}
